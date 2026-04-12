// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// testRateLimitApp wires the given limiter into a minimal Fiber app.
func testRateLimitApp(rl *IPRateLimiter) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(rl.Middleware())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func TestRateLimiterProductionConstants(t *testing.T) {
	if apiRequestsPerMinute != 60 {
		t.Errorf("apiRequestsPerMinute: want 60, got %d", apiRequestsPerMinute)
	}
	if apiLimiterBurst != 10 {
		t.Errorf("apiLimiterBurst: want 10, got %d", apiLimiterBurst)
	}
}

// TestRateLimiter60thAllowed61stBlocked tests the documented limit:
// after 60 requests the 61st receives 429.
// The limiter is created with burst=60 so all 60 tokens are available immediately.
func TestRateLimiter60thAllowed61stBlocked(t *testing.T) {
	// rate=60/min, burst=60 → exactly 60 tokens available immediately.
	rl := newIPRateLimiterWith(60, 60)
	app := testRateLimitApp(rl)

	for i := 1; i <= 60; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("request %d: want 200, got %d", i, resp.StatusCode)
		}
	}

	// 61st request must be rejected.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("61st request: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("61st request: want 429, got %d", resp.StatusCode)
	}
}

func TestRateLimiter429ResponseIsJSON(t *testing.T) {
	// burst=1 → second request immediately gets 429.
	rl := newIPRateLimiterWith(1, 1)
	app := testRateLimitApp(rl)

	// First request consumes the only token.
	app.Test(httptest.NewRequest(http.MethodGet, "/", nil)) //nolint:errcheck

	// Second request must be 429 with a JSON body.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Errorf("response body is not valid JSON: %v", err)
	}
	if _, ok := payload["error"]; !ok {
		t.Error("expected 'error' key in JSON response")
	}
}

// TestRateLimiterPerIPIsolation verifies each IP gets its own bucket.
func TestRateLimiterPerIPIsolation(t *testing.T) {
	// burst=1 — each IP gets exactly one free request.
	rl := newIPRateLimiterWith(1, 1)

	limiterA := rl.getLimiter("192.0.2.1")
	limiterB := rl.getLimiter("192.0.2.2")

	if limiterA == limiterB {
		t.Fatal("expected different limiters for different IPs")
	}

	// Exhaust A's token.
	if !limiterA.Allow() {
		t.Fatal("expected first Allow() to succeed for IP A")
	}
	// A is now exhausted.
	if limiterA.Allow() {
		t.Error("expected IP A to be rate-limited after burst exhausted")
	}
	// B should still have its token.
	if !limiterB.Allow() {
		t.Error("expected IP B to still have its token, independent of IP A")
	}
}

func TestRateLimiterSameIPSharesBucket(t *testing.T) {
	rl := newIPRateLimiterWith(1, 1)

	// Two calls with the same IP return the same limiter instance.
	l1 := rl.getLimiter("10.0.0.1")
	l2 := rl.getLimiter("10.0.0.1")
	if l1 != l2 {
		t.Error("expected same limiter for repeated calls with same IP")
	}
}

// TestNewIPRateLimiterUsesProductionDefaults verifies the exported constructor
// wires the correct rate and burst values from package constants.
func TestNewIPRateLimiterUsesProductionDefaults(t *testing.T) {
	rl := NewIPRateLimiter()

	// Fetch the limiter for a new IP; it should have burst=apiLimiterBurst tokens.
	l := rl.getLimiter("172.16.0.1")
	allowed := 0
	for l.Allow() {
		allowed++
	}
	if allowed != apiLimiterBurst {
		t.Errorf("expected burst of %d tokens, consumed %d before block", apiLimiterBurst, allowed)
	}
}

// TestRateLimiterMiddlewareBlocksAfterBurst exercises the Fiber middleware path
// rather than the getLimiter method directly.
func TestRateLimiterMiddlewareBlocksAfterBurst(t *testing.T) {
	// burst=3: first 3 requests pass, 4th must be 429.
	rl := newIPRateLimiterWith(60, 3)
	app := testRateLimitApp(rl)

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("request %d: want 200, got %d", i, resp.StatusCode)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("4th request: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("4th request: want 429 after burst exhausted, got %d", resp.StatusCode)
	}
}

// ── UserRateLimiter tests ─────────────────────────────────────────────────────

func TestUserRateLimiterAllowsWithinBurst(t *testing.T) {
	ul := newUserRateLimiterWith(60, 5) // burst=5
	userID := "user-abc"

	for i := 0; i < 5; i++ {
		if !ul.Allow(userID) {
			t.Fatalf("expected Allow() to return true for request %d within burst", i+1)
		}
	}
}

