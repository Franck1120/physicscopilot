// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

// Package sentry provides a thin wrapper around the Sentry Go SDK.
// It initialises Sentry from SENTRY_DSN and exposes helpers used by the
// server's error handler.
package sentry

import (
	"log/slog"
	"os"

	gosentry "github.com/getsentry/sentry-go"
)

// Init initialises the Sentry client. If SENTRY_DSN is empty, Sentry is
// disabled and all captures are no-ops. Safe to call multiple times.
func Init(release string) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		slog.Info("SENTRY_DSN not set — error tracking disabled")
		return
	}
	if err := gosentry.Init(gosentry.ClientOptions{
		Dsn:              dsn,
		Release:          release,
		EnableTracing:    false,
		AttachStacktrace: true,
	}); err != nil {
		slog.Warn("sentry init failed", "err", err)
	} else {
		slog.Info("sentry initialised", "release", release)
	}
}

// CaptureError sends err to Sentry. Safe to call when Sentry is disabled.
func CaptureError(err error) {
	if err == nil {
		return
	}
	gosentry.CaptureException(err)
}
