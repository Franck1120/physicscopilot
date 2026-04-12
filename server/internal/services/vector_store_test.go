package services

import (
	"testing"
)

// ── tokenize ─────────────────────────────────────────────────────────────────

func TestTokenizeFiltersShortWords(t *testing.T) {
	tokens := tokenize("a bb ccc dddd")
	for _, tok := range tokens {
		if len(tok) < 3 {
			t.Errorf("expected all tokens >= 3 chars, got %q", tok)
		}
	}
	if len(tokens) != 2 { // "ccc" and "dddd"
		t.Errorf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestTokenizeLowercases(t *testing.T) {
	tokens := tokenize("Nozzle CLOGGED Warping")
	for _, tok := range tokens {
		for _, ch := range tok {
			if ch >= 'A' && ch <= 'Z' {
				t.Errorf("expected all tokens lowercase, got %q", tok)
			}
		}
	}
}

func TestTokenizeEmpty(t *testing.T) {
	if tokens := tokenize(""); len(tokens) != 0 {
		t.Errorf("expected empty slice for empty input, got %v", tokens)
	}
}

// ── computeDocTF ─────────────────────────────────────────────────────────────

func TestComputeDocTFNormalizesToOne(t *testing.T) {
	entry := KBEntry{
		Name:        "clogged nozzle",
		Description: "nozzle blockage",
	}
	tf := computeDocTF(entry)

	total := 0.0
	for _, v := range tf {
		total += v
	}
	// Allow small floating-point tolerance
	if total < 0.99 || total > 1.01 {
		t.Errorf("expected TF weights to sum to ~1.0, got %f", total)
	}
}

func TestComputeDocTFNameWeightedHigher(t *testing.T) {
	// "nozzle" appears in Name (weight 3) and Description (weight 2)
	// The name field should dominate
	entry := KBEntry{
		Name:        "nozzle clog",
		Description: "generic extrusion problem",
	}
	tf := computeDocTF(entry)

	nozzleWeight := tf["nozzle"]
	genericWeight := tf["generic"]

	if nozzleWeight <= genericWeight {
		t.Errorf("expected 'nozzle' (in name) to have higher TF than 'generic' (in desc): nozzle=%f generic=%f", nozzleWeight, genericWeight)
	}
}

// ── MemoryVectorStore ─────────────────────────────────────────────────────────

func TestMemoryVectorStoreEmptySearchReturnsNil(t *testing.T) {
	store := NewMemoryVectorStore()
	store.Index([]KBEntry{})

	if got := store.Search("nozzle", 5); got != nil {
		t.Errorf("expected nil for empty store, got %v", got)
	}
}

func TestMemoryVectorStoreEmptyQueryReturnsNil(t *testing.T) {
	store := NewMemoryVectorStore()
	store.Index([]KBEntry{{ID: "x", Name: "Nozzle clog", Description: "blocked"}})

	if got := store.Search("", 5); got != nil {
		t.Errorf("expected nil for empty query, got %v", got)
	}
}

func TestMemoryVectorStoreSearchRanksRelevant(t *testing.T) {
	docs := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Category: "extrusion",
			Description: "Nozzle is blocked causing under-extrusion"},
		{ID: "warp", Name: "Warping Issue", Category: "bed",
			Description: "Print corners lift off the bed during printing"},
		{ID: "layer", Name: "Layer Separation", Category: "temperature",
			Description: "Layers not bonding properly due to low temperature"},
	}
	store := NewMemoryVectorStore()
	store.Index(docs)

	results := store.Search("clogged nozzle extrusion", 3)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].ID != "clog" {
		t.Errorf("expected 'clog' as top result, got %q", results[0].ID)
	}
}

func TestMemoryVectorStoreSearchRespectsTopK(t *testing.T) {
	docs := []KBEntry{
		{ID: "a", Name: "Alpha failure", Description: "device error failure mode"},
		{ID: "b", Name: "Beta failure", Description: "device error failure mode"},
		{ID: "c", Name: "Gamma failure", Description: "device error failure mode"},
	}
	store := NewMemoryVectorStore()
	store.Index(docs)

	results := store.Search("failure error", 2)
	if len(results) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(results))
	}
}

func TestMemoryVectorStoreSearchNoMatchReturnsNil(t *testing.T) {
	docs := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "nozzle blocked"},
	}
	store := NewMemoryVectorStore()
	store.Index(docs)

	// Query has no overlap with any document
	results := store.Search("bed adhesion warping temperature", 5)
	if results != nil {
		t.Errorf("expected nil when no terms match, got %v", results)
	}
}

func TestMemoryVectorStoreTFIDFPrefersDomainSpecificTerms(t *testing.T) {
	// "the" appears in all docs → low IDF; "clogged" appears in one → high IDF.
	// A query for "clogged" should still rank the clog doc highest.
	docs := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "nozzle completely clogged during the print"},
		{ID: "warp", Name: "Warping", Description: "print lifts during the process"},
		{ID: "layer", Name: "Layer Issues", Description: "layers separate during the print"},
	}
	store := NewMemoryVectorStore()
	store.Index(docs)

	results := store.Search("clogged nozzle", 3)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].ID != "clog" {
		t.Errorf("expected 'clog' ranked first, got %q", results[0].ID)
	}
}

func TestMemoryVectorStoreIndexReplacesPreviousCorpus(t *testing.T) {
	store := NewMemoryVectorStore()
	// First index
	store.Index([]KBEntry{{ID: "old", Name: "Old problem", Description: "obsolete issue"}})
	// Re-index with different corpus
	store.Index([]KBEntry{{ID: "new", Name: "New problem", Description: "current issue"}})

	results := store.Search("new problem", 5)
	if len(results) == 0 {
		t.Fatal("expected at least one result after re-index")
	}
	if results[0].ID != "new" {
		t.Errorf("expected 'new' after re-index, got %q", results[0].ID)
	}
}

// ── MemoryVectorStore — topK larger than results ────────────────────────────

func TestMemoryVectorStoreTopKLargerThanHits(t *testing.T) {
	docs := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Description: "nozzle blocked"},
	}
	store := NewMemoryVectorStore()
	store.Index(docs)

	// topK=100, but only 1 doc matches
	results := store.Search("clogged nozzle", 100)
	if len(results) != 1 {
		t.Errorf("expected 1 result (topK > hits), got %d", len(results))
	}
}

// ── MemoryVectorStore — concurrent access ───────────────────────────────────

func TestMemoryVectorStoreConcurrentSearchDuringIndex(t *testing.T) {
	store := NewMemoryVectorStore()
	docs := []KBEntry{
		{ID: "a", Name: "Alpha Failure", Description: "device error"},
		{ID: "b", Name: "Beta Failure", Description: "device error"},
	}
	store.Index(docs)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			store.Search("device failure", 5)
		}
	}()

	// Re-index while searching
	for i := 0; i < 10; i++ {
		store.Index(docs)
	}
	<-done
}

// ── computeDocTF — empty entry ──────────────────────────────────────────────

func TestComputeDocTFEmptyEntry(t *testing.T) {
	tf := computeDocTF(KBEntry{})
	if len(tf) != 0 {
		t.Errorf("expected empty TF map for empty entry, got %d entries", len(tf))
	}
}

// ── VectorStore interface compliance ─────────────────────────────────────────

// TestVectorStoreInterfaceCompliance verifies that MemoryVectorStore
// satisfies the VectorStore interface at compile time.
func TestVectorStoreInterfaceCompliance(t *testing.T) {
	var _ VectorStore = NewMemoryVectorStore()
}
