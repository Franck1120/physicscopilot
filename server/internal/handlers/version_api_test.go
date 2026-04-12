// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// newVersionAPITestApp returns a minimal Fiber application with VersionHandler
// mounted at GET /api/version using fixed, deterministic build metadata.
// Shared by all test functions in this file to avoid repeated setup boilerplate.
func newVersionAPITestApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/version", VersionHandler("0.1.0", "2026-04-12T00:00:00Z", "go1.25"))
	return app
}

func TestVersionAPIHandlerReturns200(t *testing.T) {
	app := newVersionAPITestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestVersionAPIHandlerAllFieldsPresent(t *testing.T) {
	app := newVersionAPITestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	requiredFields := []string{"version", "build_time", "go_version", "api_version"}
	for _, field := range requiredFields {
		if _, ok := body[field]; !ok {
			t.Errorf("missing required field %q in /api/version response", field)
		}
	}
}

func TestVersionAPIHandlerCacheControlHeader(t *testing.T) {
	app := newVersionAPITestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "public, max-age=3600" {
		t.Errorf("Cache-Control: want %q, got %q", "public, max-age=3600", cc)
	}
}
