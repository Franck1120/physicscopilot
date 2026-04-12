package models

import (
	"encoding/json"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Device
// ---------------------------------------------------------------------------

func TestDeviceJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	d := Device{
		ID:        "dev-1",
		UserID:    "user-1",
		Brand:     "Apple",
		Model:     "iPhone 15",
		CreatedAt: now,
	}

	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Device
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != d.ID {
		t.Errorf("ID: want %q, got %q", d.ID, got.ID)
	}
	if got.Brand != d.Brand {
		t.Errorf("Brand: want %q, got %q", d.Brand, got.Brand)
	}
	if got.Model != d.Model {
		t.Errorf("Model: want %q, got %q", d.Model, got.Model)
	}
}

func TestDeviceDisplayName(t *testing.T) {
	tests := []struct {
		brand, model, want string
	}{
		{"Apple", "iPhone 15", "Apple iPhone 15"},
		{"", "MK4", " MK4"},
		{"Prusa", "", "Prusa "},
		{"", "", " "},
	}
	for _, tc := range tests {
		d := Device{Brand: tc.brand, Model: tc.model}
		if got := d.DisplayName(); got != tc.want {
			t.Errorf("DisplayName(%q, %q): want %q, got %q", tc.brand, tc.model, tc.want, got)
		}
	}
}

func TestDeviceJSONOmitsNoFields(t *testing.T) {
	// All fields must appear in JSON output (none are omitempty).
	d := Device{ID: "x", UserID: "u", Brand: "B", Model: "M", CreatedAt: time.Now()}
	b, _ := json.Marshal(d)

	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck

	for _, key := range []string{"id", "user_id", "brand", "model", "created_at"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Session
// ---------------------------------------------------------------------------

func TestSessionJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	prob := "stringing"
	s := Session{
		ID:              "sess-1",
		UserID:          "user-1",
		Status:          SessionStatusActive,
		ProblemDetected: &prob,
		CreatedAt:       now,
	}

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Session
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != s.ID {
		t.Errorf("ID: want %q, got %q", s.ID, got.ID)
	}
	if got.Status != SessionStatusActive {
		t.Errorf("Status: want %q, got %q", SessionStatusActive, got.Status)
	}
	if got.ProblemDetected == nil || *got.ProblemDetected != prob {
		t.Errorf("ProblemDetected: want %q, got %v", prob, got.ProblemDetected)
	}
}

func TestSessionIsActive(t *testing.T) {
	tests := []struct {
		status SessionStatus
		want   bool
	}{
		{SessionStatusActive, true},
		{SessionStatusCompleted, false},
		{SessionStatusAbandoned, false},
	}
	for _, tc := range tests {
		s := Session{Status: tc.status}
		if got := s.IsActive(); got != tc.want {
			t.Errorf("IsActive(%q): want %v, got %v", tc.status, tc.want, got)
		}
	}
}

func TestSessionOptionalFieldsOmitted(t *testing.T) {
	// Nullable pointer fields must be omitted from JSON when nil.
	s := Session{ID: "x", UserID: "u", Status: SessionStatusActive, CreatedAt: time.Now()}
	b, _ := json.Marshal(s)

	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck

	for _, key := range []string{"device_id", "problem_detected", "solution_applied", "success", "duration_seconds"} {
		if _, ok := m[key]; ok {
			t.Errorf("expected key %q to be omitted when nil, but it was present", key)
		}
	}
}

func TestSessionStatusConstants(t *testing.T) {
	if SessionStatusActive != "active" {
		t.Errorf("SessionStatusActive: want 'active', got %q", SessionStatusActive)
	}
	if SessionStatusCompleted != "completed" {
		t.Errorf("SessionStatusCompleted: want 'completed', got %q", SessionStatusCompleted)
	}
	if SessionStatusAbandoned != "abandoned" {
		t.Errorf("SessionStatusAbandoned: want 'abandoned', got %q", SessionStatusAbandoned)
	}
}

// ---------------------------------------------------------------------------
// SessionStep
// ---------------------------------------------------------------------------

func TestSessionStepJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	step := SessionStep{
		ID:          "step-1",
		SessionID:   "sess-1",
		StepNumber:  3,
		Instruction: "Clean the nozzle",
		Verified:    true,
		CreatedAt:   now,
	}

	b, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got SessionStep
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.StepNumber != step.StepNumber {
		t.Errorf("StepNumber: want %d, got %d", step.StepNumber, got.StepNumber)
	}
	if got.Instruction != step.Instruction {
		t.Errorf("Instruction: want %q, got %q", step.Instruction, got.Instruction)
	}
	if got.Verified != step.Verified {
		t.Errorf("Verified: want %v, got %v", step.Verified, got.Verified)
	}
}

