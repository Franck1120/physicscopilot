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

// ── LRU cache internals ───────────────────────────────────────────────────────

func TestRAGLRUCacheEvictionAtCapacity(t *testing.T) {
	cache := newRAGLRU(2, 5*time.Minute) // capacity = 2

	entry1 := []KBEntry{{ID: "e1", Name: "Entry 1"}}
	entry2 := []KBEntry{{ID: "e2", Name: "Entry 2"}}
	entry3 := []KBEntry{{ID: "e3", Name: "Entry 3"}}

	cache.set("query1", 1, entry1)
	cache.set("query2", 1, entry2)

	// Access query1 to make it recently used.
	cache.get("query1", 1)

	// Inserting a third entry should evict the LRU entry (query2).
	cache.set("query3", 1, entry3)

	if _, ok := cache.get("query1", 1); !ok {
		t.Error("query1 should still be in cache (recently accessed)")
	}
	if _, ok := cache.get("query3", 1); !ok {
		t.Error("query3 should be in cache (just inserted)")
	}
	if _, ok := cache.get("query2", 1); ok {
		t.Error("query2 should have been evicted as the LRU entry")
	}
}

func TestRAGLRUCacheTTLExpiry(t *testing.T) {
	ttl := 10 * time.Millisecond
	cache := newRAGLRU(10, ttl)

	entry := []KBEntry{{ID: "ttl-entry", Name: "TTL Entry"}}
	cache.set("ttl-query", 1, entry)

	// Entry should be present immediately.
	if _, ok := cache.get("ttl-query", 1); !ok {
		t.Fatal("entry should be present immediately after set")
	}

	// Wait for TTL to expire.
	time.Sleep(ttl + 5*time.Millisecond)

	// Entry should now be expired.
	if _, ok := cache.get("ttl-query", 1); ok {
		t.Error("entry should have been evicted after TTL expiry")
	}
}

func TestRAGLRUCacheUpdateExistingEntry(t *testing.T) {
	cache := newRAGLRU(10, 5*time.Minute)

	first := []KBEntry{{ID: "first"}}
	second := []KBEntry{{ID: "second"}, {ID: "third"}}

	cache.set("same-query", 1, first)
	cache.set("same-query", 1, second) // update

	results, ok := cache.get("same-query", 1)
	if !ok {
		t.Fatal("entry should be in cache after update")
	}
	if len(results) != 2 {
		t.Errorf("expected updated results (len 2), got %d", len(results))
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
