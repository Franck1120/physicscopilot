// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"testing"
	"time"
)

// ── UserFrameLimiter tests ────────────────────────────────────────────────────

// TestUserFrameLimiterConstants verifies the package-level frame-rate constants
// match the documented production values.
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
	if userSessionLimiterExpiry != 5*time.Minute {
		t.Errorf("userSessionLimiterExpiry: want 5m, got %v", userSessionLimiterExpiry)
	}
}

// TestUserFrameLimiterAllowsWithinBurst verifies that the first userFrameBurst
// calls to Allow succeed for a fresh user ID.
func TestUserFrameLimiterAllowsWithinBurst(t *testing.T) {
	u := NewUserFrameLimiter()
	userID := "user-burst-ok"

	for i := 0; i < userFrameBurst; i++ {
		if !u.Allow(userID) {
			t.Fatalf("expected Allow() to return true for call %d within burst", i+1)
		}
	}
}

// TestUserFrameLimiterBlocksAfterBurst verifies that once the burst is
// exhausted, subsequent calls to Allow return false.
func TestUserFrameLimiterBlocksAfterBurst(t *testing.T) {
	u := NewUserFrameLimiter()
	userID := "user-burst-exhaust"

	// Consume all burst tokens.
	for i := 0; i < userFrameBurst; i++ {
		u.Allow(userID)
	}

	// Next call must be blocked.
	if u.Allow(userID) {
		t.Error("expected Allow() to return false after burst exhausted")
	}
}

// TestUserFrameLimiterPerUserIsolation verifies that exhausting one user's
// burst does not affect an independent user's token bucket.
func TestUserFrameLimiterPerUserIsolation(t *testing.T) {
	u := NewUserFrameLimiter()
	userA := "user-a"
	userB := "user-b"

	// Exhaust userA.
	for i := 0; i < userFrameBurst; i++ {
		u.Allow(userA)
	}
	if u.Allow(userA) {
		t.Error("expected userA to be rate-limited after burst exhausted")
	}

	// userB should still have its full burst available.
	if !u.Allow(userB) {
		t.Error("expected userB Allow() to succeed independently of userA's limit")
	}
}

// ── UserSessionTracker tests ──────────────────────────────────────────────────

// TestUserSessionTrackerAdd verifies that Add returns true for a brand-new user.
func TestUserSessionTrackerAdd(t *testing.T) {
	tr := NewUserSessionTracker()
	if !tr.Add("new-user") {
		t.Error("expected Add() to return true for a new user with no existing sessions")
	}
}

// TestUserSessionTrackerMaxSessions verifies that a user may open up to
// userMaxSessions (3) sessions and is rejected on the (userMaxSessions+1)th.
func TestUserSessionTrackerMaxSessions(t *testing.T) {
	tr := NewUserSessionTracker()
	userID := "session-user"

	for i := 0; i < userMaxSessions; i++ {
		if !tr.Add(userID) {
			t.Fatalf("expected Add() %d to return true (within limit)", i+1)
		}
	}

	// One over the limit must be rejected.
	if tr.Add(userID) {
		t.Errorf("expected Add() to return false after %d sessions", userMaxSessions)
	}
}

// TestUserSessionTrackerRemove verifies that Remove frees a slot so that a
// subsequent Add call succeeds again after the limit was reached.
func TestUserSessionTrackerRemove(t *testing.T) {
	tr := NewUserSessionTracker()
	userID := "remove-user"

	// Fill to the limit.
	for i := 0; i < userMaxSessions; i++ {
		tr.Add(userID)
	}
	// At limit — next Add must fail.
	if tr.Add(userID) {
		t.Fatalf("expected Add() to fail at limit before Remove")
	}

	// Free one slot.
	tr.Remove(userID)

	// Now Add must succeed again.
	if !tr.Add(userID) {
		t.Error("expected Add() to succeed after Remove freed a slot")
	}
}

// TestUserSessionTrackerRemoveNonExistent verifies that calling Remove for a
// user that has no tracked sessions does not panic.
func TestUserSessionTrackerRemoveNonExistent(t *testing.T) {
	tr := NewUserSessionTracker()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Remove on non-existent user panicked: %v", r)
		}
	}()

	tr.Remove("ghost-user") // must not panic
}
