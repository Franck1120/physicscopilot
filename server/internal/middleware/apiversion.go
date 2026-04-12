// Copyright (c) 2026 PhysicsCopilot contributors. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import "github.com/gofiber/fiber/v2"

// APIVersion injects a static X-API-Version header in every response
// so clients can detect server API version without parsing the body.
func APIVersion(version string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-API-Version", version)
		return c.Next()
	}
}
