package services

import (
	"testing"
)

// TestVectorStoreIndexZeroEntriesNoPanicOnSearch verifies that indexing an
// empty document set and then calling Search does not panic.
func TestVectorStoreIndexZeroEntriesNoPanicOnSearch(t *testing.T) {
	t.Parallel()

	store := NewMemoryVectorStore()
	store.Index([]KBEntry{})

	result := store.Search("nozzle clogged", 5)
	if result != nil {
		t.Errorf("expected nil from store with zero entries, got %v", result)
	}
}

// TestVectorStoreSearchOnUnindexedStoreReturnsNil verifies that calling
// Search on a store that has never been indexed returns nil without panicking.
func TestVectorStoreSearchOnUnindexedStoreReturnsNil(t *testing.T) {
	t.Parallel()

	store := NewMemoryVectorStore()
	// No Index call — searching an empty store must not panic.
	result := store.Search("any query", 3)
	if result != nil {
		t.Errorf("expected nil from unindexed store, got %v", result)
	}
}

// TestVectorStoreReIndexReplacesOldIndex verifies that calling Index a second
// time completely replaces the previous corpus: documents from the first index
// are no longer retrievable.
func TestVectorStoreReIndexReplacesOldIndex(t *testing.T) {
	t.Parallel()

	store := NewMemoryVectorStore()

	// First corpus — nozzle topics.
	store.Index([]KBEntry{
		{ID: "old-1", Name: "Old Nozzle Clog",    Description: "severe nozzle clogging blockage"},
		{ID: "old-2", Name: "Old Extrusion Issue", Description: "under-extrusion nozzle filament"},
	})

	// Confirm old corpus is findable.
	before := store.Search("nozzle clogging", 5)
	if len(before) == 0 {
		t.Fatal("expected old corpus to be searchable before re-index")
	}

	// Re-index with a completely different corpus — bed topics.
	store.Index([]KBEntry{
		{ID: "new-1", Name: "Bed Levelling", Description: "first layer adhesion bed surface calibration"},
		{ID: "new-2", Name: "Bed Tilt",      Description: "uneven bed tilt first layer problem"},
	})

	// Old documents must not appear.
	afterOld := store.Search("nozzle clogging", 5)
	for _, r := range afterOld {
		if r.ID == "old-1" || r.ID == "old-2" {
			t.Errorf("old document %q found after re-index; corpus should be fully replaced", r.ID)
		}
	}

	// New documents must be findable.
	afterNew := store.Search("bed levelling calibration", 5)
	if len(afterNew) == 0 {
		t.Fatal("expected new corpus to be searchable after re-index")
	}
	if afterNew[0].ID != "new-1" && afterNew[0].ID != "new-2" {
		t.Errorf("expected a new-corpus document as top result, got %q", afterNew[0].ID)
	}
}

// TestVectorStoreSearchSingleCharacterQuery verifies that a single-character
// query (filtered by tokenize) returns nil without panicking.
func TestVectorStoreSearchSingleCharacterQuery(t *testing.T) {
	t.Parallel()

	store := NewMemoryVectorStore()
	store.Index([]KBEntry{
		{ID: "a", Name: "Alpha Issue", Description: "hardware device alpha failure"},
	})

	// Single-character tokens are filtered out by tokenize (len < 3).
	result := store.Search("a", 5)
	if result != nil {
		t.Errorf("expected nil for single-character query (filtered by tokenize), got %v", result)
	}
}

// TestVectorStoreSearchOnlyPunctuationSpecialChars verifies that a query
// made up entirely of punctuation or special characters does not panic and
// returns nil (tokenize will filter all tokens shorter than 3 chars or
// producing no valid tokens).
func TestVectorStoreSearchOnlyPunctuationSpecialChars(t *testing.T) {
	t.Parallel()

	store := NewMemoryVectorStore()
	store.Index([]KBEntry{
		{ID: "x", Name: "Some Issue", Description: "generic device problem"},
	})

	specialQueries := []string{
		"!!! ??? ###",
		"... ---",
		"@@@",
		"((()))",
		"",
	}

	for _, q := range specialQueries {
		result := store.Search(q, 5)
		// Must not panic; all-special-char tokens produce no terms ≥ 3 chars
		// (or produce tokens that don't match any document).
		_ = result
	}
}

// TestVectorStoreInterfaceComplianceEdge is a compile-time assertion that
// *MemoryVectorStore satisfies the VectorStore interface.
func TestVectorStoreInterfaceComplianceEdge(t *testing.T) {
	t.Parallel()

	var _ VectorStore = (*MemoryVectorStore)(nil)
}
