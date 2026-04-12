// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// SessionsActive is a gauge of currently active sessions in memory.
	SessionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sessions_active_total",
		Help: "Number of currently active sessions in memory.",
	})

	// FeedbackTotal counts feedback submissions by rating label (positive/negative).
	FeedbackTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "feedback_submissions_total",
		Help: "Total feedback submissions by rating.",
	}, []string{"rating"})

	// SessionCreatedTotal counts the total number of sessions ever created.
	SessionCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sessions_created_total",
		Help: "Total number of sessions ever created.",
	})
)
