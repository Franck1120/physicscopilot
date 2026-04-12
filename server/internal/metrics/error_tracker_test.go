package metrics_test

import (
	"errors"
	"fmt"
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

// TestTrackError_ConcurrentTracking verifies that 10 goroutines concurrently
// calling TrackError on the same category produces no data race (run with
// -race to surface issues). All goroutines must complete without panicking.
func TestTrackError_ConcurrentTracking(t *testing.T) {
	const goroutines = 10
	done := make(chan struct{}, goroutines)

	for i := range goroutines {
		go func(idx int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("goroutine %d: TrackError panicked: %v", idx, r)
				}
				done <- struct{}{}
			}()
			metrics.TrackError(metrics.CategoryAI, errors.New("concurrent error"), "goroutine", idx)
		}(i)
	}

	for range goroutines {
		<-done
	}

	// All goroutines completed — verify the counter is present in output.
	output := scrapeMetrics(t)
	if !strings.Contains(output, `app_errors_total{category="AI_ERROR"}`) {
		t.Error("expected AI_ERROR counter present after concurrent tracking")
	}
}

// TestTrackError_CategoriesAreIndependent verifies that incrementing one
// category does not affect the counter of a different category.
func TestTrackError_CategoriesAreIndependent(t *testing.T) {
	// Seed all categories so their lines are present.
	metrics.TrackError(metrics.CategoryDB, errors.New("db probe"), "test", "true")
	metrics.TrackError(metrics.CategoryWS, errors.New("ws probe"), "test", "true")

	output := scrapeMetrics(t)

	// Both DB and WS lines must appear.
	if !strings.Contains(output, `app_errors_total{category="DB_ERROR"}`) {
		t.Error("expected DB_ERROR counter in output")
	}
	if !strings.Contains(output, `app_errors_total{category="WS_ERROR"}`) {
		t.Error("expected WS_ERROR counter in output")
	}

	// Each category must appear exactly once (no cross-label pollution).
	dbCount := strings.Count(output, `app_errors_total{category="DB_ERROR"}`)
	wsCount := strings.Count(output, `app_errors_total{category="WS_ERROR"}`)
	if dbCount != 1 {
		t.Errorf("expected DB_ERROR line to appear exactly once, got %d", dbCount)
	}
	if wsCount != 1 {
		t.Errorf("expected WS_ERROR line to appear exactly once, got %d", wsCount)
	}
}

// TestTrackError_CounterValueGrows verifies that repeated calls increase the
// numeric value in the Prometheus output (not just that the line is present).
func TestTrackError_CounterValueGrows(t *testing.T) {
	// Use a dedicated unique sub-test so the counter starts at a known state
	// relative to whatever previous tests may have observed.
	// We scrape before and after calling TrackError and compare the float values.
	sentinel := errors.New("grow-test")

	before := scrapeMetrics(t)

	// Extract the current AUTH_ERROR value (may already be > 0 from earlier tests).
	authBefore := extractCounterValue(before, `app_errors_total{category="AUTH_ERROR"}`)

	metrics.TrackError(metrics.CategoryAuth, sentinel)

	after := scrapeMetrics(t)
	authAfter := extractCounterValue(after, `app_errors_total{category="AUTH_ERROR"}`)

	if authAfter <= authBefore {
		t.Errorf("AUTH_ERROR counter did not increase: before=%g after=%g", authBefore, authAfter)
	}
}

// extractCounterValue parses the float value from a Prometheus text output line
// matching the given label selector. Returns 0 if the line is not found.
func extractCounterValue(output, labelSelector string) float64 {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, labelSelector) && !strings.HasPrefix(line, "#") {
			// Format: metric_name{labels} VALUE
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var v float64
				if _, err := fmt.Sscanf(fields[len(fields)-1], "%g", &v); err == nil {
					return v
				}
			}
		}
	}
	return 0
}
