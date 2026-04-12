package main

import (
	"net/http"
	"net/http/httptest"
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

