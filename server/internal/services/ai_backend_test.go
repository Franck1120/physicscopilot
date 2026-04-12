package services

import (
	"context"
	"strings"
	"testing"
)

// TestAIBackendInterfaceCompliance verifies at compile time that all backend
// types satisfy the AIBackend interface.
func TestAIBackendInterfaceCompliance(t *testing.T) {
	var _ AIBackend = (*GeminiService)(nil)
	var _ AIBackend = (*OpenAIBackend)(nil)
	var _ AIBackend = (*ClaudeBackend)(nil)
}

// TestNewAIBackendDefaultsToGemini verifies that an empty AI_BACKEND env var
// resolves to GeminiService (which works with or without an API key because
// it falls back to the CLIProxyAPI when GEMINI_API_KEY is absent).
func TestNewAIBackendDefaultsToGemini(t *testing.T) {
	t.Setenv("AI_BACKEND", "")
	t.Setenv("GEMINI_API_KEY", "") // force proxy path so no real key needed
	t.Setenv("CLIPROXY_URL", "http://localhost:18085")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("expected no error for default backend, got: %v", err)
	}
	if _, ok := backend.(*GeminiService); !ok {
		t.Errorf("expected *GeminiService, got %T", backend)
	}
}

func TestNewAIBackendGeminiExplicit(t *testing.T) {
	t.Setenv("AI_BACKEND", "gemini")
	t.Setenv("GEMINI_API_KEY", "")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if _, ok := backend.(*GeminiService); !ok {
		t.Errorf("expected *GeminiService, got %T", backend)
	}
}

func TestNewAIBackendOpenAIRequiresKey(t *testing.T) {
	t.Setenv("AI_BACKEND", "openai")
	t.Setenv("OPENAI_API_KEY", "")

	_, err := NewAIBackend()
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY is missing, got nil")
	}
}

func TestNewAIBackendOpenAISucceedsWithKey(t *testing.T) {
	t.Setenv("AI_BACKEND", "openai")
	t.Setenv("OPENAI_API_KEY", "sk-test-key")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if _, ok := backend.(*OpenAIBackend); !ok {
		t.Errorf("expected *OpenAIBackend, got %T", backend)
	}
}

func TestNewAIBackendClaudeRequiresKey(t *testing.T) {
	t.Setenv("AI_BACKEND", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "")

	_, err := NewAIBackend()
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is missing, got nil")
	}
}

func TestNewAIBackendClaudeSucceedsWithKey(t *testing.T) {
	t.Setenv("AI_BACKEND", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if _, ok := backend.(*ClaudeBackend); !ok {
		t.Errorf("expected *ClaudeBackend, got %T", backend)
	}
}

func TestNewAIBackendUnknownReturnsError(t *testing.T) {
	t.Setenv("AI_BACKEND", "mistral")

	_, err := NewAIBackend()
	if err == nil {
		t.Fatal("expected error for unknown backend, got nil")
	}
}

// ---------------------------------------------------------------------------
// Stub AnalyzeFrame calls return "not yet implemented" errors
// ---------------------------------------------------------------------------

func TestOpenAIBackendAnalyzeFrameNotImplemented(t *testing.T) {
	t.Setenv("AI_BACKEND", "openai")
	t.Setenv("OPENAI_API_KEY", "sk-test")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = backend.AnalyzeFrame(context.Background(), "frame", "ctx", "it")
	if err == nil {
		t.Fatal("expected error from OpenAI stub")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in error, got: %v", err)
	}
}

func TestClaudeBackendAnalyzeFrameNotImplemented(t *testing.T) {
	t.Setenv("AI_BACKEND", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = backend.AnalyzeFrame(context.Background(), "frame", "ctx", "it")
	if err == nil {
		t.Fatal("expected error from Claude stub")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in error, got: %v", err)
	}
}
