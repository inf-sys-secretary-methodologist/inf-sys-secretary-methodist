// Package dto contains data transfer objects for the AI module.
package dto

// FunFactResponse represents a fun fact API response
type FunFactResponse struct {
	ID       int64  `json:"id"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Source   string `json:"source,omitempty"`
}
