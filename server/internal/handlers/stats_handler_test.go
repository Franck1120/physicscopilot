package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// stubWSCounter implements WSConnCounter for tests.
type stubWSCounter struct{ n int32 }

func (s *stubWSCounter) ActiveConnections() int32 { return s.n }

// stubRAGLoader implements RAGLoader for tests.
type stubRAGLoader struct {
	loaded bool
	count  int
}

func (r *stubRAGLoader) Loaded() bool    { return r.loaded }
func (r *stubRAGLoader) EntryCount() int { return r.count }

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

	requiredFields := []string{
		"active_sessions",
		"active_ws_connections",
		"kb_loaded",
		"kb_entry_count",
		"uptime_seconds",
		"version",
		"total_messages",
	}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing required field %q in stats response", field)
		}
	}
}

// TestGetStatsFullHandlerNonNilDependencies verifies that NewStatsHandlerFull
// wires ws and rag correctly when both are non-nil.
func TestGetStatsFullHandlerNonNilDependencies(t *testing.T) {
	sessions := services.NewSessionService()
	ws := &stubWSCounter{n: 3}
	rag := &stubRAGLoader{loaded: true, count: 42}
	h := NewStatsHandlerFull(sessions, ws, rag, "1.2.3", time.Now())

	if h.ws == nil {
		t.Error("expected non-nil ws counter after NewStatsHandlerFull")
	}
	if h.rag == nil {
		t.Error("expected non-nil rag loader after NewStatsHandlerFull")
	}
	if h.version != "1.2.3" {
		t.Errorf("version: want %q, got %q", "1.2.3", h.version)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", h.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveWSConnections != 3 {
		t.Errorf("active_ws_connections: want 3, got %d", body.ActiveWSConnections)
	}
	if body.KBEntryCount != 42 {
		t.Errorf("kb_entry_count: want 42, got %d", body.KBEntryCount)
	}
}

// TestGetStatsUptimeIsPositive verifies that uptime_seconds is greater than 0
// when the handler is constructed with a start time in the past.
func TestGetStatsUptimeIsPositive(t *testing.T) {
	sessions := services.NewSessionService()
	// Start time 2 seconds in the past ensures uptime_seconds >= 1.
	past := time.Now().Add(-2 * time.Second)
	h := NewStatsHandlerFull(sessions, nil, nil, "0.1.0", past)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", h.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.UptimeSeconds <= 0 {
		t.Errorf("uptime_seconds: want > 0, got %d", body.UptimeSeconds)
	}
}

func TestGetStatsWSConnections(t *testing.T) {
	sessions := services.NewSessionService()
	ws := &stubWSCounter{n: 5}
	rag := &stubRAGLoader{loaded: true, count: 66}
	h := NewStatsHandlerFull(sessions, ws, rag, "0.1.0", time.Now())

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", h.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveWSConnections != 5 {
		t.Errorf("active_ws_connections: want 5, got %d", body.ActiveWSConnections)
	}
	if !body.KBLoaded {
		t.Error("kb_loaded: want true, got false")
	}
	if body.KBEntryCount != 66 {
		t.Errorf("kb_entry_count: want 66, got %d", body.KBEntryCount)
	}
	if body.Version != "0.1.0" {
		t.Errorf("version: want %q, got %q", "0.1.0", body.Version)
	}
}
