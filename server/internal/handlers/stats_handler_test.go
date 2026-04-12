package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// newStatsTestApp wires a StatsHandler to a fresh Fiber app and returns both.
func newStatsTestApp(t *testing.T) (*fiber.App, *services.SessionService) {
	t.Helper()
	sessions := services.NewSessionService()
	h := NewStatsHandler(sessions)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", h.GetStats)
	return app, sessions
}

func TestGetStatsEmpty(t *testing.T) {
	app, _ := newStatsTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveSessions != 0 {
		t.Errorf("active_sessions: want 0, got %d", body.ActiveSessions)
	}
	if body.TotalMessages != 0 {
		t.Errorf("total_messages: want 0, got %d", body.TotalMessages)
	}
}

func TestGetStatsWithSessions(t *testing.T) {
	app, sessions := newStatsTestApp(t)

	sessions.CreateSession("BrandA", "ModelA", "", "") //nolint:errcheck
	sessions.CreateSession("BrandB", "ModelB", "", "") //nolint:errcheck
	sessions.CreateSession("BrandC", "ModelC", "", "") //nolint:errcheck

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveSessions != 3 {
		t.Errorf("active_sessions: want 3, got %d", body.ActiveSessions)
	}
}

func TestGetStatsStatusCode200(t *testing.T) {
	app, _ := newStatsTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestGetStatsResponseIsJSON(t *testing.T) {
	app, _ := newStatsTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}

	if _, ok := raw["active_sessions"]; !ok {
		t.Error("missing required field 'active_sessions' in stats response")
	}
	if _, ok := raw["total_messages"]; !ok {
		t.Error("missing required field 'total_messages' in stats response")
	}
}
