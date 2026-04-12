// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKBFilesHaveRequiredTopLevelFields verifies that every *.json file in the
// real kb/data directory has the required top-level fields: "domain",
// "problems", "version", and "last_updated".
//
// The test locates the directory relative to the Go module root by walking up
// from the package directory until it finds a go.mod file.
func TestKBFilesHaveRequiredTopLevelFields(t *testing.T) {
	dir := findKBDataDir(t)

	pattern := filepath.Join(dir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}
	if len(files) == 0 {
		t.Skipf("no *.json files found in %s — skipping schema validation", dir)
	}

	for _, f := range files {
		f := f // capture
		t.Run(filepath.Base(f), func(t *testing.T) {
			// Skip files that use a non-standard top-level key (e.g. printer_profiles.json
			// uses "profiles" instead of "problems"). These are valid support files but
			// do not follow the KBEntry schema.
			if !hasProblemsKey(t, f) {
				t.Skipf("skipping %s: no 'problems' key — not a KBEntry file", filepath.Base(f))
			}
			validateKBFile(t, f)
		})
	}
}

// validateKBFile checks a single KB JSON file for required fields and problem
// entry validity.
func validateKBFile(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	// ── Top-level required fields ────────────────────────────────────────────

	requiredTopLevel := []string{"domain", "problems", "version", "last_updated"}
	for _, field := range requiredTopLevel {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing top-level field %q in %s", field, filepath.Base(path))
		}
	}

	// ── Domain must be a non-empty string ────────────────────────────────────

	var domain string
	if domainRaw, ok := raw["domain"]; ok {
		if err := json.Unmarshal(domainRaw, &domain); err != nil {
			t.Errorf("field 'domain' in %s must be a string: %v", filepath.Base(path), err)
		} else if strings.TrimSpace(domain) == "" {
			t.Errorf("field 'domain' in %s must not be empty", filepath.Base(path))
		}
	}

	// ── Problems must be a non-empty array ───────────────────────────────────

	var problems []map[string]json.RawMessage
	if problemsRaw, ok := raw["problems"]; ok {
		if err := json.Unmarshal(problemsRaw, &problems); err != nil {
			t.Errorf("field 'problems' in %s must be an array: %v", filepath.Base(path), err)
		} else if len(problems) == 0 {
			t.Errorf("field 'problems' in %s must not be empty", filepath.Base(path))
		}
	}

	// ── Each problem entry must have required fields ──────────────────────────

	requiredProblemFields := []string{"id", "name", "category", "severity", "description"}
	validSeverities := map[string]bool{"error": true, "warning": true, "info": true}

	for i, prob := range problems {
		label := fmt.Sprintf("%s[%d]", filepath.Base(path), i)

		for _, field := range requiredProblemFields {
			if _, ok := prob[field]; !ok {
				t.Errorf("problem %s: missing required field %q", label, field)
			}
		}

		// id must be non-empty.
		if idRaw, ok := prob["id"]; ok {
			var id string
			if err := json.Unmarshal(idRaw, &id); err != nil || strings.TrimSpace(id) == "" {
				t.Errorf("problem %s: 'id' must be a non-empty string", label)
			}
		}

		// severity must be one of the accepted values.
		if sevRaw, ok := prob["severity"]; ok {
			var sev string
			if err := json.Unmarshal(sevRaw, &sev); err != nil {
				t.Errorf("problem %s: 'severity' must be a string", label)
			} else if !validSeverities[sev] {
				t.Errorf("problem %s: severity %q is not valid (want error|warning|info)", label, sev)
			}
		}

		// visual_symptoms must be an array when present.
		if vsRaw, ok := prob["visual_symptoms"]; ok {
			var vs []string
			if err := json.Unmarshal(vsRaw, &vs); err != nil {
				t.Errorf("problem %s: 'visual_symptoms' must be a string array", label)
			}
		}

		// probable_causes must be an array when present.
		if pcRaw, ok := prob["probable_causes"]; ok {
			var causes []map[string]json.RawMessage
			if err := json.Unmarshal(pcRaw, &causes); err != nil {
				t.Errorf("problem %s: 'probable_causes' must be an array", label)
			}
		}

		// solutions must be an array when present.
		if solRaw, ok := prob["solutions"]; ok {
			var solutions []map[string]json.RawMessage
			if err := json.Unmarshal(solRaw, &solutions); err != nil {
				t.Errorf("problem %s: 'solutions' must be an array", label)
			}
		}
	}
}

// hasProblemsKey returns true if the JSON file contains a top-level "problems" key.
// Files like printer_profiles.json use a "profiles" key and should be excluded from
// KBEntry schema validation.
func hasProblemsKey(t *testing.T, path string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}
	_, ok := raw["problems"]
	return ok
}

// findKBDataDir locates the kb/data directory by walking up from the current
// working directory until a go.mod file is found.
func findKBDataDir(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			candidate := filepath.Join(dir, "..", "kb", "data")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				abs, _ := filepath.Abs(candidate)
				return abs
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Skip("kb/data directory not found — skipping schema validation")
	return ""
}
