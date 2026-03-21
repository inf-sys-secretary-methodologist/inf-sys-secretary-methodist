package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

func TestNewJWTService(t *testing.T) {
	svc := NewJWTService("secret", "refresh", time.Hour, 24*time.Hour)
	assert.NotNil(t, svc)
}

func TestJWTService_GenerateTokens(t *testing.T) {
	svc := NewJWTService("test-secret", "test-refresh", time.Hour, 24*time.Hour)

	pair, err := svc.GenerateTokens(1, "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestJWTService_GenerateTokens_EmptySecret(t *testing.T) {
	svc := NewJWTService("", "", time.Hour, 24*time.Hour)
	pair, err := svc.GenerateTokens(1, "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
}

func TestNewAuthorizationService(t *testing.T) {
	svc := NewAuthorizationService()
	assert.NotNil(t, svc)
}

func TestAuthorizationServiceImpl_CheckPermission(t *testing.T) {
	svc := &AuthorizationServiceImpl{}

	tests := []struct {
		name     string
		userCtx  *entities.UserContext
		resource domain.ResourceType
		action   domain.ActionType
		scope    *Scope
		expected bool
	}{
		{
			"nil user ctx",
			nil,
			domain.ResourceUsers, domain.ActionRead, nil, false,
		},
		{
			"admin full access",
			&entities.UserContext{UserID: 1, Role: domain.RoleSystemAdmin},
			domain.ResourceUsers, domain.ActionCreate, nil, true,
		},
		{
			"student denied",
			&entities.UserContext{UserID: 1, Role: domain.RoleStudent},
			domain.ResourceUsers, domain.ActionCreate, nil, false,
		},
		{
			"teacher limited access no scope",
			&entities.UserContext{UserID: 1, Role: domain.RoleTeacher},
			domain.ResourceUsers, domain.ActionRead, nil, true,
		},
		{
			"teacher limited access with matching faculty",
			&entities.UserContext{UserID: 1, Role: domain.RoleTeacher, FacultyID: strPtr("F1")},
			domain.ResourceUsers, domain.ActionRead, &Scope{FacultyID: strPtr("F1")}, true,
		},
		{
			"teacher limited access with non-matching faculty",
			&entities.UserContext{UserID: 1, Role: domain.RoleTeacher, FacultyID: strPtr("F1")},
			domain.ResourceUsers, domain.ActionRead, &Scope{FacultyID: strPtr("F2")}, false,
		},
		{
			"teacher limited access with matching group",
			&entities.UserContext{UserID: 1, Role: domain.RoleTeacher, GroupID: strPtr("G1")},
			domain.ResourceUsers, domain.ActionRead, &Scope{GroupID: strPtr("G1")}, true,
		},
		{
			"teacher limited access with non-matching group",
			&entities.UserContext{UserID: 1, Role: domain.RoleTeacher, GroupID: strPtr("G1")},
			domain.ResourceUsers, domain.ActionRead, &Scope{GroupID: strPtr("G2")}, false,
		},
		{
			"student own access",
			&entities.UserContext{UserID: 1, Role: domain.RoleStudent},
			domain.ResourceUsers, domain.ActionUpdate, nil, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CheckPermission(tt.userCtx, tt.resource, tt.action, tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthorizationServiceImpl_CheckOwnership(t *testing.T) {
	svc := &AuthorizationServiceImpl{}

	assert.True(t, svc.CheckOwnership(1, 1))
	assert.False(t, svc.CheckOwnership(1, 2))
}

func TestAuthorizationServiceImpl_CanApproveDocument(t *testing.T) {
	svc := &AuthorizationServiceImpl{}

	tests := []struct {
		name     string
		userCtx  *entities.UserContext
		docType  string
		step     int
		expected bool
	}{
		{"nil context", nil, "curriculum", 1, false},
		{"curriculum step 1 methodist", &entities.UserContext{Role: domain.RoleMethodist}, "curriculum", 1, true},
		{"curriculum step 1 teacher", &entities.UserContext{Role: domain.RoleTeacher}, "curriculum", 1, true},
		{"curriculum step 1 student", &entities.UserContext{Role: domain.RoleStudent}, "curriculum", 1, false},
		{"curriculum step 2 methodist", &entities.UserContext{Role: domain.RoleMethodist}, "curriculum", 2, true},
		{"curriculum step 2 teacher", &entities.UserContext{Role: domain.RoleTeacher}, "curriculum", 2, false},
		{"curriculum step 3 secretary", &entities.UserContext{Role: domain.RoleAcademicSecretary}, "curriculum", 3, true},
		{"curriculum step 3 teacher", &entities.UserContext{Role: domain.RoleTeacher}, "curriculum", 3, false},
		{"curriculum step 4 admin", &entities.UserContext{Role: domain.RoleSystemAdmin}, "curriculum", 4, true},
		{"curriculum step 4 methodist", &entities.UserContext{Role: domain.RoleMethodist}, "curriculum", 4, true},
		{"curriculum step 4 teacher", &entities.UserContext{Role: domain.RoleTeacher}, "curriculum", 4, false},
		{"curriculum step 5 invalid", &entities.UserContext{Role: domain.RoleSystemAdmin}, "curriculum", 5, false},
		{"report admin", &entities.UserContext{Role: domain.RoleSystemAdmin}, "report", 1, false},
		{"unknown doc type", &entities.UserContext{Role: domain.RoleSystemAdmin}, "unknown", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, svc.CanApproveDocument(tt.userCtx, tt.docType, tt.step))
		})
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "password123", hash)
}

func TestCheckPasswordHash(t *testing.T) {
	hash, _ := HashPassword("password123")

	err := CheckPasswordHash("password123", hash)
	assert.NoError(t, err)

	err = CheckPasswordHash("wrongpassword", hash)
	assert.Error(t, err)
}

func strPtr(s string) *string {
	return &s
}
