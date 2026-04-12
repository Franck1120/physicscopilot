package services

import (
	"context"
	"testing"
)

// TestNewAIBackendNoEnvVarsReturnsGemini verifies that when no AI_BACKEND or
// GEMINI_API_KEY env vars are set, NewAIBackend defaults to GeminiService.
// This is the "no env vars" scenario: the proxy URL is pointed at a non-
// existent address so no real network connection is made.
func TestNewAIBackendNoEnvVarsReturnsGemini(t *testing.T) {
	t.Setenv("AI_BACKEND", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("CLIPROXY_URL", "http://127.0.0.1:19998") // nothing listens here

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("NewAIBackend with no env vars: unexpected error: %v", err)
	}
	if backend == nil {
		t.Fatal("NewAIBackend returned nil backend")
	}
	if _, ok := backend.(*GeminiService); !ok {
		t.Errorf("expected *GeminiService as default, got %T", backend)
	}
}

// TestAnalyzeFrameMethodExistsOnGeminiService confirms that GeminiService
// implements the AIBackend interface and that the AnalyzeFrame method is
// callable. The call is expected to fail (no real server), but must not panic.
func TestAnalyzeFrameMethodExistsOnGeminiService(t *testing.T) {
	t.Setenv("AI_BACKEND", "gemini")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("CLIPROXY_URL", "http://127.0.0.1:19997")

	backend, err := NewAIBackend()
	if err != nil {
		t.Fatalf("NewAIBackend: %v", err)
	}

	gs, ok := backend.(*GeminiService)
	if !ok {
		t.Fatalf("expected *GeminiService, got %T", backend)
	}

	// Compile-time: verify AIBackend interface is satisfied.
	var _ AIBackend = gs

	// Runtime: cancelled context causes the call to fail immediately without
	// making a real network request.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, callErr := gs.AnalyzeFrame(ctx, "", "", "it")
	// We expect an error (cancelled ctx / no server). No panic = pass.
	_ = callErr
}

// TestAIBackendCaseSensitiveDocumented documents that AI_BACKEND matching is
// case-sensitive: only lowercase "gemini", "openai", and "claude" are valid.
// Mixed-case values must return an error.
func TestAIBackendCaseSensitiveDocumented(t *testing.T) {
	mixedCaseValues := []string{
		"Gemini", "GEMINI", "GeMiNi",
		"OpenAI", "OPENAI",
		"Claude", "CLAUDE",
	}

	for _, val := range mixedCaseValues {
		val := val
		t.Run(val, func(t *testing.T) {
			t.Setenv("AI_BACKEND", val)

			_, err := NewAIBackend()
			if err == nil {
				t.Errorf("AI_BACKEND=%q: expected error (case-sensitive match required), got nil", val)
			}
		})
	}
}

// TestNewAIBackendAllSupportedBackendsSucceed verifies that each lowercase
// backend name returns a non-nil backend when the required API key is supplied.
func TestNewAIBackendAllSupportedBackendsSucceed(t *testing.T) {
	cases := []struct {
		backend string
		keyName string
		keyVal  string
	}{
		{"gemini", "GEMINI_API_KEY", ""},        // gemini works without a real key (proxy path)
		{"openai", "OPENAI_API_KEY", "sk-test"},
		{"claude", "ANTHROPIC_API_KEY", "sk-ant-test"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.backend, func(t *testing.T) {
			t.Setenv("AI_BACKEND", tc.backend)
			t.Setenv(tc.keyName, tc.keyVal)
			if tc.backend == "gemini" {
				t.Setenv("CLIPROXY_URL", "http://127.0.0.1:19996")
			}

			backend, err := NewAIBackend()
			if err != nil {
				t.Fatalf("NewAIBackend(%q): unexpected error: %v", tc.backend, err)
			}
			if backend == nil {
				t.Errorf("NewAIBackend(%q): returned nil backend", tc.backend)
			}
		})
	}
}
