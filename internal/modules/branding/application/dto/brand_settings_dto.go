// Package dto holds JSON DTOs for the branding module. Domain
// entities never escape the application boundary directly — DTOs
// own the JSON shape and field tags.
package dto

import "time"

// BrandSettingsDTO is the JSON projection of the BrandSettings
// entity. Used by both admin endpoints (GET + PUT response) and
// the public endpoint — no field is sensitive, so the same shape
// surfaces in both contexts.
type BrandSettingsDTO struct {
	AppName        string    `json:"app_name"`
	Tagline        string    `json:"tagline"`
	LogoURL        string    `json:"logo_url"`
	FaviconURL     string    `json:"favicon_url"`
	PrimaryColor   string    `json:"primary_color"`
	SecondaryColor string    `json:"secondary_color"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UpdateBrandSettingsRequest is the body accepted by PUT
// /api/admin/branding. Every field is optional — empty strings
// clear the corresponding column.
type UpdateBrandSettingsRequest struct {
	AppName        string `json:"app_name"`
	Tagline        string `json:"tagline"`
	LogoURL        string `json:"logo_url"`
	FaviconURL     string `json:"favicon_url"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
}
