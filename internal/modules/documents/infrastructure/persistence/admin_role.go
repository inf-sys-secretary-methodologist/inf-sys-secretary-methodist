package persistence

// IsAdminRole reports whether a role string represents an admin who
// should bypass the per-user access-control WHERE clause. Stub returns
// false для всех — GREEN commit will recognize both legacy "admin" and
// canonical "system_admin" forms (case-insensitive).
//
// v0.156.0 ADR-4 (#266).
func IsAdminRole(role string) bool {
	return false
}
