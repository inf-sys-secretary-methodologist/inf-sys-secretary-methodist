// Package entities contains domain entities for the auth module.
package entities

import "time"

// Session represents a user session with refresh token
type Session struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive checks if the session is still active
func (s *Session) IsActive() bool {
	return !s.IsExpired()
}
