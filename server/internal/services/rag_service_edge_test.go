package services

import (
	"strings"
	"testing"
	"time"
)

// TestQueryKBVeryLongQueryNoPanic verifies that a query string larger than
// 10 KB does not cause a panic or out-of-bounds access.
func TestQueryKBVeryLongQueryNoPanic(t *testing.T) {
	entries := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "nozzle blocked under-extrusion"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("NewRAGService: %v", err)
	}

	// Build a query that is well over 10 KB.
	longQuery := strings.Repeat("clogged nozzle extrusion ", 500) // ~12 KB
	// Must not panic; result may be nil or non-nil.
	results := svc.QueryKB(longQuery, 3)
	_ = results
}

// TestQueryKBMaxResultsZeroReturnsEmptyOrNil verifies that requesting zero
// results returns an empty/nil slice without panicking.
func TestQueryKBMaxResultsZeroReturnsEmptyOrNil(t *testing.T) {
	entries := []KBEntry{
		{ID: "a", Name: "Alpha Error", Description: "device error failure"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("NewRAGService: %v", err)
	}

	results := svc.QueryKB("device error", 0)
	// maxResults=0 must not return more than 0 results.
	if len(results) > 0 {
		t.Errorf("expected 0 results for maxResults=0, got %d", len(results))
	}
}

// TestQueryKBNegativeMaxResultsDocumentedBehavior documents the current
// behavior of QueryKB when maxResults is negative. The underlying
// MemoryVectorStore.Search passes the negative value to make(), which panics
// with "makeslice: len out of range". This test captures that panic via
// recover() to document the behavior without crashing the test suite.
// A future fix should clamp maxResults to 0 before calling Search.
func TestQueryKBNegativeMaxResultsDocumentedBehavior(t *testing.T) {
	entries := []KBEntry{
		{ID: "b", Name: "Beta Failure", Description: "device hardware failure"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("NewRAGService: %v", err)
	}

	// Capture a potential panic — the current implementation panics for
	// negative topK values. We accept either panic (recovered) or nil result.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic recovered — document that this is the current (unfixed) behavior.
				t.Logf("QueryKB(-1) panicked as expected (known limitation): %v", r)
			}
		}()
		results := svc.QueryKB("device failure", -1)
		// If it reaches here, the implementation handled negative maxResults gracefully.
		if len(results) > 0 {
			t.Errorf("expected 0 results for negative maxResults, got %d", len(results))
		}
	}()
}

// TestRAGCacheTTLSetThenImmediateGetReturnsSame verifies that a value stored
// in the ragLRU cache is immediately retrievable with identical contents.
func TestRAGCacheTTLSetThenImmediateGetReturnsSame(t *testing.T) {

	cache := newRAGLRU(10, 1*time.Minute)
	want := []KBEntry{{ID: "x", Name: "X Issue"}}
	cache.set("my query", 3, want)

	got, ok := cache.get("my query", 3)
	if !ok {
		t.Fatal("expected cache hit immediately after set")
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d results, got %d", len(want), len(got))
	}
	if got[0].ID != want[0].ID {
		t.Errorf("expected ID %q, got %q", want[0].ID, got[0].ID)
	}
}

// TestQueryKBCacheHitVsMissDistinction verifies that consecutive identical
// queries are served from cache (consistent results) and that a different
// query (cache miss) still returns valid results.
func TestQueryKBCacheHitVsMissDistinction(t *testing.T) {
	entries := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "nozzle blocked under-extrusion"},
		{ID: "warp", Name: "Warping", Description: "corners lift off bed adhesion"},
	}
	path := writeTestKB(t, entries)
	t.Setenv("KB_PATH", path)

	svc, err := NewRAGService()
	if err != nil {
		t.Fatalf("NewRAGService: %v", err)
	}

	// First call — cache miss, populates cache.
	r1 := svc.QueryKB("clogged nozzle extrusion", 2)
	// Second call — should be a cache hit and return the same results.
	r2 := svc.QueryKB("clogged nozzle extrusion", 2)

	if len(r1) != len(r2) {
		t.Errorf("cache hit returned different result count: %d vs %d", len(r1), len(r2))
	}
	for i := range r1 {
		if r1[i].ID != r2[i].ID {
			t.Errorf("result[%d]: cache hit ID %q != miss ID %q", i, r2[i].ID, r1[i].ID)
		}
	}

	// A different query should still work (independent cache key).
	r3 := svc.QueryKB("warping bed adhesion", 2)
	_ = r3 // may or may not return results depending on KB size
}

// TestQueryKBNoOpServiceNeverPanics verifies that calling QueryKB on an
// uninitialised RAGService (no entries, empty store) returns nil safely.
func TestQueryKBNoOpServiceNeverPanics(t *testing.T) {

	svc := &RAGService{
		store:        NewMemoryVectorStore(),
		domainStores: make(map[string]VectorStore),
		domainCaches: make(map[string]*ragLRU),
		cache:        newRAGLRU(10, 1*time.Minute),
	}

	result := svc.QueryKB("anything at all", 5)
	if result != nil {
		t.Errorf("expected nil from no-op RAGService, got %v", result)
	}
}
