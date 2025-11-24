// Package fixtures provides test fixtures for testing.
package fixtures

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// UserBuilder provides a fluent interface for building test users
type UserBuilder struct {
	user *entities.User
}

// NewUserBuilder creates a new user builder with defaults
func NewUserBuilder() *UserBuilder {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test123456"), bcrypt.DefaultCost)

	return &UserBuilder{
		user: &entities.User{
			Email:     "test@example.com",
			Password:  string(hashedPassword),
			Role:      domain.RoleStudent,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// WithID sets the user ID
func (b *UserBuilder) WithID(id int64) *UserBuilder {
	b.user.ID = id
	return b
}

// WithEmail sets the user email
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

// WithPassword sets the user password (will be hashed)
func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	b.user.Password = string(hashedPassword)
	return b
}

// WithRawPassword sets the user password without hashing
func (b *UserBuilder) WithRawPassword(hashedPassword string) *UserBuilder {
	b.user.Password = hashedPassword
	return b
}

// WithRole sets the user role
func (b *UserBuilder) WithRole(role domain.RoleType) *UserBuilder {
	b.user.Role = role
	return b
}

// WithStatus sets the user status
func (b *UserBuilder) WithStatus(status entities.UserStatus) *UserBuilder {
	b.user.Status = status
	return b
}

// WithCreatedAt sets the created timestamp
func (b *UserBuilder) WithCreatedAt(t time.Time) *UserBuilder {
	b.user.CreatedAt = t
	return b
}

// WithUpdatedAt sets the updated timestamp
func (b *UserBuilder) WithUpdatedAt(t time.Time) *UserBuilder {
	b.user.UpdatedAt = t
	return b
}

// Build returns the built user
func (b *UserBuilder) Build() *entities.User {
	return b.user
}

// BuildWithContext builds and returns the user (for consistency with other fixtures)
func (b *UserBuilder) BuildWithContext(_ context.Context) *entities.User {
	return b.user
}

// Predefined fixtures for common test scenarios

// AdminUser returns a test admin user
func AdminUser() *entities.User {
	return NewUserBuilder().
		WithEmail("admin@example.com").
		WithRole(domain.RoleSystemAdmin).
		Build()
}

// SecretaryUser returns a test secretary user
func SecretaryUser() *entities.User {
	return NewUserBuilder().
		WithEmail("secretary@example.com").
		WithRole(domain.RoleAcademicSecretary).
		Build()
}

// MethodistUser returns a test methodist user
func MethodistUser() *entities.User {
	return NewUserBuilder().
		WithEmail("methodist@example.com").
		WithRole(domain.RoleMethodist).
		Build()
}

// TeacherUser returns a test teacher user
func TeacherUser() *entities.User {
	return NewUserBuilder().
		WithEmail("teacher@example.com").
		WithRole(domain.RoleTeacher).
		Build()
}

// StudentUser returns a test student user
func StudentUser() *entities.User {
	return NewUserBuilder().
		WithEmail("student@example.com").
		WithRole(domain.RoleStudent).
		Build()
}

// InactiveUser returns an inactive user
func InactiveUser() *entities.User {
	return NewUserBuilder().
		WithEmail("inactive@example.com").
		WithStatus(entities.UserStatusInactive).
		Build()
}

// UserWithKnownPassword returns a user with a known password for login tests
func UserWithKnownPassword(password string) *entities.User {
	return NewUserBuilder().
		WithPassword(password).
		Build()
}
