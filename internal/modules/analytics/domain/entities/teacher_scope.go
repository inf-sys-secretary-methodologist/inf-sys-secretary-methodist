package entities

import "errors"

// ErrAnalyticsScopeForbidden indicates that a request targets analytics
// data outside the caller's allowed scope (e.g., a teacher querying a
// group they do not teach, or a student whose group is not in their
// scheduled assignments). Handlers map this sentinel to HTTP 403.
var ErrAnalyticsScopeForbidden = errors.New("analytics: target outside teacher scope")

// TeacherScope is a value object that describes which student groups a
// teacher is permitted to query within the analytics module. The scope
// is derived from authoritative scheduling data (schedule_lessons joined
// with student_groups).
//
// Semantics for callers:
//   - A nil *TeacherScope means "unrestricted" — used for methodist,
//     academic_secretary, and system_admin roles. Use cases must treat
//     nil as a pass-through.
//   - A non-nil scope enforces a strict whitelist. An empty whitelist
//     denies every group; this is the correct behaviour for a teacher
//     who has no scheduled lessons.
type TeacherScope struct {
	teacherID  int64
	allowedSet map[string]struct{}
}

// NewTeacherScope constructs a TeacherScope from a teacher's user ID
// and the canonical list of group names that teacher teaches.
// Duplicate and empty names in groupNames are dropped silently; a nil
// or empty slice yields a deny-all scope (no group can be accessed).
func NewTeacherScope(teacherID int64, groupNames []string) *TeacherScope {
	set := make(map[string]struct{}, len(groupNames))
	for _, name := range groupNames {
		if name == "" {
			continue
		}
		set[name] = struct{}{}
	}
	return &TeacherScope{teacherID: teacherID, allowedSet: set}
}

// TeacherID returns the user ID this scope was built for.
func (s *TeacherScope) TeacherID() int64 { return s.teacherID }

// AllowsGroup reports whether name is in the whitelist. The empty string
// is always denied. Comparison is case-sensitive — group names come from
// authoritative storage where casing is canonical.
func (s *TeacherScope) AllowsGroup(name string) bool {
	if name == "" {
		return false
	}
	_, ok := s.allowedSet[name]
	return ok
}

// AllowsGroupPtr is the *string variant of AllowsGroup. A nil pointer
// is denied: a missing group on a risk record cannot be affirmed against
// the whitelist and must not silently pass the scope check.
func (s *TeacherScope) AllowsGroupPtr(name *string) bool {
	if name == nil {
		return false
	}
	return s.AllowsGroup(*name)
}

// FilterGroupNames returns the subset of in that the scope allows,
// preserving the original order. The result is non-nil even when empty
// (callers can range over it without nil checks).
func (s *TeacherScope) FilterGroupNames(in []string) []string {
	out := make([]string, 0, len(in))
	for _, name := range in {
		if s.AllowsGroup(name) {
			out = append(out, name)
		}
	}
	return out
}
