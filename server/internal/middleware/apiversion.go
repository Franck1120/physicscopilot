// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

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
