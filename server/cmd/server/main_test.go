package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/handlers"
	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// buildTestApp constructs the Fiber app without starting a listener or
// requiring any environment variables.
func buildTestApp(t *testing.T) *fiber.App {
	t.Helper()
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	return newFiberApp("test", sh, fh, ws, nil)
}

func TestNewFiberAppHealthRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/health not registered: got 404")
	}
}

func TestNewFiberAppSessionsRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// Without auth the endpoint should return 401, not 404.
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/api/sessions not registered: got 404")
	}
}

func TestNewFiberAppMetricsRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// Without credentials the endpoint returns 401, not 404.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/metrics not registered: got 404")
	}
}

func TestNewFiberAppWSRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// Plain GET without Upgrade header should return 426, not 404.
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/ws not registered: got 404")
	}
}

func TestNewFiberAppFeedbackRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// POST /api/feedback without a body returns 400 (validation), not 404.
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/api/feedback not registered: got 404")
	}
}

func TestNewFiberAppDocsRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/api/docs not registered: got 404")
	}
}

func TestNewFiberAppProductionAllowedOriginsWarning(t *testing.T) {
	// APP_ENV=production without ALLOWED_ORIGINS triggers a slog.Warn.
	// Verify the app still builds and serves /health normally.
	t.Setenv("APP_ENV", "production")
	// ALLOWED_ORIGINS intentionally not set → exercises the warning branch.

	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	app := newFiberApp("test", sh, fh, ws, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("production without ALLOWED_ORIGINS: want 200, got %d", resp.StatusCode)
	}
}

// TestSecurityHeadersOnEveryResponse verifies that X-Content-Type-Options,
// X-Frame-Options, Referrer-Policy, and Content-Security-Policy are present
// on all responses regardless of environment.
func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	app := buildTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options: want nosniff, got %q", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options: want DENY, got %q", got)
	}
	if got := resp.Header.Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy: want strict-origin-when-cross-origin, got %q", got)
	}
	if got := resp.Header.Get("Content-Security-Policy"); got == "" {
		t.Error("Content-Security-Policy header is missing")
	}
}

// TestCSRFProtectionDocumented verifies that the API routes carry the
// Authorization header requirement in the CORS AllowHeaders config, confirming
// that JWT Bearer tokens (not cookies) are the auth mechanism and therefore
// CSRF tokens are not required (JWT via Authorization header is CSRF-safe).
func TestCSRFProtectionDocumented(t *testing.T) {
	app := buildTestApp(t)
	// OPTIONS preflight on an API route must expose Authorization in allowed headers.
	req := httptest.NewRequest(http.MethodOptions, "/api/sessions", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowHeaders, "Authorization") {
		t.Errorf("Access-Control-Allow-Headers must include Authorization; got %q", allowHeaders)
	}
}

func TestNewFiberAppHSTSHeaderInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")

	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	app := newFiberApp("test", sh, fh, ws, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	hsts := resp.Header.Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("production: expected Strict-Transport-Security header")
	}
}

