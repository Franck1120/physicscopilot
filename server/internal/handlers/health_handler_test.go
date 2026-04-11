package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

func newHealthApp(version string, startTime time.Time) (*fiber.App, *WSHandler) {
	sessionSvc := services.NewSessionService(nil, nil)
	convSvc := services.NewConversationService(sessionSvc, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler(version, startTime, ws))
	return app, ws
}

func TestHealthEndpointReturns200(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestHealthEndpointHasRequiredFields(t *testing.T) {
	startTime := time.Now().Add(-5 * time.Minute)
	app, _ := newHealthApp("0.1.0", startTime)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("status: want %q, got %q", "ok", body.Status)
	}
	if body.Service != "physicscopilot" {
		t.Errorf("service: want %q, got %q", "physicscopilot", body.Service)
	}
	if body.Version != "0.1.0" {
		t.Errorf("version: want %q, got %q", "0.1.0", body.Version)
	}
	if body.Uptime == "" {
		t.Error("uptime must not be empty")
	}
	// Uptime should include at least one digit and a time unit.
	if !strings.ContainsAny(body.Uptime, "0123456789") {
		t.Errorf("uptime looks malformed: %q", body.Uptime)
	}
}

func TestHealthEndpointReportsActiveConnections(t *testing.T) {
	app, ws := newHealthApp("0.1.0", time.Now())

	// Simulate 3 open connections.
	ws.activeConns.Add(3)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveConnections != 3 {
		t.Errorf("active_connections: want 3, got %d", body.ActiveConnections)
	}
}

func TestHealthEndpointVersionPropagated(t *testing.T) {
	app, _ := newHealthApp("1.2.3", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, _ := app.Test(req)

	var body HealthResponse
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck
	if body.Version != "1.2.3" {
		t.Errorf("version: want %q, got %q", "1.2.3", body.Version)
	}
}
