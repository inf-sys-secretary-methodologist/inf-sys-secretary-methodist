package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionMatrix_AllRolesExist(t *testing.T) {
	roles := []RoleType{RoleSystemAdmin, RoleMethodist, RoleAcademicSecretary, RoleTeacher, RoleStudent}
	for _, role := range roles {
		_, exists := PermissionMatrix[role]
		assert.True(t, exists, "PermissionMatrix should contain role %s", role)
	}
}

func TestPermissionMatrix_AllResourcesExist(t *testing.T) {
	resources := []ResourceType{ResourceUsers, ResourceCurriculum, ResourceSchedule, ResourceAssignments, ResourceReports}
	for _, role := range []RoleType{RoleSystemAdmin} {
		for _, resource := range resources {
			_, exists := PermissionMatrix[role][resource]
			assert.True(t, exists, "PermissionMatrix[%s] should contain resource %s", role, resource)
		}
	}
}

func TestRoleDefinitions_AllRolesExist(t *testing.T) {
	roles := []RoleType{RoleSystemAdmin, RoleMethodist, RoleAcademicSecretary, RoleTeacher, RoleStudent}
	for _, role := range roles {
		def, exists := RoleDefinitions[role]
		assert.True(t, exists, "RoleDefinitions should contain role %s", role)
		assert.True(t, def.IsActive)
		assert.NotEmpty(t, def.Name)
		assert.NotEmpty(t, def.Description)
	}
}

func TestAccessLevelValues(t *testing.T) {
	assert.Equal(t, AccessLevel(0), AccessDenied)
	assert.Equal(t, AccessLevel(1), AccessLimited)
	assert.Equal(t, AccessLevel(2), AccessOwn)
	assert.Equal(t, AccessLevel(3), AccessFull)
}

func TestWorkflowStepStruct(t *testing.T) {
	step := WorkflowStep{
		ID:         "1",
		DocumentID: "doc1",
		StepNumber: 1,
		RoleType:   RoleMethodist,
		Action:     ActionApprove,
		Status:     "pending",
	}
	assert.Equal(t, "1", step.ID)
	assert.Equal(t, RoleMethodist, step.RoleType)
}
