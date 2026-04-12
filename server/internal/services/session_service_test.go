package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewSessionService(t *testing.T) {
	svc := NewSessionService()
	if svc == nil {
		t.Fatal("expected non-nil SessionService")
	}
	if svc.sessions == nil {
		t.Fatal("expected sessions map to be initialized")
	}
}

func TestCreateSession(t *testing.T) {
	svc := NewSessionService()

	session, err := svc.CreateSession("Creality", "Ender 3", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.SessionID == "" {
		t.Error("expected non-empty SessionID")
	}
	if session.DeviceInfo.Brand != "Creality" {
		t.Errorf("expected brand 'Creality', got %q", session.DeviceInfo.Brand)
	}
	if session.DeviceInfo.Model != "Ender 3" {
		t.Errorf("expected model 'Ender 3', got %q", session.DeviceInfo.Model)
	}
	if session.CurrentStep != 0 {
		t.Errorf("expected CurrentStep 0, got %d", session.CurrentStep)
	}
	if session.TotalSteps != 0 {
		t.Errorf("expected TotalSteps 0, got %d", session.TotalSteps)
	}
	if session.ProblemDetected != "" {
		t.Errorf("expected empty ProblemDetected, got %q", session.ProblemDetected)
	}
	if len(session.ConversationHistory) != 0 {
		t.Errorf("expected empty ConversationHistory, got %d messages", len(session.ConversationHistory))
	}
	if session.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if session.LastActivity.IsZero() {
		t.Error("expected non-zero LastActivity")
	}
}

func TestGetSession(t *testing.T) {
	svc := NewSessionService()

	session, _ := svc.CreateSession("Prusa", "MK4", "", "")

	retrieved, err := svc.GetSession(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.SessionID != session.SessionID {
		t.Errorf("expected session ID %q, got %q", session.SessionID, retrieved.SessionID)
	}
	if retrieved.DeviceInfo.Brand != "Prusa" {
		t.Errorf("expected brand 'Prusa', got %q", retrieved.DeviceInfo.Brand)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	svc := NewSessionService()

	_, err := svc.GetSession("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %q", err.Error())
	}
}

func TestAddMessage(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Bambu", "X1C", "", "")

	err := svc.AddMessage(session.SessionID, "user", "My printer is making a clicking noise", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := svc.GetSession(session.SessionID)
	if len(retrieved.ConversationHistory) != 1 {
		t.Fatalf("expected 1 message, got %d", len(retrieved.ConversationHistory))
	}

	msg := retrieved.ConversationHistory[0]
	if msg.Role != "user" {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}
	if msg.Content != "My printer is making a clicking noise" {
		t.Errorf("unexpected content: %q", msg.Content)
	}
	if msg.HasImage {
		t.Error("expected HasImage false")
	}
	if msg.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestAddMessageWithImage(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Bambu", "X1C", "", "")

	err := svc.AddMessage(session.SessionID, "user", "Here is a photo of the issue", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := svc.GetSession(session.SessionID)
	if !retrieved.ConversationHistory[0].HasImage {
		t.Error("expected HasImage true")
	}
}

func TestAddMessageRollingWindow(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Creality", "Ender 3", "", "")

	// Add 25 messages — only the last 20 should remain
	for i := 0; i < 25; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		err := svc.AddMessage(session.SessionID, role, "message", false)
		if err != nil {
			t.Fatalf("unexpected error adding message %d: %v", i, err)
		}
	}

	retrieved, _ := svc.GetSession(session.SessionID)
	if len(retrieved.ConversationHistory) != maxConversationHistory {
		t.Errorf("expected %d messages, got %d", maxConversationHistory, len(retrieved.ConversationHistory))
	}
}

func TestAddMessageNonexistentSession(t *testing.T) {
	svc := NewSessionService()

	err := svc.AddMessage("nonexistent", "user", "hello", false)
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestAddMessageUpdatesLastActivity(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Prusa", "MK4", "", "")

	before := session.LastActivity

	// Small delay to ensure time difference
	time.Sleep(1 * time.Millisecond)

	_ = svc.AddMessage(session.SessionID, "user", "test", false)

	retrieved, _ := svc.GetSession(session.SessionID)
	if !retrieved.LastActivity.After(before) {
		t.Error("expected LastActivity to be updated after AddMessage")
	}
}

func TestUpdateStep(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Creality", "Ender 3", "", "")

	err := svc.UpdateStep(session.SessionID, 3, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := svc.GetSession(session.SessionID)
	if retrieved.CurrentStep != 3 {
		t.Errorf("expected CurrentStep 3, got %d", retrieved.CurrentStep)
	}
	if retrieved.TotalSteps != 10 {
		t.Errorf("expected TotalSteps 10, got %d", retrieved.TotalSteps)
	}
}

func TestUpdateStepNonexistentSession(t *testing.T) {
	svc := NewSessionService()

	err := svc.UpdateStep("nonexistent", 1, 5)
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestSetProblemDetected(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Bambu", "X1C", "", "")

	err := svc.SetProblemDetected(session.SessionID, "Clogged nozzle detected")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := svc.GetSession(session.SessionID)
	if retrieved.ProblemDetected != "Clogged nozzle detected" {
		t.Errorf("expected 'Clogged nozzle detected', got %q", retrieved.ProblemDetected)
	}
}

func TestSetProblemDetectedNonexistentSession(t *testing.T) {
	svc := NewSessionService()

	err := svc.SetProblemDetected("nonexistent", "problem")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestBuildContextForGemini(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Prusa", "MK4", "", "")

	_ = svc.AddMessage(session.SessionID, "user", "My printer is clicking", false)
	_ = svc.AddMessage(session.SessionID, "assistant", "Can you show the extruder?", false)
	_ = svc.AddMessage(session.SessionID, "user", "Here it is", true)

	context, err := svc.BuildContextForGemini(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(context, "user: My printer is clicking") {
		t.Error("expected context to contain first user message")
	}
	if !strings.Contains(context, "assistant: Can you show the extruder?") {
		t.Error("expected context to contain assistant message")
	}
	if !strings.Contains(context, "user: Here it is") {
		t.Error("expected context to contain second user message")
	}
}

func TestBuildContextForGeminiLimitsTen(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Creality", "Ender 3", "", "")

	// Add 15 messages
	for i := 0; i < 15; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		_ = svc.AddMessage(session.SessionID, role, "msg", false)
	}

	context, err := svc.BuildContextForGemini(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count the number of "msg" occurrences — should be 10 (last 10)
	lines := strings.Split(strings.TrimSpace(context), "\n")
	if len(lines) != 10 {
		t.Errorf("expected 10 lines in context, got %d", len(lines))
	}
}

func TestBuildContextForGeminiEmptyHistory(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Bambu", "X1C", "", "")

	context, err := svc.BuildContextForGemini(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if context != "" {
		t.Errorf("expected empty context, got %q", context)
	}
}

func TestBuildContextForGeminiNonexistentSession(t *testing.T) {
	svc := NewSessionService()

	_, err := svc.BuildContextForGemini("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestDeleteSession(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Creality", "Ender 3", "", "")

	err := svc.DeleteSession(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetSession(session.SessionID)
	if err == nil {
		t.Fatal("expected error after deleting session")
	}
}

func TestDeleteSessionNonexistentSession(t *testing.T) {
	svc := NewSessionService()

	err := svc.DeleteSession("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	svc := NewSessionService()

	// Create two sessions
	session1, _ := svc.CreateSession("Prusa", "MK4", "", "")
	session2, _ := svc.CreateSession("Bambu", "X1C", "", "")

	// Manually set session1's LastActivity to 2 hours ago
	svc.mu.Lock()
	svc.sessions[session1.SessionID].LastActivity = time.Now().Add(-2 * time.Hour)
	svc.mu.Unlock()

	// Cleanup sessions inactive for more than 1 hour
	svc.CleanupExpiredSessions(1 * time.Hour)

	// session1 should be gone
	_, err := svc.GetSession(session1.SessionID)
	if err == nil {
		t.Error("expected session1 to be cleaned up")
	}

	// session2 should still exist
	_, err = svc.GetSession(session2.SessionID)
	if err != nil {
		t.Error("expected session2 to still exist")
	}
}

func TestCleanupExpiredSessionsKeepsActive(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Creality", "Ender 3", "", "")

	// Cleanup with very large maxAge — nothing should be removed
	svc.CleanupExpiredSessions(24 * time.Hour)

	_, err := svc.GetSession(session.SessionID)
	if err != nil {
		t.Error("expected session to still exist after cleanup with large maxAge")
	}
}

func TestGetSessionSnapshot(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Prusa", "MK4", "", "")
	_ = svc.UpdateStep(session.SessionID, 3, 10)

	snapshot, err := svc.GetSessionSnapshot(session.SessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snapshot.CurrentStep != 3 {
		t.Errorf("expected CurrentStep 3, got %d", snapshot.CurrentStep)
	}
	if snapshot.TotalSteps != 10 {
		t.Errorf("expected TotalSteps 10, got %d", snapshot.TotalSteps)
	}

	// Mutating the original session must not affect the snapshot
	_ = svc.UpdateStep(session.SessionID, 7, 15)

	if snapshot.CurrentStep != 3 {
		t.Errorf("snapshot mutated after UpdateStep: expected CurrentStep 3, got %d", snapshot.CurrentStep)
	}
	if snapshot.TotalSteps != 10 {
		t.Errorf("snapshot mutated after UpdateStep: expected TotalSteps 10, got %d", snapshot.TotalSteps)
	}
}

func TestGetSessionSnapshotNotFound(t *testing.T) {
	svc := NewSessionService()

	_, err := svc.GetSessionSnapshot("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// ActiveSessionCount
// ---------------------------------------------------------------------------

func TestActiveSessionCountEmpty(t *testing.T) {
	svc := NewSessionService()
	if n := svc.ActiveSessionCount(""); n != 0 {
		t.Errorf("expected 0 for empty userID, got %d", n)
	}
}

func TestActiveSessionCountMatchesCreatedSessions(t *testing.T) {
	svc := NewSessionService()

	svc.CreateSession("Prusa", "MK4", "user-A", "") //nolint:errcheck
	svc.CreateSession("Bambu", "X1C", "user-A", "") //nolint:errcheck
	svc.CreateSession("Creality", "K1", "user-B", "") //nolint:errcheck

	if n := svc.ActiveSessionCount("user-A"); n != 2 {
		t.Errorf("expected 2 sessions for user-A, got %d", n)
	}
	if n := svc.ActiveSessionCount("user-B"); n != 1 {
		t.Errorf("expected 1 session for user-B, got %d", n)
	}
	if n := svc.ActiveSessionCount("user-C"); n != 0 {
		t.Errorf("expected 0 sessions for user-C, got %d", n)
	}
}

// ---------------------------------------------------------------------------
// ListSessions
// ---------------------------------------------------------------------------

func TestListSessionsEmpty(t *testing.T) {
	svc := NewSessionService()
	list := svc.ListSessions()
	if len(list) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(list))
	}
}

func TestListSessionsReturnsAll(t *testing.T) {
	svc := NewSessionService()
	svc.CreateSession("Prusa", "MK4", "", "") //nolint:errcheck
	svc.CreateSession("Bambu", "X1C", "", "") //nolint:errcheck
	svc.CreateSession("Creality", "K1", "", "") //nolint:errcheck

	list := svc.ListSessions()
	if len(list) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(list))
	}
}

// ---------------------------------------------------------------------------
// CreateSession session limit
// ---------------------------------------------------------------------------

func TestCreateSessionRejectsOverLimit(t *testing.T) {
	svc := NewSessionService()

	// Create maxActiveSessionsPerUser sessions for user-A
	for i := 0; i < maxActiveSessionsPerUser; i++ {
		_, err := svc.CreateSession("Prusa", "MK4", "user-limit", "")
		if err != nil {
			t.Fatalf("session %d: unexpected error: %v", i, err)
		}
	}

	// The next one must fail
	_, err := svc.CreateSession("Prusa", "MK4", "user-limit", "")
	if err == nil {
		t.Fatal("expected error when exceeding session limit")
	}
	if !strings.Contains(err.Error(), "maximum") {
		t.Errorf("expected 'maximum' in error, got: %v", err)
	}
}

func TestCreateSessionLimitDoesNotAffectOtherUsers(t *testing.T) {
	svc := NewSessionService()

	// Fill up user-A
	for i := 0; i < maxActiveSessionsPerUser; i++ {
		svc.CreateSession("Prusa", "MK4", "user-A", "") //nolint:errcheck
	}

	// user-B should still be able to create sessions
	_, err := svc.CreateSession("Prusa", "MK4", "user-B", "")
	if err != nil {
		t.Fatalf("user-B should not be affected by user-A limit: %v", err)
	}
}

func TestCreateSessionLimitIgnoredForAnonymous(t *testing.T) {
	svc := NewSessionService()

	// Anonymous sessions (empty userID) should never be limited
	for i := 0; i < maxActiveSessionsPerUser+5; i++ {
		_, err := svc.CreateSession("Prusa", "MK4", "", "")
		if err != nil {
			t.Fatalf("anonymous session %d: unexpected error: %v", i, err)
		}
	}
}

// ---------------------------------------------------------------------------
// CreateSession concurrent stress test
// ---------------------------------------------------------------------------

func TestCreateSessionConcurrent(t *testing.T) {
	svc := NewSessionService()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errCh := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := svc.CreateSession("Brand", fmt.Sprintf("Model-%d", i), "", "")
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent create error: %v", err)
	}

	list := svc.ListSessions()
	if len(list) != goroutines {
		t.Errorf("expected %d sessions, got %d", goroutines, len(list))
	}
}

// ---------------------------------------------------------------------------
// CreateSession language default
// ---------------------------------------------------------------------------

func TestCreateSessionDefaultsLanguageToIT(t *testing.T) {
	svc := NewSessionService()
	sess, _ := svc.CreateSession("Prusa", "MK4", "", "")
	if sess.Language != "it" {
		t.Errorf("expected default language 'it', got %q", sess.Language)
	}
}

func TestCreateSessionPreservesExplicitLanguage(t *testing.T) {
	svc := NewSessionService()
	sess, _ := svc.CreateSession("Prusa", "MK4", "", "en")
	if sess.Language != "en" {
		t.Errorf("expected language 'en', got %q", sess.Language)
	}
}

func TestConcurrentAccess(t *testing.T) {
	svc := NewSessionService()
	session, _ := svc.CreateSession("Prusa", "MK4", "", "")

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	// Spawn 50 goroutines adding messages concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := svc.AddMessage(session.SessionID, "user", "concurrent msg", false); err != nil {
				errCh <- err
			}
		}()
	}

	// Spawn 50 goroutines reading the session concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.GetSession(session.SessionID); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent access error: %v", err)
	}

	// Verify session is still valid and has at most 20 messages
	retrieved, _ := svc.GetSession(session.SessionID)
	if len(retrieved.ConversationHistory) > maxConversationHistory {
		t.Errorf("expected at most %d messages, got %d", maxConversationHistory, len(retrieved.ConversationHistory))
	}
}

// ---------------------------------------------------------------------------
// ActiveSessionCount and ListSessions
// ---------------------------------------------------------------------------

func TestActiveSessionCountEmptyUserID(t *testing.T) {
	svc := NewSessionService()
	if count := svc.ActiveSessionCount(""); count != 0 {
		t.Errorf("empty userID: want 0, got %d", count)
	}
}

func TestActiveSessionCountNoSessions(t *testing.T) {
	svc := NewSessionService()
	if count := svc.ActiveSessionCount("user-none"); count != 0 {
		t.Errorf("no sessions for user: want 0, got %d", count)
	}
}

func TestActiveSessionCountCorrect(t *testing.T) {
	svc := NewSessionService()
	svc.CreateSession("A", "1", "user-a", "it")
	svc.CreateSession("B", "2", "user-a", "it")
	svc.CreateSession("C", "3", "user-b", "it")

	if count := svc.ActiveSessionCount("user-a"); count != 2 {
		t.Errorf("user-a: want 2 sessions, got %d", count)
	}
	if count := svc.ActiveSessionCount("user-b"); count != 1 {
		t.Errorf("user-b: want 1 session, got %d", count)
	}
}

func TestListSessionsEmpty(t *testing.T) {
	svc := NewSessionService()
	if sessions := svc.ListSessions(); len(sessions) != 0 {
		t.Errorf("empty store: want 0 sessions, got %d", len(sessions))
	}
}

func TestListSessionsReturnsAll(t *testing.T) {
	svc := NewSessionService()
	svc.CreateSession("A", "1", "", "it")
	svc.CreateSession("B", "2", "", "it")
	svc.CreateSession("C", "3", "", "it")

	if sessions := svc.ListSessions(); len(sessions) != 3 {
		t.Errorf("want 3 sessions, got %d", len(sessions))
	}
}

// ---------------------------------------------------------------------------
// Session limit enforcement
// ---------------------------------------------------------------------------

func TestSessionLimitEnforcement(t *testing.T) {
	svc := NewSessionService()
	userID := "limited-user"

	for i := 0; i < maxActiveSessionsPerUser; i++ {
		_, err := svc.CreateSession("Brand", "Model", userID, "it")
		if err != nil {
			t.Fatalf("session %d: unexpected error: %v", i+1, err)
		}
	}

	_, err := svc.CreateSession("Brand", "Model", userID, "it")
	if err == nil {
		t.Fatal("expected error when exceeding maxActiveSessionsPerUser")
	}
	if !strings.Contains(err.Error(), "maximum") {
		t.Errorf("expected 'maximum' in error, got: %v", err)
	}
}

func TestSessionLimitAnonymousUnbounded(t *testing.T) {
	svc := NewSessionService()
	// Empty userID bypasses the session limit.
	for i := 0; i < maxActiveSessionsPerUser+1; i++ {
		_, err := svc.CreateSession("Brand", "Model", "", "it")
		if err != nil {
			t.Fatalf("anonymous session %d: unexpected error: %v", i+1, err)
		}
	}
}

func TestCreateSessionDefaultsLanguageToIT(t *testing.T) {
	svc := NewSessionService()
	sess, err := svc.CreateSession("Brand", "Model", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.Language != "it" {
		t.Errorf("empty language: want default 'it', got %q", sess.Language)
	}
}

func TestCreateSessionConcurrent(t *testing.T) {
	svc := NewSessionService()
	const goroutines = 30
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.CreateSession("Brand", "Model", "", "it")
			if err != nil {
				t.Errorf("concurrent CreateSession error: %v", err)
			}
		}()
	}
	wg.Wait()

	sessions := svc.ListSessions()
	if len(sessions) != goroutines {
		t.Errorf("expected %d sessions after concurrent create, got %d", goroutines, len(sessions))
	}
}

// ---------------------------------------------------------------------------
// DB error path tests
// ---------------------------------------------------------------------------

func TestSessionServiceDBSaveError(t *testing.T) {
	// A DB save error must not fail the in-memory CreateSession.
	db := newMockDB()
	db.saveErr = fmt.Errorf("db unavailable")
	svc := NewSessionService()
	svc.SetDB(db)

	sess, err := svc.CreateSession("Prusa", "MK4", "", "it")
	if err != nil {
		t.Fatalf("CreateSession must succeed even when DB save fails: %v", err)
	}
	if sess == nil {
		t.Fatal("expected non-nil session")
	}
	// Session must be retrievable from the in-memory store.
	got, err := svc.GetSession(sess.SessionID)
	if err != nil {
		t.Fatalf("session not found in memory: %v", err)
	}
	if got.SessionID != sess.SessionID {
		t.Errorf("expected %s, got %s", sess.SessionID, got.SessionID)
	}
}

func TestSessionServiceDBDeleteError(t *testing.T) {
	// A DB delete error must not fail the in-memory DeleteSession.
	db := newMockDB()
	db.deleteErr = fmt.Errorf("db unavailable")
	svc := NewSessionService()
	svc.SetDB(db)

	sess, _ := svc.CreateSession("Bambu", "X1C", "", "it")
	if err := svc.DeleteSession(sess.SessionID); err != nil {
		t.Fatalf("DeleteSession must succeed even when DB delete fails: %v", err)
	}
	// Session must be gone from memory.
	_, err := svc.GetSession(sess.SessionID)
	if err == nil {
		t.Fatal("expected error: session should have been deleted from memory")
	}
}

func TestSessionServiceHydrateFromDBError(t *testing.T) {
	db := newMockDB()
	db.listErr = fmt.Errorf("connection refused")
	svc := NewSessionService()
	svc.SetDB(db)

	err := svc.HydrateFromDB(context.Background())
	if err == nil {
		t.Fatal("expected error when DB ListSessions fails")
	}
	if !strings.Contains(err.Error(), "hydrate sessions from DB") {
		t.Errorf("expected 'hydrate sessions from DB' in error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Concurrent access: mixed operations on the same session
// ---------------------------------------------------------------------------

func TestSessionServiceConcurrentMixedAccess(t *testing.T) {
	svc := NewSessionService()
	sess, _ := svc.CreateSession("Creality", "K1", "", "it")
	sessionID := sess.SessionID

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Concurrent AddMessage
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			_ = svc.AddMessage(sessionID, "user", fmt.Sprintf("msg %d", i), false)
		}(i)
	}

	// Concurrent GetSessionSnapshot
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = svc.GetSessionSnapshot(sessionID)
		}()
	}

	// Concurrent UpdateStep
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			_ = svc.UpdateStep(sessionID, i, goroutines)
		}(i)
	}

	wg.Wait()
}
