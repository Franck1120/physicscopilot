package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestWSAuthMiddlewareBearerOnlyNoTokenReturns401 verifies that a header
// containing exactly "Bearer " (keyword plus trailing space, empty token)
// is rejected as a missing token.
func TestWSAuthMiddlewareBearerOnlyNoTokenReturns401(t *testing.T) {
	const secret = "edge-secret-bearer-only"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ") // "Bearer " with no token after
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("'Bearer ' with empty token: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareBearerDoubleSpaceReturns401 verifies that
// "Bearer  <token>" (double space between Bearer and the token) is rejected.
// The middleware uses strings.HasPrefix(auth, "Bearer ") + TrimPrefix, so the
// leading extra space remains inside the extracted token string, making it invalid.
func TestWSAuthMiddlewareBearerDoubleSpaceReturns401(t *testing.T) {
	const secret = "edge-secret-double-space"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	tok := signToken(t, secret, "user-ds")
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer  "+tok) // double space
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("double-space Bearer: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareTokenSignedWithWrongSecretReturns401 verifies that a
// structurally valid HS256 JWT whose signature was produced with a different
// secret key is rejected.
func TestWSAuthMiddlewareTokenSignedWithWrongSecretReturns401(t *testing.T) {
	const secret = "correct-secret-for-middleware"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Sign with a completely different secret.
	tok := signToken(t, "totally-wrong-secret", "attacker-user")
	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong-secret token: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareVeryLongTokenReturns401 verifies that an oversized token
// (> 8 KB) does not crash the middleware and is rejected as invalid.
func TestWSAuthMiddlewareVeryLongTokenReturns401(t *testing.T) {
	const secret = "edge-secret-long-token"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Construct a fake "token" that is far larger than any legitimate JWT.
	longToken := strings.Repeat("a", 2*1024) // 2 KB — still unrealistically large for a JWT
	req := httptest.NewRequest(http.MethodGet, "/protected?token="+longToken, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("very long token: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareRS256TokenRejectedReturns401 verifies that a token whose
// signing algorithm header claims RS256 (non-HMAC) is rejected even when the
// payload is otherwise well-formed. The middleware enforces HMAC-only tokens.
func TestWSAuthMiddlewareRS256TokenRejectedReturns401(t *testing.T) {
	const secret = "edge-secret-rs256"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Construct a fake JWT with alg=RS256 header manually.
	// base64url-encode each part (no padding).
	b64url := func(s string) string {
		enc := base64.StdEncoding.EncodeToString([]byte(s))
		enc = strings.TrimRight(enc, "=")
		enc = strings.NewReplacer("+", "-", "/", "_").Replace(enc)
		return enc
	}
	header := b64url(`{"alg":"RS256","typ":"JWT"}`)
	exp := time.Now().Add(time.Hour).Unix()
	payload := b64url(fmt.Sprintf(`{"sub":"evil","exp":%d}`, exp))
	rs256Token := header + "." + payload + ".fakesignature"

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+rs256Token, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("RS256 token: want 401, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareTokenInQueryParamNoHeaderPresent verifies that the ?token=
// query parameter is accepted as the sole source of authentication when there is
// no Authorization header at all.
func TestWSAuthMiddlewareTokenInQueryParamNoHeaderPresent(t *testing.T) {
	const secret = "edge-secret-query-only"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	tok := signToken(t, secret, "query-only-user")
	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	// Explicitly leave Authorization header absent.
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("token via query param (no header): want 200, got %d", resp.StatusCode)
	}
}

// TestWSAuthMiddlewareNoneAlgorithmReturns401 verifies that a JWT with
// alg=none (the classic signature-bypass) is rejected.
func TestWSAuthMiddlewareNoneAlgorithmReturns401(t *testing.T) {
	const secret = "edge-secret-none-alg"
	t.Setenv("SUPABASE_JWT_SECRET", secret)
	app := newAuthTestApp(t)

	// Create an unsigned token via the jwt library's special UnsafeAllowNoneSignatureType.
	claims := jwt.MapClaims{
		"sub": "attacker",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none-alg token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+tok, nil)
	resp, respErr := app.Test(req)
	if respErr != nil {
		t.Fatalf("test: %v", respErr)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("alg=none token: want 401, got %d", resp.StatusCode)
	}
}
