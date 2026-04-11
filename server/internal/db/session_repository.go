package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRecord mirrors the sessions table row.
type SessionRecord struct {
	ID          string
	UserID      string
	DeviceBrand string
	DeviceModel string
	ProblemType string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SessionRepo handles persistence for repair sessions.
type SessionRepo struct {
	pool *pgxpool.Pool
}

// NewSessionRepo creates a SessionRepo backed by the given pool.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

// CreateSession inserts a new session row and returns the created record.
func (r *SessionRepo) CreateSession(ctx context.Context, userID, deviceBrand, deviceModel string) (*SessionRecord, error) {
	const q = `
		INSERT INTO sessions (user_id, device_brand, device_model, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id, user_id, device_brand, device_model, problem_type, status, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q, userID, deviceBrand, deviceModel)
	return scanSession(row)
}

// GetSession retrieves a session by its UUID string ID.
func (r *SessionRepo) GetSession(ctx context.Context, sessionID string) (*SessionRecord, error) {
	const q = `
		SELECT id, user_id, device_brand, device_model, problem_type, status, created_at, updated_at
		FROM sessions WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, sessionID)
	rec, err := scanSession(row)
	if err != nil {
		return nil, fmt.Errorf("get session %s: %w", sessionID, err)
	}
	return rec, nil
}

// ListUserSessions returns all sessions for a given userID, ordered by creation time descending.
func (r *SessionRepo) ListUserSessions(ctx context.Context, userID string) ([]*SessionRecord, error) {
	const q = `
		SELECT id, user_id, device_brand, device_model, problem_type, status, created_at, updated_at
		FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list sessions for user %s: %w", userID, err)
	}
	defer rows.Close()

	var sessions []*SessionRecord
	for rows.Next() {
		rec := &SessionRecord{}
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.DeviceBrand, &rec.DeviceModel,
			&rec.ProblemType, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		sessions = append(sessions, rec)
	}
	return sessions, rows.Err()
}

// UpdateSessionStatus updates the status and problem_type of a session.
func (r *SessionRepo) UpdateSessionStatus(ctx context.Context, sessionID, status, problemType string) error {
	const q = `
		UPDATE sessions SET status = $2, problem_type = $3, updated_at = NOW()
		WHERE id = $1`

	ct, err := r.pool.Exec(ctx, q, sessionID, status, problemType)
	if err != nil {
		return fmt.Errorf("update session %s: %w", sessionID, err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("session %s not found", sessionID)
	}
	return nil
}

// scanSession scans a single session row. It handles nullable problem_type.
func scanSession(row interface{ Scan(...any) error }) (*SessionRecord, error) {
	rec := &SessionRecord{}
	var problemType *string
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.DeviceBrand, &rec.DeviceModel,
		&problemType, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	if problemType != nil {
		rec.ProblemType = *problemType
	}
	return rec, nil
}
