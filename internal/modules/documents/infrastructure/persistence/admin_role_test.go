package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/infrastructure/persistence"
)

// TestIsAdminRole guards v0.156.0 ADR-4 (#266): SystemAdmin role bypass.
//
// Pre-fix document_repository_pg.go compared filter.CurrentUserRole !=
// "admin" — production JWT carries "system_admin" (auth.RoleSystemAdmin),
// so actual admins were caught by the per-user access-control WHERE and
// could not see all rows. Test pins admin-recognition policy for both
// legacy DocumentPermission rows ("admin") и canonical auth role
// ("system_admin").
func TestIsAdminRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"production system_admin recognized", "system_admin", true},
		{"legacy admin recognized", "admin", true},
		{"empty role rejected", "", false},
		{"methodist rejected", "methodist", false},
		{"student rejected", "student", false},
		{"teacher rejected", "teacher", false},
		{"academic_secretary rejected", "academic_secretary", false},
		{"random string rejected", "root", false},
		{"case-insensitive system_admin", "SYSTEM_ADMIN", true},
		{"case-insensitive admin", "ADMIN", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, persistence.IsAdminRole(tt.role),
				"role %q expected admin=%v", tt.role, tt.expected)
		})
	}
}
