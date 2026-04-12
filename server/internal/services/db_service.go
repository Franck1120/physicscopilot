package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBBackend is the persistence interface used by SessionService for optional
// Postgres write-through. Implementors: DBService (production) and any mock
// in tests.
//
// All methods accept a context so callers can enforce timeouts. Errors from
// the DB layer are non-fatal for the in-memory store: SessionService logs
// them at WARN and continues serving from memory.
type DBBackend interface {
	// SaveSession upserts the session record (create or update).
	SaveSession(ctx context.Context, s *SessionState) error
	// DeleteSession soft-deletes the session (status → 'completed').
	DeleteSession(ctx context.Context, id string) error
	// ListSessions returns all active sessions, newest first.
	ListSessions(ctx context.Context) ([]SessionState, error)
	// SaveSessionStep upserts a step record in session_steps.
	SaveSessionStep(ctx context.Context, sessionID string, stepNumber int, instruction string) error
	// SaveFeedback persists a user feedback entry for a given session step.
	SaveFeedback(ctx context.Context, f *FeedbackEntry) error
	// ExpireSession marks the session status as 'expired' in Postgres.
	// Called by the cleanup goroutine for sessions that timed out.
	ExpireSession(ctx context.Context, id string) error
	// Ping returns nil when the database is reachable.
	Ping(ctx context.Context) error
	// Close releases the connection pool.
	Close()
}

// FeedbackEntry holds a single user feedback record.
type FeedbackEntry struct {
	SessionID  string  `json:"session_id"`
	StepNumber int     `json:"step_number"`
	Rating     string  `json:"rating"` // "positive" or "negative"
	Comment    *string `json:"comment,omitempty"`
}

// DBService implements DBBackend using a pgx/v5 connection pool.
// Construct via NewDBService; call Close() on server shutdown.
type DBService struct {
	pool *pgxpool.Pool
}

// Pool tuning constants.
const (
	poolMaxConns        = 10
	poolMinConns        = 2
	poolMaxConnLifetime = 1 * time.Hour
)

// NewDBService opens a pgx pool for connString and pings the server.
// Returns an error if the connection string is malformed or the initial
// ping fails (e.g. host unreachable, bad credentials).
// Pool is configured with MaxConns=10, MinConns=2, MaxConnLifetime=1h.
func NewDBService(ctx context.Context, connString string) (*DBService, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse db pool config: %w", err)
	}
	cfg.MaxConns = poolMaxConns
	cfg.MinConns = poolMinConns
	cfg.MaxConnLifetime = poolMaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open db pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &DBService{pool: pool}, nil
}

// DBPoolStats holds a snapshot of connection pool metrics.
type DBPoolStats struct {
	TotalConns    int32 `json:"total_conns"`
	IdleConns     int32 `json:"idle_conns"`
	AcquiredConns int32 `json:"acquired_conns"`
	MaxConns      int32 `json:"max_conns"`
}

// PoolStats returns a snapshot of the current connection pool state.
func (d *DBService) PoolStats() DBPoolStats {
	s := d.pool.Stat()
	return DBPoolStats{
		TotalConns:    s.TotalConns(),
		IdleConns:     s.IdleConns(),
		AcquiredConns: s.AcquiredConns(),
		MaxConns:      s.MaxConns(),
	}
}

// Ping checks the database connection is alive.
func (d *DBService) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// Close releases all pool connections. Must be called on shutdown.
func (d *DBService) Close() {
	d.pool.Close()
}

// SaveSession upserts a session row.
// Columns device_brand, device_model, and last_activity are required by the
// migration 20260413000000_add_session_device_fields.
func (d *DBService) SaveSession(ctx context.Context, s *SessionState) error {
	t0 := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("save_session").Observe(time.Since(t0).Seconds()) }()

	_, err := d.pool.Exec(ctx, `
		INSERT INTO sessions
			(id, device_brand, device_model, status, problem_detected, created_at, last_activity)
		VALUES ($1, $2, $3, 'active', $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			problem_detected = EXCLUDED.problem_detected,
			last_activity    = EXCLUDED.last_activity
	`,
		s.SessionID,
		s.DeviceInfo.Brand,
		s.DeviceInfo.Model,
		s.ProblemDetected,
		s.CreatedAt,
		s.LastActivity,
	)
	if err != nil {
		return fmt.Errorf("save session %q: %w", s.SessionID, err)
	}
	return nil
}

