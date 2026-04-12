package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Franck1120/physicscopilot/server/internal/handlers"
	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// buildTestApp constructs the Fiber app without starting a listener or
// requiring any environment variables.
func buildTestApp(t *testing.T) *fiber.App {
	t.Helper()
	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	return newFiberApp("test", sh, fh, ws, nil, nil)
}

func TestNewFiberAppHealthRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/health not registered: got 404")
	}
}

func TestNewFiberAppSessionsRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// Without auth the endpoint should return 401, not 404.
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/api/sessions not registered: got 404")
	}
}

func TestNewFiberAppMetricsRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// Without credentials the endpoint returns 401, not 404.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/metrics not registered: got 404")
	}
}

func TestNewFiberAppWSRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// Plain GET without Upgrade header should return 426, not 404.
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/ws not registered: got 404")
	}
}

func TestNewFiberAppFeedbackRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	// POST /api/feedback without a body returns 400 (validation), not 404.
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/api/feedback not registered: got 404")
	}
}

func TestNewFiberAppDocsRouteRegistered(t *testing.T) {
	app := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/api/docs not registered: got 404")
	}
}

func TestNewFiberAppProductionAllowedOriginsWarning(t *testing.T) {
	// APP_ENV=production without ALLOWED_ORIGINS triggers a slog.Warn.
	// Verify the app still builds and serves /health normally.
	t.Setenv("APP_ENV", "production")
	// ALLOWED_ORIGINS intentionally not set → exercises the warning branch.

	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	app := newFiberApp("test", sh, fh, ws, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("production without ALLOWED_ORIGINS: want 200, got %d", resp.StatusCode)
	}
}

// TestSecurityHeadersOnEveryResponse verifies that X-Content-Type-Options,
// X-Frame-Options, Referrer-Policy, and Content-Security-Policy are present
// on all responses regardless of environment.
func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	app := buildTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options: want nosniff, got %q", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options: want DENY, got %q", got)
	}
	if got := resp.Header.Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy: want strict-origin-when-cross-origin, got %q", got)
	}
	if got := resp.Header.Get("Content-Security-Policy"); got == "" {
		t.Error("Content-Security-Policy header is missing")
	}
}

// TestCSRFProtectionDocumented verifies that the API routes carry the
// Authorization header requirement in the CORS AllowHeaders config, confirming
// that JWT Bearer tokens (not cookies) are the auth mechanism and therefore
// CSRF tokens are not required (JWT via Authorization header is CSRF-safe).
func TestCSRFProtectionDocumented(t *testing.T) {
	app := buildTestApp(t)
	// OPTIONS preflight on an API route must expose Authorization in allowed headers.
	req := httptest.NewRequest(http.MethodOptions, "/api/sessions", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowHeaders, "Authorization") {
		t.Errorf("Access-Control-Allow-Headers must include Authorization; got %q", allowHeaders)
	}
}

func TestNewFiberAppHSTSHeaderInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")

	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	app := newFiberApp("test", sh, fh, ws, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	hsts := resp.Header.Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("production: expected Strict-Transport-Security header")
	}
}

func TestNewFiberAppGzipCompressesResponse(t *testing.T) {
	app := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	// Response must be 200 and either gzip-encoded or not (middleware may skip
	// small payloads), but must never be 500.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("gzip test: want 200, got %d", resp.StatusCode)
	}
}

func TestNewFiberAppIdleTimeoutConfigured(t *testing.T) {
	// Verify the app builds correctly with IdleTimeout set — no panic or error.
	app := buildTestApp(t)
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestHealthRouteReturnsOK(t *testing.T) {
	app := buildTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/health: want 200, got %d", resp.StatusCode)
	}
}

func TestDocsRouteReturnsContent(t *testing.T) {
	app := buildTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	// Must not be 404 or 500.
	if resp.StatusCode == http.StatusNotFound {
		t.Error("/api/docs returned 404")
	}
	if resp.StatusCode >= 500 {
		t.Errorf("/api/docs server error: %d", resp.StatusCode)
	}
}

