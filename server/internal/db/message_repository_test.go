package db

import (
	"context"
	"testing"
	"time"
)

// TestNewMessageRepoNotNil verifies that NewMessageRepo returns a non-nil
// pointer even when called with a nil pool.
func TestNewMessageRepoNotNil(t *testing.T) {
	repo := NewMessageRepo(nil)
	if repo == nil {
		t.Fatal("expected NewMessageRepo(nil) to return non-nil *MessageRepo")
	}
}

// TestMessageRecordFields verifies that a manually constructed MessageRecord
// stores and exposes all fields correctly.
func TestMessageRecordFields(t *testing.T) {
	now := time.Now()
	rec := MessageRecord{
		ID:          "msg-1",
		SessionID:   "session-42",
		Role:        "user",
		Content:     "What is kinetic energy?",
		MessageType: "text",
		CreatedAt:   now,
	}

	if rec.ID != "msg-1" {
		t.Errorf("ID: want %q, got %q", "msg-1", rec.ID)
	}
	if rec.SessionID != "session-42" {
		t.Errorf("SessionID: want %q, got %q", "session-42", rec.SessionID)
	}
	if rec.Role != "user" {
		t.Errorf("Role: want %q, got %q", "user", rec.Role)
	}
	if rec.Content != "What is kinetic energy?" {
		t.Errorf("Content: want %q, got %q", "What is kinetic energy?", rec.Content)
	}
	if rec.MessageType != "text" {
		t.Errorf("MessageType: want %q, got %q", "text", rec.MessageType)
	}
	if !rec.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: want %v, got %v", now, rec.CreatedAt)
	}
}

// TestMessageRecordZeroValue verifies that the zero value of MessageRecord has
// sensible defaults (empty strings, zero time).
func TestMessageRecordZeroValue(t *testing.T) {
	rec := MessageRecord{}
	if rec.ID != "" {
		t.Errorf("expected empty ID, got %q", rec.ID)
	}
	if rec.SessionID != "" {
		t.Errorf("expected empty SessionID, got %q", rec.SessionID)
	}
	if rec.Role != "" {
		t.Errorf("expected empty Role, got %q", rec.Role)
	}
	if rec.Content != "" {
		t.Errorf("expected empty Content, got %q", rec.Content)
	}
	if rec.MessageType != "" {
		t.Errorf("expected empty MessageType, got %q", rec.MessageType)
	}
	if !rec.CreatedAt.IsZero() {
		t.Errorf("expected zero CreatedAt, got %v", rec.CreatedAt)
	}
}

// TestSaveMessageRequiresPool verifies that calling SaveMessage on a repo
// backed by a nil pool results in either a panic (recovered) or a non-nil
// error — it must never silently succeed.
func TestSaveMessageRequiresPool(t *testing.T) {
	repo := NewMessageRepo(nil)

	didPanic := false
	var returnedErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		_, returnedErr = repo.SaveMessage(context.Background(), "s1", "user", "hello", "text")
	}()

	if !didPanic && returnedErr == nil {
		t.Error("expected SaveMessage with nil pool to panic or return an error")
	}
}

// TestGetSessionMessagesRequiresPool verifies that calling GetSessionMessages
// on a repo backed by a nil pool results in either a panic (recovered) or a
// non-nil error — it must never silently succeed.
func TestGetSessionMessagesRequiresPool(t *testing.T) {
	repo := NewMessageRepo(nil)

	didPanic := false
	var returnedErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		_, returnedErr = repo.GetSessionMessages(context.Background(), "s1")
	}()

	if !didPanic && returnedErr == nil {
		t.Error("expected GetSessionMessages with nil pool to panic or return an error")
	}
}
