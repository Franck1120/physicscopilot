// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// geminiStubServer returns a test HTTP server that responds with the given
// structured JSON wrapped in the Gemini response envelope. The caller must
// close the returned server after use.
func geminiStubServer(structuredJSON string) *httptest.Server {
	envelope := validGeminiJSON(`"` + structuredJSON + `"`)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(envelope))
	}))
}

// newTestGeminiService creates a GeminiService pointing at the given test server URL.
func newTestGeminiService(serverURL string) *GeminiService {
	return &GeminiService{
		apiKey:     "test-key",
		baseURL:    serverURL,
		httpClient: &http.Client{},
	}
}

// setupConversationTest creates a ConversationService backed by a real
// SessionService and a stub Gemini server. Returns the service, the
// created session ID, and a cleanup function to close the server.
func setupConversationTest(t *testing.T, structuredJSON string) (*ConversationService, string, func()) {
	t.Helper()

	server := geminiStubServer(structuredJSON)

	sessions := NewSessionService()
	session, err := sessions.CreateSession("Prusa", "MK4", "", "")
	if err != nil {
		server.Close()
		t.Fatalf("create session: %v", err)
	}

	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, nil)

	return svc, session.SessionID, server.Close
}

func TestNewConversationService(t *testing.T) {
	sessions := NewSessionService()
	var ai AIBackend = &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}

	svc := NewConversationService(sessions, ai, nil)
	if svc == nil {
		t.Fatal("expected non-nil ConversationService")
	}
	if svc.sessions != sessions {
		t.Error("expected sessions field to match")
	}
	if svc.ai != ai {
		t.Error("expected ai field to match")
	}
	if svc.frameHashes == nil {
		t.Error("expected frameHashes map to be initialized")
	}
}

