// Package messages contains user-facing strings returned by auth HTTP handlers.
// Centralizing these here keeps Clean Architecture's UI/messaging concern out
// of usecase and handler logic, and makes future i18n a single point of change.
package messages

// Errors returned to clients of the auth API.
const (
	// RoleNotAllowedForSelfRegistration is shown when a user tries to register
	// with a privileged role that is reserved for admin-created accounts.
	RoleNotAllowedForSelfRegistration = "Эта роль недоступна для самостоятельной регистрации"
)
