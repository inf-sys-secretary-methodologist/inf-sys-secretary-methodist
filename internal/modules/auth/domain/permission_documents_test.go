package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPermissionMatrix_DocumentsResourceForAllRoles verifies that
// ResourceDocuments is defined for every role with explicit access levels
// for create / read / update / delete actions. AUDIT_REPORT
// (academic_secretary section, table) flagged that ResourceDocuments was
// missing entirely from the permission matrix despite documents being a
// core entity.
func TestPermissionMatrix_DocumentsResourceForAllRoles(t *testing.T) {
	roles := []RoleType{
		RoleSystemAdmin,
		RoleMethodist,
		RoleAcademicSecretary,
		RoleTeacher,
		RoleStudent,
	}
	requiredActions := []ActionType{
		ActionCreate,
		ActionRead,
		ActionUpdate,
		ActionDelete,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			actions, exists := PermissionMatrix[role][ResourceDocuments]
			assert.True(t, exists,
				"PermissionMatrix[%s] must contain ResourceDocuments", role)
			for _, action := range requiredActions {
				_, ok := actions[action]
				assert.True(t, ok,
					"PermissionMatrix[%s][ResourceDocuments] must define action %s", role, action)
			}
		})
	}
}

// TestPermissionMatrix_DocumentsAccessLevels documents the intended access
// per role and locks it down. If anyone weakens these levels later, this test
// fails — preventing accidental privilege escalation.
func TestPermissionMatrix_DocumentsAccessLevels(t *testing.T) {
	cases := []struct {
		role     RoleType
		action   ActionType
		expected AccessLevel
	}{
		// system_admin — full power
		{RoleSystemAdmin, ActionCreate, AccessFull},
		{RoleSystemAdmin, ActionRead, AccessFull},
		{RoleSystemAdmin, ActionUpdate, AccessFull},
		{RoleSystemAdmin, ActionDelete, AccessFull},

		// methodist — full CRUD on documents (methodical materials)
		{RoleMethodist, ActionCreate, AccessFull},
		{RoleMethodist, ActionRead, AccessFull},
		{RoleMethodist, ActionUpdate, AccessFull},
		{RoleMethodist, ActionDelete, AccessFull},

		// academic_secretary — full CRUD (administrative paperwork)
		{RoleAcademicSecretary, ActionCreate, AccessFull},
		{RoleAcademicSecretary, ActionRead, AccessFull},
		{RoleAcademicSecretary, ActionUpdate, AccessFull},
		{RoleAcademicSecretary, ActionDelete, AccessFull},

		// teacher — owns documents (own scope), can read shared
		{RoleTeacher, ActionCreate, AccessFull},
		{RoleTeacher, ActionRead, AccessLimited},
		{RoleTeacher, ActionUpdate, AccessOwn},
		{RoleTeacher, ActionDelete, AccessOwn},

		// student — read only what is shared via ACL, no writes
		{RoleStudent, ActionCreate, AccessDenied},
		{RoleStudent, ActionRead, AccessLimited},
		{RoleStudent, ActionUpdate, AccessDenied},
		{RoleStudent, ActionDelete, AccessDenied},
	}

	for _, tc := range cases {
		t.Run(string(tc.role)+"_"+string(tc.action), func(t *testing.T) {
			actual := PermissionMatrix[tc.role][ResourceDocuments][tc.action]
			assert.Equal(t, tc.expected, actual,
				"role=%s action=%s should be %d (got %d)",
				tc.role, tc.action, tc.expected, actual)
		})
	}
}
