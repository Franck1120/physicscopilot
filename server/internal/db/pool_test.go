package db

import (
	"context"
	"os"
	"testing"
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
