// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import "github.com/gofiber/fiber/v2"

// VersionResponse is the JSON payload returned by GET /version.
type VersionResponse struct {
	Version    string `json:"version"`
	BuildTime  string `json:"build_time"`
	GoVersion  string `json:"go_version"`
	CommitHash string `json:"commit_hash"`
}

// VersionHandler returns a Fiber handler for GET /version.
// It reports the application version, build time, Go runtime version, and
// the Git commit hash baked in at build time (or "dev" when not set).
func VersionHandler(version, buildTime, goVersion, commitHash string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(VersionResponse{
			Version:    version,
			BuildTime:  buildTime,
			GoVersion:  goVersion,
			CommitHash: commitHash,
		})
	}
}
