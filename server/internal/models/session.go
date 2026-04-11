package models

import "time"

// TODO: Full session model with steps, device info, and AI analysis results

type Session struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	DeviceID        string    `json:"device_id"`
	Status          string    `json:"status"`
	ProblemDetected string    `json:"problem_detected,omitempty"`
	SolutionApplied string    `json:"solution_applied,omitempty"`
	Success         bool      `json:"success"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
}
