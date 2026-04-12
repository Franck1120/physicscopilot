// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

package handlers

import "github.com/gofiber/fiber/v2"

// VersionHandler returns a fiber.Handler that serves build version information.
// version, buildTime, and goVersion are baked in at startup; api_version is
// always "v1" for the current generation of this service.
//
// The response is publicly cacheable for one hour — version metadata changes
// only on deployment, so clients may cache it aggressively.
func VersionHandler(version, buildTime, goVersion string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "public, max-age=3600")
		return c.JSON(fiber.Map{
			"version":     version,
			"build_time":  buildTime,
			"go_version":  goVersion,
			"api_version": "v1",
		})
	}
}
