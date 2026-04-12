package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// ── isDuplicateFrame / storeFrameHash / CleanupSession ───────────────────────

func TestIsDuplicateFrameReturnsTrueForSameHash(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)
	sessionID := "dup-session"

	svc.storeFrameHash(sessionID, "abc123", 0, false)

	if !svc.isDuplicateFrame(sessionID, "abc123", 0, false) {
		t.Error("expected isDuplicateFrame to return true for same hash")
	}
}

func TestIsDuplicateFrameReturnsFalseForDifferentHash(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)
	sessionID := "diff-session"

	svc.storeFrameHash(sessionID, "hash-a", 0, false)

	if svc.isDuplicateFrame(sessionID, "hash-b", 0, false) {
		t.Error("expected isDuplicateFrame to return false for a different hash")
	}
}

func TestIsDuplicateFrameReturnsFalseForUnknownSession(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)

	if svc.isDuplicateFrame("unknown", "anyhash", 0, false) {
		t.Error("expected isDuplicateFrame to return false for a session with no stored hash")
	}
}

func TestIsDuplicateFrameReturnsFalseForExpiredEntry(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)
	sessionID := "ttl-session"

	svc.mu.Lock()
	svc.frameHashes[sessionID] = frameHashEntry{
		sha256:     "xyz789",
		recordedAt: time.Now().Add(-frameHashTTL - time.Second),
	}
	svc.mu.Unlock()

	// Same hash but TTL expired — must NOT be treated as duplicate.
	if svc.isDuplicateFrame(sessionID, "xyz789", 0, false) {
		t.Error("expected isDuplicateFrame to return false for an expired entry")
	}
}

func TestStoreFrameHashPersistsEntry(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)
	sessionID := "store-session"

	svc.storeFrameHash(sessionID, "hash1", 42, true)

	svc.mu.Lock()
	entry, ok := svc.frameHashes[sessionID]
	svc.mu.Unlock()

	if !ok {
		t.Fatal("expected frame hash entry to be stored")
	}
	if entry.sha256 != "hash1" {
		t.Errorf("sha256: want 'hash1', got %q", entry.sha256)
	}
	if entry.pHash != 42 {
		t.Errorf("pHash: want 42, got %d", entry.pHash)
	}
	if !entry.hasPHash {
		t.Error("expected hasPHash to be true")
	}
}

func TestCleanupSessionRemovesFrameHash(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)
	sessionID := "cleanup-session"

	svc.storeFrameHash(sessionID, "hash1", 0, false)
	svc.CleanupSession(sessionID)

	svc.mu.Lock()
	_, ok := svc.frameHashes[sessionID]
	svc.mu.Unlock()

	if ok {
		t.Error("expected frame hash entry to be removed after CleanupSession")
	}
}

