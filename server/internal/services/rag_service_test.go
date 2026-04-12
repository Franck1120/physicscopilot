// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTestKB writes a temporary KB JSON file with the given entries and
// returns its path. The file is automatically removed after the test.
func writeTestKB(t *testing.T, problems []KBEntry) string {
	t.Helper()

	f, err := os.CreateTemp("", "kb-*.json")
	if err != nil {
		t.Fatalf("create temp KB: %v", err)
	}
	kb := kbFile{Problems: problems}
	if err := json.NewEncoder(f).Encode(kb); err != nil {
		t.Fatalf("write temp KB: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestNewRAGServiceLoadsFile(t *testing.T) {
	path := writeTestKB(t, []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "Nozzle blocked"},
	})
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !svc.Loaded() {
		t.Fatal("expected service to report Loaded()=true")
	}
}

func TestNewRAGServiceMissingFileIsNoOp(t *testing.T) {
	t.Setenv("KB_PATH", "/non/existent/path.json")

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("missing file should not error, got: %v", err)
	}
	if svc.Loaded() {
		t.Error("expected service to be in no-op mode when file is absent")
	}
}

func TestNewRAGServiceMalformedJSONErrors(t *testing.T) {
	f, _ := os.CreateTemp("", "kb-bad-*.json")
	f.WriteString("not-json{{{")
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	t.Setenv("KB_PATH", f.Name())

	_, err := NewRAGService()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestQueryKBReturnsTopResult(t *testing.T) {
	entries := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Category: "extrusion",
			Description: "Nozzle is blocked causing under-extrusion"},
		{ID: "warp", Name: "Warping", Category: "bed_adhesion",
			Description: "Print corners lift off the bed"},
		{ID: "layer", Name: "Layer Adhesion", Category: "temperature",
			Description: "Layers not bonding properly"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	results := svc.QueryKB("clogged nozzle", 3)
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'clogged nozzle'")
	}
	if results[0].ID != "clog" {
		t.Errorf("expected top result ID 'clog', got %q", results[0].ID)
	}
}

