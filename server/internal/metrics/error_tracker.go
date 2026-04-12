package metrics

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ErrorCategory classifies application errors for structured tracking and
// Prometheus aggregation. Keep categories coarse-grained; fine-grained detail
// goes in the structured log fields, not in label cardinality.
type ErrorCategory string

const (
	// CategoryAI covers Gemini inference failures and AI-pipeline errors.
	CategoryAI ErrorCategory = "AI_ERROR"
	// CategoryDB covers Postgres / Supabase query and connectivity failures.
	CategoryDB ErrorCategory = "DB_ERROR"
	// CategoryAuth covers JWT validation, missing credentials, and auth middleware failures.
	CategoryAuth ErrorCategory = "AUTH_ERROR"
	// CategoryWS covers WebSocket protocol, connection-limit, and I/O errors.
	CategoryWS ErrorCategory = "WS_ERROR"
)

// appErrorsTotal counts every tracked error grouped by its high-level category.
// High-cardinality fields (session_id, user_id) belong in the log, not here.
var appErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "app_errors_total",
		Help: "Total application errors grouped by category (AI_ERROR, DB_ERROR, AUTH_ERROR, WS_ERROR).",
	},
	[]string{"category"},
)

// TrackError increments the app_errors_total Prometheus counter for the given
// category and emits a structured log line at ERROR level via slog.
//
// Use TrackError instead of plain slog.Error whenever an error should be
// visible in Prometheus dashboards or alerts — for example, failed Gemini
// calls, database timeouts, auth rejections, and WebSocket I/O errors.
// Use plain slog.Error (without TrackError) for informational or expected
// conditions (e.g. a client disconnecting cleanly) that do not warrant a
// counter increment.
//
// category must be one of the predefined [ErrorCategory] constants:
//   - [CategoryAI]   — Gemini inference and AI-pipeline failures
//   - [CategoryDB]   — Postgres / Supabase query and connectivity failures
//   - [CategoryAuth] — JWT validation, missing credentials, auth middleware
//   - [CategoryWS]   — WebSocket protocol, connection-limit, and I/O errors
//
// logAttrs are forwarded verbatim to slog as additional key-value pairs.
// High-cardinality identifiers (session_id, user_id) belong here, not as
// Prometheus label values, to keep label cardinality bounded.
//
// Example:
//
//	metrics.TrackError(metrics.CategoryAI, err, "session_id", sessionID, "msg_type", "frame")
func TrackError(category ErrorCategory, err error, logAttrs ...any) {
	appErrorsTotal.WithLabelValues(string(category)).Inc()
	args := make([]any, 0, 4+len(logAttrs))
	args = append(args, "error_category", string(category), "err", err)
	args = append(args, logAttrs...)
	slog.Error("application error", args...)
}
