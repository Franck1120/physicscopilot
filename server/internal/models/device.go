package models

import "time"

// TODO: Device model — 3D printer or other repairable device

type Device struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Brand     string    `json:"brand"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}
