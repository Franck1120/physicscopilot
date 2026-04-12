// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

// Package middleware provides reusable Fiber middleware components for the
// PhysicsCopilot server: structured request logging, per-IP and per-user
// rate limiting, HTTP Basic Auth for the metrics endpoint, and user-level
// session and frame limits.
package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const metricsUser = "admin"

// MetricsBasicAuth returns a Fiber middleware that enforces HTTP Basic
// authentication on the /metrics endpoint.
//
// The password is read exclusively from the METRICS_PASSWORD environment
// variable. If the variable is not set, the endpoint is disabled and every
// request returns 503 Service Unavailable — no hardcoded fallback is used
// so that misconfigured deployments fail loudly rather than silently exposing
// metrics behind a known-weak password.
//
// Uses constant-time comparison to prevent timing attacks.
func MetricsBasicAuth() fiber.Handler {
	password := os.Getenv("METRICS_PASSWORD")
	disabled := password == ""

	return func(c *fiber.Ctx) error {
		if disabled {
			return fiber.NewError(
				fiber.StatusServiceUnavailable,
				"metrics endpoint is disabled: set METRICS_PASSWORD to enable it",
			)
		}

		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			c.Set("WWW-Authenticate", `Basic realm="PhysicsCopilot Metrics"`)
			return fiber.NewError(fiber.StatusUnauthorized, "metrics endpoint requires authentication")
		}

		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header")
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header")
		}

		userMatch := subtle.ConstantTimeCompare([]byte(parts[0]), []byte(metricsUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(parts[1]), []byte(password)) == 1

		if !userMatch || !passMatch {
			c.Set("WWW-Authenticate", `Basic realm="PhysicsCopilot Metrics"`)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}

		return c.Next()
	}
}
