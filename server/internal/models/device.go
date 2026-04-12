package models

import "time"

// Device represents any repairable or maintainable physical device owned by a user.
// Maps directly to the `devices` table in Supabase.
type Device struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Brand     string    `json:"brand"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

// DisplayName returns the human-readable label shown in the UI: "Brand Model".
func (d Device) DisplayName() string {
	return d.Brand + " " + d.Model
}
