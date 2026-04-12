package services

import (
	"context"
	"strings"
	"testing"
)

// TestProcessFrameEmptyFrameBytesHandledGracefully verifies that passing an
// empty string as the frame data does not panic and either returns a result
// or a non-nil error — both are acceptable as long as the server stays alive.
func TestProcessFrameEmptyFrameBytesHandledGracefully(t *testing.T) {
	t.Parallel()

	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"continue\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	// Empty frame: should not panic. We accept either nil result or an error.
	result, err := svc.ProcessFrame(context.Background(), sessionID, "", "")
	if err != nil {
		// A well-wrapped error is fine.
		return
	}
	// nil result is also acceptable (treated as no-op/duplicate by the dedup layer).
	_ = result
}

// TestProcessFrameNonexistentSessionIDHandledWithoutPanic verifies that a
// completely unknown session ID propagates an error without panicking.
func TestProcessFrameNonexistentSessionIDHandledWithoutPanic(t *testing.T) {
	t.Parallel()

	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"ok\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, _, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	_, err := svc.ProcessFrame(context.Background(), "totally-unknown-session-id", "framedata", "hello")
	if err == nil {
		t.Fatal("expected error for unknown session, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestToVoiceTextOnlyWhitespaceReturnsEmpty verifies that a string containing
// only whitespace characters collapses to an empty string.
func TestToVoiceTextOnlyWhitespaceReturnsEmpty(t *testing.T) {
	t.Parallel()

	cases := []string{
		"   ",
		"\t\t\t",
		"\n\n",
		"  \t  \n  ",
	}
	for _, input := range cases {
		got := toVoiceText(input)
		if got != "" {
			t.Errorf("toVoiceText(%q) = %q, want empty string", input, got)
		}
	}
}

// TestBuildContextForGeminiNoKBStillReturnsValidContext verifies that
// a ConversationService with a nil RAGService still produces a valid (possibly
// empty) conversation context without panicking.
func TestBuildContextForGeminiNoKBStillReturnsValidContext(t *testing.T) {
	t.Parallel()

	sessions := NewSessionService()
	sess, err := sessions.CreateSession("Prusa", "MK4", "", "it")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_ = sessions.AddMessage(sess.SessionID, "user", "what is wrong?", false)

	// BuildContextForGemini is on SessionService, not ConversationService, but
	// this test verifies that ConversationService.ProcessTextMessage works fine
	// without a RAG service (rag=nil) and does not panic.
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"ok\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	server := geminiStubServer(structured)
	defer server.Close()

	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, nil) // nil RAG

	result, err := svc.ProcessTextMessage(context.Background(), sess.SessionID, "check the nozzle")
	if err != nil {
		t.Fatalf("ProcessTextMessage with nil RAG: unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result even without RAG")
	}
}

// TestToVoiceTextAllMarkdownTokensStripped verifies that each individual
// Markdown token is removed, leaving clean text.
func TestToVoiceTextAllMarkdownTokensStripped(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		token string
	}{
		{"**bold**", "**"},
		{"__bold__", "__"},
		{"*italic*", "*"},
		{"_italic_", "_"},
		{"```code```", "```"},
		{"`code`", "`"},
		{"### h3", "###"},
		{"## h2", "##"},
		{"# h1", "#"},
	}

	for _, tc := range cases {
		got := toVoiceText(tc.input)
		if strings.Contains(got, tc.token) {
			t.Errorf("toVoiceText(%q) still contains token %q: got %q", tc.input, tc.token, got)
		}
	}
}

// TestHashFrameEmptyStringProducesStableHash verifies that hashFrame on an
// empty string does not panic and returns a consistent, non-empty hash.
func TestHashFrameEmptyStringProducesStableHash(t *testing.T) {
	t.Parallel()

	h1 := hashFrame("")
	h2 := hashFrame("")
	if h1 == "" {
		t.Error("hashFrame(\"\") returned empty string, want non-empty SHA-256 hex")
	}
	if h1 != h2 {
		t.Errorf("hashFrame(\"\") is not deterministic: %q vs %q", h1, h2)
	}
}
