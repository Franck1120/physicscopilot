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
	svc := &RAGService{entries: []KBEntry{{ID: "x", Name: "X"}}}
	if results := svc.QueryKB("", 5); results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

func TestQueryKBNoOpServiceReturnsNil(t *testing.T) {
	svc := &RAGService{} // no entries
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

func TestScoreEntryWeightsNameHigher(t *testing.T) {
	nameEntry := KBEntry{ID: "warp", Name: "warping problem", Description: "generic"}
	descEntry := KBEntry{ID: "other", Name: "other", Description: "warping problem here"}

	terms := tokenize("warping problem")
	nameScore := scoreEntry(nameEntry, terms)
	descScore := scoreEntry(descEntry, terms)

	if nameScore <= descScore {
		t.Errorf("expected name match to score higher (%d) than desc match (%d)", nameScore, descScore)
	}
}
