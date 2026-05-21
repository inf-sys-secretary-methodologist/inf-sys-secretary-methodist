// Package dto contains Data Transfer Objects for the auth module.
package dto

// RegisterInput represents the input data for user registration.
// v0.159.0 ADR-3 Tier 2: the Role binding tag is restricted to roles
// that may self-register (student, teacher) so a hostile request
// asking for system_admin / methodist / academic_secretary is rejected
// at the Gin binding boundary with 400 — never reaches the use case.
// IsAllowedForSelfRegistration in the domain remains the authoritative
// guard (defense in depth); the tag tightens the front line.
type RegisterInput struct {
	Name     string `json:"name" binding:"required,min=2,max=255" validate:"no_xss"`
	Email    string `json:"email" binding:"required,email,max=255" validate:"no_xss,no_sql_injection"`
	Password string `json:"password" binding:"required,min=8,max=128" validate:"strong_password"`
	Role     string `json:"role" binding:"required,oneof=student teacher" validate:"no_xss,no_sql_injection"`
}

// LoginInput represents the input data for user login.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255" validate:"no_xss,no_sql_injection"`
	Password string `json:"password" binding:"required,min=1,max=128"`
}

// RefreshTokenInput represents the input data for token refresh.
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required" validate:"no_xss,no_sql_injection"`
}
