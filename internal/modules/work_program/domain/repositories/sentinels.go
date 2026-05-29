// Package repositories holds repository-contract sentinels and query-
// shape DTOs for the work_program bounded context. Persistence ports
// themselves live in application/usecases per DIP (see CLAUDE.md
// "Repository interfaces — в пакете-потребителе (usecase/)" gate);
// sentinels live here so handlers and read-model consumers can chain
// errors.Is against them without depending on the persistence port.
package repositories

import "errors"

// ErrWorkProgramNotFound signals that no WorkProgram row matched the
// given id, identity tuple, or filter. Handlers map this sentinel to
// HTTP 404.
var ErrWorkProgramNotFound = errors.New("work_program: not found")

// ErrWorkProgramIdentityExists signals that a WorkProgram with the
// same identity tuple (discipline_id, specialty_code, applicable_from_year)
// already exists — uniqueness mirrors the uq_work_programs_identity
// constraint в migration 047. Handlers map to HTTP 409 (Conflict).
var ErrWorkProgramIdentityExists = errors.New("work_program: identity already exists")

// ErrWorkProgramVersionConflict signals a stale-version Update attempt
// against optimistic-lock check (WHERE id = ? AND version = ?). The
// caller MUST re-read the latest state and retry the edit. Handlers
// map to HTTP 409 (Conflict).
var ErrWorkProgramVersionConflict = errors.New("work_program: version conflict, re-read and retry")

// ErrMinobrnaukiOrderNotFound signals that no minobrnauki_orders row
// matched the given id (приказ Минобрнауки per ADR-11). Handlers map
// this sentinel to HTTP 404.
var ErrMinobrnaukiOrderNotFound = errors.New("work_program: minobrnauki order not found")
