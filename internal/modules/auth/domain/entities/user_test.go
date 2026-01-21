package entities

import (
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
)

func TestNewUser(t *testing.T) {
	email := "test@example.com"
	passwordHash := "hashedpassword123"
	name := "Test User"
	role := domain.RoleStudent

	user := NewUser(email, passwordHash, name, role)

	if user.Email != email {
		t.Errorf("expected email %q, got %q", email, user.Email)
	}
	if user.Password != passwordHash {
		t.Errorf("expected password hash %q, got %q", passwordHash, user.Password)
	}
	if user.Name != name {
		t.Errorf("expected name %q, got %q", name, user.Name)
	}
	if user.Role != role {
		t.Errorf("expected role %q, got %q", role, user.Role)
	}
	if user.Status != UserStatusActive {
		t.Errorf("expected status %q, got %q", UserStatusActive, user.Status)
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestUser_Activate(t *testing.T) {
	user := NewUser("test@example.com", "hash", "Test", domain.RoleStudent)
	user.Status = UserStatusInactive

	user.Activate()

	if user.Status != UserStatusActive {
		t.Errorf("expected status %q, got %q", UserStatusActive, user.Status)
	}
}

func TestUser_Deactivate(t *testing.T) {
	user := NewUser("test@example.com", "hash", "Test", domain.RoleStudent)

	user.Deactivate()

	if user.Status != UserStatusInactive {
		t.Errorf("expected status %q, got %q", UserStatusInactive, user.Status)
	}
}

func TestUser_Block(t *testing.T) {
	user := NewUser("test@example.com", "hash", "Test", domain.RoleStudent)

	user.Block()

	if user.Status != UserStatusBlocked {
		t.Errorf("expected status %q, got %q", UserStatusBlocked, user.Status)
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"active user", UserStatusActive, true},
		{"inactive user", UserStatusInactive, false},
		{"blocked user", UserStatusBlocked, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test@example.com", "hash", "Test", domain.RoleStudent)
			user.Status = tt.status

			got := user.IsActive()
			if got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name    string
		status  UserStatus
		wantErr error
	}{
		{"active user can login", UserStatusActive, nil},
		{"inactive user cannot login", UserStatusInactive, ErrAccountNotActive},
		// Note: blocked users return ErrAccountNotActive because IsActive() check comes first
		{"blocked user cannot login", UserStatusBlocked, ErrAccountNotActive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test@example.com", "hash", "Test", domain.RoleStudent)
			user.Status = tt.status

			err := user.CanLogin()
			if err != tt.wantErr {
				t.Errorf("CanLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_ToUserContext(t *testing.T) {
	user := NewUser("test@example.com", "hash", "Test", domain.RoleTeacher)
	user.ID = 42

	ctx := user.ToUserContext()

	if ctx.UserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, ctx.UserID)
	}
	if ctx.Role != user.Role {
		t.Errorf("expected role %q, got %q", user.Role, ctx.Role)
	}
}

func TestUserContext_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		role     domain.RoleType
		resource domain.ResourceType
		action   domain.ActionType
		want     bool
	}{
		{"admin can create users", domain.RoleSystemAdmin, domain.ResourceUsers, domain.ActionCreate, true},
		{"admin can read users", domain.RoleSystemAdmin, domain.ResourceUsers, domain.ActionRead, true},
		{"student cannot create users", domain.RoleStudent, domain.ResourceUsers, domain.ActionCreate, false},
		{"student can read own assignments", domain.RoleStudent, domain.ResourceAssignments, domain.ActionRead, true},
		{"teacher can create assignments", domain.RoleTeacher, domain.ResourceAssignments, domain.ActionCreate, true},
		{"methodist can read curriculum", domain.RoleMethodist, domain.ResourceCurriculum, domain.ActionRead, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &UserContext{
				UserID: 1,
				Role:   tt.role,
			}

			got := ctx.HasPermission(tt.resource, tt.action)
			if got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserContext_HasPermission_InvalidRole(t *testing.T) {
	ctx := &UserContext{
		UserID: 1,
		Role:   "invalid_role",
	}

	if ctx.HasPermission(domain.ResourceUsers, domain.ActionRead) {
		t.Error("expected no permission for invalid role")
	}
}

func TestUserContext_GetAccessLevel(t *testing.T) {
	tests := []struct {
		name     string
		role     domain.RoleType
		resource domain.ResourceType
		action   domain.ActionType
		want     domain.AccessLevel
	}{
		{"admin full access to users", domain.RoleSystemAdmin, domain.ResourceUsers, domain.ActionCreate, domain.AccessFull},
		{"student own access to assignments", domain.RoleStudent, domain.ResourceAssignments, domain.ActionRead, domain.AccessOwn},
		{"student denied access to reports", domain.RoleStudent, domain.ResourceReports, domain.ActionRead, domain.AccessDenied},
		{"teacher limited access to curriculum update", domain.RoleTeacher, domain.ResourceCurriculum, domain.ActionUpdate, domain.AccessLimited},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &UserContext{
				UserID: 1,
				Role:   tt.role,
			}

			got := ctx.GetAccessLevel(tt.resource, tt.action)
			if got != tt.want {
				t.Errorf("GetAccessLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserContext_GetAccessLevel_InvalidRole(t *testing.T) {
	ctx := &UserContext{
		UserID: 1,
		Role:   "invalid_role",
	}

	if ctx.GetAccessLevel(domain.ResourceUsers, domain.ActionRead) != domain.AccessDenied {
		t.Error("expected AccessDenied for invalid role")
	}
}

func TestUserContext_GetAccessLevel_InvalidResource(t *testing.T) {
	ctx := &UserContext{
		UserID: 1,
		Role:   domain.RoleSystemAdmin,
	}

	if ctx.GetAccessLevel("invalid_resource", domain.ActionRead) != domain.AccessDenied {
		t.Error("expected AccessDenied for invalid resource")
	}
}

func TestUserContext_GetAccessLevel_InvalidAction(t *testing.T) {
	ctx := &UserContext{
		UserID: 1,
		Role:   domain.RoleSystemAdmin,
	}

	if ctx.GetAccessLevel(domain.ResourceUsers, "invalid_action") != domain.AccessDenied {
		t.Error("expected AccessDenied for invalid action")
	}
}

func TestUserStatusConstants(t *testing.T) {
	if UserStatusActive != "active" {
		t.Errorf("expected UserStatusActive to be %q, got %q", "active", UserStatusActive)
	}
	if UserStatusInactive != "inactive" {
		t.Errorf("expected UserStatusInactive to be %q, got %q", "inactive", UserStatusInactive)
	}
	if UserStatusBlocked != "blocked" {
		t.Errorf("expected UserStatusBlocked to be %q, got %q", "blocked", UserStatusBlocked)
	}
}
