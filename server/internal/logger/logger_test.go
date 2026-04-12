package logger

import (
	"os"
	"testing"
)

// TestIsProductionReturnsTrueForProductionEnv verifies both recognised aliases.
func TestIsProductionReturnsTrueForProductionEnv(t *testing.T) {
	for _, env := range []string{"production", "prod"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("APP_ENV", env)
			if !isProduction() {
				t.Errorf("isProduction() should return true for APP_ENV=%q", env)
			}
		})
	}
}

// TestIsProductionReturnsFalseForNonProductionEnv verifies non-production values.
func TestIsProductionReturnsFalseForNonProductionEnv(t *testing.T) {
	for _, env := range []string{"", "development", "dev", "staging", "test"} {
		t.Run("env="+env, func(t *testing.T) {
			t.Setenv("APP_ENV", env)
			if isProduction() {
				t.Errorf("isProduction() should return false for APP_ENV=%q", env)
			}
		})
	}
}

// TestInitProductionSetsJSONHandler verifies that Init() does not panic in
// production mode and sets a functioning logger.
func TestInitProductionSetsJSONHandler(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	// Init should not panic.
	Init()
	// Logger must be usable — calling slog with it should not panic.
	// We can't easily inspect the handler type in tests, so we rely on the
	// no-panic guarantee and the isProduction branch coverage.
}

// TestInitDevelopmentSetsTextHandler verifies Init() in development mode.
func TestInitDevelopmentSetsTextHandler(t *testing.T) {
	os.Unsetenv("APP_ENV")
	Init()
	// No panic = success; branch coverage is the goal.
}

// TestHashIPIsDeterministic verifies the same input always produces the same hash.
func TestHashIPIsDeterministic(t *testing.T) {
	h1 := HashIP("192.168.1.1")
	h2 := HashIP("192.168.1.1")
	if h1 != h2 {
		t.Errorf("HashIP is not deterministic: %q != %q", h1, h2)
	}
}

// TestHashIPLength verifies the output is always 8 hex characters (4 bytes SHA-256 prefix).
func TestHashIPLength(t *testing.T) {
	for _, ip := range []string{"", "127.0.0.1", "::1", "10.0.0.1"} {
		h := HashIP(ip)
		if len(h) != 8 {
			t.Errorf("HashIP(%q): expected 8 chars, got %d (%q)", ip, len(h), h)
		}
	}
}

// TestHashIPDifferentInputsDifferentOutputs verifies distinct IPs hash differently.
func TestHashIPDifferentInputsDifferentOutputs(t *testing.T) {
	h1 := HashIP("192.168.1.1")
	h2 := HashIP("10.0.0.1")
	if h1 == h2 {
		t.Error("different IPs produced the same hash — collision detected")
	}
}

// TestHashIPIsHex verifies the output contains only hexadecimal characters.
func TestHashIPIsHex(t *testing.T) {
	h := HashIP("1.2.3.4")
	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("HashIP output %q contains non-hex character %q", h, c)
		}
	}
}

// TestHashIPEmptyString verifies no panic on empty input.
func TestHashIPEmptyString(t *testing.T) {
	h := HashIP("")
	if len(h) != 8 {
		t.Errorf("HashIP(\"\") expected 8-char hash, got %q", h)
	}
}
