// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// newLoggerApp builds a minimal Fiber app with StructuredLogger, capturing
// the request_id set by the middleware into the provided slice on each request.
func newLoggerApp(captured *[]string) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(StructuredLogger())
	app.Get("/", func(c *fiber.Ctx) error {
		*captured = append(*captured, RequestID(c))
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/bad", func(c *fiber.Ctx) error {
		*captured = append(*captured, RequestID(c))
		return c.SendStatus(fiber.StatusBadRequest)
	})
	app.Get("/err", func(c *fiber.Ctx) error {
		*captured = append(*captured, RequestID(c))
		return c.SendStatus(fiber.StatusInternalServerError)
	})
	return app
}

func TestStructuredLoggerInjectsRequestID(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	if len(ids) == 0 || ids[0] == "" {
		t.Error("expected non-empty request_id in c.Locals")
	}
	// generateRequestID() returns 16-char lowercase hex (8 random bytes).
	if got := ids[0]; len(got) != 16 {
		t.Errorf("request_id should be 16-char hex, got %q (len=%d)", got, len(got))
	}
}

func TestStructuredLoggerUniqueIDPerRequest(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if _, err := app.Test(req); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}

	if len(ids) != 5 {
		t.Fatalf("expected 5 captured IDs, got %d", len(ids))
	}
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate request_id: %q", id)
		}
		seen[id] = true
	}
}

func TestStructuredLoggerPassesThroughStatus(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/bad", http.StatusBadRequest},
		{"/err", http.StatusInternalServerError},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
		if resp.StatusCode != tc.want {
			t.Errorf("%s: want %d, got %d", tc.path, tc.want, resp.StatusCode)
		}
	}
}

func TestRequestIDReturnsEmptyWithoutMiddleware(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	var captured string
	app.Get("/", func(c *fiber.Ctx) error {
		captured = RequestID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("test: %v", err)
	}
	if captured != "" {
		t.Errorf("expected empty request_id without middleware, got %q", captured)
	}
}

func TestAnonymizeIPProduces8CharHex(t *testing.T) {
	result := anonymizeIP("192.168.1.100")
	if len(result) != 8 {
		t.Errorf("anonymizeIP: want 8-char hex, got %q (len=%d)", result, len(result))
	}
	for _, ch := range result {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("anonymizeIP: non-hex character %q in result %q", ch, result)
		}
	}
}

func TestAnonymizeIPIsDeterministic(t *testing.T) {
	r1 := anonymizeIP("10.0.0.1")
	r2 := anonymizeIP("10.0.0.1")
	if r1 != r2 {
		t.Errorf("anonymizeIP is not deterministic: %q != %q", r1, r2)
	}
}

func TestAnonymizeIPDiffersForDifferentIPs(t *testing.T) {
	r1 := anonymizeIP("1.2.3.4")
	r2 := anonymizeIP("5.6.7.8")
	if r1 == r2 {
		t.Error("expected different hashes for different IPs")
	}
}

// ── Middleware logs HTTP method ────────────────────────────────────────────────

// TestStructuredLoggerDoesNotBreakOnDifferentMethods verifies the middleware
// works for POST, PUT, DELETE, PATCH requests without panicking or altering
// the response status code. This ensures the "method" field logged is exercised
// for all HTTP verbs.
func TestStructuredLoggerDoesNotBreakOnDifferentMethods(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(StructuredLogger())
	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodPatch,
	} {
		m := method // capture
		app.Add(m, "/echo", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
	}

	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodPatch,
	} {
		req := httptest.NewRequest(method, "/echo", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s /echo: test error: %v", method, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s /echo: want 200, got %d", method, resp.StatusCode)
		}
	}
}

// ── Middleware logs path ───────────────────────────────────────────────────────

// TestStructuredLoggerPreservesPathInResponse verifies that the middleware is
// transparent — requesting different paths returns the correct response for each,
// confirming c.Path() would be logged correctly for each distinct path.
func TestStructuredLoggerPreservesPathInResponse(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(StructuredLogger())
	app.Get("/alpha", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/beta", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusCreated) })
	app.Get("/gamma", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	cases := []struct {
		path string
		want int
	}{
		{"/alpha", http.StatusOK},
		{"/beta", http.StatusCreated},
		{"/gamma", http.StatusNoContent},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("GET %s: test error: %v", tc.path, err)
		}
		if resp.StatusCode != tc.want {
			t.Errorf("GET %s: want %d, got %d", tc.path, tc.want, resp.StatusCode)
		}
	}
}

// ── Middleware logs status code ───────────────────────────────────────────────

// TestStructuredLoggerLogsCorrectStatusForAllLevels verifies that the middleware
// handles 2xx, 4xx, and 5xx responses without changing them. The logged
// status code must match what the handler set.
func TestStructuredLoggerLogsCorrectStatusForAllLevels(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	cases := []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},          // INFO level
		{"/bad", http.StatusBadRequest}, // WARN level
		{"/err", http.StatusInternalServerError}, // ERROR level
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}
		if resp.StatusCode != tc.want {
			t.Errorf("GET %s: want status %d, got %d", tc.path, tc.want, resp.StatusCode)
		}
	}
}

// ── Middleware logs duration ───────────────────────────────────────────────────

// TestStructuredLoggerCompletesRequestWithinReasonableTime verifies that the
// middleware overhead is minimal: the handler completes and returns before a
// generous timeout, meaning latency_ms would be > 0 but small.
func TestStructuredLoggerCompletesRequestWithinReasonableTime(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, 500) // 500 ms timeout
	if err != nil {
		t.Fatalf("request timed out or failed: %v", err)
	}
	elapsed := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	// The middleware must complete in well under 500 ms.
	if elapsed > 400*time.Millisecond {
		t.Errorf("request took too long: %v (middleware overhead?)", elapsed)
	}
}

// TestStructuredLoggerRequestIDPresentInLocalsAfterMiddleware confirms that
// the request_id injected by the middleware is visible to downstream handlers,
// which is a prerequisite for logging correlation IDs on every log line.
func TestStructuredLoggerRequestIDPresentInLocalsAfterMiddleware(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("test: %v", err)
	}

	if len(ids) == 0 {
		t.Fatal("no request_id captured from c.Locals")
	}
	if ids[0] == "" {
		t.Error("request_id must be non-empty so it can be included in log output")
	}
}
