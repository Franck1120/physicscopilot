package services

import (
	"encoding/json"
	"fmt"
	"os"
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

// RAGService loads a knowledge-base JSON file at startup and performs
// retrieval-augmented generation (RAG) to enrich AI prompts with relevant
// problem context before sending frames for analysis.
//
// The retrieval strategy is determined by the configured VectorStore.
// The default is MemoryVectorStore which uses TF-IDF scoring.
// Future implementations (e.g. SQLiteVectorStore using sqlite-vec) can be
// swapped in without changing RAGService callers.
//
// The knowledge-base path is read from the KB_PATH env var.
// Default: ../kb/data/problems.json (relative to server working directory).
// If the file is absent the service starts in no-op mode: QueryKB always
// returns nil and the server continues to function without KB context.
type RAGService struct {
	entries []KBEntry
	store   VectorStore
}

// NewRAGService loads the knowledge-base from the path given by KB_PATH
// (default ../kb/data/problems.json) and indexes documents into a
// MemoryVectorStore using TF-IDF scoring.
//
// A missing file is not an error — the service starts in no-op mode and
// QueryKB always returns nil.
func NewRAGService() (*RAGService, error) {
	return newRAGServiceWith(NewMemoryVectorStore())
}

// newRAGServiceWith creates a RAGService with a custom VectorStore.
// Intended for tests and future extension points.
func newRAGServiceWith(store VectorStore) (*RAGService, error) {
	path := os.Getenv("KB_PATH")
	if path == "" {
		path = "../kb/data/problems.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// KB is optional: log-worthy but not fatal.
		return &RAGService{store: store}, nil
	}

	var kb kbFile
	if err := json.Unmarshal(data, &kb); err != nil {
		return nil, fmt.Errorf("parse knowledge base at %s: %w", path, err)
	}

	store.Index(kb.Problems)
	return &RAGService{entries: kb.Problems, store: store}, nil
}

// Loaded reports whether the knowledge base was successfully loaded.
func (r *RAGService) Loaded() bool { return len(r.entries) > 0 }

// QueryKB returns the top maxResults entries most relevant to query,
// ordered by descending TF-IDF relevance score. Returns nil when the KB is
// empty or no terms match.
func (r *RAGService) QueryKB(query string, maxResults int) []KBEntry {
	if len(r.entries) == 0 || query == "" {
		return nil
	}
	return r.store.Search(query, maxResults)
}

// FormatForPrompt formats KB results as a concise text block for injection
// into an AI conversation context. Returns an empty string when entries is nil.
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
