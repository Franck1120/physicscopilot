package services

import (
	"encoding/json"
	"os"
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
