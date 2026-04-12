package metrics_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Franck1120/physicscopilot/server/internal/metrics"
)

// scrapeMetrics returns the full Prometheus text output from a fresh Fiber app.
func scrapeMetrics(t *testing.T) string {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("metrics scrape: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body)
}

// countOccurrences counts how many times substr appears in s.
func countOccurrences(s, substr string) int {
	return strings.Count(s, substr)
}

func TestTrackError_IncrementsCounter(t *testing.T) {
	sentinel := errors.New("test error")

	// Observe baseline before the call.
	before := scrapeMetrics(t)
	baselineAI := countOccurrences(before, `app_errors_total{category="AI_ERROR"}`)

	metrics.TrackError(metrics.CategoryAI, sentinel, "session_id", "s1", "user_id", "u1")

	after := scrapeMetrics(t)
	afterAI := countOccurrences(after, `app_errors_total{category="AI_ERROR"}`)

	// The counter line must appear exactly once after the first observation.
	if afterAI < 1 {
		t.Errorf("expected app_errors_total{category=\"AI_ERROR\"} in metrics output, got:\n%s", after)
	}
	// Counter must have grown (or appeared for the first time).
	if afterAI < baselineAI && baselineAI > 0 {
		t.Errorf("counter did not increase: before=%d after=%d", baselineAI, afterAI)
	}
}

func TestTrackError_AllCategories(t *testing.T) {
	cases := []struct {
		category metrics.ErrorCategory
		label    string
	}{
		{metrics.CategoryAI, `app_errors_total{category="AI_ERROR"}`},
		{metrics.CategoryDB, `app_errors_total{category="DB_ERROR"}`},
		{metrics.CategoryAuth, `app_errors_total{category="AUTH_ERROR"}`},
		{metrics.CategoryWS, `app_errors_total{category="WS_ERROR"}`},
	}

	for _, tc := range cases {
		metrics.TrackError(tc.category, errors.New("probe"), "test", "true")
	}

	output := scrapeMetrics(t)
	for _, tc := range cases {
		if !strings.Contains(output, tc.label) {
			t.Errorf("expected label %q in /metrics output", tc.label)
		}
	}
}

func TestTrackError_HelpLineRegistered(t *testing.T) {
	// Pre-seed so the HELP line is emitted.
	metrics.TrackError(metrics.CategoryDB, errors.New("seed"), "test", "true")

	output := scrapeMetrics(t)
	if !strings.Contains(output, "# HELP app_errors_total") {
		t.Error("expected '# HELP app_errors_total' in /metrics output")
	}
	if !strings.Contains(output, "# TYPE app_errors_total counter") {
		t.Error("expected '# TYPE app_errors_total counter' in /metrics output")
	}
}

func TestTrackError_NilError(t *testing.T) {
	// TrackError must not panic when err is nil (e.g. defensive call sites).
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TrackError panicked with nil error: %v", r)
		}
	}()
	metrics.TrackError(metrics.CategoryWS, nil)
}

func TestTrackError_ExtraAttrsForwarded(t *testing.T) {
	// TrackError must not panic with an odd number of extra attrs.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TrackError panicked with odd attrs: %v", r)
		}
	}()
	metrics.TrackError(metrics.CategoryAuth, errors.New("odd"), "key_without_value")
}
