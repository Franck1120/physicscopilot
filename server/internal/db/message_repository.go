// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageRecord mirrors the messages table row.
type MessageRecord struct {
	ID          string
	SessionID   string
	Role        string
	Content     string
	MessageType string
	CreatedAt   time.Time
}

// MessageRepo handles persistence for conversation messages.
type MessageRepo struct {
	pool pgxPool
}

// NewMessageRepo creates a MessageRepo backed by the given pool.
// Passing nil results in a nil interface pool field (panics at query time).
func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	if pool == nil {
		return &MessageRepo{}
	}
	return &MessageRepo{pool: pool}
}

// SaveMessage inserts a new message row.
func (r *MessageRepo) SaveMessage(ctx context.Context, sessionID, role, content, messageType string) (*MessageRecord, error) {
	const q = `
		INSERT INTO messages (session_id, role, content, message_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, session_id, role, content, message_type, created_at`

	rec := &MessageRecord{}
	err := r.pool.QueryRow(ctx, q, sessionID, role, content, messageType).
		Scan(&rec.ID, &rec.SessionID, &rec.Role, &rec.Content, &rec.MessageType, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("save message for session %s: %w", sessionID, err)
	}
	return rec, nil
}

// GetSessionMessages returns all messages for a session, ordered by creation time ascending.
func (r *MessageRepo) GetSessionMessages(ctx context.Context, sessionID string) ([]*MessageRecord, error) {
	const q = `
		SELECT id, session_id, role, content, message_type, created_at
		FROM messages WHERE session_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get messages for session %s: %w", sessionID, err)
	}
	defer rows.Close()

	var msgs []*MessageRecord
	for rows.Next() {
		rec := &MessageRecord{}
		if err := rows.Scan(&rec.ID, &rec.SessionID, &rec.Role, &rec.Content, &rec.MessageType, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		msgs = append(msgs, rec)
	}
	return msgs, rows.Err()
}
