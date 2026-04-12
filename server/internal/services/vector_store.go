// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"math"
	"sort"
	"strings"
	"sync"
)

// VectorStore is the interface for similarity search over KBEntry documents.
// Implementations may use keyword matching, TF-IDF, or dense vector embeddings.
//
// Implementations must be safe for concurrent use.
type VectorStore interface {
	// Index loads documents into the store, replacing any previous corpus.
	// Must be called before Search.
	Index(docs []KBEntry)

	// Search returns up to topK documents most relevant to query,
	// ordered by descending relevance score.
	// Returns nil when the store is empty or no terms match.
	Search(query string, topK int) []KBEntry
}

// ── MemoryVectorStore ─────────────────────────────────────────────────────────

// MemoryVectorStore is the default VectorStore implementation.
// It stores documents in memory and scores them using TF-IDF weighting
// with field-priority boosts:
//
//	name + id + category → 3×
//	description          → 2×
//	visual_symptoms      → 1×
//
// The entire corpus is scanned on each Search call.
// At the current KB size (~66 entries) each search completes in < 1 ms.
//
// To swap in a persistent or embedding-based backend in future, implement
// VectorStore and pass it to RAGService via WithVectorStore.
//
// SQLiteVectorStore (sqlite-vec) will be the next implementation when
// the corpus grows past ~5 000 entries.
type MemoryVectorStore struct {
	mu      sync.RWMutex
	docs    []KBEntry
	vectors []docVector          // TF-IDF vectors per document
	idf     map[string]float64   // global IDF weights
}

// docVector maps term → TF-IDF weight within one document.
type docVector map[string]float64

// NewMemoryVectorStore returns an empty MemoryVectorStore.
// Call Index before searching.
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{}
}

// Index computes TF-IDF vectors for the given documents and stores them
// for fast retrieval. Concurrent calls to Search during Index are safe
// (protected by a write lock).
func (m *MemoryVectorStore) Index(docs []KBEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.docs = make([]KBEntry, len(docs))
	copy(m.docs, docs)

	// Step 1: term-frequency maps per document (weighted by field priority).
	tfs := make([]map[string]float64, len(docs))
	for i, doc := range docs {
		tfs[i] = computeDocTF(doc)
	}

	// Step 2: document frequency — how many docs contain each term.
	df := make(map[string]int, 256)
	for _, tf := range tfs {
		for term := range tf {
			df[term]++
		}
	}

	// Step 3: IDF = log(1 + N / (1 + df(t)))
	// The smoothing (+1) prevents division by zero and reduces the impact
	// of terms that appear in every document.
	N := float64(len(docs))
	m.idf = make(map[string]float64, len(df))
	for term, count := range df {
		m.idf[term] = math.Log(1 + N/float64(1+count))
	}

	// Step 4: TF-IDF = TF * IDF per term per document.
	m.vectors = make([]docVector, len(docs))
	for i, tf := range tfs {
		vec := make(docVector, len(tf))
		for term, tfVal := range tf {
			vec[term] = tfVal * m.idf[term]
		}
		m.vectors[i] = vec
	}
}

// Search returns up to topK documents ranked by TF-IDF cosine similarity
// to the query. Thread-safe.
func (m *MemoryVectorStore) Search(query string, topK int) []KBEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	terms := tokenize(query)
	if len(m.docs) == 0 || len(terms) == 0 {
		return nil
	}

	type hit struct {
		idx   int
		score float64
	}
	hits := make([]hit, 0, len(m.docs))

	for i, vec := range m.vectors {
		score := 0.0
		for _, term := range terms {
			score += vec[term]
		}
		if score > 0 {
			hits = append(hits, hit{idx: i, score: score})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].score > hits[j].score
	})

	if len(hits) == 0 {
		return nil
	}

	if topK > len(hits) {
		topK = len(hits)
	}

	out := make([]KBEntry, topK)
	for i := range out {
		out[i] = m.docs[hits[i].idx]
	}
	return out
}

// ── helpers ───────────────────────────────────────────────────────────────────

// computeDocTF builds the normalized, field-weighted term-frequency map
// for a KBEntry. Field weights: name/id/category=3, description=2, symptoms=1.
//
// The result is L1-normalised (total weight sums to 1) so documents of
// different lengths compare fairly.
func computeDocTF(entry KBEntry) map[string]float64 {
	freq := make(map[string]float64, 32)

	addTokens := func(text string, weight float64) {
		for _, tok := range tokenize(text) {
			freq[tok] += weight
		}
	}

	addTokens(entry.Name+" "+entry.ID+" "+entry.Category, 3)
	addTokens(entry.Description, 2)
	addTokens(strings.Join(entry.VisualSymptoms, " "), 1)

	// L1 normalisation: divide each weight by the sum of all weights.
	total := 0.0
	for _, v := range freq {
		total += v
	}
	if total > 0 {
		for term := range freq {
			freq[term] /= total
		}
	}
	return freq
}

// tokenize splits s into lowercase tokens, filtering words shorter than 3 chars.
// Kept in this file because MemoryVectorStore is its primary consumer.
func tokenize(s string) []string {
	words := strings.Fields(strings.ToLower(s))
	tokens := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) >= 3 {
			tokens = append(tokens, w)
		}
	}
	return tokens
}
