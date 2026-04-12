// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/metrics"
)

const (
	ragCacheCapacity = 128
	ragCacheTTL      = 5 * time.Minute
)

type ragCacheKey struct {
	queryHash  string
	maxResults int
}

type ragCacheEntry struct {
	key      ragCacheKey
	results  []KBEntry
	expireAt time.Time
	element  *list.Element
}

// ragLRU is a thread-safe LRU cache for RAG KB query results.
// It uses container/list for O(1) eviction and a map for O(1) lookup.
type ragLRU struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration
	items    map[ragCacheKey]*ragCacheEntry
	order    *list.List // front = most recently used
}

// newRAGLRU creates a new LRU cache with the given capacity and TTL.
func newRAGLRU(capacity int, ttl time.Duration) *ragLRU {
	return &ragLRU{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[ragCacheKey]*ragCacheEntry, capacity),
		order:    list.New(),
	}
}

// cacheKey builds a ragCacheKey by hashing the query string (SHA-256) so that
// map keys stay small regardless of query length.
func cacheKey(query string, maxResults int) ragCacheKey {
	h := sha256.Sum256([]byte(query))
	return ragCacheKey{queryHash: hex.EncodeToString(h[:]), maxResults: maxResults}
}

// get returns cached results if present and not expired.
func (c *ragLRU) get(query string, maxResults int) ([]KBEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(query, maxResults)
	entry, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expireAt) {
		c.order.Remove(entry.element)
		delete(c.items, key)
		return nil, false
	}

	c.order.MoveToFront(entry.element)
	return entry.results, true
}

// set stores results in the cache, evicting the LRU entry if at capacity.
func (c *ragLRU) set(query string, maxResults int, results []KBEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(query, maxResults)

	// Update existing entry.
	if entry, ok := c.items[key]; ok {
		c.order.MoveToFront(entry.element)
		entry.results = results
		entry.expireAt = time.Now().Add(c.ttl)
		return
	}

	// Evict oldest entry when at capacity.
	if c.order.Len() >= c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*ragCacheEntry).key)
		}
	}

	entry := &ragCacheEntry{
		key:      key,
		results:  results,
		expireAt: time.Now().Add(c.ttl),
	}
	entry.element = c.order.PushFront(entry)
	c.items[key] = entry
}

// kbFile mirrors the top-level structure of the knowledge-base JSON.
type kbFile struct {
	Domain   string    `json:"domain"`
	Problems []KBEntry `json:"problems"`
}

// KBEntry is a single problem record from the knowledge base.
type KBEntry struct {
	Domain         string       `json:"domain,omitempty"`
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Category       string       `json:"category"`
	Severity       string       `json:"severity"`
	Description    string       `json:"description"`
	VisualSymptoms []string     `json:"visual_symptoms"`
	ProbableCauses []KBCause    `json:"probable_causes"`
	Solutions      []KBSolution `json:"solutions"`
}

// KBCause is a single probable-cause entry within a KBEntry.
type KBCause struct {
	Cause       string  `json:"cause"`
	Probability float64 `json:"probability"`
	Test        string  `json:"test"`
}

// KBSolution contains a difficulty-rated set of repair steps.
type KBSolution struct {
	Difficulty string   `json:"difficulty"`
	Steps      []KBStep `json:"steps"`
}

// KBStep is a single repair instruction with a verification note.
type KBStep struct {
	Instruction  string `json:"instruction"`
	Verification string `json:"verification"`
	TimeSeconds  int    `json:"time_seconds"`
}

// RAGService loads knowledge-base JSON files at startup and performs
// retrieval-augmented generation (RAG) to enrich AI prompts with relevant
// problem context before sending frames for analysis.
//
// The retrieval strategy is determined by the configured VectorStore.
// The default is MemoryVectorStore which uses TF-IDF scoring.
// Future implementations (e.g. SQLiteVectorStore using sqlite-vec) can be
// swapped in without changing RAGService callers.
//
// All *.json files in the directory given by KB_DATA_DIR are loaded. Each file
// must have a top-level "domain" field; that value is stamped onto every
// KBEntry loaded from that file so callers can filter by domain.
// Default directory: ../kb/data/ (relative to server working directory).
// If no files are found the service starts in no-op mode: QueryKB always
// returns nil and the server continues to function without KB context.
type RAGService struct {
	entries           []KBEntry
	store             VectorStore            // all-domains store
	domainStores      map[string]VectorStore // per-domain stores
	domainEntryCounts map[string]int         // number of entries per domain
	domains           []string               // sorted list of loaded domain names
	cache             *ragLRU
	domainCaches      map[string]*ragLRU
}

// NewRAGService loads all *.json knowledge-base files from the directory given
// by KB_DATA_DIR (default ../kb/data/) and indexes documents into
// MemoryVectorStore instances using TF-IDF scoring — one global store and one
// per-domain store.
//
// A missing directory or no matching files is not an error — the service starts
// in no-op mode and QueryKB always returns nil.
func NewRAGService() (*RAGService, error) {
	return newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
}

