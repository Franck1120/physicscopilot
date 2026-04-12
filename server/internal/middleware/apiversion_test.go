package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestAPIVersionHeaderPresent verifies that X-API-Version is set on every response.
func TestAPIVersionHeaderPresent(t *testing.T) {
	const ver = "1.2.3"
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(APIVersion(ver))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	got := resp.Header.Get("X-API-Version")
	if got != ver {
		t.Errorf("expected X-API-Version %q, got %q", ver, got)
	}
}

// TestAPIVersionHeaderValue verifies the header value exactly matches the
// version string passed to APIVersion(), without any transformation.
func TestAPIVersionHeaderValue(t *testing.T) {
	for _, ver := range []string{"0.1.0", "2.0.0-beta", "v3"} {
		t.Run(ver, func(t *testing.T) {
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(APIVersion(ver))
			app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("test: %v", err)
			}

			got := resp.Header.Get("X-API-Version")
			if got != ver {
				t.Errorf("want %q, got %q", ver, got)
			}
		})
	}
}

// TestAPIVersionHeaderPresentOnErrorResponse verifies the header is included
// even when the handler returns a non-200 status code.
func TestAPIVersionHeaderPresentOnErrorResponse(t *testing.T) {
	const ver = "1.0.0"
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(APIVersion(ver))
	app.Get("/err", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusInternalServerError) })

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	got := resp.Header.Get("X-API-Version")
	if got != ver {
		t.Errorf("expected X-API-Version %q on error response, got %q", ver, got)
	}
}
