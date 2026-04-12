// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

// requestIDKey is the Locals key used to pass the correlation ID down the chain.
const requestIDKey = "request_id"

// ctxRequestIDKey is the context key type used to store the request ID in a
// context.Context so that handler-level slog calls can retrieve it via
// WithRequestID / RequestIDFromCtx.
type ctxRequestIDKeyType struct{}

var ctxKey = ctxRequestIDKeyType{}

// WithRequestID returns a new context carrying the given request ID.
// Handlers can retrieve it with RequestIDFromCtx to annotate slog calls with
// the same correlation ID that is logged by StructuredLogger.
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ctxKey, reqID)
}

// RequestIDFromCtx retrieves the request ID stored by WithRequestID.
// Returns an empty string when the context carries no request ID.
func RequestIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey).(string)
	return id
}

// StructuredLogger returns a Fiber middleware that:
//   - generates a random 8-byte hex request ID per request
//   - stores it in c.Locals(requestIDKey) for downstream handlers
//   - propagates the ID into the request's user context via WithRequestID
//   - logs method, path, status code, and latency via the global slog logger
//
// Log level escalates automatically: INFO for 2xx/3xx, WARN for 4xx, ERROR for 5xx.
// The output format (JSON or text) is determined by the global slog handler
// configured in logger.Init() — JSON in production, text in development.
// IP addresses are anonymized (SHA-256, first 4 bytes) before logging.
func StructuredLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		// Reuse the ID injected by RequestIDMiddleware when present; otherwise
		// generate one here so StructuredLogger remains self-contained when used
		// without the RequestIDMiddleware in front of it.
		reqID, _ := c.Locals(requestIDKey).(string)
		if reqID == "" {
			reqID = generateRequestID()
			c.Locals(requestIDKey, reqID)
		}
		// Propagate the request ID into the user context so handler-level slog
		// calls can attach the same correlation ID without accessing c.Locals.
		c.SetUserContext(WithRequestID(c.UserContext(), reqID))

		err := c.Next()

		status := c.Response().StatusCode()
		latency := time.Since(start)

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		slog.Log(context.Background(), level, "request",
			"request_id", reqID,
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip_hash", anonymizeIP(c.IP()),
		)

		return err
	}
}

// RequestID retrieves the correlation ID injected by StructuredLogger.
// Returns an empty string when the middleware has not been applied.
func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals(requestIDKey).(string)
	return id
}

// generateRequestID returns a random 8-byte hex string (16 hex chars).
// Falls back to a static string on the rare crypto/rand failure.
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "00000000deadbeef"
	}
	return hex.EncodeToString(b)
}

// anonymizeIP hashes the raw IP with SHA-256 and returns the first 4 bytes
// as an 8-character hex string. This is sufficient for log correlation while
// avoiding storage of personally identifiable network addresses.
func anonymizeIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:4])
}
