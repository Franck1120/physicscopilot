// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

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

// TrackError increments the app_errors_total counter for category and emits a
// structured log line at ERROR level. Callers pass extra key-value pairs (e.g.
// "session_id", id, "user_id", uid) that are forwarded to slog verbatim.
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
