package handlers

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Franck1120/physicscopilot/server/internal/services"
)

// TestSessionStressConcurrentCreate creates 100 sessions concurrently using
// goroutines and a WaitGroup, verifies every session got a unique ID, then
// calls CleanupExpiredSessions(0) and verifies all are removed.
func TestSessionStressConcurrentCreate(t *testing.T) {
	t.Parallel()

	const numSessions = 100

	sessionSvc := services.NewSessionService()
	_ = NewSessionHandler(sessionSvc)

	ids := make([]string, numSessions)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			sess, err := sessionSvc.CreateSession(
				fmt.Sprintf("Brand%d", idx),
				fmt.Sprintf("Model%d", idx),
				"",
				"",
			)
			if err != nil {
				t.Errorf("CreateSession[%d]: %v", idx, err)
				return
			}
			if sess.SessionID == "" {
				t.Errorf("CreateSession[%d]: got empty SessionID", idx)
				return
			}
			mu.Lock()
			ids[idx] = sess.SessionID
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all IDs are unique (no duplicates).
	seen := make(map[string]int, numSessions)
	for idx, id := range ids {
		if id == "" {
			// goroutine already reported this failure via t.Errorf
			continue
		}
		if prev, exists := seen[id]; exists {
			t.Errorf("duplicate session ID %q at index %d and %d", id, prev, idx)
		}
		seen[id] = idx
	}

	// Verify all sessions are present.
	all := sessionSvc.ListSessions()
	if len(all) != numSessions {
		t.Errorf("expected %d sessions, got %d", numSessions, len(all))
	}

	// CleanupExpiredSessions with 0 duration uses cutoff = now, so sessions whose
	// LastActivity is before now are expired. Add a tiny sleep to ensure
	// LastActivity is strictly in the past.
	time.Sleep(time.Millisecond)

	cleaned := sessionSvc.CleanupExpiredSessions(0)
	if cleaned != numSessions {
		t.Errorf("CleanupExpiredSessions: want %d cleaned, got %d", numSessions, cleaned)
	}

	remaining := sessionSvc.ListSessions()
	if len(remaining) != 0 {
		t.Errorf("expected 0 sessions after cleanup, got %d", len(remaining))
	}
}
