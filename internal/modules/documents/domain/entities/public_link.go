package entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// PublicLinkPermission represents the access level for a public link
type PublicLinkPermission string

const (
	PublicLinkRead     PublicLinkPermission = "read"
	PublicLinkDownload PublicLinkPermission = "download"
)

// PublicLink represents a public sharing link for a document
type PublicLink struct {
	ID           int64                `json:"id"`
	DocumentID   int64                `json:"document_id"`
	Token        string               `json:"token"`
	Permission   PublicLinkPermission `json:"permission"`
	CreatedBy    int64                `json:"created_by"`
	ExpiresAt    *time.Time           `json:"expires_at,omitempty"`
	MaxUses      *int                 `json:"max_uses,omitempty"`
	UseCount     int                  `json:"use_count"`
	PasswordHash *string              `json:"-"`
	IsActive     bool                 `json:"is_active"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`

	// Populated via JOIN
	DocumentTitle *string `json:"document_title,omitempty"`
	CreatedByName *string `json:"created_by_name,omitempty"`
}

// GenerateToken creates a secure random token for the public link
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsExpired checks if the link has expired
func (l *PublicLink) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// IsUsageLimitReached checks if the link has reached its usage limit
func (l *PublicLink) IsUsageLimitReached() bool {
	if l.MaxUses == nil {
		return false
	}
	return l.UseCount >= *l.MaxUses
}

// IsValid checks if the link is valid and can be used
func (l *PublicLink) IsValid() bool {
	return l.IsActive && !l.IsExpired() && !l.IsUsageLimitReached()
}

// HasPassword checks if the link is password protected
func (l *PublicLink) HasPassword() bool {
	return l.PasswordHash != nil && *l.PasswordHash != ""
}

// CanDownload checks if the link allows downloading
func (l *PublicLink) CanDownload() bool {
	return l.Permission == PublicLinkDownload
}

// IncrementUseCount increments the use count
func (l *PublicLink) IncrementUseCount() {
	l.UseCount++
	l.UpdatedAt = time.Now()
}

// Deactivate deactivates the link
func (l *PublicLink) Deactivate() {
	l.IsActive = false
	l.UpdatedAt = time.Now()
}

// Activate activates the link
func (l *PublicLink) Activate() {
	l.IsActive = true
	l.UpdatedAt = time.Now()
}
