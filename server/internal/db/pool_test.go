package db

import (
	"context"
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// NewPool tests
// ---------------------------------------------------------------------------

func TestNewPoolFailsWithoutDatabaseURL(t *testing.T) {
	// Ensure DATABASE_URL is not set for this test.
	original := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if original != "" {
			os.Setenv("DATABASE_URL", original)
		}
	}()

	_, err := NewPool(context.Background())
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is not set")
	}
	if err.Error() != "DATABASE_URL environment variable is not set" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestNewPoolFailsWithInvalidURL(t *testing.T) {
	original := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "not-a-valid-url")
	defer func() {
		if original != "" {
			os.Setenv("DATABASE_URL", original)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	_, err := NewPool(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid DATABASE_URL")
	}
}

// ---------------------------------------------------------------------------
// envInt32 tests
// ---------------------------------------------------------------------------

func TestEnvInt32_Fallback(t *testing.T) {
	os.Unsetenv("TEST_ENVINT32_KEY")
	got := envInt32("TEST_ENVINT32_KEY", 42)
	if got != 42 {
		t.Errorf("expected fallback 42, got %d", got)
	}
}

func TestEnvInt32_ValidPositive(t *testing.T) {
	os.Setenv("TEST_ENVINT32_KEY", "25")
	defer os.Unsetenv("TEST_ENVINT32_KEY")

	got := envInt32("TEST_ENVINT32_KEY", 42)
	if got != 25 {
		t.Errorf("expected 25, got %d", got)
	}
}

func TestEnvInt32_NonNumeric(t *testing.T) {
	os.Setenv("TEST_ENVINT32_KEY", "abc")
	defer os.Unsetenv("TEST_ENVINT32_KEY")

	got := envInt32("TEST_ENVINT32_KEY", 42)
	if got != 42 {
		t.Errorf("expected fallback 42 for non-numeric value, got %d", got)
	}
}

func TestEnvInt32_Negative(t *testing.T) {
	os.Setenv("TEST_ENVINT32_KEY", "-5")
	defer os.Unsetenv("TEST_ENVINT32_KEY")

	// Negative is not a positive integer, so fallback is returned.
	got := envInt32("TEST_ENVINT32_KEY", 42)
	if got != 42 {
		t.Errorf("expected fallback 42 for negative value, got %d", got)
	}
}

func TestEnvInt32_Zero(t *testing.T) {
	os.Setenv("TEST_ENVINT32_KEY", "0")
	defer os.Unsetenv("TEST_ENVINT32_KEY")

	// Zero is not > 0, so fallback is returned.
	got := envInt32("TEST_ENVINT32_KEY", 42)
	if got != 42 {
		t.Errorf("expected fallback 42 for zero value, got %d", got)
	}
}

func TestEnvInt32_DefaultMaxConns(t *testing.T) {
	os.Unsetenv("DB_POOL_MAX_CONNS")
	got := envInt32("DB_POOL_MAX_CONNS", defaultMaxConns)
	if got != defaultMaxConns {
		t.Errorf("expected defaultMaxConns %d, got %d", defaultMaxConns, got)
	}
}

func TestEnvInt32_DefaultMinConns(t *testing.T) {
	os.Unsetenv("DB_POOL_MIN_CONNS")
	got := envInt32("DB_POOL_MIN_CONNS", defaultMinConns)
	if got != defaultMinConns {
		t.Errorf("expected defaultMinConns %d, got %d", defaultMinConns, got)
	}
}
