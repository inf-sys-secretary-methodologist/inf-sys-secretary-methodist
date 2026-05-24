package entities

import (
	"errors"
	"fmt"
)

// Role constants matching the project-wide role matrix
// (per docs/roles-and-flows.md). Declared locally to keep
// extracurricular bounded context free от cross-module imports.
const (
	roleSystemAdmin       = "system_admin"
	roleMethodist         = "methodist"
	roleAcademicSecretary = "academic_secretary"
	roleTeacher           = "teacher"
	roleStudent           = "student"
)

// Stub anchors keeping role constants + scopeForbidden referenced
// during Pair 4 RED so the unused-linter stays quiet. Pair 4 GREEN
// switches on these directly and removes the var.
//
//nolint:gochecknoglobals // anchor only
var _ = []string{roleSystemAdmin, roleMethodist, roleAcademicSecretary, roleTeacher, roleStudent}

// AuthorizeEventCreate gates POST /api/v1/extracurricular/events.
// Pair 4 RED stub — always returns "not implemented"; GREEN restricts
// to system_admin / methodist / academic_secretary.
func AuthorizeEventCreate(actorRole string, isAdmin bool) error {
	_ = scopeForbidden("")
	return errors.New("not implemented (Pair 4 RED stub)")
}

// AuthorizeEventEdit gates PUT/DELETE /api/v1/extracurricular/events/:id.
// Pair 4 RED stub.
func AuthorizeEventEdit(actorID, organizerID int64, actorRole string, isAdmin bool) error {
	return errors.New("not implemented (Pair 4 RED stub)")
}

// CanViewEvent reports whether a caller in actorRole может see an event
// targeted at the given audience. Used by repository List query +
// GetByID handler 404. Pair 4 RED stub always returns true.
func CanViewEvent(actorRole string, audience TargetAudience) bool {
	return true
}

// scopeForbidden wraps ErrEventScopeForbidden с context for handlers'
// 403 mapping. Internal helper to avoid duplicating wrap pattern.
func scopeForbidden(reason string) error {
	return fmt.Errorf("%w: %s", ErrEventScopeForbidden, reason)
}
