// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"testing"
)

// BenchmarkSessionCreate measures the cost of creating a new session per
// iteration, including UUID generation and in-memory map insertion.
func BenchmarkSessionCreate(b *testing.B) {
	b.ReportAllocs()

	svc := NewSessionService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.CreateSession("Creality", "Ender 3", "", "it")
		if err != nil {
			b.Fatalf("CreateSession: %v", err)
		}
	}
}

// BenchmarkSessionGetExisting measures the hot-path read cost of GetSession
// for a single pre-created session. The map and RWMutex overhead is isolated
// from allocation cost — all reads should return without error.
func BenchmarkSessionGetExisting(b *testing.B) {
	b.ReportAllocs()

	svc := NewSessionService()
	sess, err := svc.CreateSession("Prusa", "MK4", "", "it")
	if err != nil {
		b.Fatalf("CreateSession: %v", err)
	}
	sessionID := sess.SessionID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.GetSession(sessionID); err != nil {
			b.Fatalf("GetSession: %v", err)
		}
	}
}

// BenchmarkSessionCRUD measures the complete session lifecycle — create, get,
// delete — in a single iteration. Represents the end-to-end cost for a short
// repair session with no conversation messages.
func BenchmarkSessionCRUD(b *testing.B) {
	b.ReportAllocs()

	svc := NewSessionService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sess, err := svc.CreateSession("Bambu", "X1C", "", "it")
		if err != nil {
			b.Fatalf("CreateSession: %v", err)
		}

		if _, err := svc.GetSession(sess.SessionID); err != nil {
			b.Fatalf("GetSession: %v", err)
		}

		if err := svc.DeleteSession(sess.SessionID); err != nil {
			b.Fatalf("DeleteSession: %v", err)
		}
	}
}

// BenchmarkSessionListWith100 pre-populates 100 sessions and then measures the
// cost of ListSessions, which acquires a read lock, copies every SessionState,
// and returns the slice. Exercises the O(n) snapshot path.
func BenchmarkSessionListWith100(b *testing.B) {
	b.ReportAllocs()

	svc := NewSessionService()

	const sessionCount = 100
	for i := 0; i < sessionCount; i++ {
		brand := fmt.Sprintf("Brand%03d", i)
		model := fmt.Sprintf("Model%03d", i)
		if _, err := svc.CreateSession(brand, model, "", "it"); err != nil {
			b.Fatalf("CreateSession %d: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions := svc.ListSessions()
		if len(sessions) != sessionCount {
			b.Fatalf("ListSessions: expected %d, got %d", sessionCount, len(sessions))
		}
	}
}
