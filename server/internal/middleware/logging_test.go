package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// newLoggerApp builds a minimal Fiber app with StructuredLogger, capturing
// the request_id set by the middleware into the provided slice on each request.
func newLoggerApp(captured *[]string) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(StructuredLogger())
	app.Get("/", func(c *fiber.Ctx) error {
		*captured = append(*captured, RequestID(c))
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/bad", func(c *fiber.Ctx) error {
		*captured = append(*captured, RequestID(c))
		return c.SendStatus(fiber.StatusBadRequest)
	})
	app.Get("/err", func(c *fiber.Ctx) error {
		*captured = append(*captured, RequestID(c))
		return c.SendStatus(fiber.StatusInternalServerError)
	})
	return app
}

func TestStructuredLoggerInjectsRequestID(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	if len(ids) == 0 || ids[0] == "" {
		t.Error("expected non-empty request_id in c.Locals")
	}
	// generateRequestID() returns 16-char lowercase hex (8 random bytes).
	if got := ids[0]; len(got) != 16 {
		t.Errorf("request_id should be 16-char hex, got %q (len=%d)", got, len(got))
	}
}

func TestStructuredLoggerUniqueIDPerRequest(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if _, err := app.Test(req); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}

	if len(ids) != 5 {
		t.Fatalf("expected 5 captured IDs, got %d", len(ids))
	}
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate request_id: %q", id)
		}
		seen[id] = true
	}
}

func TestStructuredLoggerPassesThroughStatus(t *testing.T) {
	var ids []string
	app := newLoggerApp(&ids)

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/bad", http.StatusBadRequest},
		{"/err", http.StatusInternalServerError},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
		if resp.StatusCode != tc.want {
			t.Errorf("%s: want %d, got %d", tc.path, tc.want, resp.StatusCode)
		}
	}
}

func TestRequestIDReturnsEmptyWithoutMiddleware(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	var captured string
	app.Get("/", func(c *fiber.Ctx) error {
		captured = RequestID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("test: %v", err)
	}
	if captured != "" {
		t.Errorf("expected empty request_id without middleware, got %q", captured)
	}
}

func TestAnonymizeIPProduces8CharHex(t *testing.T) {
	result := anonymizeIP("192.168.1.100")
	if len(result) != 8 {
		t.Errorf("anonymizeIP: want 8-char hex, got %q (len=%d)", result, len(result))
	}
	for _, ch := range result {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("anonymizeIP: non-hex character %q in result %q", ch, result)
		}
	}
}

func TestAnonymizeIPIsDeterministic(t *testing.T) {
	r1 := anonymizeIP("10.0.0.1")
	r2 := anonymizeIP("10.0.0.1")
	if r1 != r2 {
		t.Errorf("anonymizeIP is not deterministic: %q != %q", r1, r2)
	}
}

func TestAnonymizeIPDiffersForDifferentIPs(t *testing.T) {
	r1 := anonymizeIP("1.2.3.4")
	r2 := anonymizeIP("5.6.7.8")
	if r1 == r2 {
		t.Error("expected different hashes for different IPs")
	}
}