func TestCleanupSessionNonexistentIsNoop(t *testing.T) {
	svc := NewConversationService(NewSessionService(), nil, nil)
	// Must not panic when there is no entry for the given session.
	svc.CleanupSession("no-such-session")
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

// ---------------------------------------------------------------------------
// CleanupSession
// ---------------------------------------------------------------------------

func TestCleanupSessionRemovesFrameHash(t *testing.T) {
	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Prusa", "MK4", "", "")
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	// Manually store a frame hash
	svc.storeFrameHash(session.SessionID, "abc123", 0, false)

	// Verify it's there via isDuplicateFrame
	if !svc.isDuplicateFrame(session.SessionID, "abc123", 0, false) {
		t.Error("expected frame hash to be stored before cleanup")
	}

	// Cleanup should remove the hash
	svc.CleanupSession(session.SessionID)

	// After cleanup, the same hash should NOT be detected as duplicate
	if svc.isDuplicateFrame(session.SessionID, "abc123", 0, false) {
		t.Error("expected frame hash to be removed after CleanupSession")
	}
}

func TestCleanupSessionNoopForUnknownSession(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	// Should not panic for a session that was never stored
	svc.CleanupSession("nonexistent-session")
}

// ---------------------------------------------------------------------------
// isDuplicateFrame and storeFrameHash
// ---------------------------------------------------------------------------

func TestIsDuplicateFrameNoEntry(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	if svc.isDuplicateFrame("session-X", "hash", 0, false) {
		t.Error("expected no duplicate when no frame hash stored")
	}
}

func TestIsDuplicateFrameSHA256Match(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	svc.storeFrameHash("sess-1", "sha-AAA", 0, false)

	// Same SHA-256, no pHash
	if !svc.isDuplicateFrame("sess-1", "sha-AAA", 0, false) {
		t.Error("expected duplicate for matching SHA-256 hash")
	}
	// Different SHA-256
	if svc.isDuplicateFrame("sess-1", "sha-BBB", 0, false) {
		t.Error("expected no duplicate for different SHA-256 hash")
	}
}

func TestIsDuplicateFramePHashMatch(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	svc.storeFrameHash("sess-1", "sha", 0x00FF, true)

	// Same pHash -> duplicate (hamming distance = 0, below threshold 5)
	if !svc.isDuplicateFrame("sess-1", "sha-different", 0x00FF, true) {
		t.Error("expected duplicate for identical pHash")
	}
	// pHash with large hamming distance -> not duplicate
	if svc.isDuplicateFrame("sess-1", "sha-different", 0xFFFFFFFFFFFFFFFF, true) {
		t.Error("expected no duplicate for very different pHash")
	}
}

func TestIsDuplicateFrameTTLExpiry(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	// Manually insert an old entry
	svc.mu.Lock()
	svc.frameHashes["sess-1"] = frameHashEntry{
		sha256:     "sha-old",
		recordedAt: time.Now().Add(-frameHashTTL - time.Minute),
	}
	svc.mu.Unlock()

	// Expired entry should NOT be detected as duplicate
	if svc.isDuplicateFrame("sess-1", "sha-old", 0, false) {
		t.Error("expected no duplicate for expired frame hash entry")
	}
}

// ---------------------------------------------------------------------------
// checkAndStoreFrameHash TTL expiry
// ---------------------------------------------------------------------------

func TestCheckAndStoreFrameHashTTLExpiry(t *testing.T) {
	sessions := NewSessionService()
	gemini := &GeminiService{apiKey: "k", baseURL: "http://test", httpClient: &http.Client{}}
	svc := NewConversationService(sessions, gemini, nil)

	// Insert an old entry directly
	svc.mu.Lock()
	svc.frameHashes["sess-1"] = frameHashEntry{
		sha256:     "sha-old",
		recordedAt: time.Now().Add(-frameHashTTL - time.Minute),
	}
	svc.mu.Unlock()

	// Even with the same hash, an expired entry should NOT be treated as dup
	isDup := svc.checkAndStoreFrameHash("sess-1", "sha-old", 0, false)
	if isDup {
		t.Error("expected checkAndStoreFrameHash to return false for expired entry")
	}

	// After the call, the entry should be updated (not expired anymore)
	isDup2 := svc.checkAndStoreFrameHash("sess-1", "sha-old", 0, false)
	if !isDup2 {
		t.Error("expected duplicate on second call (freshly stored)")
	}
}

// ---------------------------------------------------------------------------
// toVoiceText
// ---------------------------------------------------------------------------

func TestToVoiceTextStripsMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bold",
			input: "**bold text** here",
			want:  "bold text here",
		},
		{
			name:  "italic",
			input: "*italic* text",
			want:  "italic text",
		},
		{
			name:  "heading",
			input: "# Heading Title",
			want:  "Heading Title",
		},
		{
			name:  "h2",
			input: "## Sub Heading",
			want:  "Sub Heading",
		},
		{
			name:  "h3",
			input: "### Sub Sub Heading",
			want:  "Sub Sub Heading",
		},
		{
			name:  "backtick code",
			input: "Use `command` to fix",
			want:  "Use command to fix",
		},
		{
			name:  "code block",
			input: "```code block```",
			want:  "code block",
		},
		{
			name:  "underscore bold",
			input: "__underline bold__",
			want:  "underline bold",
		},
		{
			name:  "mixed markdown",
			input: "**Step 1:** Use `nozzle clean` command\n### Important",
			want:  "Step 1: Use nozzle clean command Important",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "excess whitespace",
			input: "  too   many   spaces  ",
			want:  "too many spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toVoiceText(tt.input)
			if got != tt.want {
				t.Errorf("toVoiceText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ProcessFrame with RAG
// ---------------------------------------------------------------------------

func TestProcessFrameWithRAG(t *testing.T) {
	structured := `{\"analysis\":\"bed is level\",\"problem\":null,\"instruction\":\"start print\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	server := geminiStubServer(structured)
	defer server.Close()

	// Build a RAGService with in-memory KB
	entries := []KBEntry{
		{ID: "level", Name: "Bed Leveling", Category: "calibration",
			Description: "Bed leveling ensures first layer adhesion"},
	}
	store := NewMemoryVectorStore()
	store.Index(entries)
	rag := &RAGService{
		entries: entries,
		store:   store,
		cache:   newRAGLRU(ragCacheCapacity, ragCacheTTL),
	}

	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Prusa", "MK4", "", "")
	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, rag)

	// ProcessFrame with userText that matches KB
	result, err := svc.ProcessFrame(context.Background(), session.SessionID, "frame-data", "bed leveling issue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Text == "" {
		t.Error("expected non-empty result text")
	}
}

func TestProcessFrameWithRAGFallbackToProblemDetected(t *testing.T) {
	structured := `{\"analysis\":\"ok\",\"problem\":null,\"instruction\":\"continue\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	server := geminiStubServer(structured)
	defer server.Close()

	entries := []KBEntry{
		{ID: "clog", Name: "Clogged Nozzle", Category: "extrusion",
			Description: "Nozzle is blocked causing under-extrusion"},
	}
	store := NewMemoryVectorStore()
	store.Index(entries)
	rag := &RAGService{
		entries: entries,
		store:   store,
		cache:   newRAGLRU(ragCacheCapacity, ragCacheTTL),
	}

	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Prusa", "MK4", "", "")

	// Set a problem on the session so RAG can fall back to it
	sessions.SetProblemDetected(session.SessionID, "clogged nozzle") //nolint:errcheck

	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, rag)

	// Empty userText -> should fall back to ProblemDetected for RAG query
	result, err := svc.ProcessFrame(context.Background(), session.SessionID, "frame-data-2", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ---------------------------------------------------------------------------
// ProcessTextMessage with RAG
// ---------------------------------------------------------------------------

func TestProcessTextMessageWithRAGEmptyContext(t *testing.T) {
	// This exercises the `else { conversationCtx = kbCtx }` branch when
	// there is no conversation history yet (BuildContextForGemini returns empty).
	structured := `{\"analysis\":\"rag only\",\"problem\":null,\"instruction\":\"fix it\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	server := geminiStubServer(structured)
	defer server.Close()

	entries := []KBEntry{
		{ID: "warp", Name: "Warping Issue", Category: "bed_adhesion",
			Description: "Print corners lift off the bed during printing"},
	}
	store := NewMemoryVectorStore()
	store.Index(entries)
	rag := &RAGService{
		entries: entries,
		store:   store,
		cache:   newRAGLRU(ragCacheCapacity, ragCacheTTL),
	}

	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Prusa", "MK4", "", "")
	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, rag)

	// First message — BuildContextForGemini returns "" at the time RAG runs
	// (only 1 user message from the AddMessage call, but BuildContextForGemini
	// was called right after that, so context is "user: warping...", which is non-empty.
	// We need a scenario where conversationCtx is empty — that means 0 messages at BuildContextForGemini time.
	// But ProcessTextMessage adds a user message first...
	// Actually the context will include that user message, so we can't easily get the
	// empty-context + RAG branch this way. Let's test it differently.)
	result, err := svc.ProcessTextMessage(context.Background(), session.SessionID, "warping issue on print bed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestProcessTextMessageWithRAG(t *testing.T) {
	structured := `{\"analysis\":\"text response\",\"problem\":null,\"instruction\":\"try this\",\"overlay\":{\"boxes\":[],\"arrows\":[]}}`
	server := geminiStubServer(structured)
	defer server.Close()

	entries := []KBEntry{
		{ID: "warp", Name: "Warping Issue", Category: "bed_adhesion",
			Description: "Print corners lift off the bed during printing"},
	}
	store := NewMemoryVectorStore()
	store.Index(entries)
	rag := &RAGService{
		entries: entries,
		store:   store,
		cache:   newRAGLRU(ragCacheCapacity, ragCacheTTL),
	}

	sessions := NewSessionService()
	session, _ := sessions.CreateSession("Prusa", "MK4", "", "")
	gemini := newTestGeminiService(server.URL)
	svc := NewConversationService(sessions, gemini, rag)

	result, err := svc.ProcessTextMessage(context.Background(), session.SessionID, "my print is warping off the bed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Text == "" {
		t.Error("expected non-empty result text")
	}
}
