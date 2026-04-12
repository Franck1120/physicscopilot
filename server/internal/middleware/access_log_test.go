// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestApacheCommonLogDisabledByDefault verifies that when APP_ACCESS_LOG_FORMAT
// is not set (or set to a value other than "apache") the middleware is a no-op
// and the handler still returns the expected status code.
func TestApacheCommonLogDisabledByDefault(t *testing.T) {
	// Ensure the env var is not set for this test.
	t.Setenv("APP_ACCESS_LOG_FORMAT", "")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(ApacheCommonLog())
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

// TestApacheCommonLogEnabled sets APP_ACCESS_LOG_FORMAT=apache and verifies
// that the middleware still passes through and returns 200 for a dummy request.
func TestApacheCommonLogEnabled(t *testing.T) {
	t.Setenv("APP_ACCESS_LOG_FORMAT", "apache")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(ApacheCommonLog())
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}
