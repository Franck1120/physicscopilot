package logger

import (
	"log/slog"
	"testing"
)

// ---------------------------------------------------------------------------
// isProduction tests (unexported, accessible within package logger)
// ---------------------------------------------------------------------------

// TestIsProductionWithProd verifies that APP_ENV=prod is treated as production.
func TestIsProductionWithProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	if !isProduction() {
		t.Error("isProduction(): want true for APP_ENV=prod, got false")
	}
}

// TestIsProductionWithProduction verifies that APP_ENV=production is production.
func TestIsProductionWithProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	if !isProduction() {
		t.Error("isProduction(): want true for APP_ENV=production, got false")
	}
}

// TestIsProductionWithDev verifies that APP_ENV=development is not production.
func TestIsProductionWithDev(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	if isProduction() {
		t.Error("isProduction(): want false for APP_ENV=development, got true")
	}
}

// TestIsProductionWithEmpty verifies that an empty APP_ENV is not production.
func TestIsProductionWithEmpty(t *testing.T) {
	t.Setenv("APP_ENV", "")
	if isProduction() {
		t.Error("isProduction(): want false for APP_ENV=\"\", got true")
	}
}

// TestIsProductionWithStaging verifies that APP_ENV=staging is not production.
func TestIsProductionWithStaging(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	if isProduction() {
		t.Error("isProduction(): want false for APP_ENV=staging, got true")
	}
}

// ---------------------------------------------------------------------------
// Init tests
// ---------------------------------------------------------------------------

// TestInitProductionMode verifies that Init() runs without panic in production
// mode and installs a non-nil global logger.
func TestInitProductionMode(t *testing.T) {
	old := slog.Default()
	defer slog.SetDefault(old)

	t.Setenv("APP_ENV", "production")
	Init()

	if slog.Default() == nil {
		t.Error("Init() with APP_ENV=production: slog.Default() is nil")
	}
}

// TestInitDevMode verifies that Init() runs without panic in development mode
// and installs a non-nil global logger.
func TestInitDevMode(t *testing.T) {
	old := slog.Default()
	defer slog.SetDefault(old)

	t.Setenv("APP_ENV", "development")
	Init()

	if slog.Default() == nil {
		t.Error("Init() with APP_ENV=development: slog.Default() is nil")
	}
}

// TestInitNoEnv verifies that Init() runs without panic when APP_ENV is unset.
func TestInitNoEnv(t *testing.T) {
	old := slog.Default()
	defer slog.SetDefault(old)

	t.Setenv("APP_ENV", "")
	Init()

	if slog.Default() == nil {
		t.Error("Init() with APP_ENV=\"\": slog.Default() is nil")
	}
}

// TestInitChangesLogger verifies that Init() replaces the global logger with a
// new instance (the pointer changes between calls with different environments).
func TestInitChangesLogger(t *testing.T) {
	old := slog.Default()
	defer slog.SetDefault(old)

	t.Setenv("APP_ENV", "development")
	Init()
	devLogger := slog.Default()

	t.Setenv("APP_ENV", "production")
	Init()
	prodLogger := slog.Default()

	if devLogger == prodLogger {
		t.Error("Init() should install a new logger on each call, but dev and prod loggers are the same pointer")
	}
}