// newRAGServiceWith creates a RAGService with a custom VectorStore factory.
// The factory is called once for the global store and once per loaded domain.
// Intended for tests and future extension points.
func newRAGServiceWith(storeFactory func() VectorStore) (*RAGService, error) {
	dir := os.Getenv("KB_DATA_DIR")
	if dir == "" {
		// Legacy single-file fallback: when KB_DATA_DIR is unset but KB_PATH is
		// set, derive the data directory from the file's parent directory so
		// that tests and deployments that still use KB_PATH keep working.
		if legacyPath := os.Getenv("KB_PATH"); legacyPath != "" {
			dir = filepath.Dir(legacyPath)
		}
	}
	if dir == "" {
		dir = "../kb/data/"
	}

	pattern := filepath.Join(dir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		// Malformed pattern — not expected, but treat as no-op.
		return &RAGService{
			store:             storeFactory(),
			domainStores:      make(map[string]VectorStore),
			domainEntryCounts: make(map[string]int),
			domainCaches:      make(map[string]*ragLRU),
			cache:             newRAGLRU(ragCacheCapacity, ragCacheTTL),
		}, nil
	}

	// Accumulate all entries across files.
	var allEntries []KBEntry
	domainEntries := make(map[string][]KBEntry)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			// Individual missing/unreadable file is non-fatal.
			continue
		}

		var kb kbFile
		if err := json.Unmarshal(data, &kb); err != nil {
			return nil, fmt.Errorf("parse knowledge base at %s: %w", f, err)
		}

		// Stamp domain onto each entry.
		for i := range kb.Problems {
			kb.Problems[i].Domain = kb.Domain
		}

		allEntries = append(allEntries, kb.Problems...)
		domainEntries[kb.Domain] = append(domainEntries[kb.Domain], kb.Problems...)
	}

	// Build global store.
	globalStore := storeFactory()
	globalStore.Index(allEntries)

	// Build per-domain stores and caches.
	domainStores := make(map[string]VectorStore, len(domainEntries))
	domainCaches := make(map[string]*ragLRU, len(domainEntries))
	domainEntryCounts := make(map[string]int, len(domainEntries))
	domains := make([]string, 0, len(domainEntries))
	for domain, entries := range domainEntries {
		ds := storeFactory()
		ds.Index(entries)
		domainStores[domain] = ds
		domainCaches[domain] = newRAGLRU(ragCacheCapacity, ragCacheTTL)
		domainEntryCounts[domain] = len(entries)
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	return &RAGService{
		entries:           allEntries,
		store:             globalStore,
		domainStores:      domainStores,
		domainEntryCounts: domainEntryCounts,
		domains:           domains,
		cache:             newRAGLRU(ragCacheCapacity, ragCacheTTL),
		domainCaches:      domainCaches,
	}, nil
}

// Loaded reports whether the knowledge base was successfully loaded.
func (r *RAGService) Loaded() bool { return len(r.entries) > 0 }

// QueryKB returns the top maxResults entries most relevant to query,
// ordered by descending TF-IDF relevance score. Results are served from an
// in-memory LRU cache when the same (query, maxResults) pair is repeated
// within the TTL window. Returns nil when the KB is empty or no terms match.
func (r *RAGService) QueryKB(query string, maxResults int) []KBEntry {
	if len(r.entries) == 0 || query == "" {
		return nil
	}

	if cached, ok := r.cache.get(query, maxResults); ok {
		metrics.RagCacheHitsTotal.Inc()
		return cached
	}

	metrics.RagCacheMissesTotal.Inc()
	results := r.store.Search(query, maxResults)
	r.cache.set(query, maxResults, results)
	return results
}

// KBDomains returns the sorted list of domain names currently loaded.
// Returns an empty slice when no files were loaded.
func (r *RAGService) KBDomains() []string { return r.domains }

// DomainEntryCount returns the number of KBEntry records loaded for the given
// domain. Returns 0 when the domain was not found or no KB was loaded.
func (r *RAGService) DomainEntryCount(domain string) int {
	return r.domainEntryCounts[domain]
}

// QueryKBByDomain performs RAG scoped to a single knowledge-base domain.
//
// Domain routing: each domain has its own TF-IDF index built at startup from
// the JSON files in KB_DATA_DIR. Queries are executed directly against the
// domain-specific index — results are not post-filtered from the global index
// — so relevance scores are calibrated to the domain vocabulary.
//
// Fallback behaviour:
//   - domain == ""               → delegates to [QueryKB] (global search)
//   - domain not in loaded set   → delegates to [QueryKB] (global search)
//
// Cache: each domain has its own LRU cache (capacity [ragCacheCapacity],
// TTL [ragCacheTTL]) independent of the global cache used by [QueryKB].
// Cache keys are derived from (SHA-256(query), maxResults), so different
// maxResults values for the same query are cached separately.
//
// Thread safety: safe for concurrent use. The domain stores and caches are
// read-only after construction; individual [ragLRU] instances are protected
// by their own mutex.
//
// Returns nil when query is empty or no terms match the domain index.
func (r *RAGService) QueryKBByDomain(domain, query string, maxResults int) []KBEntry {
	if domain == "" {
		return r.QueryKB(query, maxResults)
	}

	ds, ok := r.domainStores[domain]
	if !ok {
		return r.QueryKB(query, maxResults)
	}

	dc, hasDC := r.domainCaches[domain]
	if !hasDC {
		// Fallback: query without domain cache.
		return ds.Search(query, maxResults)
	}

	if len(query) == 0 {
		return nil
	}

	if cached, ok := dc.get(query, maxResults); ok {
		metrics.RagCacheHitsTotal.Inc()
		return cached
	}

	metrics.RagCacheMissesTotal.Inc()
	results := ds.Search(query, maxResults)
	dc.set(query, maxResults, results)
	return results
}

// FormatForPrompt formats KB results as a concise text block for injection
// into an AI conversation context. Returns an empty string when entries is nil.
func (r *RAGService) FormatForPrompt(entries []KBEntry) string {
	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("RELEVANT KNOWN ISSUES:\n")
	for _, e := range entries {
		sb.WriteString("- ")
		sb.WriteString(e.Name)
		sb.WriteString(": ")
		sb.WriteString(e.Description)
		if len(e.VisualSymptoms) > 0 {
			sb.WriteString(" Symptoms: ")
			sb.WriteString(strings.Join(e.VisualSymptoms, "; "))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
