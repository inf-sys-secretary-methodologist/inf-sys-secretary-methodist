package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleTypeConstants(t *testing.T) {
	assert.Equal(t, RoleType("system_admin"), RoleSystemAdmin)
	assert.Equal(t, RoleType("methodist"), RoleMethodist)
	assert.Equal(t, RoleType("academic_secretary"), RoleAcademicSecretary)
	assert.Equal(t, RoleType("teacher"), RoleTeacher)
	assert.Equal(t, RoleType("student"), RoleStudent)
}

func TestResourceTypeConstants(t *testing.T) {
	assert.Equal(t, ResourceType("users"), ResourceUsers)
	assert.Equal(t, ResourceType("curriculum"), ResourceCurriculum)
	assert.Equal(t, ResourceType("schedule"), ResourceSchedule)
	assert.Equal(t, ResourceType("assignments"), ResourceAssignments)
	assert.Equal(t, ResourceType("reports"), ResourceReports)
}

func TestActionTypeConstants(t *testing.T) {
	assert.Equal(t, ActionType("create"), ActionCreate)
	assert.Equal(t, ActionType("read"), ActionRead)
	assert.Equal(t, ActionType("update"), ActionUpdate)
	assert.Equal(t, ActionType("delete"), ActionDelete)
	assert.Equal(t, ActionType("deactivate"), ActionDeactivate)
	assert.Equal(t, ActionType("approve"), ActionApprove)
	assert.Equal(t, ActionType("execute"), ActionExecute)
	assert.Equal(t, ActionType("export"), ActionExport)
}

func TestRoleStruct(t *testing.T) {
	r := Role{
		Type:        RoleTeacher,
		Name:        "Teacher",
		Description: "A teacher",
		IsActive:    true,
	}
	assert.Equal(t, RoleTeacher, r.Type)
	assert.True(t, r.IsActive)
}

func TestPermissionStruct(t *testing.T) {
	p := Permission{
		Resource:    ResourceUsers,
		Action:      ActionRead,
		AccessLevel: AccessFull,
		Description: "Full read access to users",
	}
	assert.Equal(t, ResourceUsers, p.Resource)
	assert.Equal(t, AccessFull, p.AccessLevel)
}

func TestUserRoleStruct(t *testing.T) {
	ur := UserRole{
		UserID:   "u1",
		RoleID:   "r1",
		IsActive: true,
	}
	assert.Equal(t, "u1", ur.UserID)
	assert.True(t, ur.IsActive)
}

func TestScopeStruct(t *testing.T) {
	fid := "f1"
	s := Scope{FacultyID: &fid}
	assert.Equal(t, "f1", *s.FacultyID)
}