// ---------------------------------------------------------------------------
// Session — all JSON tags present
// ---------------------------------------------------------------------------

// TestSessionAllJSONTagsPresent verifies that every exported field of Session
// has the expected JSON key in the serialised output.
func TestSessionAllJSONTagsPresent(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	devID := "dev-42"
	prob := "layer shift"
	sol := "re-level bed"
	ok := true
	dur := 120

	s := Session{
		ID:              "sess-99",
		UserID:          "user-99",
		DeviceID:        &devID,
		Status:          SessionStatusActive,
		ProblemDetected: &prob,
		SolutionApplied: &sol,
		Success:         &ok,
		DurationSeconds: &dur,
		CreatedAt:       now,
	}

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	required := []string{
		"id", "user_id", "device_id", "status",
		"problem_detected", "solution_applied", "success",
		"duration_seconds", "created_at",
	}
	for _, key := range required {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q to be present in Session output", key)
		}
	}
}

// TestSessionJSONRoundTripAllFields verifies lossless round-trip for all
// optional pointer fields of Session.
func TestSessionJSONRoundTripAllFields(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	devID := "dev-1"
	prob := "stringing"
	sol := "lower temp"
	success := false
	dur := 300

	original := Session{
		ID:              "sess-rt",
		UserID:          "user-rt",
		DeviceID:        &devID,
		Status:          SessionStatusCompleted,
		ProblemDetected: &prob,
		SolutionApplied: &sol,
		Success:         &success,
		DurationSeconds: &dur,
		CreatedAt:       now,
	}

	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Session
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != original.ID {
		t.Errorf("ID: want %q, got %q", original.ID, got.ID)
	}
	if got.Status != SessionStatusCompleted {
		t.Errorf("Status: want completed, got %q", got.Status)
	}
	if got.DeviceID == nil || *got.DeviceID != devID {
		t.Errorf("DeviceID: want %q, got %v", devID, got.DeviceID)
	}
	if got.ProblemDetected == nil || *got.ProblemDetected != prob {
		t.Errorf("ProblemDetected: want %q, got %v", prob, got.ProblemDetected)
	}
	if got.SolutionApplied == nil || *got.SolutionApplied != sol {
		t.Errorf("SolutionApplied: want %q, got %v", sol, got.SolutionApplied)
	}
	if got.Success == nil || *got.Success != success {
		t.Errorf("Success: want %v, got %v", success, got.Success)
	}
	if got.DurationSeconds == nil || *got.DurationSeconds != dur {
		t.Errorf("DurationSeconds: want %d, got %v", dur, got.DurationSeconds)
	}
}

// ---------------------------------------------------------------------------
// Device — all JSON tags present
// ---------------------------------------------------------------------------

// TestDeviceAllJSONTagsPresent verifies that every exported field of Device
// carries the correct JSON key name.
func TestDeviceAllJSONTagsPresent(t *testing.T) {
	d := Device{
		ID:        "dev-tag",
		UserID:    "user-tag",
		Brand:     "Prusa",
		Model:     "MK4",
		CreatedAt: time.Now(),
	}

	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	required := []string{"id", "user_id", "brand", "model", "created_at"}
	for _, key := range required {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q to be present in Device output", key)
		}
	}
}

// TestDeviceJSONRoundTripAllFields verifies a lossless Device round-trip
// covering all fields including the UserID and CreatedAt timestamp.
func TestDeviceJSONRoundTripAllFields(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	original := Device{
		ID:        "dev-rt",
		UserID:    "user-rt",
		Brand:     "Bambu",
		Model:     "X1C",
		CreatedAt: now,
	}

	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Device
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != original.ID {
		t.Errorf("ID: want %q, got %q", original.ID, got.ID)
	}
	if got.UserID != original.UserID {
		t.Errorf("UserID: want %q, got %q", original.UserID, got.UserID)
	}
	if got.Brand != original.Brand {
		t.Errorf("Brand: want %q, got %q", original.Brand, got.Brand)
	}
	if got.Model != original.Model {
		t.Errorf("Model: want %q, got %q", original.Model, got.Model)
	}
	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt: want %v, got %v", original.CreatedAt, got.CreatedAt)
	}
}
