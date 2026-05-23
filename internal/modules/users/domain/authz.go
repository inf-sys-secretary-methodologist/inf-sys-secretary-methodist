// Package domain contains domain primitives for the users module.
//
// Authorization rules for cross-user operations live here as free
// functions so they can be reused across usecases without leaking
// repository or middleware concerns.
package domain

import (
	"errors"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
)

// ErrProfileEditForbidden is returned when an actor attempts to edit
// another user's profile without being a system_admin override.
//
// Closes v1.0.0 batch 2 TIER 0 finding (#283 ADR-1): PUT
// /api/users/:id/profile previously accepted ANY caller — no
// actor==target check at the usecase layer. Audit row also wrote
// user_id=target instead of actor_user_id, so attackers were
// untraceable.
var ErrProfileEditForbidden = errors.New("profile edit forbidden: actor is not target user and lacks system_admin override")

// AuthorizeProfileEdit returns nil if the actor may edit the target
// user's profile, ErrProfileEditForbidden otherwise.
//
// Rule: actor must be the target user (self-edit) OR system_admin
// (override).
//
// RED-commit stub: returns nil unconditionally. The GREEN commit
// implements the actor==target / role==system_admin branches and
// flips the new failing test to passing.
func AuthorizeProfileEdit(actorID, targetID int64, actorRole authDomain.RoleType) error {
	_ = actorID
	_ = targetID
	_ = actorRole
	return nil
}