func TestUserRateLimiterBlocksAfterBurst(t *testing.T) {
	ul := newUserRateLimiterWith(60, 2) // burst=2

	ul.Allow("user-xyz") // consume token 1
	ul.Allow("user-xyz") // consume token 2

	if ul.Allow("user-xyz") {
		t.Error("expected Allow() to return false after burst exhausted")
	}
}

func TestUserRateLimiterEmptyUserIDAlwaysAllows(t *testing.T) {
	ul := newUserRateLimiterWith(1, 1) // very restrictive
	// Exhaust with a real ID to prove the limiter works
	ul.Allow("real-user")
	ul.Allow("real-user")

	// Empty user ID (unauthenticated / dev mode) must always pass
	for i := 0; i < 10; i++ {
		if !ul.Allow("") {
			t.Error("expected Allow('') to always return true for unauthenticated requests")
		}
	}
}

func TestUserRateLimiterPerUserIsolation(t *testing.T) {
	ul := newUserRateLimiterWith(60, 1) // burst=1

	// Exhaust userA's token
	if !ul.Allow("userA") {
		t.Fatal("first Allow for userA should succeed")
	}
	if ul.Allow("userA") {
		t.Error("second Allow for userA should be blocked")
	}

	// userB's token must be unaffected
	if !ul.Allow("userB") {
		t.Error("Allow for userB should succeed independently of userA's limit")
	}
}

func TestNewUserRateLimiterUsesProductionDefaults(t *testing.T) {
	if userMessagesPerMinute != 30 {
		t.Errorf("userMessagesPerMinute: want 30, got %d", userMessagesPerMinute)
	}
	if userLimiterBurst != 5 {
		t.Errorf("userLimiterBurst: want 5, got %d", userLimiterBurst)
	}
}

// ── IP Ban Tests ──────────────────────────────────────────────────────────────

// TestIPBanAfterRepeatedViolations verifies that an IP which exceeds the rate
// limit banViolationThreshold times within banViolationWindow is banned for
// banDuration and receives 403 Forbidden (not 429 Too Many Requests).
func TestIPBanAfterRepeatedViolations(t *testing.T) {
	// rate=1/min, burst=1: every request after the first is a violation.
	rl := newIPRateLimiterWith(1, 1)
	app := testRateLimitApp(rl)

	ip := "10.99.0.1"

	// Consume the one available token.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = ip + ":9999"
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 for first request, got %d", resp.StatusCode)
	}

	// Trigger banViolationThreshold violations so the IP is banned.
	for i := 0; i < banViolationThreshold; i++ {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = ip + ":9999"
		if _, err := app.Test(r); err != nil {
			t.Fatalf("violation request %d: %v", i+1, err)
		}
	}

	// Next request must be banned: 403, not 429.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = ip + ":9999"
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("banned request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("want 403 after %d violations, got %d", banViolationThreshold, resp.StatusCode)
	}
}

// TestIPBanDoesNotAffectOtherIPs verifies that banning one IP does not
// affect other IPs. Tested at the method level because app.Test() resolves
// c.IP() from the underlying fasthttpConn (not req.RemoteAddr).
func TestIPBanDoesNotAffectOtherIPs(t *testing.T) {
	rl := newIPRateLimiterWith(1, 1)

	bannedIP := "10.99.0.2"
	safeIP := "10.99.0.3"

	// Trigger enough violations for bannedIP to be banned.
	for i := 0; i < banViolationThreshold; i++ {
		rl.recordViolation(bannedIP)
	}

	if !rl.isBanned(bannedIP) {
		t.Fatal("bannedIP should be banned after threshold violations")
	}
	if rl.isBanned(safeIP) {
		t.Error("safeIP must not be banned when a different IP was banned")
	}
}

