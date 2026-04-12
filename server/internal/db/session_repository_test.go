package db

import (
	"testing"
)

func TestNewSessionRepoNonNil(t *testing.T) {
	// Verify constructor returns a non-nil repo even with a nil pool.
	// Actual DB operations would panic, but the constructor itself must succeed.
	repo := NewSessionRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil SessionRepo")
	}
}

func TestNewMessageRepoNonNil(t *testing.T) {
	repo := NewMessageRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil MessageRepo")
	}
}

// TestNewSessionRepoFields verifies that passing a nil pool to NewSessionRepo
// does not panic at construction time (panics only at query time).
func TestNewSessionRepoFields(t *testing.T) {
	repo := NewSessionRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
	// The pool field should be nil since we passed nil.
	if repo.pool != nil {
		t.Error("expected nil pool field for nil-pool repo")
	}
}

// TestNewMessageRepoFields verifies that the pool field is correctly stored.
func TestNewMessageRepoFields(t *testing.T) {
	repo := NewMessageRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
	if repo.pool != nil {
		t.Error("expected nil pool field for nil-pool repo")
	}
}
