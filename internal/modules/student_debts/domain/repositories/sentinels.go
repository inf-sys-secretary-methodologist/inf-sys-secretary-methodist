// Package repositories holds repository-contract sentinels and query-
// shape DTOs for the student_debts bounded context. The persistence
// ports themselves live in application/usecases per DIP (see CLAUDE.md
// "Repository interfaces — в пакете-потребителе (usecase/)" gate);
// sentinels live here so handlers and read-model consumers can chain
// errors.Is against them without depending on the persistence port.
package repositories

import "errors"

// ErrStudentDebtNotFound signals that no student_debts row matched the
// given id or filter. Handlers map this sentinel to HTTP 404.
var ErrStudentDebtNotFound = errors.New("student_debts: not found")

// ErrStudentDebtIdentityExists signals that a debt with the same natural
// key (group_name, student_full_name, discipline_name, semester) already
// exists — uniqueness mirrors the uq_student_debts_identity constraint in
// migration 050. Handlers map this to HTTP 409 (Conflict); the importer
// treats it as the signal to upsert rather than insert.
var ErrStudentDebtIdentityExists = errors.New("student_debts: identity already exists")

// ErrStudentDebtVersionConflict signals a stale-version Update against the
// optimistic-lock guard (WHERE id = ? AND version = ?). The caller MUST
// re-read the latest state and retry the edit. Handlers map this to HTTP
// 409 (Conflict).
var ErrStudentDebtVersionConflict = errors.New("student_debts: version conflict, re-read and retry")
