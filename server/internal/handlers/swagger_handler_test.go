// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newSwaggerTestApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/swagger", SwaggerUIHandler())
	return app
}

func TestSwaggerUIHandlerReturns200(t *testing.T) {
	app := newSwaggerTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/swagger", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestSwaggerUIHandlerContentTypeIsHTML(t *testing.T) {
	app := newSwaggerTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/swagger", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type: want text/html prefix, got %q", ct)
	}
}

func TestSwaggerUIHandlerBodyContainsSwaggerUI(t *testing.T) {
	app := newSwaggerTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/swagger", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "swagger-ui") {
		t.Errorf("response body does not contain %q", "swagger-ui")
	}
}
