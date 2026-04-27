package entities

import "time"

type Classroom struct {
	ID          int64          `json:"id"`
	Building    string         `json:"building"`
	Number      string         `json:"number"`
	Name        *string        `json:"name,omitempty"`
	Capacity    int            `json:"capacity"`
	Type        *string        `json:"type,omitempty"`
	Equipment   map[string]any `json:"equipment,omitempty"`
	IsAvailable bool           `json:"is_available"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
