package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
