package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Named import so TestCustomMetricsRegistered can pre-seed Vec metrics.
	"github.com/Franck1120/physicscopilot/server/internal/metrics"
)

func metricsApp() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	return app
}

func TestMetricsEndpointReturns200(t *testing.T) {
	app := metricsApp()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestMetricsEndpointPrometheusTextFormat(t *testing.T) {
	app := metricsApp()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	s := string(body)

	// Prometheus text format always contains HELP and TYPE comment lines.
	if !strings.Contains(s, "# HELP") {
		t.Error("expected '# HELP' lines in Prometheus output")
	}
	if !strings.Contains(s, "# TYPE") {
		t.Error("expected '# TYPE' lines in Prometheus output")
	}
}

func TestCustomMetricsRegistered(t *testing.T) {
	// Vec metrics (CounterVec, HistogramVec) only emit output — including
	// # HELP lines — after at least one label combination has been observed.
	// Pre-seed each Vec metric with a zero-value observation so the test
	// exercises the real registration without polluting production counters.
	metrics.HttpRequestsTotal.WithLabelValues("GET", "/_test", "200").Add(0)
	metrics.HttpRequestDuration.WithLabelValues("GET", "/_test").Observe(0)
	metrics.WsMessagesTotal.WithLabelValues("_test").Add(0)
	metrics.DBQueryDuration.WithLabelValues("_test").Observe(0)

	app := metricsApp()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	s := string(body)

	// # HELP lines are always emitted once a metric family has any observation.
	want := []string{
		"# HELP http_requests_total",
		"# HELP http_request_duration_seconds",
		"# HELP ws_active_connections",
		"# HELP ws_messages_total",
		"# HELP ai_inference_duration_seconds",
		"# HELP db_query_duration_seconds",
	}
	for _, name := range want {
		if !strings.Contains(s, name) {
			t.Errorf("metric HELP line %q not found in /metrics output", name)
		}
	}
}
