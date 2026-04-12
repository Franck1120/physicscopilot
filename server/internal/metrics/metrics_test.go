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

// TestAllMetricFamiliesPresent checks that every named metric family defined in
// metrics.go has a corresponding HELP line in the Prometheus output after
// being seeded with at least one observation.
func TestAllMetricFamiliesPresent(t *testing.T) {
	// Seed all Vec metrics so their HELP lines are emitted.
	metrics.HttpRequestsTotal.WithLabelValues("GET", "/_seed", "200").Add(0)
	metrics.HttpRequestDuration.WithLabelValues("GET", "/_seed").Observe(0)
	metrics.WsMessagesTotal.WithLabelValues("_seed").Add(0)
	metrics.GeminiErrorsTotal.WithLabelValues("frame").Add(0)
	metrics.WsActiveSessionsByLanguage.WithLabelValues("it").Set(0)
	metrics.DBQueryDuration.WithLabelValues("_seed").Observe(0)

	app := metricsApp()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	s := string(body)

	families := []string{
		"# HELP http_requests_total",
		"# HELP http_request_duration_seconds",
		"# HELP ws_active_connections",
		"# HELP ws_messages_total",
		"# HELP ai_inference_duration_seconds",
		"# HELP gemini_errors_total",
		"# HELP ws_frames_processed_total",
		"# HELP ws_active_sessions_by_language",
		"# HELP session_started_total",
		"# HELP session_completed_total",
		"# HELP session_abandoned_total",
		"# HELP session_time_to_first_response_seconds",
		"# HELP sessions_expired_total",
		"# HELP process_heap_alloc_bytes",
		"# HELP process_sys_bytes",
		"# HELP process_num_gc_total",
		"# HELP rag_cache_hits_total",
		"# HELP rag_cache_misses_total",
		"# HELP db_query_duration_seconds",
	}
	for _, f := range families {
		if !strings.Contains(s, f) {
			t.Errorf("metric family HELP %q missing from /metrics output", f)
		}
	}
}

// TestVecMetricLabelCardinality ensures that Vec metrics expose the correct
// label names in the TYPE line and that label values produce distinct series.
func TestVecMetricLabelCardinality(t *testing.T) {
	// HttpRequestsTotal has labels: method, path, status
	metrics.HttpRequestsTotal.WithLabelValues("POST", "/api/seed-cardinality", "404").Add(0)
	// WsMessagesTotal has label: type
	metrics.WsMessagesTotal.WithLabelValues("frame").Add(0)
	metrics.WsMessagesTotal.WithLabelValues("text").Add(0)
	// GeminiErrorsTotal has label: type
	metrics.GeminiErrorsTotal.WithLabelValues("text").Add(0)
	// DBQueryDuration has label: operation
	metrics.DBQueryDuration.WithLabelValues("list_sessions").Observe(0.001)

	app := metricsApp()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	s := string(body)

	// Each seeded label combination must produce a distinct metric line.
	wantLines := []string{
		`ws_messages_total{type="frame"}`,
		`ws_messages_total{type="text"}`,
		`gemini_errors_total{type="frame"}`,
		`gemini_errors_total{type="text"}`,
		`db_query_duration_seconds_count{operation="list_sessions"}`,
	}
	for _, line := range wantLines {
		if !strings.Contains(s, line) {
			t.Errorf("expected metric line containing %q in /metrics output", line)
		}
	}
}

// TestCounterIncrementNoPanic verifies that incrementing a freshly-seeded
// counter via the package-level variable does not panic.
func TestCounterIncrementNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("counter increment panicked: %v", r)
		}
	}()

	metrics.WsFramesProcessedTotal.Inc()
	metrics.SessionStartedTotal.Inc()
	metrics.SessionCompletedTotal.Inc()
	metrics.SessionAbandonedTotal.Inc()
	metrics.RagCacheHitsTotal.Inc()
	metrics.RagCacheMissesTotal.Inc()
	metrics.SessionsExpiredTotal.Inc()
}

// TestGaugeSetAndObserveNoPanic verifies that setting gauges and recording
// histogram observations does not panic.
func TestGaugeSetAndObserveNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("gauge/histogram operation panicked: %v", r)
		}
	}()

	metrics.WsActiveConnections.Set(5)
	metrics.WsActiveConnections.Inc()
	metrics.WsActiveConnections.Dec()
	metrics.WsActiveSessionsByLanguage.WithLabelValues("en").Set(3)
	metrics.MemHeapAllocBytes.Set(1024 * 1024)
	metrics.MemSysBytes.Set(2 * 1024 * 1024)
	metrics.MemNumGCTotal.Set(42)

	metrics.AiInferenceDuration.Observe(0.5)
	metrics.TimeToFirstResponseSeconds.Observe(1.2)
	metrics.HttpRequestDuration.WithLabelValues("GET", "/health").Observe(0.003)
	metrics.DBQueryDuration.WithLabelValues("save_session").Observe(0.01)
}