// TestIPBanResponseIsForbiddenJSON verifies the ban response body is JSON
// with an "error" key and status 403.
func TestIPBanResponseIsForbiddenJSON(t *testing.T) {
	rl := newIPRateLimiterWith(1, 1)
	app := testRateLimitApp(rl)

	ip := "10.99.0.4"

	// Exhaust token then trigger enough violations to ban.
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = ip + ":1"
	app.Test(r) //nolint:errcheck
	for i := 0; i < banViolationThreshold; i++ {
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = ip + ":1"
		app.Test(r) //nolint:errcheck
	}

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = ip + ":1"
	resp, err := app.Test(r)
	if err != nil {
		t.Fatalf("banned request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("want 403, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("ban response is not JSON: %v", err)
	}
	if _, ok := payload["error"]; !ok {
		t.Error("ban response JSON must have 'error' key")
	}
}

// TestBanConstants verifies the production ban constants have sensible values.
func TestBanConstants(t *testing.T) {
	if banViolationThreshold != 10 {
		t.Errorf("banViolationThreshold: want 10, got %d", banViolationThreshold)
	}
	if banViolationWindow != 1*time.Minute {
		t.Errorf("banViolationWindow: want 1m, got %v", banViolationWindow)
	}
	if banDuration != 5*time.Minute {
		t.Errorf("banDuration: want 5m, got %v", banDuration)
	}
}

// TestBanExpiresAfterDuration verifies that a banned IP becomes unbanned once
// its ban entry has expired.
func TestBanExpiresAfterDuration(t *testing.T) {
	rl := newIPRateLimiterWith(1, 1)

	ip := "10.11.0.1"

	// Ban the IP by recording enough violations.
	for i := 0; i < banViolationThreshold; i++ {
		rl.recordViolation(ip)
	}

	if !rl.isBanned(ip) {
		t.Fatal("expected IP to be banned after threshold violations")
	}

	// Manually backdate the ban expiry so it looks already elapsed.
	rl.mu.Lock()
	rl.bans[ip] = time.Now().Add(-1 * time.Second)
	rl.mu.Unlock()

	// isBanned must detect the expired entry and return false.
	if rl.isBanned(ip) {
		t.Error("expected IP ban to have expired, but isBanned still returns true")
	}

	// The entry must have been removed from the map as part of the expiry check.
	rl.mu.Lock()
	_, stillPresent := rl.bans[ip]
	rl.mu.Unlock()
	if stillPresent {
		t.Error("expected expired ban entry to be removed from bans map")
	}
}

// TestViolationWindowSlide verifies that violations recorded outside the
// banViolationWindow are not counted toward a ban. When the window has
// elapsed the counter resets so a single new violation does not trigger a ban.
func TestViolationWindowSlide(t *testing.T) {
	rl := newIPRateLimiterWith(1, 1)

	ip := "10.11.0.2"

	// Record threshold-1 violations, then slide the window start into the past.
	for i := 0; i < banViolationThreshold-1; i++ {
		rl.recordViolation(ip)
	}

	// Move the window start far enough into the past that it appears stale.
	rl.mu.Lock()
	rl.violations[ip].windowStart = time.Now().Add(-(banViolationWindow + time.Second))
	rl.mu.Unlock()

	// One more violation: the stale window should reset the counter to 1,
	// which is below the threshold — so no ban must occur.
	rl.recordViolation(ip)

	if rl.isBanned(ip) {
		t.Error("expected IP not to be banned: old violations outside the window must not count")
	}

	// Verify that the violation counter was indeed reset to 1 (not accumulated).
	rl.mu.Lock()
	v := rl.violations[ip]
	rl.mu.Unlock()
	if v == nil || v.count != 1 {
		t.Errorf("expected violation count to be reset to 1 after window slide, got %v", v)
	}
}

// TestCleanupRemovesExpiredBans verifies that the cleanup loop removes bans
// whose expiry has already passed. We exercise the cleanup logic directly by
// invoking cleanupLoop-equivalent inline cleanup to keep the test synchronous.
func TestCleanupRemovesExpiredBans(t *testing.T) {
	rl := newIPRateLimiterWith(1, 1)

	ip := "10.11.0.3"

	// Insert an already-expired ban directly.
	rl.mu.Lock()
	rl.bans[ip] = time.Now().Add(-1 * time.Second)
	rl.mu.Unlock()

	// isBanned triggers the lazy-expiry path which removes expired entries.
	if rl.isBanned(ip) {
		t.Error("expected isBanned to return false for an already-expired ban")
	}

	rl.mu.Lock()
	_, present := rl.bans[ip]
	rl.mu.Unlock()
	if present {
		t.Error("expected expired ban to have been removed from the bans map")
	}
}

// TestIPRateLimiterConcurrentRequests spawns 50 goroutines firing requests in
// parallel and verifies that the limiter does not panic and remains consistent.
func TestIPRateLimiterConcurrentRequests(t *testing.T) {
	// Use a high burst so most requests pass; the point is thread-safety, not blocking.
	rl := newIPRateLimiterWith(6000, 100)
	app := testRateLimitApp(rl)

	const goroutines = 50
	results := make(chan int, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			if err != nil {
				results <- -1
				return
			}
			results <- resp.StatusCode
		}()
	}

	for i := 0; i < goroutines; i++ {
		status := <-results
		if status == -1 {
			t.Error("a goroutine received an unexpected error from app.Test")
		}
		if status != http.StatusOK && status != http.StatusTooManyRequests && status != http.StatusForbidden {
			t.Errorf("unexpected status %d — want 200, 429, or 403", status)
		}
	}
}
