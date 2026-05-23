// Package domain contains domain primitives for the files module.
//
// Authorization rules for file access live here as free functions so
// they can be reused across usecases without coupling to repository
// or middleware concerns.
package domain

import (
	"errors"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

// FileAction enumerates the actions that AuthorizeFileAccess gates.
type FileAction string

const (
	// FileActionRead covers GET file metadata and GET download URL.
	FileActionRead FileAction = "read"
	// FileActionAttach covers attaching a temporary file to a document,
	// task, or announcement (one-shot; permanent ownership transfer).
	FileActionAttach FileAction = "attach"
	// FileActionCreateVersion covers uploading a new version of an
	// already-attached file.
	FileActionCreateVersion FileAction = "create_version"
	// FileActionDelete covers soft-deleting a file.
	FileActionDelete FileAction = "delete"
)

// ErrFileAccessDenied is returned when an actor is not authorised for
// the requested action against a file.
//
// Closes v1.0.0 batch 2 TIER 0 findings (#290 ADR-1 + ADR-2): files
// usecases previously accepted only an `id int64` and never compared
// against `FileMetadata.UploadedBy`, so any authenticated user could
// read, download, attach, version, or delete any file via sequential
// BIGSERIAL enumeration.
var ErrFileAccessDenied = errors.New("file access denied")

// AuthorizeFileAccess returns nil if `actor` is permitted to perform
// `action` against `file`, ErrFileAccessDenied otherwise.
//
// STUB — replace with real rule in GREEN commit.
func AuthorizeFileAccess(
	actorID int64,
	actorRole authDomain.RoleType,
	file *entities.FileMetadata,
	action FileAction,
) error {
	_ = actorID
	_ = actorRole
	_ = file
	_ = action
	return nil
}
