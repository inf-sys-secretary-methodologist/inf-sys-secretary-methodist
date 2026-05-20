package persistence

import "strings"

// adminRoles captures the role tokens treated as admin для access-control
// bypass: legacy DocumentPermission rows ("admin") + canonical auth
// module role ("system_admin"). Stored lower-case for case-insensitive
// lookup.
var adminRoles = map[string]struct{}{
	"admin":        {},
	"system_admin": {},
}

// IsAdminRole reports whether a role string represents an admin who
// should bypass the per-user access-control WHERE clause.
//
// v0.156.0 ADR-4 (#266): repo previously compared CurrentUserRole !=
// "admin" — production JWT carries "system_admin" so actual admins were
// caught by the access filter и could not see all rows.
func IsAdminRole(role string) bool {
	if role == "" {
		return false
	}
	_, ok := adminRoles[strings.ToLower(strings.TrimSpace(role))]
	return ok
}
