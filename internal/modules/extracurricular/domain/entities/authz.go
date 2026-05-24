package entities

import "fmt"

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

// AuthorizeEventCreate gates POST /api/v1/extracurricular/events.
// Allowed roles per ADR-6: system_admin (via isAdmin flag) + methodist +
// academic_secretary. Teacher / student / unknown denied.
func AuthorizeEventCreate(actorRole string, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	switch actorRole {
	case roleMethodist, roleAcademicSecretary:
		return nil
	}
	return scopeForbidden(fmt.Sprintf("role %q not allowed to create events", actorRole))
}

// AuthorizeEventEdit gates PUT/DELETE /api/v1/extracurricular/events/:id.
// Allowed: admin (any event); methodist|academic_secretary self-edit
// (actorID == organizerID). Teacher / student denied even для own
// events (they can't create, so own-event scenario is hypothetical).
func AuthorizeEventEdit(actorID, organizerID int64, actorRole string, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	switch actorRole {
	case roleMethodist, roleAcademicSecretary:
		if actorID == organizerID {
			return nil
		}
		return scopeForbidden(fmt.Sprintf("actor %d is not organizer %d", actorID, organizerID))
	}
	return scopeForbidden(fmt.Sprintf("role %q not allowed to edit events", actorRole))
}

// CanViewEvent reports whether a caller in actorRole может see an event
// targeted at the given audience per ADR-6 matrix:
//
//	TargetAudienceAll      → any role
//	TargetAudienceStudents → student only
//	TargetAudienceTeachers → teacher only
//	TargetAudienceStaff    → methodist | academic_secretary | system_admin
//
// Admins are not "target" но have read-everything via separate isAdmin
// flag in handler; this function is the audience-aware filter applied
// после the admin override.
func CanViewEvent(actorRole string, audience TargetAudience) bool {
	switch audience {
	case TargetAudienceAll:
		return true
	case TargetAudienceStudents:
		return actorRole == roleStudent
	case TargetAudienceTeachers:
		return actorRole == roleTeacher
	case TargetAudienceStaff:
		return actorRole == roleMethodist || actorRole == roleAcademicSecretary || actorRole == roleSystemAdmin
	}
	return false
}

// scopeForbidden wraps ErrEventScopeForbidden с context for handlers'
// 403 mapping. Internal helper to avoid duplicating wrap pattern.
func scopeForbidden(reason string) error {
	return fmt.Errorf("%w: %s", ErrEventScopeForbidden, reason)
}
