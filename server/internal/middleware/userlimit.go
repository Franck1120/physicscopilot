// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// userFramesPerMinute is the maximum number of camera frames a single
	// authenticated user may send across ALL their concurrent connections.
	userFramesPerMinute = 100
	// userFrameBurst allows short bursts above the steady-state frame rate.
	userFrameBurst = 20
	// userMaxSessions is the maximum number of concurrent WebSocket sessions
	// allowed per authenticated user.
	userMaxSessions = 3
	// userSessionLimiterExpiry removes idle per-user limiters to prevent memory growth.
	userSessionLimiterExpiry = 5 * time.Minute
)

// ---------------------------------------------------------------------------
// UserFrameLimiter — per-user frame rate limiting across all WS connections
// ---------------------------------------------------------------------------

// UserFrameLimiter enforces a per-user frame rate limit (100 frames/min)
// shared across all WebSocket connections belonging to the same user.
type UserFrameLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipEntry // ipEntry is defined in ratelimit.go (same package)
}

// NewUserFrameLimiter creates a limiter with the production defaults and
// starts a background cleanup goroutine.
func NewUserFrameLimiter() *UserFrameLimiter {
	u := &UserFrameLimiter{limiters: make(map[string]*ipEntry)}
	go u.cleanupLoop()
	return u
}

// Allow returns true if userID is within the frame rate limit.
func (u *UserFrameLimiter) Allow(userID string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	e, ok := u.limiters[userID]
	if !ok {
		e = &ipEntry{
			limiter: rate.NewLimiter(
				rate.Every(time.Minute/time.Duration(userFramesPerMinute)),
				userFrameBurst,
			),
		}
		u.limiters[userID] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

// cleanupLoop periodically removes per-user frame limiters that have been idle
// for longer than userSessionLimiterExpiry. Runs as a background goroutine.
func (u *UserFrameLimiter) cleanupLoop() {
	ticker := time.NewTicker(userSessionLimiterExpiry)
	defer ticker.Stop()
	for range ticker.C {
		u.mu.Lock()
		for id, e := range u.limiters {
			if time.Since(e.lastSeen) > userSessionLimiterExpiry {
				delete(u.limiters, id)
			}
		}
		u.mu.Unlock()
	}
}

// ---------------------------------------------------------------------------
// UserSessionTracker — per-user concurrent session limit
// ---------------------------------------------------------------------------

// UserSessionTracker enforces a per-user limit on concurrent WebSocket sessions.
// Unlike the IP-based tracker, this limits authenticated users regardless of
// which IP address they connect from.
type UserSessionTracker struct {
	mu       sync.Mutex
	sessions map[string]int
}

// NewUserSessionTracker creates a tracker enforcing userMaxSessions (3)
// concurrent sessions per user.
func NewUserSessionTracker() *UserSessionTracker {
	return &UserSessionTracker{sessions: make(map[string]int)}
}

// Add increments the session count for userID and returns true if within the
// limit. Returns false if the limit is already reached.
func (t *UserSessionTracker) Add(userID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sessions[userID] >= userMaxSessions {
		return false
	}
	t.sessions[userID]++
	return true
}

// Remove decrements the session count for userID.
func (t *UserSessionTracker) Remove(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sessions[userID] > 0 {
		t.sessions[userID]--
	}
}
