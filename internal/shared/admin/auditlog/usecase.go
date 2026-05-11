// Package auditlog hosts the admin-only read API for the audit_logs
// table. The write path (AuditLogger + AuditLogRepositoryPG) lives in
// internal/shared/infrastructure/logging — this package consumes its
// AuditLogReader port via DIP and exposes a thin use case + HTTP
// handler pair gated by RequireRole(system_admin).
//
// "shared/admin" houses cross-cutting administrative features that do
// not belong to any single bounded context (audit-log, backup/restore,
// global settings). New admin features should add a subpackage here
// rather than introducing top-level admin modules.
package auditlog

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// DefaultLimit is the page size used when a caller omits ?limit= or
// passes a non-positive value.
const DefaultLimit = 50

// MaxLimit caps the per-page size regardless of request input.
const MaxLimit = 200

// ListInput is the validated, parsed input to AdminAuditLogUseCase.List.
// Time fields are pointers so the empty/unset case is unambiguous; an
// explicit zero-time would silently filter all rows older than
// 0001-01-01 (i.e. everything).
type ListInput struct {
	Action   string
	Resource string
	UserID   *int64
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// ErrInvalidTimeRange signals From >= To — the half-open semantic
// [from, to) requires strict inequality to return any rows.
var ErrInvalidTimeRange = errors.New("audit-log list: from must be strictly before to")

// AdminAuditLogUseCase is the application-layer collaborator that
// clamps pagination, validates the time range, and delegates to the
// reader port.
type AdminAuditLogUseCase struct {
	reader logging.AuditLogReader
}

// NewAdminAuditLogUseCase wires the use case against an
// AuditLogReader. Panics on a nil reader so misconfigured DI fails
// at construction rather than on the first request.
func NewAdminAuditLogUseCase(reader logging.AuditLogReader) *AdminAuditLogUseCase {
	if reader == nil {
		panic("auditlog: nil AuditLogReader")
	}
	return &AdminAuditLogUseCase{reader: reader}
}

// List clamps the page size, validates the time range, and forwards
// the filter to the reader.
//
// Stub: behavior deferred to the matching GREEN commit; the signature
// + return shape are declared so the RED test file compiles against
// the use-case port without breaking the package build.
func (uc *AdminAuditLogUseCase) List(ctx context.Context, in ListInput) (logging.AuditLogListResult, error) {
	_ = ctx
	_ = in
	_ = uc.reader
	return logging.AuditLogListResult{}, errAdminListNotImplemented
}

// errAdminListNotImplemented marks the RED stub. Removed by the GREEN
// commit when the real body lands.
var errAdminListNotImplemented = errors.New("admin audit-log list: not implemented")
