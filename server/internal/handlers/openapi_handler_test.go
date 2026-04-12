package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// testOpenAPIApp wires OpenAPIHandler into a minimal Fiber app for testing.
func testOpenAPIApp() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/docs", OpenAPIHandler())
	return app
}

// TestOpenAPIHandlerStatusOK verifies that GET /api/docs returns HTTP 200.
func TestOpenAPIHandlerStatusOK(t *testing.T) {
	app := testOpenAPIApp()

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}
}

// TestOpenAPIHandlerContentType verifies that the response carries the correct
// Content-Type header for a YAML document.
func TestOpenAPIHandlerContentType(t *testing.T) {
	app := testOpenAPIApp()

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/yaml; charset=utf-8" {
		t.Errorf("Content-Type: want %q, got %q", "application/yaml; charset=utf-8", ct)
	}
}

// TestOpenAPIHandlerBodyNonEmpty verifies that the response body is not empty.
func TestOpenAPIHandlerBodyNonEmpty(t *testing.T) {
	app := testOpenAPIApp()

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		t.Error("expected non-empty response body, got empty")
	}
}

// TestOpenAPIHandlerBodyIsYAML verifies that the response body looks like a
// valid OpenAPI YAML document by checking for known top-level keys.
func TestOpenAPIHandlerBodyIsYAML(t *testing.T) {
	app := testOpenAPIApp()

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	content := string(body)

	if !strings.Contains(content, "openapi:") {
		t.Error("expected body to contain \"openapi:\" key")
	}
	if !strings.Contains(content, "info:") {
		t.Error("expected body to contain \"info:\" key")
	}
}
