// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDHeader is the canonical HTTP header name for request correlation.
const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware injects a unique request ID into every response and
// stores it in the Fiber context locals under key "request_id".
// If the client sends X-Request-ID, that value is reused (allows
// end-to-end correlation); otherwise a new UUID v4 is generated.
//
// This middleware must be registered before StructuredLogger so that the
// logger picks up the ID set here via c.Locals("request_id").
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		c.Locals(requestIDKey, id)
		c.Set(RequestIDHeader, id)
		return c.Next()
	}
}
