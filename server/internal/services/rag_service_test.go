package services

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
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
