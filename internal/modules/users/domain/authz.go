// Package domain contains domain primitives for the users module.
//
// Authorization rules for cross-user operations live here as free
// functions so they can be reused across usecases without leaking
// repository or middleware concerns.
package domain

import (
	"errors"
	"fmt"
	"strings"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
)

// AvatarPrefixFormat is the storage-key prefix per user, parameterised
// by the user id. Mirrors the format emitted by the avatar Upload
// handler (avatar_handler.go: "avatars/%d_%s%s") so the validator
// stays in sync with the writer.
const AvatarPrefixFormat = "avatars/%d_"

// ErrProfileEditForbidden is returned when an actor attempts to edit
// another user's profile without being a system_admin override.
//
// Closes v1.0.0 batch 2 TIER 0 finding (#283 ADR-1): PUT
// /api/users/:id/profile previously accepted ANY caller — no
// actor==target check at the usecase layer. Audit row also wrote
// user_id=target instead of actor_user_id, so attackers were
// untraceable.
var ErrProfileEditForbidden = errors.New("profile edit forbidden: actor is not target user and lacks system_admin override")

// ErrCannotDeleteSelf is returned when an actor attempts to delete
// their own account. Closes #283 ADR-4 (Tier 1): self-deletion is
// unrecoverable and bricks the actor's session — guarded
// unconditionally regardless of role.
var ErrCannotDeleteSelf = errors.New("user cannot delete their own account")

// ErrLastAdminProtected is returned when deleting the target user
// would leave the system with zero system_admin accounts, locking
// nobody out of administrative recovery. Closes #283 ADR-4 (Tier 1).
var ErrLastAdminProtected = errors.New("cannot delete the last remaining system_admin")

// ErrInvalidAvatarKey is returned when an avatar storage key does not
// belong to the target user's avatar prefix.
//
// Closes v1.0.0 batch 2 TIER 0 finding (#283 ADR-3): UpdateProfile
// accepted any string in the Avatar field, persisting an arbitrary
// MinIO object key that the avatar GET endpoint later signed as a
// presigned URL — letting any user point their avatar at HR records
// or exam reports stored in the same bucket.
var ErrInvalidAvatarKey = errors.New("invalid avatar storage key: must belong to target user's avatar prefix")

// ValidateAvatarKey checks that an avatar storage key belongs to the
// target user's avatar prefix.
//
// Rule: empty key (clearing the avatar) is always accepted; non-empty
// keys must start with "avatars/{targetID}_" — the same prefix the
// avatar Upload handler emits (avatar_handler.go AvatarFolder + user
// id + "_" + uuid + ext).
//
// Empty key is accepted (clearing the avatar is a legitimate self-edit).
// Non-empty keys must start with the avatar prefix of the target user
// — any other key signals a write attempt against another user's
// avatar storage area, or a write at an arbitrary S3 object outside
// the avatars folder.
func ValidateAvatarKey(key string, targetID int64) error {
	if key == "" {
		return nil
	}
	prefix := fmt.Sprintf(AvatarPrefixFormat, targetID)
	if !strings.HasPrefix(key, prefix) {
		return ErrInvalidAvatarKey
	}
	return nil
}

// AuthorizeProfileEdit returns nil if the actor may edit the target
// user's profile, ErrProfileEditForbidden otherwise.
//
// Rule: actor must be the target user (self-edit) OR system_admin
// (override). All other actor/target combinations — including admins
// of other kinds (methodist, academic_secretary, teacher) editing
// somebody else — are rejected. Cross-user profile mutation is a
// privileged operation that audit-tracks via actor_user_id and must
// not fall through silently.
func AuthorizeProfileEdit(actorID, targetID int64, actorRole authDomain.RoleType) error {
	if actorID == targetID {
		return nil
	}
	if actorRole == authDomain.RoleSystemAdmin {
		return nil
	}
	return ErrProfileEditForbidden
}

// AuthorizeUserDelete returns nil if the actor may delete the target
// user, an explanatory sentinel otherwise.
//
// Two guards:
//  1. Self-delete (actorID == targetID) is unconditionally forbidden:
//     no role gets to remove its own account. Returns ErrCannotDeleteSelf.
//  2. Removing the last remaining system_admin would brick the
//     administrative recovery path. Returns ErrLastAdminProtected when
//     the target is a system_admin AND the current admin headcount is
//     1 (the actor must look up the count via the repository — this
//     function takes it as an argument to stay pure and testable).
//
func AuthorizeUserDelete(
	actorID, targetID int64,
	targetRole authDomain.RoleType,
	adminHeadcount int,
) error {
	if actorID == targetID {
		return ErrCannotDeleteSelf
	}
	if targetRole == authDomain.RoleSystemAdmin && adminHeadcount <= 1 {
		return ErrLastAdminProtected
	}
	return nil
}
