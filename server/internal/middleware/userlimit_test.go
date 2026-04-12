package middleware

import (
	"sync"
	"testing"
)

// ── UserFrameLimiter ──────────────────────────────────────────────────────────

func TestNewUserFrameLimiter(t *testing.T) {
	u := NewUserFrameLimiter()
	if u == nil {
		t.Fatal("NewUserFrameLimiter returned nil")
	}
	if u.limiters == nil {
		t.Fatal("expected limiters map to be initialized")
	}
}

func TestUserFrameLimiterFirstAllowSucceeds(t *testing.T) {
	u := NewUserFrameLimiter()
	if !u.Allow("user-first") {
		t.Error("first Allow for a new user should always succeed")
	}
}

func TestUserFrameLimiterBlocksAfterBurst(t *testing.T) {
	u := NewUserFrameLimiter()
	userID := "burst-user"

	// Exhaust all burst tokens.
	for i := 0; i < userFrameBurst; i++ {
		if !u.Allow(userID) {
			t.Fatalf("expected Allow to succeed for request %d (burst=%d)", i+1, userFrameBurst)
		}
	}

	// Next call must be blocked.
	if u.Allow(userID) {
		t.Errorf("expected Allow to be blocked after exhausting %d burst tokens", userFrameBurst)
	}
}

func TestUserFrameLimiterPerUserIsolation(t *testing.T) {
	u := NewUserFrameLimiter()

	// Exhaust user1's burst tokens.
	for i := 0; i < userFrameBurst; i++ {
		u.Allow("frame-user-1")
	}

	// user2 must have an independent bucket.
	if !u.Allow("frame-user-2") {
		t.Error("user2 should not be affected by user1's rate limit")
	}
}

func TestUserFrameLimiterConstants(t *testing.T) {
	if userFramesPerMinute != 100 {
		t.Errorf("userFramesPerMinute: want 100, got %d", userFramesPerMinute)
	}
	if userFrameBurst != 20 {
		t.Errorf("userFrameBurst: want 20, got %d", userFrameBurst)
	}
}

// ── UserSessionTracker ────────────────────────────────────────────────────────

func TestNewUserSessionTracker(t *testing.T) {
	tr := NewUserSessionTracker()
	if tr == nil {
		t.Fatal("NewUserSessionTracker returned nil")
	}
	if tr.sessions == nil {
		t.Fatal("expected sessions map to be initialized")
	}
}

func TestUserSessionTrackerAddWithinLimit(t *testing.T) {
	tr := NewUserSessionTracker()
	userID := "session-user"

	for i := 0; i < userMaxSessions; i++ {
		if !tr.Add(userID) {
			t.Fatalf("expected Add to succeed for session %d (limit=%d)", i+1, userMaxSessions)
		}
	}
}

func TestUserSessionTrackerBlocksAtLimit(t *testing.T) {
	tr := NewUserSessionTracker()
	userID := "limit-user"

	for i := 0; i < userMaxSessions; i++ {
		tr.Add(userID)
	}

	if tr.Add(userID) {
		t.Errorf("expected Add to fail after reaching the limit of %d", userMaxSessions)
	}
}

func TestUserSessionTrackerRemoveDecrementsCount(t *testing.T) {
	tr := NewUserSessionTracker()
	userID := "remove-user"

	for i := 0; i < userMaxSessions; i++ {
		tr.Add(userID)
	}
	// At limit — next Add must fail.
	if tr.Add(userID) {
		t.Fatal("expected Add to fail at limit before Remove")
	}

	tr.Remove(userID)

	// One slot freed — Add must succeed again.
	if !tr.Add(userID) {
		t.Error("expected Add to succeed after Remove freed a slot")
	}
}

func TestUserSessionTrackerRemoveAtZeroIsNoop(t *testing.T) {
	tr := NewUserSessionTracker()

	// Remove on a user with no sessions must not panic or produce negative counts.
	tr.Remove("no-sessions-user")

	tr.mu.Lock()
	count := tr.sessions["no-sessions-user"]
	tr.mu.Unlock()

	if count < 0 {
		t.Errorf("session count must not go negative, got %d", count)
	}
}

func TestUserSessionTrackerPerUserIsolation(t *testing.T) {
	tr := NewUserSessionTracker()

	for i := 0; i < userMaxSessions; i++ {
		tr.Add("iso-user-1")
	}

	// user-2 must have an independent counter.
	if !tr.Add("iso-user-2") {
		t.Error("user-2 should not be affected by user-1's session limit")
	}
}

func TestUserSessionTrackerConcurrent(t *testing.T) {
	tr := NewUserSessionTracker()
	userID := "concurrent-tracker-user"

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			tr.Add(userID)
		}()
	}
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			tr.Remove(userID)
		}()
	}

	wg.Wait()

	tr.mu.Lock()
	count := tr.sessions[userID]
	tr.mu.Unlock()

	if count < 0 {
		t.Errorf("session count must not go negative after concurrent access, got %d", count)
	}
}

func TestUserMaxSessionsConstant(t *testing.T) {
	if userMaxSessions != 3 {
		t.Errorf("userMaxSessions: want 3, got %d", userMaxSessions)
	}
}
