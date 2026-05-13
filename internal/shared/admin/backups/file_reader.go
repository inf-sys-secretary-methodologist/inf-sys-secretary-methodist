// Package backups hosts the admin-only read-API for the database +
// MinIO backup files produced by the /backup sidecar container. The
// sidecar (see /backup/Dockerfile + /backup/scripts) is the source of
// truth for backup lifecycle (cron schedule, encryption, retention,
// offsite sync). This package reads the shared backup_data volume +
// Prometheus textfile metrics, exposes a thin use case + HTTP handler
// pair gated by RequireRole(system_admin), and does NOT trigger or
// mutate backups itself.
//
// "shared/admin" houses cross-cutting administrative features (see
// auditlog for the canonical precedent).
package backups

import (
	"context"
	"errors"
)

// BackupType discriminates the two artifact families produced by the
// sidecar — full PostgreSQL dump (custom-format pg_dump) versus MinIO
// bucket tarball.
type BackupType string

// BackupTypePostgres / BackupTypeMinIO are the two artifact families
// the sidecar writes under <rootDir>/postgres/ and <rootDir>/minio/
// respectively.
const (
	BackupTypePostgres BackupType = "postgres"
	BackupTypeMinIO    BackupType = "minio"
)

// EncryptionScheme reflects the optional encryption applied by the
// sidecar at write-time. The admin UI surfaces this so an operator
// knows whether the downloaded file is directly usable or requires
// an age / GPG key first.
type EncryptionScheme string

// EncryptionNone is the default scheme — file extension lacks a
// known suffix and the artifact is directly usable. EncryptionAge /
// EncryptionGPG correspond to .age and .gpg suffixes appended by the
// sidecar's encryption step.
const (
	EncryptionNone EncryptionScheme = ""
	EncryptionAge  EncryptionScheme = "age"
	EncryptionGPG  EncryptionScheme = "gpg"
)

// BackupFile is the read-only projection of one artifact in the
// backup_data volume.
type BackupFile struct {
	Name       string
	Type       BackupType
	Size       int64
	ModifiedAt int64
	Encryption EncryptionScheme
}

// ErrInvalidRoot signals an empty / non-existent backup root passed to
// the constructor. We fail fast at DI time rather than per-request so
// misconfiguration surfaces at startup.
var ErrInvalidRoot = errors.New("backups: invalid root directory")

// FileReader lists artifacts under <rootDir>/{postgres,minio}/ that
// match the sidecar's filename grammar.
type FileReader struct {
	rootDir string
}

// NewFileReader constructs a reader rooted at the backup_data mount
// path (typically /var/backups inside the backend container).
func NewFileReader(rootDir string) (*FileReader, error) {
	if rootDir == "" {
		return nil, ErrInvalidRoot
	}
	return &FileReader{rootDir: rootDir}, nil
}

// List returns every artifact matching the sidecar's known filename
// grammar under <rootDir>/postgres/ and <rootDir>/minio/. Entries are
// sorted newest-first by ModifiedAt.
//
// Pair-1 stub: returns ErrNotImplemented until the GREEN commit wires
// the filesystem walk + whitelist + encryption-suffix detection.
func (r *FileReader) List(ctx context.Context) ([]BackupFile, error) {
	return nil, errNotImplemented
}

var errNotImplemented = errors.New("backups: file reader not implemented (RED stub)")
