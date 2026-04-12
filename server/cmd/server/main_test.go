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
	return newFiberApp("test", sh, ws, nil)
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
