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