// GetSession retrieves a single active session by ID.
// Returns a wrapped pgx.ErrNoRows error when not found.
func (d *DBService) GetSession(ctx context.Context, id string) (*SessionState, error) {
	row := d.pool.QueryRow(ctx, `
		SELECT id, device_brand, device_model, problem_detected, created_at, last_activity
		FROM   sessions
		WHERE  id = $1 AND status = 'active'
	`, id)

	s, err := scanSessionRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("session %q not found", id)
		}
		return nil, fmt.Errorf("get session %q: %w", id, err)
	}
	return s, nil
}

// ListSessions returns all active sessions ordered newest-first.
func (d *DBService) ListSessions(ctx context.Context) ([]SessionState, error) {
	t0 := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("list_sessions").Observe(time.Since(t0).Seconds()) }()

	rows, err := d.pool.Query(ctx, `
		SELECT id, device_brand, device_model, problem_detected, created_at, last_activity
		FROM   sessions
		WHERE  status = 'active'
		ORDER  BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionState
	for rows.Next() {
		s, err := scanSessionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		sessions = append(sessions, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	return sessions, nil
}

// ExpireSession marks the session status as 'expired'.
// Used by the background cleanup goroutine; no error is returned when the
// row is already non-active (idempotent for concurrent cleanup calls).
func (d *DBService) ExpireSession(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE sessions
		SET    status = 'expired', last_activity = $2
		WHERE  id = $1 AND status = 'active'
	`, id, time.Now())
	if err != nil {
		return fmt.Errorf("expire session %q: %w", id, err)
	}
	return nil
}

// DeleteSession soft-deletes the session (status → 'completed').
// Returns an error if no active session with that ID exists.
func (d *DBService) DeleteSession(ctx context.Context, id string) error {
	t0 := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("delete_session").Observe(time.Since(t0).Seconds()) }()

	tag, err := d.pool.Exec(ctx, `
		UPDATE sessions
		SET    status = 'completed', last_activity = $2
		WHERE  id = $1 AND status = 'active'
	`, id, time.Now())
	if err != nil {
		return fmt.Errorf("delete session %q: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session %q not found", id)
	}
	return nil
}

// SaveSessionStep upserts a step into session_steps.
// The UNIQUE constraint on (session_id, step_number) is defined in migration
// 20260412000000_add_constraints.sql; conflicts update the instruction field.
func (d *DBService) SaveSessionStep(ctx context.Context, sessionID string, stepNumber int, instruction string) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO session_steps (session_id, step_number, instruction, verified, created_at)
		VALUES ($1, $2, $3, false, now())
		ON CONFLICT (session_id, step_number) DO UPDATE SET
			instruction = EXCLUDED.instruction
	`, sessionID, stepNumber, instruction)
	if err != nil {
		return fmt.Errorf("save step %d for session %q: %w", stepNumber, sessionID, err)
	}
	return nil
}

// SaveFeedback inserts a feedback row. comment may be nil for anonymous ratings.
func (d *DBService) SaveFeedback(ctx context.Context, f *FeedbackEntry) error {
	t0 := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("save_feedback").Observe(time.Since(t0).Seconds()) }()

	_, err := d.pool.Exec(ctx, `
		INSERT INTO feedback (session_id, step_number, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, now())
	`, f.SessionID, f.StepNumber, f.Rating, f.Comment)
	if err != nil {
		return fmt.Errorf("save feedback for session %q step %d: %w", f.SessionID, f.StepNumber, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// rowScanner is satisfied by both *pgx.Row and pgx.Rows, allowing scanSessionRow
// to be called from both QueryRow (single) and Query (iteration) paths.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSessionRow(r rowScanner) (*SessionState, error) {
	var s SessionState
	err := r.Scan(
		&s.SessionID,
		&s.DeviceInfo.Brand,
		&s.DeviceInfo.Model,
		&s.ProblemDetected,
		&s.CreatedAt,
		&s.LastActivity,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
