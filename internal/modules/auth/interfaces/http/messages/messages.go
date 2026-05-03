// Package messages contains user-facing strings returned by auth HTTP handlers.
// Centralizing these here keeps Clean Architecture's UI/messaging concern out
// of usecase and handler logic, and makes future i18n a single point of change.
package messages

// Errors returned to clients of the auth API.
const (
	// RoleNotAllowedForSelfRegistration is shown when a user tries to register
	// with a privileged role that is reserved for admin-created accounts.
	RoleNotAllowedForSelfRegistration = "Эта роль недоступна для самостоятельной регистрации"

	// PasswordResetEmailRequired is shown when the request body lacks an
	// email field (or it is empty).
	PasswordResetEmailRequired = "Email обязателен"

	// PasswordResetMalformedRequest is shown for unparseable JSON bodies
	// on the password-reset endpoints.
	PasswordResetMalformedRequest = "Некорректный формат запроса"

	// PasswordResetTokenExpired is shown when a password-reset token is
	// missing, expired, or otherwise no longer valid (verify or confirm).
	PasswordResetTokenExpired = "Срок действия ссылки для сброса пароля истёк"

	// PasswordResetWeakPassword is shown when the new password fails the
	// backend minimum-length check.
	PasswordResetWeakPassword = "Пароль не соответствует требованиям безопасности"
)
