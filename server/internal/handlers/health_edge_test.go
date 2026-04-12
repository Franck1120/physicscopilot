package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// errorDBPinger always returns an error from Ping, simulating an unavailable DB.
type errorDBPinger struct{ msg string }

func (e *errorDBPinger) Ping(_ context.Context) error {
	return errors.New(e.msg)
}

// newHealthEdgeApp builds a Fiber app with a HealthHandler wired to the given
// DB pinger. It mirrors the helper in health_handler_test.go but accepts a
// custom DBPinger to exercise error paths.
func newHealthEdgeApp(t *testing.T, version string, startTime time.Time, db DBPinger) *fiber.App {
	t.Helper()
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := NewWSHandler(convSvc, sessionSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/health", NewHealthHandler(version, startTime, ws, db))
	return app
}

// TestHealthEdgeDBPingErrorReturnsUnavailableStatus verifies that when the
// injected DB pinger returns an error the health response still comes back
// with HTTP 200 but the "db_status" field is "unavailable".
func TestHealthEdgeDBPingErrorReturnsUnavailableStatus(t *testing.T) {
	t.Parallel()
	app := newHealthEdgeApp(t, "1.0.0", time.Now(), &errorDBPinger{msg: "connection refused"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	// Overall HTTP status must still be 200 — the endpoint is the health check
	// itself; an unavailable DB does not mean the server should return 5xx.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200 even with DB error, got %d", resp.StatusCode)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBStatus != "unavailable" {
		t.Errorf("db_status: want %q, got %q", "unavailable", body.DBStatus)
	}
}

// TestHealthEdgeResponseTimeIsReasonable verifies that the /health endpoint
// responds within 100 ms even with a nil DB (no network I/O).
func TestHealthEdgeResponseTimeIsReasonable(t *testing.T) {
	t.Parallel()
	app := newHealthEdgeApp(t, "1.0.0", time.Now(), nil)

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	const maxAllowed = 100 * time.Millisecond
	if elapsed > maxAllowed {
		t.Errorf("response time %v exceeds 100 ms limit", elapsed)
	}
}

// TestHealthEdgeVersionFieldMatchesExpectedFormat verifies that the "version"
// field in the health response is a non-empty string. When a semver-like string
// is passed, it is returned verbatim (no truncation or transformation).
func TestHealthEdgeVersionFieldMatchesExpectedFormat(t *testing.T) {
	t.Parallel()
	const version = "2.3.4"
	app := newHealthEdgeApp(t, version, time.Now(), nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Version == "" {
		t.Error("version field must not be empty")
	}
	if body.Version != version {
		t.Errorf("version: want %q, got %q", version, body.Version)
	}

	// Verify the value looks like a semantic version (X.Y.Z) as an extra guard.
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+`)
	if !semverPattern.MatchString(body.Version) {
		t.Errorf("version %q does not look like a semantic version (X.Y.Z)", body.Version)
	}
}

// TestHealthEdgeMultipleDBErrors verifies that the health handler correctly
// reports "unavailable" on every call, not just the first, when the DB pinger
// consistently fails.
func TestHealthEdgeMultipleDBErrors(t *testing.T) {
	t.Parallel()
	app := newHealthEdgeApp(t, "1.0.0", time.Now(), &errorDBPinger{msg: "timeout"})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		var body HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode %d: %v", i, err)
		}
		if body.DBStatus != "unavailable" {
			t.Errorf("request %d db_status: want 'unavailable', got %q", i, body.DBStatus)
		}
	}
}
