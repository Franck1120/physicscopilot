package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// newStatsEdgeApp is a test helper that builds a Fiber app with a fully-wired
// StatsHandlerFull so edge-case dependencies (ws, rag, version) can be
// controlled independently.
func newStatsEdgeApp(
	t *testing.T,
	sessions *services.SessionService,
	ws WSConnCounter,
	rag RAGLoader,
	version string,
	startTime time.Time,
) *fiber.App {
	t.Helper()
	h := NewStatsHandlerFull(sessions, ws, rag, version, startTime)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", h.GetStats)
	return app
}

// TestStatsEdgeZeroConnectionsReportsZero verifies that when no WebSocket
// counter is injected (ws == nil) the active_ws_connections field is 0 and
// the response is still HTTP 200.
func TestStatsEdgeZeroConnectionsReportsZero(t *testing.T) {
	t.Parallel()
	sessions := services.NewSessionService()
	app := newStatsEdgeApp(t, sessions, nil, nil, "1.0.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveWSConnections != 0 {
		t.Errorf("active_ws_connections: want 0 (nil ws), got %d", body.ActiveWSConnections)
	}
	if body.ActiveSessions != 0 {
		t.Errorf("active_sessions: want 0, got %d", body.ActiveSessions)
	}
}

// TestStatsEdgeMaxInt32ConnectionsReported verifies that the handler correctly
// reports a very large (near int32 max) connection count without overflow or
// truncation.
func TestStatsEdgeMaxInt32ConnectionsReported(t *testing.T) {
	t.Parallel()
	sessions := services.NewSessionService()
	// Use math.MaxInt32 / 2 to stay safely within int32 range.
	const bigCount = int32(math.MaxInt32 / 2)
	ws := &stubWSCounter{n: bigCount}
	app := newStatsEdgeApp(t, sessions, ws, nil, "1.0.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ActiveWSConnections != bigCount {
		t.Errorf("active_ws_connections: want %d, got %d", bigCount, body.ActiveWSConnections)
	}
}

// TestStatsEdgeResponseJSONStructureValidation verifies that every field
// required by the stats contract is present in the JSON response and has the
// correct type.
func TestStatsEdgeResponseJSONStructureValidation(t *testing.T) {
	t.Parallel()
	sessions := services.NewSessionService()
	ws := &stubWSCounter{n: 2}
	rag := &stubRAGLoader{loaded: true, count: 42}
	app := newStatsEdgeApp(t, sessions, ws, rag, "0.9.0", time.Now().Add(-30*time.Second))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}

	// Verify all required top-level keys exist.
	required := []string{
		"active_sessions",
		"active_ws_connections",
		"total_sessions_started",
		"kb_loaded",
		"kb_entry_count",
		"uptime_seconds",
		"version",
		"total_messages",
	}
	for _, key := range required {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing required field %q in stats response", key)
		}
	}

	// Spot-check types.
	if _, ok := raw["active_sessions"].(float64); !ok {
		t.Errorf("active_sessions: expected numeric, got %T", raw["active_sessions"])
	}
	if _, ok := raw["kb_loaded"].(bool); !ok {
		t.Errorf("kb_loaded: expected bool, got %T", raw["kb_loaded"])
	}
	if _, ok := raw["version"].(string); !ok {
		t.Errorf("version: expected string, got %T", raw["version"])
	}
}

// TestStatsEdgeVersionFieldReflectsInjectedValue verifies that the version
// string passed to NewStatsHandlerFull is returned verbatim in the response.
func TestStatsEdgeVersionFieldReflectsInjectedValue(t *testing.T) {
	t.Parallel()
	sessions := services.NewSessionService()
	const version = "3.14.159"
	app := newStatsEdgeApp(t, sessions, nil, nil, version, time.Now())

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Version != version {
		t.Errorf("version: want %q, got %q", version, body.Version)
	}
}

// TestStatsEdgeContentTypeIsJSON verifies that the response has application/json
// Content-Type.
func TestStatsEdgeContentTypeIsJSON(t *testing.T) {
	t.Parallel()
	sessions := services.NewSessionService()
	h := NewStatsHandler(sessions)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/stats", h.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}
}