func TestProcessFrameSuccess(t *testing.T) {
	structured := `{\"analysis\":\"bed is level\",\"problem\":null,\"instruction\":\"start print\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	result, err := svc.ProcessFrame(context.Background(), sessionID, "base64framedata", "check my bed level")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	expectedText := "bed is level\n\nstart print"
	if result.Text != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, result.Text)
	}

	// Verify user and assistant messages were recorded
	session, _ := svc.sessions.GetSession(sessionID)
	if len(session.ConversationHistory) != 2 {
		t.Fatalf("expected 2 messages in history, got %d", len(session.ConversationHistory))
	}
	if session.ConversationHistory[0].Role != "user" {
		t.Errorf("expected first message role 'user', got %q", session.ConversationHistory[0].Role)
	}
	if session.ConversationHistory[0].Content != "check my bed level" {
		t.Errorf("expected first message content 'check my bed level', got %q", session.ConversationHistory[0].Content)
	}
	if session.ConversationHistory[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got %q", session.ConversationHistory[1].Role)
	}
	if session.ConversationHistory[1].Content != expectedText {
		t.Errorf("expected second message content %q, got %q", expectedText, session.ConversationHistory[1].Content)
	}
}

func TestProcessFrameSkipsDuplicateFrame(t *testing.T) {
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"continue\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	// First call should succeed
	result1, err := svc.ProcessFrame(context.Background(), sessionID, "sameframe", "")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if result1 == nil {
		t.Fatal("expected non-nil result on first call")
	}

	// Second call with same frame should be skipped
	result2, err := svc.ProcessFrame(context.Background(), sessionID, "sameframe", "")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if result2 != nil {
		t.Error("expected nil result for duplicate frame")
	}

	// Third call with different frame should succeed
	result3, err := svc.ProcessFrame(context.Background(), sessionID, "differentframe", "")
	if err != nil {
		t.Fatalf("third call error: %v", err)
	}
	if result3 == nil {
		t.Fatal("expected non-nil result for different frame")
	}
}

func TestProcessFrameWithProblemDetected(t *testing.T) {
	structured := `{\"analysis\":\"nozzle clogged\",\"problem\":\"clogged nozzle\",\"instruction\":\"clean nozzle\",\"overlay\":{\"boxes\":[{\"x\":0.1,\"y\":0.2,\"w\":0.3,\"h\":0.4,\"label\":\"clog\"}],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	result, err := svc.ProcessFrame(context.Background(), sessionID, "framedata", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify problem was recorded in session
	session, _ := svc.sessions.GetSession(sessionID)
	if session.ProblemDetected != "clogged nozzle" {
		t.Errorf("expected ProblemDetected 'clogged nozzle', got %q", session.ProblemDetected)
	}

	// Verify overlay data is passed through
	if len(result.Overlay.Boxes) != 1 {
		t.Fatalf("expected 1 bounding box, got %d", len(result.Overlay.Boxes))
	}
	if result.Overlay.Boxes[0].Label != "clog" {
		t.Errorf("expected box label 'clog', got %q", result.Overlay.Boxes[0].Label)
	}
}

func TestProcessFrameWithoutUserText(t *testing.T) {
	structured := `{\"analysis\":\"all good\",\"problem\":null,\"instruction\":\"keep going\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	result, err := svc.ProcessFrame(context.Background(), sessionID, "framedata", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Only the assistant message should be in history (no user text)
	session, _ := svc.sessions.GetSession(sessionID)
	if len(session.ConversationHistory) != 1 {
		t.Fatalf("expected 1 message (assistant only), got %d", len(session.ConversationHistory))
	}
	if session.ConversationHistory[0].Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", session.ConversationHistory[0].Role)
	}
}

func TestProcessFrameStepInfo(t *testing.T) {
	structured := `{\"analysis\":\"step 3\",\"problem\":null,\"instruction\":\"do this\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	// Set step info on the session
	_ = svc.sessions.UpdateStep(sessionID, 3, 10)

	result, err := svc.ProcessFrame(context.Background(), sessionID, "framedata", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Step.Current != 3 {
		t.Errorf("expected step current 3, got %d", result.Step.Current)
	}
	if result.Step.Total != 10 {
		t.Errorf("expected step total 10, got %d", result.Step.Total)
	}
}

func TestProcessFrameNonexistentSession(t *testing.T) {
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"ok\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, _, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	_, err := svc.ProcessFrame(context.Background(), "nonexistent-session", "frame", "hello")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestProcessTextMessageSuccess(t *testing.T) {
	structured := `{\"analysis\":\"text response\",\"problem\":null,\"instruction\":\"try this\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	result, err := svc.ProcessTextMessage(context.Background(), sessionID, "my nozzle is clicking")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	expectedText := "text response\n\ntry this"
	if result.Text != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, result.Text)
	}

	// Verify both messages in history
	session, _ := svc.sessions.GetSession(sessionID)
	if len(session.ConversationHistory) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(session.ConversationHistory))
	}
	if session.ConversationHistory[0].Role != "user" {
		t.Errorf("expected first role 'user', got %q", session.ConversationHistory[0].Role)
	}
	if session.ConversationHistory[1].Role != "assistant" {
		t.Errorf("expected second role 'assistant', got %q", session.ConversationHistory[1].Role)
	}
}

func TestProcessTextMessageNonexistentSession(t *testing.T) {
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"ok\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, _, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	_, err := svc.ProcessTextMessage(context.Background(), "nonexistent-session", "hello")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestGetSessionStep(t *testing.T) {
	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Bambu", "X1C", "", "")
	_ = sessions.UpdateStep(session.SessionID, 5, 12)

	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	step, err := svc.GetSessionStep(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step.Current != 5 {
		t.Errorf("expected current step 5, got %d", step.Current)
	}
	if step.Total != 12 {
		t.Errorf("expected total steps 12, got %d", step.Total)
	}
}

func TestGetSessionStepNonexistentSession(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	_, err := svc.GetSessionStep("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestHashFrameTruncatesLongInput(t *testing.T) {
	shortFrame := "abc"
	longFrame := strings.Repeat("x", 2048)

	// Short frame hashes the entire input
	hash1 := hashFrame(shortFrame)
	if hash1 == "" {
		t.Fatal("expected non-empty hash for short frame")
	}

	// Long frame hashes only the first 1024 bytes, so two strings that
	// share the same prefix should produce the same hash
	longFrame2 := strings.Repeat("x", 2048) + "different-suffix"
	hash2 := hashFrame(longFrame)
	hash3 := hashFrame(longFrame2)
	if hash2 != hash3 {
		t.Errorf("expected same hash for frames sharing first 1024 bytes, got %q vs %q", hash2, hash3)
	}

	// Different prefix should produce different hash
	differentFrame := strings.Repeat("y", 2048)
	hash4 := hashFrame(differentFrame)
	if hash2 == hash4 {
		t.Error("expected different hashes for different frame prefixes")
	}
}

func TestProcessFrameGeminiError(t *testing.T) {
	// Gemini returns 400 (non-retryable error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Creality", "Ender 3", "", "")
	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, nil)

	_, err := svc.ProcessFrame(context.Background(), session.SessionID, "frame", "")
	if err == nil {
		t.Fatal("expected error when Gemini returns error")
	}
	if !strings.Contains(err.Error(), "analyze frame") {
		t.Errorf("expected 'analyze frame' in error, got: %v", err)
	}
}

func TestProcessFrameClearsHashOnGeminiError(t *testing.T) {
	callCount := 0
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"continue\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`

	// First call fails, second succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"bad request"}`))
			return
		}
		envelope := validGeminiJSON(`"` + structured + `"`)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(envelope))
	}))
	defer server.Close()

	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Creality", "Ender 3", "", "")
	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, nil)

	// First call with "frame" fails -- hash should be cleared
	_, err := svc.ProcessFrame(context.Background(), session.SessionID, "frame", "")
	if err == nil {
		t.Fatal("expected error on first call")
	}

	// Retry with the same frame data should NOT be skipped as duplicate
	result, err := svc.ProcessFrame(context.Background(), session.SessionID, "frame", "")
	if err != nil {
		t.Fatalf("expected success on retry, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result on retry (hash should have been cleared)")
	}
	if callCount != 2 {
		t.Errorf("expected 2 Gemini calls (fail + retry), got %d", callCount)
	}
}

