package usecases

import (
	"context"

	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// UserAccountRepository is the consumer-side narrow port that the users
// module needs from the auth user store: just the lifecycle operations
// (read by id, save, delete) — never the auth-only methods (password
// hash lookup, MFA secret, list-by-email). DIP per CLAUDE.md: interfaces
// live in the consumer package, concrete implementations from
// auth/infrastructure satisfy this structurally.
//
// Note: the entity parameter is still authEntities.User to avoid a
// users-owned DTO + adapter layer in this release; eliminating that
// last cross-module type dependency is out of v0.139.0 scope.
type UserAccountRepository interface {
	GetByID(ctx context.Context, id int64) (*authEntities.User, error)
	Save(ctx context.Context, user *authEntities.User) error
	Delete(ctx context.Context, id int64) error
}
