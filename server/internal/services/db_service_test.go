package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// mockDB — in-memory DBBackend for unit tests
// ---------------------------------------------------------------------------

type mockDB struct {
	sessions  map[string]*SessionState
	steps     []struct{ sid string; num int; instr string }
	pingErr   error
	saveErr   error
	deleteErr error
	listErr   error
}

func newMockDB() *mockDB {
	return &mockDB{sessions: make(map[string]*SessionState)}
}

func (m *mockDB) SaveSession(_ context.Context, s *SessionState) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	cp := *s
	m.sessions[s.SessionID] = &cp
	return nil
}

func (m *mockDB) DeleteSession(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sessions, id)
	return nil
}

func (m *mockDB) ListSessions(_ context.Context) ([]SessionState, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	out := make([]SessionState, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, *s)
	}
	return out, nil
}

func (m *mockDB) SaveSessionStep(_ context.Context, sid string, num int, instr string) error {
	m.steps = append(m.steps, struct{ sid string; num int; instr string }{sid, num, instr})
	return nil
}

func (m *mockDB) SaveFeedback(_ context.Context, _ *FeedbackEntry) error  { return nil }
func (m *mockDB) ExpireSession(_ context.Context, id string) error        { delete(m.sessions, id); return nil }
func (m *mockDB) Ping(_ context.Context) error                            { return m.pingErr }
func (m *mockDB) Close()                                                   {}

// ---------------------------------------------------------------------------
// SessionService ↔ DBBackend integration
// ---------------------------------------------------------------------------

