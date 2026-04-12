package middleware

import (
	"sync"
	"testing"
	"time"
)

// ── UserFrameLimiter ─────────────────────────────────────────────────────────

func TestUserFrameLimiterConstants(t *testing.T) {
	if userFramesPerMinute != 100 {
		t.Errorf("userFramesPerMinute: want 100, got %d", userFramesPerMinute)
	}
	if userFrameBurst != 20 {
		t.Errorf("userFrameBurst: want 20, got %d", userFrameBurst)
	}
	if userMaxSessions != 3 {
		t.Errorf("userMaxSessions: want 3, got %d", userMaxSessions)
	}
}

func TestNewUserFrameLimiterReturnsNonNil(t *testing.T) {
	ufl := NewUserFrameLimiter()
	if ufl == nil {
		t.Fatal("NewUserFrameLimiter() returned nil")
	}
}

func TestUserFrameLimiterAllowsWithinBurst(t *testing.T) {
	ufl := NewUserFrameLimiter()

	for i := 0; i < userFrameBurst; i++ {
		if !ufl.Allow("user-burst") {
			t.Fatalf("Allow() returned false on request %d, expected true within burst of %d", i+1, userFrameBurst)
		}
	}
}

func TestUserFrameLimiterBlocksAfterBurst(t *testing.T) {
	ufl := NewUserFrameLimiter()
	userID := "user-block"

	// Exhaust all burst tokens.
	for i := 0; i < userFrameBurst; i++ {
		ufl.Allow(userID)
	}

	if ufl.Allow(userID) {
		t.Error("Allow() should return false after burst exhausted")
	}
}

func TestUserFrameLimiterPerUserIsolation(t *testing.T) {
	ufl := NewUserFrameLimiter()

	// Exhaust userA's burst.
	for i := 0; i < userFrameBurst; i++ {
		ufl.Allow("userA")
	}
	if ufl.Allow("userA") {
		t.Error("userA should be blocked after exhausting burst")
	}

	// userB must still have tokens.
	if !ufl.Allow("userB") {
		t.Error("userB should be unaffected by userA's exhaustion")
	}
}

func TestUserFrameLimiterCleanupRemovesIdleEntries(t *testing.T) {
	ufl := NewUserFrameLimiter()

	// Trigger creation of a limiter entry.
	ufl.Allow("idle-user")

	// Verify entry exists.
	ufl.mu.Lock()
	if _, ok := ufl.limiters["idle-user"]; !ok {
		ufl.mu.Unlock()
		t.Fatal("expected limiter entry for idle-user after Allow()")
	}

	// Backdate lastSeen beyond userSessionLimiterExpiry so the cleanup considers it stale.
	ufl.limiters["idle-user"].lastSeen = time.Now().Add(-2 * userSessionLimiterExpiry)
	ufl.mu.Unlock()

	// Simulate one cleanup pass (same logic as cleanupLoop body).
	ufl.mu.Lock()
	for id, e := range ufl.limiters {
		if time.Since(e.lastSeen) > userSessionLimiterExpiry {
			delete(ufl.limiters, id)
		}
	}
	ufl.mu.Unlock()

	ufl.mu.Lock()
	defer ufl.mu.Unlock()
	if _, ok := ufl.limiters["idle-user"]; ok {
		t.Error("idle-user limiter should have been cleaned up")
	}
}

func TestUserFrameLimiterConcurrentAccess(t *testing.T) {
	ufl := NewUserFrameLimiter()
	var wg sync.WaitGroup
	const goroutines = 50

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ufl.Allow("concurrent-user")
		}()
	}
	wg.Wait()

	// No race or panic means success. The -race flag catches data races.
}

// ── UserSessionTracker ───────────────────────────────────────────────────────

func TestNewUserSessionTrackerReturnsNonNil(t *testing.T) {
	tracker := NewUserSessionTracker()
	if tracker == nil {
		t.Fatal("NewUserSessionTracker() returned nil")
	}
}

func TestUserSessionTrackerAddWithinLimit(t *testing.T) {
	tracker := NewUserSessionTracker()

	for i := 0; i < userMaxSessions; i++ {
		if !tracker.Add("user-ok") {
			t.Fatalf("Add() returned false on session %d, expected true (limit=%d)", i+1, userMaxSessions)
		}
	}
}

func TestUserSessionTrackerRejectsOverLimit(t *testing.T) {
	tracker := NewUserSessionTracker()

	// Fill up to the max.
	for i := 0; i < userMaxSessions; i++ {
		tracker.Add("user-full")
	}

	// Next add should be rejected.
	if tracker.Add("user-full") {
		t.Errorf("Add() should return false when %d sessions already active", userMaxSessions)
	}
}

func TestUserSessionTrackerRemoveAllowsNewSession(t *testing.T) {
	tracker := NewUserSessionTracker()

	// Fill to capacity.
	for i := 0; i < userMaxSessions; i++ {
		tracker.Add("user-rm")
	}

	// Remove one session.
	tracker.Remove("user-rm")

	// Should be able to add one more.
	if !tracker.Add("user-rm") {
		t.Error("Add() should succeed after Remove() freed a slot")
	}
}

func TestUserSessionTrackerRemoveDoesNotGoBelowZero(t *testing.T) {
	tracker := NewUserSessionTracker()

	// Remove on a user that was never added should not panic or underflow.
	tracker.Remove("nonexistent")

	tracker.mu.Lock()
	count := tracker.sessions["nonexistent"]
	tracker.mu.Unlock()

	if count != 0 {
		t.Errorf("session count for nonexistent user: want 0, got %d", count)
	}
}

func TestUserSessionTrackerPerUserIsolation(t *testing.T) {
	tracker := NewUserSessionTracker()

	// Fill userA to capacity.
	for i := 0; i < userMaxSessions; i++ {
		tracker.Add("userA")
	}

	// userB should still be able to add.
	if !tracker.Add("userB") {
		t.Error("userB should be able to add sessions regardless of userA's count")
	}
}

func TestUserSessionTrackerConcurrentAccess(t *testing.T) {
	tracker := NewUserSessionTracker()
	var wg sync.WaitGroup
	const goroutines = 50

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			tracker.Add("concurrent-user")
			tracker.Remove("concurrent-user")
		}()
	}
	wg.Wait()

	// No race or panic means success.
}
