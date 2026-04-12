package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ---------------------------------------------------------------------------
// mockRow implements pgx.Row
//
// vals must match the destination types that the repo passes to Scan.
// For nullable *string columns (problem_type) the val must be *string or nil.
// ---------------------------------------------------------------------------

type mockRow struct {
	vals []any
	err  error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	if len(dest) != len(m.vals) {
		return errors.New("mockRow: scan dest/vals length mismatch")
	}
	for i, d := range dest {
		if err := assignAny(d, m.vals[i]); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// assignAny copies val into dest (a pointer).  Only the types used by the
// repository scan calls need to be handled.
// ---------------------------------------------------------------------------

func assignAny(dest, val any) error {
	switch p := dest.(type) {
	case **string:
		if val == nil {
			*p = nil
		} else {
			s := val.(string)
			*p = &s
		}
	case *string:
		if val == nil {
			*p = ""
		} else {
			*p = val.(string)
		}
	case *time.Time:
		*p = val.(time.Time)
	default:
		// Unknown type — silently skip; the test will verify the result anyway.
	}
	return nil
}

// ---------------------------------------------------------------------------
// mockRows implements pgx.Rows.
//
// Each inner []any slice must match the concrete types that repo.Scan
// expects: strings for string fields, *string for nullable fields, time.Time
// for timestamps.
// ---------------------------------------------------------------------------

type mockRows struct {
	rows   [][]any
	idx    int
	rowErr error // returned by Err()
}

func (m *mockRows) Close() {}

func (m *mockRows) Err() error { return m.rowErr }

func (m *mockRows) CommandTag() pgconn.CommandTag { return pgconn.NewCommandTag("SELECT") }

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (m *mockRows) Next() bool {
	if m.idx < len(m.rows) {
		m.idx++
		return true
	}
	return false
}

func (m *mockRows) Scan(dest ...any) error {
	row := m.rows[m.idx-1]
	if len(dest) != len(row) {
		return errors.New("mockRows: scan dest/row length mismatch")
	}
	for i, d := range dest {
		if err := assignAny(d, row[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockRows) Values() ([]any, error)  { return nil, nil }
func (m *mockRows) RawValues() [][]byte      { return nil }
func (m *mockRows) Conn() *pgx.Conn         { return nil }

// ---------------------------------------------------------------------------
// mockRowsError wraps mockRows and overrides Scan to always return an error.
// ---------------------------------------------------------------------------

type mockRowsError struct {
	mockRows
	scanErr error
}

func (m *mockRowsError) Scan(_ ...any) error { return m.scanErr }

// ---------------------------------------------------------------------------
// mockPool implements pgxPool
// ---------------------------------------------------------------------------

type mockPool struct {
	row      pgx.Row
	rows     pgx.Rows
	queryErr error
	tag      pgconn.CommandTag
	execErr  error
}

func (m *mockPool) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return m.row
}

func (m *mockPool) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return m.rows, m.queryErr
}

func (m *mockPool) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return m.tag, m.execErr
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testTime() time.Time { return time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) }

// sessionRowNullable builds a []any for CreateSession / GetSession scans where
// the problem_type destination is **string (i.e. var pt *string; row.Scan(&pt)).
func sessionRowNullable(id, userID, brand, model string, problemType *string, status string) []any {
	var pt any
	if problemType != nil {
		pt = *problemType
	}
	// We pass pt as the raw string or nil; assignAny handles **string dest correctly.
	return []any{id, userID, brand, model, pt, status, testTime(), testTime()}
}

// sessionRowList builds a []any for ListUserSessions scans where
// problem_type destination is *string (i.e. &rec.ProblemType, a plain string field).
func sessionRowList(id, userID, brand, model, problemType, status string) []any {
	return []any{id, userID, brand, model, problemType, status, testTime(), testTime()}
}

// ---------------------------------------------------------------------------
// Constructor test
// ---------------------------------------------------------------------------

func TestNewSessionRepoNonNil(t *testing.T) {
	repo := NewSessionRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil SessionRepo")
	}
}

// ---------------------------------------------------------------------------
// CreateSession tests
// ---------------------------------------------------------------------------

func TestCreateSession_Success(t *testing.T) {
	pt := "electrical"
	row := &mockRow{vals: sessionRowNullable("sess-1", "user-1", "Apple", "iPhone 14", &pt, "active")}
	repo := &SessionRepo{pool: &mockPool{row: row}}

	rec, err := repo.CreateSession(context.Background(), "user-1", "Apple", "iPhone 14")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "sess-1" {
		t.Errorf("ID: got %q, want %q", rec.ID, "sess-1")
	}
	if rec.UserID != "user-1" {
		t.Errorf("UserID: got %q, want %q", rec.UserID, "user-1")
	}
	if rec.DeviceBrand != "Apple" {
		t.Errorf("DeviceBrand: got %q, want %q", rec.DeviceBrand, "Apple")
	}
	if rec.DeviceModel != "iPhone 14" {
		t.Errorf("DeviceModel: got %q, want %q", rec.DeviceModel, "iPhone 14")
	}
	if rec.ProblemType != "electrical" {
		t.Errorf("ProblemType: got %q, want %q", rec.ProblemType, "electrical")
	}
	if rec.Status != "active" {
		t.Errorf("Status: got %q, want %q", rec.Status, "active")
	}
}

func TestCreateSession_NilProblemType(t *testing.T) {
	row := &mockRow{vals: sessionRowNullable("sess-2", "user-2", "Samsung", "Galaxy S23", nil, "active")}
	repo := &SessionRepo{pool: &mockPool{row: row}}

	rec, err := repo.CreateSession(context.Background(), "user-2", "Samsung", "Galaxy S23")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ProblemType != "" {
		t.Errorf("ProblemType should be empty when nil, got %q", rec.ProblemType)
	}
}

func TestCreateSession_ScanError(t *testing.T) {
	scanErr := errors.New("scan failed")
	row := &mockRow{err: scanErr}
	repo := &SessionRepo{pool: &mockPool{row: row}}

	_, err := repo.CreateSession(context.Background(), "u", "b", "m")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, scanErr) {
		t.Errorf("expected wrapped scanErr, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetSession tests
// ---------------------------------------------------------------------------

func TestGetSession_Success(t *testing.T) {
	pt := "screen"
	row := &mockRow{vals: sessionRowNullable("sess-3", "user-3", "Google", "Pixel 7", &pt, "closed")}
	repo := &SessionRepo{pool: &mockPool{row: row}}

	rec, err := repo.GetSession(context.Background(), "sess-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "sess-3" {
		t.Errorf("ID: got %q, want %q", rec.ID, "sess-3")
	}
	if rec.ProblemType != "screen" {
		t.Errorf("ProblemType: got %q, want %q", rec.ProblemType, "screen")
	}
}

func TestGetSession_ScanError(t *testing.T) {
	scanErr := errors.New("db error")
	row := &mockRow{err: scanErr}
	repo := &SessionRepo{pool: &mockPool{row: row}}

	_, err := repo.GetSession(context.Background(), "sess-3")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, scanErr) {
		t.Errorf("expected wrapped scanErr, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ListUserSessions tests
// ---------------------------------------------------------------------------

func TestListUserSessions_Success(t *testing.T) {
	rows := &mockRows{
		rows: [][]any{
			sessionRowList("s1", "u1", "Apple", "iPhone 13", "", "active"),
			sessionRowList("s2", "u1", "Apple", "iPhone 14", "battery", "closed"),
		},
	}
	repo := &SessionRepo{pool: &mockPool{rows: rows}}

	sessions, err := repo.ListUserSessions(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].ID != "s1" {
		t.Errorf("first session ID: got %q, want %q", sessions[0].ID, "s1")
	}
	if sessions[1].ID != "s2" {
		t.Errorf("second session ID: got %q, want %q", sessions[1].ID, "s2")
	}
	if sessions[1].ProblemType != "battery" {
		t.Errorf("second session ProblemType: got %q, want %q", sessions[1].ProblemType, "battery")
	}
}

func TestListUserSessions_Empty(t *testing.T) {
	rows := &mockRows{rows: [][]any{}}
	repo := &SessionRepo{pool: &mockPool{rows: rows}}

	sessions, err := repo.ListUserSessions(context.Background(), "u-nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestListUserSessions_QueryError(t *testing.T) {
	qErr := errors.New("connection lost")
	repo := &SessionRepo{pool: &mockPool{queryErr: qErr}}

	_, err := repo.ListUserSessions(context.Background(), "u1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, qErr) {
		t.Errorf("expected wrapped queryErr, got: %v", err)
	}
}

func TestListUserSessions_ScanError(t *testing.T) {
	scanErr := errors.New("scan blow-up")
	rows := &mockRowsError{
		mockRows: mockRows{
			rows: [][]any{
				sessionRowList("s1", "u1", "Apple", "iPhone 13", "", "active"),
			},
		},
		scanErr: scanErr,
	}
	repo := &SessionRepo{pool: &mockPool{rows: rows}}

	_, err := repo.ListUserSessions(context.Background(), "u1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, scanErr) {
		t.Errorf("expected wrapped scanErr, got: %v", err)
	}
}

func TestListUserSessions_RowsErr(t *testing.T) {
	rowErr := errors.New("rows iteration error")
	rows := &mockRows{
		rows:   [][]any{},
		rowErr: rowErr,
	}
	repo := &SessionRepo{pool: &mockPool{rows: rows}}

	_, err := repo.ListUserSessions(context.Background(), "u1")
	if err == nil {
		t.Fatal("expected error from rows.Err(), got nil")
	}
	if !errors.Is(err, rowErr) {
		t.Errorf("expected rowErr, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateSessionStatus tests
// ---------------------------------------------------------------------------

func TestUpdateSessionStatus_Success(t *testing.T) {
	pt := "battery"
	repo := &SessionRepo{pool: &mockPool{tag: pgconn.NewCommandTag("UPDATE 1")}}

	err := repo.UpdateSessionStatus(context.Background(), "sess-1", "closed", &pt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateSessionStatus_ExecError(t *testing.T) {
	execErr := errors.New("exec failed")
	repo := &SessionRepo{pool: &mockPool{execErr: execErr}}

	err := repo.UpdateSessionStatus(context.Background(), "sess-1", "closed", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, execErr) {
		t.Errorf("expected wrapped execErr, got: %v", err)
	}
}

func TestUpdateSessionStatus_NotFound(t *testing.T) {
	repo := &SessionRepo{pool: &mockPool{tag: pgconn.NewCommandTag("UPDATE 0")}}

	err := repo.UpdateSessionStatus(context.Background(), "nonexistent", "closed", nil)
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}