func TestSessionServiceSyncsCreateToDB(t *testing.T) {
	db := newMockDB()
	svc := NewSessionService()
	svc.SetDB(db)

	sess, err := svc.CreateSession("Apple", "iPhone 15", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, ok := db.sessions[sess.SessionID]; !ok {
		t.Error("expected session to be saved in mock DB after create")
	}
}

func TestSessionServiceSyncsDeleteToDB(t *testing.T) {
	db := newMockDB()
	svc := NewSessionService()
	svc.SetDB(db)

	sess, _ := svc.CreateSession("Samsung", "Galaxy S24", "")
	if err := svc.DeleteSession(sess.SessionID); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	if _, ok := db.sessions[sess.SessionID]; ok {
		t.Error("expected session to be removed from mock DB after delete")
	}
}

func TestSessionServiceDBWriteErrorDoesNotFailInMemory(t *testing.T) {
	// Even when the DB write fails, the in-memory store must succeed.
	db := newMockDB()
	db.saveErr = fmt.Errorf("connection refused")
	svc := NewSessionService()
	svc.SetDB(db)

	sess, err := svc.CreateSession("Google", "Pixel 9", "")
	if err != nil {
		t.Fatalf("expected in-memory create to succeed despite DB error, got: %v", err)
	}
	// Session must be retrievable from memory.
	if _, err := svc.GetSession(sess.SessionID); err != nil {
		t.Errorf("GetSession after failed DB write: %v", err)
	}
}

func TestSessionServiceHydrateFromDB(t *testing.T) {
	db := newMockDB()
	svc := NewSessionService()
	svc.SetDB(db)

	// Pre-populate the mock DB directly (simulates a restart scenario).
	now := time.Now()
	db.sessions["pre-existing-id"] = &SessionState{
		SessionID:    "pre-existing-id",
		DeviceInfo:   DeviceInfo{Brand: "Bambu", Model: "X1C"},
		CreatedAt:    now,
		LastActivity: now,
	}

	if err := svc.HydrateFromDB(context.Background()); err != nil {
		t.Fatalf("HydrateFromDB: %v", err)
	}

	if _, err := svc.GetSession("pre-existing-id"); err != nil {
		t.Errorf("session from DB not loaded into memory: %v", err)
	}
}

func TestSessionServiceHydrateNoopWhenNoDB(t *testing.T) {
	svc := NewSessionService() // no DB set
	if err := svc.HydrateFromDB(context.Background()); err != nil {
		t.Errorf("HydrateFromDB should be a no-op when db is nil, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DBService — real Postgres (skipped unless TEST_DATABASE_URL is set)
// ---------------------------------------------------------------------------

func TestDBServicePing_RealDB(t *testing.T) {
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping real DB test")
	}

	ctx := context.Background()
	svc, err := NewDBService(ctx, connStr)
	if err != nil {
		t.Fatalf("NewDBService: %v", err)
	}
	defer svc.Close()

	if err := svc.Ping(ctx); err != nil {
		t.Errorf("Ping: %v", err)
	}
}

func TestDBServiceNewDBService_InvalidConnString(t *testing.T) {
	// A clearly malformed URL must fail without panicking.
	ctx := context.Background()
	_, err := NewDBService(ctx, "not-a-postgres-url")
	if err == nil {
		t.Error("expected error for invalid connection string")
	}
}

// ---------------------------------------------------------------------------
// scanSessionRow helper
// ---------------------------------------------------------------------------

// fakeRow is a minimal rowScanner for unit-testing scanSessionRow.
type fakeRow struct{ vals []any; err error }

func (f *fakeRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	for i, d := range dest {
		switch v := d.(type) {
		case *string:
			*v = f.vals[i].(string)
		case *time.Time:
			*v = f.vals[i].(time.Time)
		}
	}
	return nil
}

func TestScanSessionRow_Happy(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	row := &fakeRow{vals: []any{
		"sess-1", "Apple", "iPhone", "stringing", now, now,
	}}
	s, err := scanSessionRow(row)
	if err != nil {
		t.Fatalf("scanSessionRow: %v", err)
	}
	if s.SessionID != "sess-1" {
		t.Errorf("SessionID: want 'sess-1', got %q", s.SessionID)
	}
	if s.DeviceInfo.Brand != "Apple" {
		t.Errorf("Brand: want 'Apple', got %q", s.DeviceInfo.Brand)
	}
	if s.ProblemDetected != "stringing" {
		t.Errorf("ProblemDetected: want 'stringing', got %q", s.ProblemDetected)
	}
}

func TestScanSessionRow_Error(t *testing.T) {
	row := &fakeRow{err: fmt.Errorf("scan error")}
	if _, err := scanSessionRow(row); err == nil {
		t.Error("expected error from scanSessionRow on scan failure")
	}
}

// ---------------------------------------------------------------------------
// Connection pool config tests
// ---------------------------------------------------------------------------

func TestDBServicePoolConstants(t *testing.T) {
	// Verify the tuning constants are set to the required values.
	if poolMaxConns != 10 {
		t.Errorf("poolMaxConns: want 10, got %d", poolMaxConns)
	}
	if poolMinConns != 2 {
		t.Errorf("poolMinConns: want 2, got %d", poolMinConns)
	}
	if poolMaxConnLifetime != 1*time.Hour {
		t.Errorf("poolMaxConnLifetime: want 1h, got %v", poolMaxConnLifetime)
	}
}

func TestDBServicePoolConfig_RealDB(t *testing.T) {
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping real DB pool test")
	}

	ctx := context.Background()
	svc, err := NewDBService(ctx, connStr)
	if err != nil {
		t.Fatalf("NewDBService: %v", err)
	}
	defer svc.Close()

	stats := svc.PoolStats()

	if stats.MaxConns != poolMaxConns {
		t.Errorf("MaxConns: want %d, got %d", poolMaxConns, stats.MaxConns)
	}
	if stats.TotalConns < 1 {
		t.Error("expected at least 1 total connection after ping")
	}
}

// ---------------------------------------------------------------------------
// Session expiry tests
// ---------------------------------------------------------------------------

func TestCleanupExpiredSessionsMarksDBExpired(t *testing.T) {
	db := newMockDB()
	svc := NewSessionService()
	svc.SetDB(db)

	// Create two sessions; make one look old by pushing LastActivity back.
	old, _ := svc.CreateSession("Bambu", "A1", "")
	fresh, _ := svc.CreateSession("Prusa", "MK4", "")

	// Manipulate LastActivity directly via the in-memory pointer.
	svc.mu.Lock()
	svc.sessions[old.SessionID].LastActivity = time.Now().Add(-2 * time.Hour)
	svc.mu.Unlock()

	n := svc.CleanupExpiredSessions(1 * time.Hour)

	if n != 1 {
		t.Errorf("expected 1 session cleaned, got %d", n)
	}
	// Old session must be gone from memory and from mockDB.
	if _, err := svc.GetSession(old.SessionID); err == nil {
		t.Error("expired session should have been removed from memory")
	}
	if _, exists := db.sessions[old.SessionID]; exists {
		t.Error("expired session should have been removed from mockDB")
	}
	// Fresh session must still be alive.
	if _, err := svc.GetSession(fresh.SessionID); err != nil {
		t.Errorf("fresh session should still be in memory: %v", err)
	}
}

func TestCleanupExpiredSessionsNoOp(t *testing.T) {
	svc := NewSessionService()
	svc.CreateSession("Apple", "iPhone", "") //nolint:errcheck
	n := svc.CleanupExpiredSessions(30 * time.Minute)
	if n != 0 {
		t.Errorf("expected 0 cleanups for fresh sessions, got %d", n)
	}
}
