// Package metrics defines and registers all Prometheus metrics for the
// PhysicsCopilot server. Import this package for its side-effects to ensure
// metrics are registered with the default registry before any handler runs.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal counts every HTTP request served, labelled by method,
	// path, and HTTP status code.
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// HttpRequestDuration tracks the latency distribution of HTTP requests,
	// labelled by method and path (status excluded to keep cardinality low).
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distribution.",
			Buckets: prometheus.DefBuckets, // .005 .01 .025 .05 .1 .25 .5 1 2.5 5 10
		},
		[]string{"method", "path"},
	)

	// WsActiveConnections is a gauge of currently open WebSocket connections.
	WsActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ws_active_connections",
			Help: "Number of currently active WebSocket connections.",
		},
	)

	// WsMessagesTotal counts incoming WebSocket messages by application type
	// ("frame", "text", "ping", "unknown").
	WsMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ws_messages_total",
			Help: "Total WebSocket messages received, by message type.",
		},
		[]string{"type"},
	)

	// AiInferenceDuration tracks how long AI inference calls take (seconds).
	// Buckets are tuned for Gemini latencies (100 ms – 10 s range).
	AiInferenceDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ai_inference_duration_seconds",
			Help:    "Time spent waiting for AI inference (Gemini) to respond.",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 3.0, 5.0, 10.0},
		},
	)

	// GeminiErrorsTotal counts failed Gemini inference calls, labelled by
	// the message type that triggered the call ("frame" or "text").
	GeminiErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gemini_errors_total",
			Help: "Total number of Gemini AI inference errors, by message type.",
		},
		[]string{"type"},
	)

	// WsFramesProcessedTotal counts camera frames that passed the rate limiter
	// and were successfully forwarded to Gemini for analysis.
	WsFramesProcessedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ws_frames_processed_total",
			Help: "Camera frames forwarded to Gemini (after rate limiting, successful only).",
		},
	)

	// WsActiveSessionsByLanguage is a per-language gauge of open WebSocket
	// sessions. The language label holds the BCP-47 code (e.g. "it", "en").
	WsActiveSessionsByLanguage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ws_active_sessions_by_language",
			Help: "Number of active WebSocket sessions grouped by user language.",
		},
		[]string{"language"},
	)

	// RagCacheHitsTotal counts KB query cache hits (identical query served from in-memory LRU).
	RagCacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rag_cache_hits_total",
			Help: "Total number of RAG KB query cache hits.",
		},
	)

	// RagCacheMissesTotal counts KB query cache misses (query not found in LRU, forwarded to vector store).
	RagCacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rag_cache_misses_total",
			Help: "Total number of RAG KB query cache misses.",
		},
	)
)
