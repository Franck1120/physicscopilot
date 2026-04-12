package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// newRequestIDApp builds a minimal Fiber app with RequestIDMiddleware applied.
func newRequestIDApp() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(RequestIDMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

// TestRequestIDGeneratesUUIDWhenMissing verifies that a UUID v4 is generated
// and written to the X-Request-ID response header when the client omits it.
func TestRequestIDGeneratesUUIDWhenMissing(t *testing.T) {
	app := newRequestIDApp()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	id := resp.Header.Get(RequestIDHeader)
	if id == "" {
		t.Fatal("expected X-Request-ID in response, got empty string")
	}
	// UUID v4 canonical form: 8-4-4-4-12 hex chars separated by hyphens = 36 chars.
	if len(id) != 36 {
		t.Errorf("expected UUID v4 (36 chars), got %q (len=%d)", id, len(id))
	}
}

// TestRequestIDReusesClientValue verifies that the middleware propagates an
// existing X-Request-ID supplied by the client to the response header.
func TestRequestIDReusesClientValue(t *testing.T) {
	app := newRequestIDApp()
	const clientID = "my-trace-id-abc123"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, clientID)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	got := resp.Header.Get(RequestIDHeader)
	if got != clientID {
		t.Errorf("expected X-Request-ID %q, got %q", clientID, got)
	}
}

// TestRequestIDPresentInResponseHeader verifies the header is present on every
// response regardless of the handler's status code.
func TestRequestIDPresentInResponseHeader(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(RequestIDMiddleware())
	app.Get("/ok", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/err", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusInternalServerError) })

	for _, path := range []string{"/ok", "/err"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if resp.Header.Get(RequestIDHeader) == "" {
			t.Errorf("GET %s: X-Request-ID missing from response", path)
		}
	}
}

// TestRequestIDStoredInLocals verifies the ID is stored in c.Locals so that
// downstream middleware (e.g. StructuredLogger) can read it without parsing headers.
func TestRequestIDStoredInLocals(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(RequestIDMiddleware())

	var capturedLocal string
	app.Get("/", func(c *fiber.Ctx) error {
		capturedLocal = RequestID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	headerID := resp.Header.Get(RequestIDHeader)
	if capturedLocal == "" {
		t.Error("expected non-empty request_id in c.Locals")
	}
	if capturedLocal != headerID {
		t.Errorf("locals value %q does not match header value %q", capturedLocal, headerID)
	}
}
