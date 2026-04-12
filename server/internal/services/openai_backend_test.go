// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"strings"
	"testing"
)

// Compile-time interface compliance checks — these will fail to build if either
// backend stops satisfying the AIBackend interface.
var _ AIBackend = (*OpenAIBackend)(nil)
var _ AIBackend = (*ClaudeBackend)(nil)

// TestOpenAIBackendImplementsAIBackend documents the compile-time constraint
// above as a named, discoverable test.
func TestOpenAIBackendImplementsAIBackend(t *testing.T) {
	// Satisfied at compile time via the package-level var declaration above.
}

// TestClaudeBackendImplementsAIBackend documents the compile-time constraint
// above as a named, discoverable test.
func TestClaudeBackendImplementsAIBackend(t *testing.T) {
	// Satisfied at compile time via the package-level var declaration above.
}

// ── OpenAIBackend ─────────────────────────────────────────────────────────────

// TestOpenAIBackendRequiresAPIKey verifies that NewOpenAIBackend returns a
// non-nil error when OPENAI_API_KEY is not set.
func TestOpenAIBackendRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	backend, err := NewOpenAIBackend()
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY is empty, got nil")
	}
	if backend != nil {
		t.Error("expected nil backend when OPENAI_API_KEY is empty")
	}
}

// TestOpenAIBackendCreatedWithKey verifies that NewOpenAIBackend succeeds and
// returns a non-nil *OpenAIBackend when OPENAI_API_KEY is provided.
func TestOpenAIBackendCreatedWithKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-openai-key")

	backend, err := NewOpenAIBackend()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if backend == nil {
		t.Fatal("expected non-nil backend, got nil")
	}
}

// TestOpenAIBackendAnalyzeFrameNotImplemented verifies that AnalyzeFrame
// returns a non-nil error containing "not yet implemented".
func TestOpenAIBackendAnalyzeFrameNotImplemented(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-openai-key")

	backend, err := NewOpenAIBackend()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, analyzeErr := backend.AnalyzeFrame(context.Background(), "", "", "en")
	if analyzeErr == nil {
		t.Fatal("expected error from AnalyzeFrame, got nil")
	}
	if !strings.Contains(analyzeErr.Error(), "not yet implemented") {
		t.Errorf("error message: want substring %q, got %q", "not yet implemented", analyzeErr.Error())
	}
}

// ── ClaudeBackend ─────────────────────────────────────────────────────────────

// TestClaudeBackendRequiresAPIKey verifies that NewClaudeBackend returns a
// non-nil error when ANTHROPIC_API_KEY is not set.
func TestClaudeBackendRequiresAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	backend, err := NewClaudeBackend()
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is empty, got nil")
	}
	if backend != nil {
		t.Error("expected nil backend when ANTHROPIC_API_KEY is empty")
	}
}

// TestClaudeBackendCreatedWithKey verifies that NewClaudeBackend succeeds and
// returns a non-nil *ClaudeBackend when ANTHROPIC_API_KEY is provided.
func TestClaudeBackendCreatedWithKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")

	backend, err := NewClaudeBackend()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if backend == nil {
		t.Fatal("expected non-nil backend, got nil")
	}
}

// TestClaudeBackendAnalyzeFrameNotImplemented verifies that AnalyzeFrame
// returns a non-nil error containing "not yet implemented".
func TestClaudeBackendAnalyzeFrameNotImplemented(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")

	backend, err := NewClaudeBackend()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, analyzeErr := backend.AnalyzeFrame(context.Background(), "", "", "en")
	if analyzeErr == nil {
		t.Fatal("expected error from AnalyzeFrame, got nil")
	}
	if !strings.Contains(analyzeErr.Error(), "not yet implemented") {
		t.Errorf("error message: want substring %q, got %q", "not yet implemented", analyzeErr.Error())
	}
}
