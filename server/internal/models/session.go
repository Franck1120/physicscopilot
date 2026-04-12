// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import "time"

// SessionStatus represents the lifecycle state of a repair session.
// Maps to the `status` column in the `sessions` table.
type SessionStatus string

const (
	// SessionStatusActive is set when the session is ongoing.
	SessionStatusActive SessionStatus = "active"
	// SessionStatusCompleted is set when the repair was finished successfully.
	SessionStatusCompleted SessionStatus = "completed"
	// SessionStatusAbandoned is set when the session was closed without resolution.
	SessionStatusAbandoned SessionStatus = "abandoned"
)

// Session represents a repair or maintenance session for a specific device.
// Maps directly to the `sessions` table in Supabase.
type Session struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	DeviceID        *string       `json:"device_id,omitempty"`   // nullable — device may be unlinked
	Status          SessionStatus `json:"status"`
	ProblemDetected *string       `json:"problem_detected,omitempty"` // nullable
	SolutionApplied *string       `json:"solution_applied,omitempty"` // nullable
	Success         *bool         `json:"success,omitempty"`          // nullable
	DurationSeconds *int          `json:"duration_seconds,omitempty"` // nullable
	CreatedAt       time.Time     `json:"created_at"`
}

// IsActive reports whether the session is still in progress.
func (s Session) IsActive() bool { return s.Status == SessionStatusActive }

// SessionStep represents a single AI-generated instruction within a session.
// Maps directly to the `session_steps` table in Supabase.
type SessionStep struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	StepNumber  int       `json:"step_number"`
	Instruction string    `json:"instruction"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
}
