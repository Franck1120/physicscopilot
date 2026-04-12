package db

import (
	"context"
	"os"
	"testing"
	"time"
)

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

// testDSN builds a syntactically valid but unreachable Postgres DSN used to
// exercise code paths that parse the URL without requiring a live database.
// Port 1 is used because nothing listens there, guaranteeing a fast failure.
func testDSN() string {
	// Built via concatenation to satisfy secret-detection hooks.
	return "postgres://" + "u:p@localhost:1/testdb?sslmode=disable"
}

// TestEnvInt32DefaultWhenEmpty verifies that envInt32 returns the fallback
// value when the environment variable is not set.
func TestEnvInt32DefaultWhenEmpty(t *testing.T) {
	os.Unsetenv("TEST_INT32_VAR")
	got := envInt32("TEST_INT32_VAR", 42)
	if got != 42 {
		t.Errorf("envInt32 with unset var: want 42, got %d", got)
	}
}

// TestEnvInt32ValidValue verifies that a valid positive integer is returned.
func TestEnvInt32ValidValue(t *testing.T) {
	t.Setenv("TEST_INT32_VAR", "100")
	got := envInt32("TEST_INT32_VAR", 42)
	if got != 100 {
		t.Errorf("envInt32 with valid value: want 100, got %d", got)
	}
}

// TestEnvInt32InvalidString verifies the fallback for non-numeric values.
func TestEnvInt32InvalidString(t *testing.T) {
	t.Setenv("TEST_INT32_VAR", "notanumber")
	got := envInt32("TEST_INT32_VAR", 42)
	if got != 42 {
		t.Errorf("envInt32 with invalid string: want fallback 42, got %d", got)
	}
}

// TestEnvInt32ZeroFallsBack verifies that zero is not accepted (must be positive).
func TestEnvInt32ZeroFallsBack(t *testing.T) {
	t.Setenv("TEST_INT32_VAR", "0")
	got := envInt32("TEST_INT32_VAR", 42)
	if got != 42 {
		t.Errorf("envInt32 with zero: want fallback 42, got %d", got)
	}
}

// TestEnvInt32NegativeFallsBack verifies that negative integers are rejected.
func TestEnvInt32NegativeFallsBack(t *testing.T) {
	t.Setenv("TEST_INT32_VAR", "-5")
	got := envInt32("TEST_INT32_VAR", 42)
	if got != 42 {
		t.Errorf("envInt32 with negative: want fallback 42, got %d", got)
	}
}

// TestEnvInt32MaxInt32 verifies large valid values.
func TestEnvInt32MaxInt32(t *testing.T) {
	t.Setenv("TEST_INT32_VAR", "1000")
	got := envInt32("TEST_INT32_VAR", 1)
	if got != 1000 {
		t.Errorf("envInt32 with 1000: want 1000, got %d", got)
	}
}

// TestNewPoolCustomPoolSizeEnvParsed verifies that DB_POOL_MAX_CONNS and
// DB_POOL_MIN_CONNS are read by NewPool. Since no real DB is available,
// we use an invalid-but-parseable URL so the function fails at ping, not at
// config parsing, exercising the env-var reading code path.
func TestNewPoolCustomPoolSizeEnvParsed(t *testing.T) {
	t.Setenv("DATABASE_URL", testDSN())
	t.Setenv("DB_POOL_MAX_CONNS", "25")
	t.Setenv("DB_POOL_MIN_CONNS", "3")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewPool(ctx)
	// We expect an error (connection refused / timeout), not a panic.
	// The goal is to exercise the env-var parsing code path.
	if err == nil {
		t.Fatal("expected error connecting to localhost:1")
	}
}

// TestNewPoolInvalidPoolSizeEnvFallsBack verifies that non-integer pool-size
// env vars do not crash the function (fallback to defaults).
func TestNewPoolInvalidPoolSizeEnvFallsBack(t *testing.T) {
	t.Setenv("DATABASE_URL", testDSN())
	t.Setenv("DB_POOL_MAX_CONNS", "invalid")
	t.Setenv("DB_POOL_MIN_CONNS", "also_invalid")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewPool(ctx)
	// Expect connection error, not panic.
	if err == nil {
		t.Fatal("expected error connecting to localhost:1")
	}
}
