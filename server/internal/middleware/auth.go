package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const metricsUser = "admin"
const defaultMetricsPassword = "metrics-secret"

// MetricsBasicAuth returns a Fiber middleware that enforces HTTP Basic
// authentication on the /metrics endpoint.
//
// Credentials: username "admin", password from METRICS_PASSWORD env var
// (falls back to "metrics-secret" when the var is not set).
//
// Uses constant-time comparison to prevent timing attacks.
func MetricsBasicAuth() fiber.Handler {
	password := os.Getenv("METRICS_PASSWORD")
	if password == "" {
		password = defaultMetricsPassword
	}

	return func(c *fiber.Ctx) error {
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
