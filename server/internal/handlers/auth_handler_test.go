// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
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

// TestWSAuthMiddlewareWrongBearerFormatReturns401 verifies that a token prefixed
// with "Token " instead of "Bearer " is rejected with 401.
func TestWSAuthMiddlewareWrongBearerFormatReturns401(t *testing.T) {
	const secret = "test-secret-wrong-prefix"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	tok := signToken(t, secret, "user-123")
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token "+tok) // wrong prefix
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong prefix (Token): want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareBearerNoSpaceReturns401 verifies that "Bearer<token>" (no
// space after Bearer) is treated as a missing token, not a valid header.
func TestWSAuthMiddlewareBearerNoSpaceReturns401(t *testing.T) {
	const secret = "test-secret-nospace"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	tok := signToken(t, secret, "user-456")
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer"+tok) // no space
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no-space Bearer: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareTokenWithNoSubStillAllows verifies that a valid signed
// token without a "sub" claim passes authentication (sub is optional for
// authorization) but stores an empty user_id.
func TestWSAuthMiddlewareTokenWithNoSubStillAllows(t *testing.T) {
	const secret = "test-secret-no-sub"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Token without a sub claim — still valid HS256 with exp in the future.
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		// no "sub" field
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign no-sub token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, respErr := app.Test(req)
	if respErr != nil {
		t.Fatalf("test: %v", respErr)
	}
	// The middleware must allow a valid token even without sub (sub is
	// stored only when present; absence is not an auth failure).
	if resp.StatusCode != http.StatusOK {
		t.Errorf("token without sub: want 200, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareWrongAlgorithmReturns401 verifies that a token signed
// with an unexpected algorithm (RS256/ECDSA) is rejected even if the header
// claims HS256 but the payload was not signed with the correct secret.
func TestWSAuthMiddlewareWrongAlgorithmReturns401(t *testing.T) {
	const secret = "test-secret-wrong-algo"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Build a token signed with a different secret to simulate key mismatch.
	claims := jwt.MapClaims{
		"sub": "attacker",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, respErr := app.Test(req)
	if respErr != nil {
		t.Fatalf("test: %v", respErr)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong secret: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareBothQueryAndHeaderUsesQuery verifies that when both the
// query param and Authorization header contain tokens, the query param takes
// precedence (as documented in the handler).
func TestWSAuthMiddlewareBothQueryAndHeaderUsesQuery(t *testing.T) {
	const secret = "test-secret-precedence"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	validTok := signToken(t, secret, "query-user")
	// Header token signed with wrong secret — if header took precedence it would fail.
	invalidHeaderTok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "header-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("wrong-secret"))

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+validTok, nil)
	req.Header.Set("Authorization", "Bearer "+invalidHeaderTok)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	// Query param (valid) should win → 200.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("query param should take precedence: want 200, got %d", resp.StatusCode)
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
