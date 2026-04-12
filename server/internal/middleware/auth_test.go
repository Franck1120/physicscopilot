// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// newMetricsApp builds a minimal Fiber app with MetricsBasicAuth protecting /metrics.
// The middleware is constructed inside the function so t.Setenv takes effect before
// MetricsBasicAuth() reads METRICS_PASSWORD from the environment.
func newMetricsApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/metrics", MetricsBasicAuth(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

// basicAuthHeader returns the Base64-encoded "user:pass" header value.
func basicAuthHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

// ── MetricsBasicAuth ──────────────────────────────────────────────────────────

func TestMetricsBasicAuthDisabledWhenNoPassword(t *testing.T) {
	t.Setenv("METRICS_PASSWORD", "")
	app := newMetricsApp(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("no password: want 503, got %d", resp.StatusCode)
	}
}

func TestMetricsBasicAuthReturns401WithNoHeader(t *testing.T) {
	t.Setenv("METRICS_PASSWORD", "supersecret")
	app := newMetricsApp(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no header: want 401, got %d", resp.StatusCode)
	}
	if wwwAuth := resp.Header.Get("WWW-Authenticate"); wwwAuth == "" {
		t.Error("expected WWW-Authenticate header on 401")
	}
}

func TestMetricsBasicAuthAllowsValidCredentials(t *testing.T) {
	const pass = "correctpassword"
	t.Setenv("METRICS_PASSWORD", pass)
	app := newMetricsApp(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", basicAuthHeader(metricsUser, pass))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid credentials: want 200, got %d", resp.StatusCode)
	}
}

func TestMetricsBasicAuthReturns401WithWrongPassword(t *testing.T) {
	t.Setenv("METRICS_PASSWORD", "correct")
	app := newMetricsApp(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", basicAuthHeader(metricsUser, "wrong"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong password: want 401, got %d", resp.StatusCode)
	}
}

func TestMetricsBasicAuthReturns401WithWrongUser(t *testing.T) {
	const pass = "mypassword"
	t.Setenv("METRICS_PASSWORD", pass)
	app := newMetricsApp(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", basicAuthHeader("notadmin", pass))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong user: want 401, got %d", resp.StatusCode)
	}
}

func TestMetricsBasicAuthReturns401WithInvalidBase64(t *testing.T) {
	t.Setenv("METRICS_PASSWORD", "secret")
	app := newMetricsApp(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Basic not!valid!base64!!!")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("invalid base64: want 401, got %d", resp.StatusCode)
	}
}

func TestMetricsBasicAuthReturns401WithMalformedCredentials(t *testing.T) {
	t.Setenv("METRICS_PASSWORD", "secret")
	app := newMetricsApp(t)

	// Base64 of "nocolon" — no colon separator.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("nocolon")))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no colon: want 401, got %d", resp.StatusCode)
	}
}

// ── NewUserRateLimiter ────────────────────────────────────────────────────────

func TestNewUserRateLimiterReturnsNonNil(t *testing.T) {
	ul := NewUserRateLimiter()
	if ul == nil {
		t.Fatal("NewUserRateLimiter() returned nil")
	}
}

func TestNewUserRateLimiterHasProductionDefaults(t *testing.T) {
	ul := NewUserRateLimiter()
	// First Allow for a fresh user must succeed — production burst is > 0.
	if !ul.Allow("fresh-test-user") {
		t.Error("first Allow for a fresh user should succeed with production limiter")
	}
}
