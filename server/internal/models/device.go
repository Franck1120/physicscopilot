package models

import "time"

// Device represents any repairable or maintainable physical device owned by a user.

type Device struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Brand     string    `json:"brand"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}
