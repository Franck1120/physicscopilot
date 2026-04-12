package db

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Constructor test
// ---------------------------------------------------------------------------

func TestNewMessageRepoNonNil(t *testing.T) {
	repo := NewMessageRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil MessageRepo")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func msgRow(id, sessionID, role, content, msgType string) []any {
	return []any{id, sessionID, role, content, msgType, testTime()}
}

// ---------------------------------------------------------------------------
// SaveMessage tests
// ---------------------------------------------------------------------------

func TestSaveMessage_Success(t *testing.T) {
	row := &mockRow{
		vals: msgRow("msg-1", "sess-1", "user", "Hello world", "text"),
	}
	repo := &MessageRepo{pool: &mockPool{row: row}}

	rec, err := repo.SaveMessage(context.Background(), "sess-1", "user", "Hello world", "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "msg-1" {
		t.Errorf("ID: got %q, want %q", rec.ID, "msg-1")
	}
	if rec.SessionID != "sess-1" {
		t.Errorf("SessionID: got %q, want %q", rec.SessionID, "sess-1")
	}
	if rec.Role != "user" {
		t.Errorf("Role: got %q, want %q", rec.Role, "user")
	}
	if rec.Content != "Hello world" {
		t.Errorf("Content: got %q, want %q", rec.Content, "Hello world")
	}
	if rec.MessageType != "text" {
		t.Errorf("MessageType: got %q, want %q", rec.MessageType, "text")
	}
	if rec.CreatedAt != testTime() {
		t.Errorf("CreatedAt: got %v, want %v", rec.CreatedAt, testTime())
	}
}

func TestSaveMessage_ScanError(t *testing.T) {
	scanErr := errors.New("scan error")
	row := &mockRow{err: scanErr}
	repo := &MessageRepo{pool: &mockPool{row: row}}

	_, err := repo.SaveMessage(context.Background(), "sess-1", "user", "hello", "text")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, scanErr) {
		t.Errorf("expected wrapped scanErr, got: %v", err)
	}
}

func TestSaveMessage_AssistantRole(t *testing.T) {
	row := &mockRow{
		vals: msgRow("msg-2", "sess-1", "assistant", "Sure, I can help.", "text"),
	}
	repo := &MessageRepo{pool: &mockPool{row: row}}

	rec, err := repo.SaveMessage(context.Background(), "sess-1", "assistant", "Sure, I can help.", "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Role != "assistant" {
		t.Errorf("Role: got %q, want %q", rec.Role, "assistant")
	}
}

// ---------------------------------------------------------------------------
// GetSessionMessages tests
// ---------------------------------------------------------------------------

func TestGetSessionMessages_Success(t *testing.T) {
	rows := &mockRows{
		rows: [][]any{
			msgRow("m1", "sess-1", "user", "What is wrong?", "text"),
			msgRow("m2", "sess-1", "assistant", "Checking now.", "text"),
		},
	}
	repo := &MessageRepo{pool: &mockPool{rows: rows}}

	msgs, err := repo.GetSessionMessages(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].ID != "m1" {
		t.Errorf("first message ID: got %q, want %q", msgs[0].ID, "m1")
	}
	if msgs[0].Role != "user" {
		t.Errorf("first message role: got %q, want %q", msgs[0].Role, "user")
	}
	if msgs[1].ID != "m2" {
		t.Errorf("second message ID: got %q, want %q", msgs[1].ID, "m2")
	}
	if msgs[1].Role != "assistant" {
		t.Errorf("second message role: got %q, want %q", msgs[1].Role, "assistant")
	}
}

func TestGetSessionMessages_Empty(t *testing.T) {
	rows := &mockRows{rows: [][]any{}}
	repo := &MessageRepo{pool: &mockPool{rows: rows}}

	msgs, err := repo.GetSessionMessages(context.Background(), "sess-empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestGetSessionMessages_QueryError(t *testing.T) {
	qErr := errors.New("query error")
	repo := &MessageRepo{pool: &mockPool{queryErr: qErr}}

	_, err := repo.GetSessionMessages(context.Background(), "sess-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, qErr) {
		t.Errorf("expected wrapped queryErr, got: %v", err)
	}
}

func TestGetSessionMessages_ScanError(t *testing.T) {
	scanErr := errors.New("scan error")
	rows := &mockRowsError{
		mockRows: mockRows{
			rows: [][]any{
				msgRow("m1", "sess-1", "user", "hello", "text"),
			},
		},
		scanErr: scanErr,
	}
	repo := &MessageRepo{pool: &mockPool{rows: rows}}

	_, err := repo.GetSessionMessages(context.Background(), "sess-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, scanErr) {
		t.Errorf("expected wrapped scanErr, got: %v", err)
	}
}

func TestGetSessionMessages_RowsErr(t *testing.T) {
	rowErr := errors.New("rows error")
	rows := &mockRows{
		rows:   [][]any{},
		rowErr: rowErr,
	}
	repo := &MessageRepo{pool: &mockPool{rows: rows}}

	_, err := repo.GetSessionMessages(context.Background(), "sess-1")
	if err == nil {
		t.Fatal("expected error from rows.Err(), got nil")
	}
	if !errors.Is(err, rowErr) {
		t.Errorf("expected rowErr, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestSaveMessage_ImageType(t *testing.T) {
	row := &mockRow{
		vals: msgRow("msg-3", "sess-2", "user", "base64data...", "image"),
	}
	repo := &MessageRepo{pool: &mockPool{row: row}}

	rec, err := repo.SaveMessage(context.Background(), "sess-2", "user", "base64data...", "image")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.MessageType != "image" {
		t.Errorf("MessageType: got %q, want %q", rec.MessageType, "image")
	}
}

func TestGetSessionMessages_TimestampOrder(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)

	rows := &mockRows{
		rows: [][]any{
			{"m1", "sess-1", "user", "first", "text", t1},
			{"m2", "sess-1", "assistant", "second", "text", t2},
		},
	}
	repo := &MessageRepo{pool: &mockPool{rows: rows}}

	msgs, err := repo.GetSessionMessages(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !msgs[0].CreatedAt.Before(msgs[1].CreatedAt) {
		t.Errorf("expected msgs[0].CreatedAt < msgs[1].CreatedAt")
	}
}
