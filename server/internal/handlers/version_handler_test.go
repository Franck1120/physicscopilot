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

func newVersionTestApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/version", VersionHandler("0.1.0", "2026-04-12T00:00:00Z", "go1.25"))
	return app
}

func TestVersionHandlerReturns200(t *testing.T) {
	app := newVersionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestVersionHandlerAllFieldsPresent(t *testing.T) {
	app := newVersionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
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
			t.Errorf("missing required field %q in version response", field)
		}
	}
}

func TestVersionHandlerFieldValues(t *testing.T) {
	app := newVersionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body["version"] != "0.1.0" {
		t.Errorf("version: want %q, got %v", "0.1.0", body["version"])
	}
	if body["build_time"] != "2026-04-12T00:00:00Z" {
		t.Errorf("build_time: want %q, got %v", "2026-04-12T00:00:00Z", body["build_time"])
	}
	if body["go_version"] != "go1.25" {
		t.Errorf("go_version: want %q, got %v", "go1.25", body["go_version"])
	}
	if body["api_version"] != "v1" {
		t.Errorf("api_version: want %q, got %v", "v1", body["api_version"])
	}
}

func TestVersionHandlerCacheControlHeader(t *testing.T) {
	app := newVersionTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc != "public, max-age=3600" {
		t.Errorf("Cache-Control: want %q, got %q", "public, max-age=3600", cc)
	}
}
