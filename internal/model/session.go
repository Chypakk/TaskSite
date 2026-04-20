package model

import "time"

type Session struct {
	ID           int       `json:"id"`
	Token        string    `json:"token"`
	Username     string    `json:"username"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}
