// Package entities contains domain entities for the AI module.
package entities

import "time"

// FunFact represents an educational fun fact
type FunFact struct {
	ID         int64      `json:"id"`
	Content    string     `json:"content"`
	Category   string     `json:"category"`
	Source     string     `json:"source,omitempty"`
	SourceURL  string     `json:"source_url,omitempty"`
	Language   string     `json:"language"`
	IsApproved bool       `json:"is_approved"`
	UsedCount  int        `json:"used_count"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
