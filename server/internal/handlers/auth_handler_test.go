package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// newAuthTestApp builds a minimal Fiber app protected by WSAuthMiddleware.
// The middleware is constructed inside the function so t.Setenv takes effect.
func newAuthTestApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(WSAuthMiddleware())
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"user_id": userID})
	})
	return app
}

// signToken creates an HS256 JWT with the given secret and subject.
func signToken(t *testing.T, secret, subject string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": subject,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

// ── WSAuthMiddleware ──────────────────────────────────────────────────────────

func TestWSAuthMiddlewareDevModeAlwaysPasses(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", "")
	app := newAuthTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("dev mode (no secret): want 200, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareMissingTokenReturns401(t *testing.T) {
	const secret = "test-secret-for-missing-token"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("missing token: want 401, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareInvalidTokenReturns401(t *testing.T) {
	const secret = "test-secret-for-invalid-token"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/protected?token=notavalidjwt", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("invalid token: want 401, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareValidQueryTokenReturns200(t *testing.T) {
	const secret = "test-secret-for-valid-token"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	tok := signToken(t, secret, "user-abc")
	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid token (query): want 200, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareValidBearerHeaderReturns200(t *testing.T) {
	const secret = "test-secret-for-bearer"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	tok := signToken(t, secret, "user-xyz")
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid token (header): want 200, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareExpiredTokenReturns401(t *testing.T) {
	const secret = "test-secret-for-expired"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Token expired one hour ago.
	claims := jwt.MapClaims{
		"sub": "user-old",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expired token: want 401, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareWrongSigningMethodReturns401(t *testing.T) {
	const secret = "test-secret-signing-method"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Craft a raw JWT claiming RS256 (non-HMAC) with a dummy signature.
	// The middleware keyFunc rejects non-HMAC methods before verifying the signature,
	// so the exact signature value does not matter.
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"x","exp":9999999999}`))
	fakeToken := header + "." + payload + ".fakesig"

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+fakeToken, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong signing method: want 401, got %d", resp.StatusCode)
	}
}

func TestWSAuthMiddlewareTokenWithoutSubDoesNotSetUserID(t *testing.T) {
	const secret = "test-secret-no-sub"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Sign a valid token without a "sub" claim — request should pass but
	// user_id must not be set in c.Locals.
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		// no "sub" claim
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("no-sub token: want 200 (request allowed), got %d", resp.StatusCode)
	}
}

// ── OpenAPIHandler ────────────────────────────────────────────────────────────

func TestOpenAPIHandlerReturnsYAML(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/docs", OpenAPIHandler())

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "yaml") {
		t.Errorf("Content-Type: want yaml, got %q", ct)
	}
}
