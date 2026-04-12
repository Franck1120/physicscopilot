package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// testSwaggerApp builds a minimal Fiber app with the /docs route registered.
func testSwaggerApp() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.SendString(SwaggerUIHTML("/api/docs"))
	})
	return app
}

// TestSwaggerUIStatus verifies that GET /docs returns HTTP 200.
func TestSwaggerUIStatus(t *testing.T) {
	app := testSwaggerApp()

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}
}

// TestSwaggerUIContentType verifies that the response Content-Type is text/html.
func TestSwaggerUIContentType(t *testing.T) {
	app := testSwaggerApp()

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type: want text/html prefix, got %q", ct)
	}
}

// TestSwaggerUIBodyContainsSwaggerUI verifies that the response body references
// the swagger-ui element so the browser can mount the widget.
func TestSwaggerUIBodyContainsSwaggerUI(t *testing.T) {
	app := testSwaggerApp()

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	content := string(body)
	if !strings.Contains(content, "swagger-ui") {
		t.Error("expected body to contain \"swagger-ui\"")
	}
}

// TestSwaggerUIBodyContainsSpecURL verifies that the generated HTML embeds the
// specURL passed to SwaggerUIHTML so the client fetches the correct spec.
func TestSwaggerUIBodyContainsSpecURL(t *testing.T) {
	const specURL = "/api/docs"
	html := SwaggerUIHTML(specURL)
	if !strings.Contains(html, specURL) {
		t.Errorf("expected HTML to contain spec URL %q", specURL)
	}
}
