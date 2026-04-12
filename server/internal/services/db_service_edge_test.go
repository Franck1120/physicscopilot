package services

import (
	"context"
	"strings"
	"testing"
)

// TestNewDBServiceEmptyURLReturnsError verifies that passing an empty
// connection string to NewDBService returns an error without panicking.
func TestNewDBServiceEmptyURLReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := NewDBService(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty connection string, got nil")
	}
}

// TestNewDBServiceInvalidURLFormatReturnsError verifies that a clearly
// malformed URL (no scheme, no host) returns an error from NewDBService.
func TestNewDBServiceInvalidURLFormatReturnsError(t *testing.T) {
	t.Parallel()

	invalidURLs := []string{
		"not-a-postgres-url",
		"http://wrong-scheme/db",
		"://missing-scheme",
		"postgres://",
		"ftp://localhost/db",
	}

	ctx := context.Background()
	for _, url := range invalidURLs {
		url := url
		t.Run(url, func(t *testing.T) {
			t.Parallel()
			_, err := NewDBService(ctx, url)
			if err == nil {
				t.Errorf("NewDBService(%q): expected error for invalid URL, got nil", url)
			}
		})
	}
}

// TestNewDBServiceErrorMessageIsWrapped verifies that the error returned by
// NewDBService for an invalid URL is wrapped (contains ":") so callers can
// inspect the context chain.
func TestNewDBServiceErrorMessageIsWrapped(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := NewDBService(ctx, "not-a-postgres-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, ":") {
		t.Errorf("expected wrapped error message containing ':', got: %q", msg)
	}
}

// TestDBServiceCloseOnMockDoesNotPanic verifies that calling Close on a
// DBBackend that wraps a mock (DBService-like) does not panic. We use mockDB
// here because DBService requires a real Postgres connection; mockDB.Close is
// a no-op that models the same contract.
func TestDBServiceCloseOnMockDoesNotPanic(t *testing.T) {
	t.Parallel()

	db := newMockDB()
	// Close must be callable without panic, even if called multiple times.
	db.Close()
	db.Close()
}

// TestDBBackendInterfaceComplianceMockDB is a compile-time check that mockDB
// satisfies the DBBackend interface.
func TestDBBackendInterfaceComplianceMockDB(t *testing.T) {
	t.Parallel()

	var _ DBBackend = (*mockDB)(nil)
}
