// Copyright (c) 2026 PhysicsCopilot contributors. All rights reserved.
// SPDX-License-Identifier: MIT

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

	// ── Funnel analytics ─────────────────────────────────────────────────────

	// SessionStartedTotal counts every WebSocket session that successfully
	// completed creation (session row written, first message loop entered).
	SessionStartedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "session_started_total",
			Help: "Total WebSocket sessions started.",
		},
	)

	// SessionCompletedTotal counts sessions that ended with a clean close
	// (CloseNormalClosure or CloseGoingAway from the server side on shutdown).
	SessionCompletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "session_completed_total",
			Help: "Total WebSocket sessions ended with a clean close.",
		},
	)

	// SessionAbandonedTotal counts sessions that ended unexpectedly (network
	// drop, unexpected close code, or read error after at least one message).
	SessionAbandonedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "session_abandoned_total",
			Help: "Total WebSocket sessions that ended unexpectedly (network drop or error).",
		},
	)

	// TimeToFirstResponseSeconds measures the wall-clock time from session
	// creation to the first AI response sent to the client (TTFR).
	// Buckets cover 100 ms – 15 s, which spans typical Gemini cold/warm paths.
	TimeToFirstResponseSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "session_time_to_first_response_seconds",
			Help:    "Time from session start to first AI response (TTFR).",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 3.0, 5.0, 10.0, 15.0},
		},
	)

	// SessionsExpiredTotal counts in-memory sessions removed by the
	// background cleanup goroutine because they exceeded the idle timeout.
	SessionsExpiredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sessions_expired_total",
			Help: "Total number of sessions removed by the expiry cleanup goroutine.",
		},
	)

	// MemHeapAllocBytes tracks the current bytes allocated on the Go heap.
	MemHeapAllocBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_heap_alloc_bytes",
			Help: "Heap memory currently allocated (live objects) in bytes.",
		},
	)

	// MemSysBytes tracks the total memory obtained from the OS.
	MemSysBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_sys_bytes",
			Help: "Total memory obtained from the OS in bytes.",
		},
	)

	// MemNumGCTotal tracks the cumulative number of completed GC cycles.
	MemNumGCTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_num_gc_total",
			Help: "Cumulative number of completed garbage collection cycles.",
		},
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

	// DBQueryDuration tracks the latency distribution of Postgres queries,
	// labelled by operation name (e.g. "save_session", "list_sessions").
	// Buckets are tuned for typical Postgres latencies (1 ms - 1 s).
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Latency distribution of Postgres DB queries by operation.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"operation"},
	)
)