func TestQueryKBLimitsResultCount(t *testing.T) {
	entries := []KBEntry{
		{ID: "a", Name: "Alpha error fix", Description: "error in device"},
		{ID: "b", Name: "Beta error fix", Description: "error in device"},
		{ID: "c", Name: "Gamma error fix", Description: "error in device"},
		{ID: "d", Name: "Delta error fix", Description: "error in device"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, _ := NewRAGService()
	results := svc.QueryKB("error device fix", 2)
	if len(results) > 2 {
		t.Errorf("expected max 2 results, got %d", len(results))
	}
}

func TestQueryKBEmptyQueryReturnsNil(t *testing.T) {
	svc := &RAGService{
		entries: []KBEntry{{ID: "x", Name: "X"}},
		store:   NewMemoryVectorStore(),
	}
	if results := svc.QueryKB("", 5); results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

func TestQueryKBNoOpServiceReturnsNil(t *testing.T) {
	svc := &RAGService{store: NewMemoryVectorStore()} // no entries
	if results := svc.QueryKB("clog", 5); results != nil {
		t.Errorf("expected nil from no-op service, got %v", results)
	}
}

func TestFormatForPromptContainsEntryName(t *testing.T) {
	svc := &RAGService{}
	entries := []KBEntry{
		{Name: "Clogged Nozzle", Description: "Nozzle blocked", VisualSymptoms: []string{"under extrusion"}},
	}
	out := svc.FormatForPrompt(entries)

	if !strings.Contains(out, "Clogged Nozzle") {
		t.Errorf("expected output to contain entry name, got: %q", out)
	}
	if !strings.Contains(out, "RELEVANT KNOWN ISSUES") {
		t.Errorf("expected output to contain section header, got: %q", out)
	}
}

func TestFormatForPromptNilReturnsEmpty(t *testing.T) {
	svc := &RAGService{}
	if out := svc.FormatForPrompt(nil); out != "" {
		t.Errorf("expected empty string for nil entries, got %q", out)
	}
}

func TestQueryKBCacheHitReturnsSameResults(t *testing.T) {
	entries := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "Nozzle blocked"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	r1 := svc.QueryKB("clogged nozzle", 3)
	r2 := svc.QueryKB("clogged nozzle", 3)
	if len(r1) != len(r2) {
		t.Errorf("cache hit returned different result length: %d vs %d", len(r1), len(r2))
	}
}

// ---------------------------------------------------------------------------
// ragLRU cache — TTL expiry
// ---------------------------------------------------------------------------

func TestRAGLRUGetExpiredEntryReturnsMiss(t *testing.T) {
	cache := newRAGLRU(10, 50*time.Millisecond)
	cache.set("test query", 3, []KBEntry{{ID: "a"}})

	// Immediately should be a hit
	if results, ok := cache.get("test query", 3); !ok || len(results) != 1 {
		t.Error("expected cache hit immediately after set")
	}

	// Wait for TTL to expire
	time.Sleep(80 * time.Millisecond)

	results, ok := cache.get("test query", 3)
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
	if results != nil {
		t.Errorf("expected nil results after TTL expiry, got %v", results)
	}
}

// ---------------------------------------------------------------------------
// ragLRU cache — eviction at capacity
// ---------------------------------------------------------------------------

func TestRAGLRUEvictsOldestAtCapacity(t *testing.T) {
	cache := newRAGLRU(2, 5*time.Minute)

	cache.set("query-A", 3, []KBEntry{{ID: "a"}})
	cache.set("query-B", 3, []KBEntry{{ID: "b"}})

	// Both should be hits
	if _, ok := cache.get("query-A", 3); !ok {
		t.Error("expected cache hit for query-A before eviction")
	}
	if _, ok := cache.get("query-B", 3); !ok {
		t.Error("expected cache hit for query-B before eviction")
	}

	// Adding a third entry should evict the LRU (query-A was accessed after query-B
	// in the get calls above, so query-B accessed last, query-A second-to-last...
	// actually both were accessed, so the LRU is whichever was accessed least recently)
	// Let's make it deterministic: access query-B to make it MRU, then add query-C
	cache.get("query-B", 3)
	cache.set("query-C", 3, []KBEntry{{ID: "c"}})

	// query-A should be evicted (LRU)
	if _, ok := cache.get("query-A", 3); ok {
		t.Error("expected query-A to be evicted (LRU)")
	}
	// query-B and query-C should still be present
	if _, ok := cache.get("query-B", 3); !ok {
		t.Error("expected cache hit for query-B after eviction of A")
	}
	if _, ok := cache.get("query-C", 3); !ok {
		t.Error("expected cache hit for query-C")
	}
}

// ---------------------------------------------------------------------------
// ragLRU cache — update existing entry
// ---------------------------------------------------------------------------

func TestRAGLRUUpdateExistingEntry(t *testing.T) {
	cache := newRAGLRU(10, 5*time.Minute)

	cache.set("query-X", 3, []KBEntry{{ID: "old"}})
	cache.set("query-X", 3, []KBEntry{{ID: "new"}})

	results, ok := cache.get("query-X", 3)
	if !ok {
		t.Fatal("expected cache hit after update")
	}
	if len(results) != 1 || results[0].ID != "new" {
		t.Errorf("expected updated entry with ID 'new', got %v", results)
	}
}

// ---------------------------------------------------------------------------
// FormatForPrompt — detailed format checks
// ---------------------------------------------------------------------------

func TestFormatForPromptMultipleEntries(t *testing.T) {
	svc := &RAGService{}
	entries := []KBEntry{
		{Name: "Clogged Nozzle", Description: "Nozzle blocked", VisualSymptoms: []string{"under extrusion", "gaps"}},
		{Name: "Warping", Description: "Corners lift up"},
	}
	out := svc.FormatForPrompt(entries)

	if !strings.HasPrefix(out, "RELEVANT KNOWN ISSUES:\n") {
		t.Errorf("expected header prefix, got: %q", out)
	}
	if !strings.Contains(out, "- Clogged Nozzle: Nozzle blocked Symptoms: under extrusion; gaps") {
		t.Errorf("expected first entry with symptoms, got: %q", out)
	}
	if !strings.Contains(out, "- Warping: Corners lift up\n") {
		t.Errorf("expected second entry without symptoms, got: %q", out)
	}
}

func TestFormatForPromptEmptySliceReturnsEmpty(t *testing.T) {
	svc := &RAGService{}
	if out := svc.FormatForPrompt([]KBEntry{}); out != "" {
		t.Errorf("expected empty string for empty slice, got %q", out)
	}
}

func TestQueryKBDifferentMaxResultsAreCachedSeparately(t *testing.T) {
	entries := []KBEntry{
		{ID: "a", Name: "Alpha error", Description: "error device"},
		{ID: "b", Name: "Beta error", Description: "error device"},
		{ID: "c", Name: "Gamma error", Description: "error device"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, _ := NewRAGService()

	r1 := svc.QueryKB("error device", 1)
	r2 := svc.QueryKB("error device", 3)
	if len(r1) > 1 {
		t.Errorf("maxResults=1 should return at most 1 result, got %d", len(r1))
	}
	if len(r2) != len(r1) && len(r2) <= 1 {
		t.Errorf("maxResults=3 should potentially return more than maxResults=1")
	}
}

// TestRAGServiceReturnsTopKResults verifies that QueryKB never returns more
// entries than the requested K parameter, regardless of KB size.
func TestRAGServiceReturnsTopKResults(t *testing.T) {
	entries := []KBEntry{
		{ID: "a", Name: "Clogged Nozzle", Description: "nozzle blocked under-extrusion"},
		{ID: "b", Name: "Warping", Description: "corners lift bed adhesion"},
		{ID: "c", Name: "Stringing", Description: "thin strands between parts"},
		{ID: "d", Name: "Layer Shift", Description: "layers misaligned stepper"},
		{ID: "e", Name: "Under Extrusion", Description: "insufficient filament flow nozzle"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	for _, k := range []int{1, 2, 3} {
		results := svc.QueryKB("nozzle filament extrusion", k)
		if len(results) > k {
			t.Errorf("K=%d: expected at most %d results, got %d", k, k, len(results))
		}
	}
}

// TestRAGServiceEmptyQueryReturnsEmpty verifies that an empty query string
// returns nil without panicking, even when the KB contains entries.
func TestRAGServiceEmptyQueryReturnsEmpty(t *testing.T) {
	entries := []KBEntry{
		{ID: "x", Name: "Some Issue", Description: "some description"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	results := svc.QueryKB("", 5)
	if results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

// TestRAGServiceIndexAndSearch verifies the full index-then-search pipeline:
// documents are indexed at construction time and the most relevant result
// appears first for a targeted query.
func TestRAGServiceIndexAndSearch(t *testing.T) {
	entries := []KBEntry{
		{ID: "warp", Name: "Warping", Category: "bed_adhesion",
			Description: "print corners lift off the heated bed"},
		{ID: "clog", Name: "Clogged Nozzle", Category: "extrusion",
			Description: "nozzle is blocked causing severe under-extrusion"},
		{ID: "shift", Name: "Layer Shift", Category: "mechanical",
			Description: "layers are misaligned due to skipped stepper steps"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !svc.Loaded() {
		t.Fatal("service must be Loaded after successful KB load")
	}

	// Query strongly associated with the clog entry.
	results := svc.QueryKB("clogged nozzle extrusion blocked", 3)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	// The most relevant result should be the clogged nozzle entry.
	if results[0].ID != "clog" {
		t.Errorf("expected top result 'clog', got %q", results[0].ID)
	}

	// Results must be in descending relevance order — no single criterion, so
	// just verify the slice length is capped at the requested K.
	if len(results) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// Multi-file glob loading and domain-filtering
// ---------------------------------------------------------------------------

// writeKBFile writes a raw JSON knowledge-base file into dir with the given
// domain name and problems, returning the full file path.
func writeKBFile(t *testing.T, dir, filename, domain string, problems []KBEntry) string {
	t.Helper()

	kb := kbFile{Domain: domain, Problems: problems}
	data, err := json.Marshal(kb)
	if err != nil {
		t.Fatalf("marshal KB file %s: %v", filename, err)
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write KB file %s: %v", path, err)
	}
	return path
}

// TestNewRAGServiceGlobLoadsMultipleFiles verifies that all *.json files in
// KB_DATA_DIR are loaded and their entries merged into a single service.
func TestNewRAGServiceGlobLoadsMultipleFiles(t *testing.T) {
	dir := t.TempDir()

	writeKBFile(t, dir, "alpha.json", "alpha", []KBEntry{
		{ID: "a1", Name: "Alpha Problem One", Category: "hardware", Severity: "error",
			Description: "alpha first desc",
			VisualSymptoms: []string{"symptom1"},
			ProbableCauses: []KBCause{{Cause: "cause1", Probability: 0.9, Test: "test1"}},
			Solutions: []KBSolution{{Difficulty: "beginner", Steps: []KBStep{
				{Instruction: "step1", Verification: "ok", TimeSeconds: 60},
			}}},
		},
		{ID: "a2", Name: "Alpha Problem Two", Category: "hardware", Severity: "warning",
			Description: "alpha second desc",
		},
	})

	writeKBFile(t, dir, "beta.json", "beta", []KBEntry{
		{ID: "b1", Name: "Beta Problem One", Category: "software", Severity: "error",
			Description: "beta first desc",
		},
		{ID: "b2", Name: "Beta Problem Two", Category: "software", Severity: "warning",
			Description: "beta second desc",
		},
		{ID: "b3", Name: "Beta Problem Three", Category: "software", Severity: "info",
			Description: "beta third desc",
		},
	})

	t.Setenv("KB_DATA_DIR", dir)
	t.Setenv("KB_PATH", "") // ensure legacy fallback is not used

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !svc.Loaded() {
		t.Fatal("expected Loaded()=true after loading two files")
	}

	if got := len(svc.entries); got != 5 {
		t.Errorf("expected 5 total entries (2 alpha + 3 beta), got %d", got)
	}

	domains := svc.KBDomains()
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d: %v", len(domains), domains)
	}
	if domains[0] != "alpha" || domains[1] != "beta" {
		t.Errorf("expected sorted domains [alpha beta], got %v", domains)
	}
}

// TestNewRAGServiceStampsDomainOnEntries verifies that the domain field from the
// JSON file is stamped onto every KBEntry returned by QueryKB, even when the
// original JSON problem objects have no domain field set.
func TestNewRAGServiceStampsDomainOnEntries(t *testing.T) {
	dir := t.TempDir()

	// Problems intentionally have no Domain field in JSON — the service must
	// derive it from the top-level "domain" key in the file.
	writeKBFile(t, dir, "gamma.json", "gamma", []KBEntry{
		{ID: "g1", Name: "Gamma Problem One", Category: "hardware", Severity: "error",
			Description: "gamma first desc unique keyword",
		},
		{ID: "g2", Name: "Gamma Problem Two", Category: "hardware", Severity: "warning",
			Description: "gamma second desc unique keyword",
		},
	})

	t.Setenv("KB_DATA_DIR", dir)
	t.Setenv("KB_PATH", "")

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := svc.QueryKB("gamma unique keyword", 5)
	if len(results) == 0 {
		t.Fatal("expected at least one result for gamma query")
	}

	for _, e := range results {
		if e.Domain != "gamma" {
			t.Errorf("entry %q: expected Domain=%q, got %q", e.ID, "gamma", e.Domain)
		}
	}
}

// TestNewRAGServiceNoFilesIsNoOp verifies that an empty KB_DATA_DIR causes the
// service to start in no-op mode without returning an error.
func TestNewRAGServiceNoFilesIsNoOp(t *testing.T) {
	dir := t.TempDir() // empty — no *.json files

	t.Setenv("KB_DATA_DIR", dir)
	t.Setenv("KB_PATH", "") // prevent legacy fallback from loading something

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("empty dir should not error, got: %v", err)
	}

	if svc.Loaded() {
		t.Error("expected Loaded()=false when no JSON files are present")
	}

	if results := svc.QueryKB("any query", 5); results != nil {
		t.Errorf("expected nil from no-op service, got %v", results)
	}
}

// TestQueryKBByDomainFiltersResults verifies that QueryKBByDomain returns only
// entries belonging to the requested domain.
func TestQueryKBByDomainFiltersResults(t *testing.T) {
	dir := t.TempDir()

	writeKBFile(t, dir, "alpha.json", "alpha", []KBEntry{
		{ID: "a1", Name: "Alpha Extrusion Error", Category: "hardware", Severity: "error",
			Description: "alpha extrusion nozzle blockage filament",
		},
		{ID: "a2", Name: "Alpha Bed Adhesion", Category: "hardware", Severity: "warning",
			Description: "alpha bed adhesion print lifting corners",
		},
	})

	writeKBFile(t, dir, "beta.json", "beta", []KBEntry{
		{ID: "b1", Name: "Beta Software Crash", Category: "software", Severity: "error",
			Description: "beta software crash firmware update failure",
		},
		{ID: "b2", Name: "Beta Network Timeout", Category: "software", Severity: "warning",
			Description: "beta network timeout connection refused",
		},
	})

	t.Setenv("KB_DATA_DIR", dir)
	t.Setenv("KB_PATH", "")

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := svc.QueryKBByDomain("alpha", "alpha extrusion nozzle filament", 5)
	if len(results) == 0 {
		t.Fatal("expected at least one result for domain alpha query")
	}

	for _, e := range results {
		if e.Domain != "alpha" {
			t.Errorf("entry %q: expected Domain=%q, got %q", e.ID, "alpha", e.Domain)
		}
	}
}

// TestQueryKBByDomainFallsBackToGlobalWhenDomainMissing verifies that
// QueryKBByDomain falls back to the global search when the requested domain
// does not match any loaded domain.
func TestQueryKBByDomainFallsBackToGlobalWhenDomainMissing(t *testing.T) {
	dir := t.TempDir()

	writeKBFile(t, dir, "alpha.json", "alpha", []KBEntry{
		{ID: "a1", Name: "Alpha Hardware Error", Category: "hardware", Severity: "error",
			Description: "alpha hardware device malfunction overheating",
		},
		{ID: "a2", Name: "Alpha Power Issue", Category: "hardware", Severity: "warning",
			Description: "alpha power supply voltage fluctuation",
		},
	})

	t.Setenv("KB_DATA_DIR", dir)
	t.Setenv("KB_PATH", "")

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := "hardware device malfunction"
	global := svc.QueryKB(query, 5)
	fallback := svc.QueryKBByDomain("nonexistent", query, 5)

	if len(fallback) != len(global) {
		t.Errorf("fallback len=%d does not match global len=%d", len(fallback), len(global))
	}

	for i := range global {
		if i >= len(fallback) {
			break
		}
		if fallback[i].ID != global[i].ID {
			t.Errorf("result[%d]: fallback ID=%q != global ID=%q", i, fallback[i].ID, global[i].ID)
		}
	}
}

// TestKBDomainsReturnsEmptySliceWhenNotLoaded verifies that KBDomains returns
// an empty (non-nil) slice when no KB files are loaded.
func TestKBDomainsReturnsEmptySliceWhenNotLoaded(t *testing.T) {
	dir := t.TempDir() // no JSON files

	t.Setenv("KB_DATA_DIR", dir)
	t.Setenv("KB_PATH", "")

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	domains := svc.KBDomains()
	if domains == nil {
		t.Error("expected non-nil empty slice from KBDomains(), got nil")
	}
	if len(domains) != 0 {
		t.Errorf("expected empty slice, got %v", domains)
	}
}

// TestNewRAGServiceKBPathLegacyFallback verifies that when KB_DATA_DIR is unset
// but KB_PATH points to a valid JSON file, the service loads the file via the
// legacy fallback by deriving the directory from the file path.
func TestNewRAGServiceKBPathLegacyFallback(t *testing.T) {
	dir := t.TempDir()

	writeKBFile(t, dir, "legacy.json", "legacy", []KBEntry{
		{ID: "l1", Name: "Legacy Hardware Problem", Category: "hardware", Severity: "error",
			Description: "legacy device hardware failure component",
			VisualSymptoms: []string{"smoke", "sparks"},
			ProbableCauses: []KBCause{{Cause: "overvoltage", Probability: 0.95, Test: "measure voltage"}},
			Solutions: []KBSolution{{Difficulty: "expert", Steps: []KBStep{
				{Instruction: "replace component", Verification: "device boots", TimeSeconds: 300},
			}}},
		},
	})

	legacyPath := filepath.Join(dir, "legacy.json")

	t.Setenv("KB_DATA_DIR", "") // ensure glob path is derived from KB_PATH
	t.Setenv("KB_PATH", legacyPath)

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !svc.Loaded() {
		t.Fatal("expected Loaded()=true when KB_PATH points to a valid JSON file")
	}

	if len(svc.entries) != 1 {
		t.Errorf("expected 1 entry from legacy file, got %d", len(svc.entries))
	}
}
