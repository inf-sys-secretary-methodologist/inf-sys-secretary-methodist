package domain

import "errors"

// ErrInvalidWorkProgram signals violation of a construction or update
// invariant on the WorkProgram aggregate (empty title/specialty_code,
// year out of [2000, 2100], non-positive author_id, etc.). Handlers
// map this sentinel to HTTP 422 (request is well-formed but rejected
// by domain rules).
var ErrInvalidWorkProgram = errors.New("work_program: invalid invariant")

// ErrInvalidStatusTransition signals an attempt to transition the
// aggregate to a status that the FSM does not permit from the current
// state. Handlers map to HTTP 422.
var ErrInvalidStatusTransition = errors.New("work_program: invalid status transition")

// ErrWorkProgramScopeForbidden indicates that the caller is not
// authorized to operate on this particular WorkProgram per ADR-5 role
// matrix — typically because the caller is not the author and not an
// admin/methodist. Handlers map to HTTP 403.
var ErrWorkProgramScopeForbidden = errors.New("work_program: caller cannot operate on this work program")

// ErrCannotEditFrozenStatus indicates that the WorkProgram is in a
// status (pending_approval / approved / archived) that does not allow
// content edits. Only `draft` and `needs_revision` permit edits per
// ADR-2 FSM. Handlers map to HTTP 422.
var ErrCannotEditFrozenStatus = errors.New("work_program: cannot edit, status is frozen")

// ErrRejectReasonRequired signals an Reject() call without a non-empty
// reason. Methodist must explain rejection so the author can revise.
// Handlers map to HTTP 422.
var ErrRejectReasonRequired = errors.New("work_program: reject reason is required")