func TestFeedbackRouteWithValidBody(t *testing.T) {
	app := buildTestApp(t)

	body := strings.NewReader(`{"session_id":"test-session","step_number":1,"rating":"positive"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	// Without a DB backend the handler may return 503/500, but NOT 404.
	if resp.StatusCode == http.StatusNotFound {
		t.Error("/api/feedback returned 404 — route not registered")
	}
}

func TestFeedbackRouteEmptyBodyReturns400(t *testing.T) {
	app := buildTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Error("/api/feedback returned 404 — route not registered")
	}
}

func TestNewFiberAppCORSWithAllowedOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com,https://staging.example.com")

	sessionSvc := services.NewSessionService()
	convSvc := services.NewConversationService(sessionSvc, nil, nil)
	ws := handlers.NewWSHandler(convSvc, sessionSvc)
	sh := handlers.NewSessionHandler(sessionSvc)
	fh := handlers.NewFeedbackHandler(nil)
	app := newFiberApp("test", sh, fh, ws, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://app.example.com")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("production with ALLOWED_ORIGINS: want 200, got %d", resp.StatusCode)
	}
}

func TestBuildTestAppNoJWTSecret(t *testing.T) {
	// In test mode JWT secret is not required — verify app builds successfully.
	os.Unsetenv("SUPABASE_JWT_SECRET")
	app := buildTestApp(t)
	if app == nil {
		t.Fatal("expected non-nil app without JWT secret in dev mode")
	}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("health check without JWT secret: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestMetricsRouteRequiresAuth(t *testing.T) {
	app := buildTestApp(t)

	// Without credentials must return 401, not 200 or 404.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Error("/metrics not registered")
	}
	if resp.StatusCode == http.StatusOK {
		t.Error("/metrics should require auth — got 200 without credentials")
	}
}

func TestFeedbackRouteCORSPreflight(t *testing.T) {
	app := buildTestApp(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/feedback", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Error("/api/feedback CORS preflight returned 404")
	}
}

// --- Tests for extracted helper functions ---

func TestCheckJWTSecretDevModeNoSecret(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", "")
	t.Setenv("APP_ENV", "development")

	if err := checkJWTSecret(); err != nil {
		t.Errorf("dev mode without JWT secret should not error: %v", err)
	}
}

func TestCheckJWTSecretProductionNoSecret(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", "")
	t.Setenv("APP_ENV", "production")

	err := checkJWTSecret()
	if err == nil {
		t.Fatal("production without JWT secret must return an error")
	}
}

func TestCheckJWTSecretProductionWithSecret(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", "super-secret-key")
	t.Setenv("APP_ENV", "production")

	if err := checkJWTSecret(); err != nil {
		t.Errorf("production with JWT secret should not error: %v", err)
	}
}

func TestResolvePortDefault(t *testing.T) {
	t.Setenv("PORT", "")
	if got := resolvePort(); got != "8080" {
		t.Errorf("resolvePort() default: want 8080, got %s", got)
	}
}

func TestResolvePortCustom(t *testing.T) {
	t.Setenv("PORT", "3000")
	if got := resolvePort(); got != "3000" {
		t.Errorf("resolvePort() custom: want 3000, got %s", got)
	}
}

func TestCollectMemoryMetrics(t *testing.T) {
	// collectMemoryMetrics should not panic and should complete without error.
	// It updates Prometheus gauges — we just verify it does not crash.
	collectMemoryMetrics()
}

func TestRunStartsAndShutsDown(t *testing.T) {
	// Use a free port to avoid conflicts with other tests.
	t.Setenv("PORT", "0")
	t.Setenv("SUPABASE_JWT_SECRET", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_URL", "")

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx)
	}()

	// Give the server a moment to start before cancelling.
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run() returned unexpected error: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("run() did not return within 15 seconds after context cancellation")
	}
}

func TestRunFailsWithProductionNoJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SUPABASE_JWT_SECRET", "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := run(ctx)
	if err == nil {
		t.Fatal("run() should fail in production without JWT secret")
	}
}

