package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// kbFile mirrors the top-level structure of the knowledge-base JSON.
type kbFile struct {
	Problems []KBEntry `json:"problems"`
}

// KBEntry is a single problem record from the knowledge base.
type KBEntry struct {
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

// kbResult pairs a KBEntry with its relevance score for internal sorting.
type kbResult struct {
	entry KBEntry
	score int
}

// RAGService loads a knowledge-base JSON file at startup and performs
// keyword-based retrieval to enrich Gemini prompts with relevant problem
// context before sending frames for analysis.
//
// Scalability note: the entire KB is loaded into memory as a slice of
// KBEntry structs and scanned linearly on every QueryKB call.
// At the current size (~66 entries, ~500 KB JSON) this is negligible —
// each query completes in < 1 ms. If the KB grows past ~5 000 entries,
// consider switching to an inverted index or a vector store (pgvector)
// to keep latency bounded. The QueryKB API signature is stable and
// can be replaced without changing callers.
//
// The knowledge-base path is read from the KB_PATH env var.
// Default: ../kb/data/problems.json (relative to server working directory).
// If the file is absent the service starts in no-op mode: QueryKB always
// returns nil and the server continues to function without KB context.
type RAGService struct {
	entries []KBEntry
}

// NewRAGService loads the knowledge-base from the path given by KB_PATH
// (default ../kb/data/problems.json). A missing file is not an error —
// the service starts in no-op mode and QueryKB always returns nil.
func NewRAGService() (*RAGService, error) {
	path := os.Getenv("KB_PATH")
	if path == "" {
		path = "../kb/data/problems.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// KB is optional: log-worthy but not fatal.
		return &RAGService{}, nil
	}

	var kb kbFile
	if err := json.Unmarshal(data, &kb); err != nil {
		return nil, fmt.Errorf("parse knowledge base at %s: %w", path, err)
	}

	return &RAGService{entries: kb.Problems}, nil
}

// Loaded reports whether the knowledge base was successfully loaded.
func (r *RAGService) Loaded() bool { return len(r.entries) > 0 }

// QueryKB returns the top maxResults entries most relevant to query,
// ordered by descending relevance score. Returns nil when the KB is
// empty or no terms match.
func (r *RAGService) QueryKB(query string, maxResults int) []KBEntry {
	if len(r.entries) == 0 || query == "" {
		return nil
	}

	terms := tokenize(query)
	if len(terms) == 0 {
		return nil
	}

	scored := make([]kbResult, 0, len(r.entries))
	for _, entry := range r.entries {
		if s := scoreEntry(entry, terms); s > 0 {
			scored = append(scored, kbResult{entry: entry, score: s})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if maxResults > len(scored) {
		maxResults = len(scored)
	}

	result := make([]KBEntry, maxResults)
	for i := range result {
		result[i] = scored[i].entry
	}
	return result
}

// FormatForPrompt formats KB results as a concise text block for injection
// into a Gemini conversation context. Returns an empty string when entries is nil.
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

// tokenize splits s into lowercase tokens, filtering words shorter than 3 chars.
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

// scoreEntry scores entry by counting query term occurrences across its
// searchable text fields with field-priority weights:
//
//	name/id/category → 3 pts per term
//	description      → 2 pts per term
//	visual_symptoms  → 1 pt per term
func scoreEntry(entry KBEntry, terms []string) int {
	name := strings.ToLower(entry.Name + " " + entry.ID + " " + entry.Category)
	desc := strings.ToLower(entry.Description)
	symptoms := strings.ToLower(strings.Join(entry.VisualSymptoms, " "))

	score := 0
	for _, term := range terms {
		switch {
		case strings.Contains(name, term):
			score += 3
		case strings.Contains(desc, term):
			score += 2
		case strings.Contains(symptoms, term):
			score += 1
		}
	}
	return score
}