func TestComputeFrameFingerprintNonJPEGFallsBackToSHA256(t *testing.T) {
	sha256Hash, _, hasPHash := computeFrameFingerprint("not-a-jpeg-base64")
	if sha256Hash == "" {
		t.Error("expected non-empty SHA-256 fallback hash")
	}
	if hasPHash {
		t.Error("expected hasPHash=false for non-JPEG input")
	}
}

// ── toVoiceText ────────────────────────────────────────────────────────────────

func TestToVoiceTextStripsMarkdownBold(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"**bold text**", "bold text"},
		{"__also bold__", "also bold"},
		{"*italic*", "italic"},
		{"_italic_", "italic"},
		{"```code block```", "code block"},
		{"`inline code`", "inline code"},
		{"### heading 3", "heading 3"},
		{"## heading 2", "heading 2"},
		{"# heading 1", "heading 1"},
	}

	for _, tc := range cases {
		got := toVoiceText(tc.input)
		if got != tc.want {
			t.Errorf("toVoiceText(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestToVoiceTextCollapsesWhitespace(t *testing.T) {
	input := "  turn   the   screw   clockwise  "
	want := "turn the screw clockwise"
	got := toVoiceText(input)
	if got != want {
		t.Errorf("toVoiceText(%q) = %q, want %q", input, got, want)
	}
}

func TestToVoiceTextEmptyInputReturnsEmpty(t *testing.T) {
	if got := toVoiceText(""); got != "" {
		t.Errorf("toVoiceText(\"\") = %q, want empty", got)
	}
}

func TestToVoiceTextPlainTextUnchanged(t *testing.T) {
	input := "Remove the nozzle and clean it with a wire brush."
	got := toVoiceText(input)
	if got != input {
		t.Errorf("toVoiceText plain text changed: got %q, want %q", got, input)
	}
}

func TestToVoiceTextMixedMarkdown(t *testing.T) {
	input := "**Step 1**: remove the `nozzle`\n### Clean it\n*carefully*"
	got := toVoiceText(input)

	// No markdown syntax must remain.
	for _, token := range []string{"**", "__", "*", "_", "`", "###", "##", "#"} {
		if strings.Contains(got, token) {
			t.Errorf("toVoiceText result still contains %q: %q", token, got)
		}
	}
	// Collapsed to single line with single spaces.
	if strings.Contains(got, "\n") {
		t.Errorf("toVoiceText result contains newline: %q", got)
	}
	if strings.Contains(got, "  ") {
		t.Errorf("toVoiceText result contains consecutive spaces: %q", got)
	}
}

// ── BuildContextForGemini ────────────────────────────────────────────────────

func TestBuildContextForGeminiEmpty(t *testing.T) {
	sessions := NewSessionService()
	session, err := sessions.CreateSession("Brand", "Model", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	ctx, err := sessions.BuildContextForGemini(session.SessionID)
	if err != nil {
		t.Fatalf("BuildContextForGemini: %v", err)
	}
	if ctx != "" {
		t.Errorf("expected empty context for fresh session, got %q", ctx)
	}
}

func TestBuildContextForGeminiFormatsRoleContent(t *testing.T) {
	sessions := NewSessionService()
	session, err := sessions.CreateSession("Brand", "Model", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	_ = sessions.AddMessage(session.SessionID, "user", "hello", false)
	_ = sessions.AddMessage(session.SessionID, "assistant", "hi there", false)

	ctx, err := sessions.BuildContextForGemini(session.SessionID)
	if err != nil {
		t.Fatalf("BuildContextForGemini: %v", err)
	}

	if !strings.Contains(ctx, "user: hello") {
		t.Errorf("expected 'user: hello' in context, got: %q", ctx)
	}
	if !strings.Contains(ctx, "assistant: hi there") {
		t.Errorf("expected 'assistant: hi there' in context, got: %q", ctx)
	}
}

func TestBuildContextForGeminiTruncatesToMaxMessages(t *testing.T) {
	sessions := NewSessionService()
	session, err := sessions.CreateSession("Brand", "Model", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Add more messages than maxContextMessages (10).
	for i := range 15 {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		_ = sessions.AddMessage(session.SessionID, role, strings.Repeat("x", i+1), false)
	}

	ctx, err := sessions.BuildContextForGemini(session.SessionID)
	if err != nil {
		t.Fatalf("BuildContextForGemini: %v", err)
	}

	// The context must contain exactly maxContextMessages lines (10).
	lines := strings.Split(strings.TrimSpace(ctx), "\n")
	if len(lines) != maxContextMessages {
		t.Errorf("expected %d lines in context (max), got %d", maxContextMessages, len(lines))
	}
}

func TestBuildContextForGeminiNonexistentSession(t *testing.T) {
	sessions := NewSessionService()

	_, err := sessions.BuildContextForGemini("no-such-session")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// ── CleanupSession / clearFrameHash ──────────────────────────────────────────

func TestCleanupSessionRemovesFrameHash(t *testing.T) {
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"continue\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	svc, sessionID, cleanup := setupConversationTest(t, structured)
	defer cleanup()

	// Process a frame to populate the hash map.
	_, err := svc.ProcessFrame(context.Background(), sessionID, "frame-data", "")
	if err != nil {
		t.Fatalf("ProcessFrame: %v", err)
	}

	// Same frame should now be a duplicate.
	dup, err := svc.ProcessFrame(context.Background(), sessionID, "frame-data", "")
	if err != nil {
		t.Fatalf("second ProcessFrame: %v", err)
	}
	if dup != nil {
		t.Fatal("expected nil (duplicate) on second ProcessFrame before cleanup")
	}

	// After cleanup, the same frame should be processed again.
	svc.CleanupSession(sessionID)

	result, err := svc.ProcessFrame(context.Background(), sessionID, "frame-data", "")
	if err != nil {
		t.Fatalf("ProcessFrame after cleanup: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result after CleanupSession (hash was cleared)")
	}
}
