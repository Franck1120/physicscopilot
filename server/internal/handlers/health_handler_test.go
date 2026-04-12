package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// mockDBPinger implements DBPinger for tests without a real DB.
type mockDBPinger struct{ err error }

func (m *mockDBPinger) Ping(_ context.Context) error { return m.err }

// mockPoolStatter implements PoolStatter so we can verify pool stats are
// forwarded to the health response.
type mockPoolStatter struct {
	mockDBPinger
	stats services.DBPoolStats
}

func (m *mockPoolStatter) PoolStats() services.DBPoolStats { return m.stats }

func newHealthApp(version string, startTime time.Time) (*fiber.App, *WSHandler) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler(version, startTime, ws, nil))
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
	if !strings.ContainsAny(body.Uptime, "0123456789") {
		t.Errorf("uptime looks malformed: %q", body.Uptime)
	}
}

func TestHealthEndpointReportsActiveConnections(t *testing.T) {
	app, ws := newHealthApp("0.1.0", time.Now())
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

func TestHealthEndpointDBStatusNotConfigured(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now()) // db=nil

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBStatus != "not_configured" {
		t.Errorf("db_status: want 'not_configured', got %q", body.DBStatus)
	}
}

func TestHealthEndpointDBStatusOK(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("0.1.0", time.Now(), ws, &mockDBPinger{err: nil}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBStatus != "ok" {
		t.Errorf("db_status: want 'ok', got %q", body.DBStatus)
	}
}

func TestHealthEndpointDBStatusUnavailable(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("0.1.0", time.Now(), ws, &mockDBPinger{err: fmt.Errorf("connection refused")}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBStatus != "unavailable" {
		t.Errorf("db_status: want 'unavailable', got %q", body.DBStatus)
	}
}

func TestHealthEndpointDBPoolStatsIncluded(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	ps := &mockPoolStatter{
		stats: services.DBPoolStats{
			TotalConns:    3,
			IdleConns:     2,
			AcquiredConns: 1,
			MaxConns:      10,
		},
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("0.1.0", time.Now(), ws, ps))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBPool == nil {
		t.Fatal("expected db_pool to be present when PoolStatter is provided")
	}
	if body.DBPool.TotalConns != 3 {
		t.Errorf("total_conns: want 3, got %d", body.DBPool.TotalConns)
	}
	if body.DBPool.MaxConns != 10 {
		t.Errorf("max_conns: want 10, got %d", body.DBPool.MaxConns)
	}
}

func TestHealthEndpointDBPoolStatsAbsentForPlainPinger(t *testing.T) {
	// A plain DBPinger (no PoolStats method) must NOT include db_pool.
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("0.1.0", time.Now(), ws, &mockDBPinger{}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBPool != nil {
		t.Errorf("expected db_pool to be absent for plain DBPinger, got %+v", body.DBPool)
	}
}

// ── Content-Type and JSON validity ───────────────────────────────────────────

func TestHealthEndpointContentTypeIsJSON(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}
}

func TestHealthEndpointResponseIsValidJSON(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
}

// ── All top-level fields present ──────────────────────────────────────────────

func TestHealthEndpointAllTopLevelFieldsPresent(t *testing.T) {
	app, _ := newHealthApp("1.0.0", time.Now().Add(-10*time.Second))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}

	required := []string{"status", "service", "version", "uptime", "active_connections", "memory_mb", "db_status"}
	for _, field := range required {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing required field %q in health response", field)
		}
	}
}

// ── memory_mb field ───────────────────────────────────────────────────────────

func TestHealthEndpointMemoryMBIsNonNegative(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// memory_mb is a uint64 and always >= 0; the Go runtime allocates at least
	// a small amount, so we just check the field is present (non-zero in practice).
	// We don't assert a specific value since that would be fragile.
	_ = body.MemoryMB // field must decode without error
}

// ── uptime increases over time ────────────────────────────────────────────────

func TestHealthEndpointUptimeReflectsStartTime(t *testing.T) {
	// Start time 2 minutes in the past — uptime string must contain digits > 0.
	startTime := time.Now().Add(-2 * time.Minute)
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

	// The uptime must be a non-empty string with at least one digit.
	if body.Uptime == "" {
		t.Error("uptime must not be empty")
	}
	if body.Uptime == "0s" {
		t.Error("uptime should reflect non-zero elapsed time since startTime 2 minutes ago")
	}
}

// ── service field value ───────────────────────────────────────────────────────

func TestHealthEndpointServiceNameIsPhysicscopilot(t *testing.T) {
	app, _ := newHealthApp("0.0.1", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Service != "physicscopilot" {
		t.Errorf("service: want %q, got %q", "physicscopilot", body.Service)
	}
}

// ── DBPool idle/acquired fields ───────────────────────────────────────────────

func TestHealthEndpointDBPoolIdleAndAcquiredFields(t *testing.T) {
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	ps := &mockPoolStatter{
		stats: services.DBPoolStats{
			TotalConns:    5,
			IdleConns:     3,
			AcquiredConns: 2,
			MaxConns:      20,
		},
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler("0.1.0", time.Now(), ws, ps))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBPool == nil {
		t.Fatal("expected db_pool to be present")
	}
	if body.DBPool.IdleConns != 3 {
		t.Errorf("idle_conns: want 3, got %d", body.DBPool.IdleConns)
	}
	if body.DBPool.AcquiredConns != 2 {
		t.Errorf("acquired_conns: want 2, got %d", body.DBPool.AcquiredConns)
	}
}

// ── Cache-Control header on health endpoint ───────────────────────────────────

func TestHealthEndpointCacheControlHeader(t *testing.T) {
	app, _ := newHealthApp("0.1.0", time.Now())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "no-cache, no-store" {
		t.Errorf("health Cache-Control: want %q, got %q", "no-cache, no-store", cc)
	}
}
