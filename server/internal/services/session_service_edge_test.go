package services

import (
	"sync"
	"testing"
	"time"
)

// TestCreateSession1000ConcurrentUniqueIDs verifies that calling CreateSession
// 1000 times concurrently produces 1000 distinct session IDs with no panics.
func TestCreateSession1000ConcurrentUniqueIDs(t *testing.T) {
	t.Parallel()

	svc := NewSessionService()

	const workers = 1000
	ids := make([]string, workers)
	errs := make([]error, workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			sess, err := svc.CreateSession("Creality", "Ender3", "", "it")
			if err != nil {
				errs[i] = err
				return
			}
			ids[i] = sess.SessionID
		}()
	}
	wg.Wait()

	// All creates must have succeeded.
	for i, err := range errs {
		if err != nil {
			t.Errorf("worker %d: unexpected error: %v", i, err)
		}
	}

	// All IDs must be non-empty and unique.
	seen := make(map[string]int, workers)
	for i, id := range ids {
		if id == "" {
			t.Errorf("worker %d: got empty session ID", i)
			continue
		}
		if prev, dup := seen[id]; dup {
			t.Errorf("duplicate session ID %q produced by workers %d and %d", id, prev, i)
		}
		seen[id] = i
	}
}

// TestGetSessionExpiredAfterManualExpiry verifies that a session manually
// backdated beyond the cleanup window is evicted by CleanupExpiredSessions
// and subsequently returns nil from GetSession.
func TestGetSessionExpiredAfterManualExpiry(t *testing.T) {
	t.Parallel()

	svc := NewSessionService()
	sess, err := svc.CreateSession("Prusa", "MK4", "", "it")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Backdate the session so it looks 2 hours old.
	svc.mu.Lock()
	svc.sessions[sess.SessionID].LastActivity = time.Now().Add(-2 * time.Hour)
	svc.mu.Unlock()

	removed := svc.CleanupExpiredSessions(1 * time.Hour)
	if removed != 1 {
		t.Errorf("expected 1 session removed, got %d", removed)
	}

	retrieved, err := svc.GetSession(sess.SessionID)
	if err == nil {
		t.Errorf("expected error for expired session, got session: %+v", retrieved)
	}
}

// TestCleanupExpiredSessionsZeroMaxAgeEvictsEverything verifies that a
// zero-duration maxAge treats every session as expired and removes them all.
func TestCleanupExpiredSessionsZeroMaxAgeEvictsEverything(t *testing.T) {
	t.Parallel()

	svc := NewSessionService()

	const n = 5
	for i := 0; i < n; i++ {
		if _, err := svc.CreateSession("Bambu", "X1C", "", "it"); err != nil {
			t.Fatalf("CreateSession %d: %v", i, err)
		}
	}

	// With maxAge=0, cutoff = time.Now() — sessions created even 1 ns ago
	// have LastActivity.Before(cutoff), so all must be evicted.
	// Sleep 1 ms to ensure LastActivity < cutoff.
	time.Sleep(1 * time.Millisecond)

	removed := svc.CleanupExpiredSessions(0)
	if removed != n {
		t.Errorf("expected %d sessions removed with maxAge=0, got %d", n, removed)
	}

	remaining := svc.ListSessions()
	if len(remaining) != 0 {
		t.Errorf("expected 0 sessions remaining after zero-maxAge cleanup, got %d", len(remaining))
	}
}

// TestListSessionsConsistentCount verifies that ListSessions returns the
// correct count after a series of creates and deletes, and that each
// returned entry carries a non-empty SessionID.
func TestListSessionsConsistentCount(t *testing.T) {
	t.Parallel()

	svc := NewSessionService()

	// Create 6 sessions.
	ids := make([]string, 6)
	for i := range ids {
		sess, err := svc.CreateSession("Prusa", "MK4", "", "it")
		if err != nil {
			t.Fatalf("CreateSession %d: %v", i, err)
		}
		ids[i] = sess.SessionID
	}

	// Delete 2 of them.
	for _, id := range ids[:2] {
		if err := svc.DeleteSession(id); err != nil {
			t.Fatalf("DeleteSession %q: %v", id, err)
		}
	}

	all := svc.ListSessions()
	if len(all) != 4 {
		t.Errorf("expected 4 sessions, got %d", len(all))
	}

	// Verify every returned entry has a non-empty ID.
	for i, s := range all {
		if s.SessionID == "" {
			t.Errorf("ListSessions entry %d has empty SessionID", i)
		}
	}

	// Deleted IDs must not appear.
	deletedSet := map[string]bool{ids[0]: true, ids[1]: true}
	for _, s := range all {
		if deletedSet[s.SessionID] {
			t.Errorf("deleted session %q still in ListSessions", s.SessionID)
		}
	}
}

// TestGetSessionSnapshotIsolation verifies that modifying the original session
// state does not affect a snapshot already obtained.
func TestGetSessionSnapshotIsolation(t *testing.T) {
	t.Parallel()

	svc := NewSessionService()
	sess, _ := svc.CreateSession("Bambu", "A1", "", "it")
	_ = svc.UpdateStep(sess.SessionID, 2, 8)

	snapshot, err := svc.GetSessionSnapshot(sess.SessionID)
	if err != nil {
		t.Fatalf("GetSessionSnapshot: %v", err)
	}

	// Mutate the live session.
	_ = svc.UpdateStep(sess.SessionID, 99, 99)

	if snapshot.CurrentStep != 2 {
		t.Errorf("snapshot.CurrentStep was mutated: expected 2, got %d", snapshot.CurrentStep)
	}
	if snapshot.TotalSteps != 8 {
		t.Errorf("snapshot.TotalSteps was mutated: expected 8, got %d", snapshot.TotalSteps)
	}
}
