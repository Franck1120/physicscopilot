package services

import (
	"context"
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

// TestNewAIBackendCaseSensitive verifies that AI_BACKEND values are case-sensitive:
// "Gemini", "GEMINI" are unknown backends — only lowercase "gemini" is valid.
func TestNewAIBackendCaseSensitive(t *testing.T) {
	cases := []struct {
		value string
	}{
		{"Gemini"},
		{"GEMINI"},
		{"OpenAI"},
		{"OPENAI"},
		{"Claude"},
		{"CLAUDE"},
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("AI_BACKEND", tc.value)

			_, err := NewAIBackend()
			if err == nil {
				t.Errorf("expected error for mixed-case AI_BACKEND %q, got nil", tc.value)
			}
		})
	}
}

// TestNewAIBackendWhitespaceReturnsError verifies that a backend name with
// surrounding whitespace is treated as unknown (no implicit trimming).
func TestNewAIBackendWhitespaceReturnsError(t *testing.T) {
	cases := []string{
		" gemini",
		"gemini ",
		" gemini ",
		"\tgemini",
		"gemini\n",
	}

	for _, val := range cases {
		t.Run("whitespace:"+val, func(t *testing.T) {
			t.Setenv("AI_BACKEND", val)

			_, err := NewAIBackend()
			if err == nil {
				t.Errorf("expected error for whitespace-padded AI_BACKEND %q, got nil", val)
			}
		})
	}
}

// TestGeminiServiceAnalyzeFrameCallable confirms that GeminiService satisfies
// the AIBackend interface and that AnalyzeFrame is callable. The call will
// fail (no real server) but must not panic.
func TestGeminiServiceAnalyzeFrameCallable(t *testing.T) {
	t.Setenv("AI_BACKEND", "gemini")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("CLIPROXY_URL", "http://127.0.0.1:19999") // nothing listens → fast error

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("NewAIBackend: %v", err)
	}

	gs, ok := backend.(*GeminiService)
	if !ok {
		t.Fatalf("expected *GeminiService, got %T", backend)
	}

	// Create a context that is already cancelled so the rate-limiter Wait
	// returns immediately with an error instead of blocking or making a real
	// network call.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The method must exist, compile, and return an error without panicking.
	_, callErr := gs.AnalyzeFrame(ctx, "", "", "it")
	if callErr == nil {
		// Unlikely (cancelled ctx / no server), but not a correctness failure.
		t.Log("AnalyzeFrame unexpectedly succeeded; no panic = OK")
	}
}
