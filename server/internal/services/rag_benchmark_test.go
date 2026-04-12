// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// setupBenchRAGService creates a temporary KB directory with the given domain
// files and returns a ready-to-use RAGService. The temp dir is cleaned up via
// b.Cleanup. Calls b.Helper() so failures point to the caller.
func setupBenchRAGService(b *testing.B, domains map[string][]KBEntry) *RAGService {
	b.Helper()

	dir := b.TempDir()

	for domain, entries := range domains {
		kb := kbFile{Domain: domain, Problems: entries}
		data, err := json.Marshal(kb)
		if err != nil {
			b.Fatalf("marshal KB for domain %q: %v", domain, err)
		}
		path := filepath.Join(dir, domain+".json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			b.Fatalf("write KB file %q: %v", path, err)
		}
	}

	b.Setenv("KB_DATA_DIR", dir)
	b.Setenv("KB_PATH", "")

	svc, err := newRAGServiceWith(func() VectorStore { return NewMemoryVectorStore() })
	if err != nil {
		b.Fatalf("newRAGServiceWith: %v", err)
	}
	return svc
}

// benchKBEntries returns a small, realistic set of KB entries for benchmarks.
func benchKBEntries() []KBEntry {
	return []KBEntry{
		{
			ID:          "clog",
			Name:        "Clogged Nozzle",
			Category:    "extrusion",
			Severity:    "error",
			Description: "Nozzle is blocked causing severe under-extrusion",
			VisualSymptoms: []string{
				"thin filament strands",
				"gaps in layers",
				"no extrusion at all",
			},
		},
		{
			ID:          "warp",
			Name:        "Warping",
			Category:    "bed_adhesion",
			Severity:    "warning",
			Description: "Print corners lift off the heated bed during printing",
			VisualSymptoms: []string{
				"lifted corners",
				"detached base layer",
			},
		},
		{
			ID:          "layer",
			Name:        "Layer Adhesion Failure",
			Category:    "temperature",
			Severity:    "error",
			Description: "Layers not bonding properly due to insufficient temperature",
			VisualSymptoms: []string{
				"layer separation",
				"delamination",
			},
		},
	}
}

// BenchmarkRAGCacheHit measures repeated queries to the same key after the
// cache has been warmed. All b.N iterations should be served from the LRU cache.
// b.RunParallel stresses the mutex for concurrent read access.
func BenchmarkRAGCacheHit(b *testing.B) {
	b.ReportAllocs()

	svc := setupBenchRAGService(b, map[string][]KBEntry{
		"printers": benchKBEntries(),
	})

	const warmQuery = "clogged nozzle extrusion blocked"
	const maxResults = 3

	// Warm the cache with a single call before timing.
	_ = svc.QueryKB(warmQuery, maxResults)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = svc.QueryKB(warmQuery, maxResults)
		}
	})
}

// BenchmarkRAGCacheMiss forces a cache miss on every iteration by appending
// the iteration index to the query string, ensuring each call hits the vector
// store and inserts a new cache entry.
func BenchmarkRAGCacheMiss(b *testing.B) {
	b.ReportAllocs()

	svc := setupBenchRAGService(b, map[string][]KBEntry{
		"printers": benchKBEntries(),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("clogged nozzle extrusion %d", i)
		_ = svc.QueryKB(query, 3)
	}
}

// BenchmarkRAGFormatForPrompt measures the throughput of FormatForPrompt with a
// fixed 10-entry slice, isolating the string-building path.
func BenchmarkRAGFormatForPrompt(b *testing.B) {
	b.ReportAllocs()

	svc := &RAGService{}

	entries := make([]KBEntry, 10)
	for i := range entries {
		entries[i] = KBEntry{
			ID:          fmt.Sprintf("id-%02d", i),
			Name:        fmt.Sprintf("Problem %02d", i),
			Category:    "hardware",
			Severity:    "error",
			Description: fmt.Sprintf("Description for problem number %d in the knowledge base", i),
			VisualSymptoms: []string{
				fmt.Sprintf("symptom-a-%d", i),
				fmt.Sprintf("symptom-b-%d", i),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.FormatForPrompt(entries)
	}
}

// BenchmarkQueryKBByDomain measures domain-filtered RAG queries against a
// two-domain KB with a fixed query string (cache warms after the first hit).
func BenchmarkQueryKBByDomain(b *testing.B) {
	b.ReportAllocs()

	alphaEntries := benchKBEntries()
	betaEntries := []KBEntry{
		{
			ID:          "sw-crash",
			Name:        "Firmware Crash",
			Category:    "software",
			Severity:    "critical",
			Description: "Firmware crashes during print job due to memory overflow",
		},
		{
			ID:          "net-timeout",
			Name:        "Network Timeout",
			Category:    "connectivity",
			Severity:    "warning",
			Description: "Connection to remote server times out during slicing upload",
		},
	}

	svc := setupBenchRAGService(b, map[string][]KBEntry{
		"alpha": alphaEntries,
		"beta":  betaEntries,
	})

	const targetDomain = "alpha"
	const query = "extrusion nozzle filament"
	const maxResults = 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.QueryKBByDomain(targetDomain, query, maxResults)
	}
}
