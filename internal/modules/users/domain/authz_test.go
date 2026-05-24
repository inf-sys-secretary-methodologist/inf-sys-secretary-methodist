package domain_test

import (
	"errors"
	"testing"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	usersDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain"
)

// v0.160.1 polish Item 7 — direct tests for the 3 domain free
// functions previously exercised only via usecase tests. Table-driven
// per CLAUDE.md ≥3-variant gate.

func TestValidateAvatarKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		key      string
		targetID int64
		want     error
	}{
		{
			name:     "empty_key_allowed_for_clear",
			key:      "",
			targetID: 42,
			want:     nil,
		},
		{
			name:     "valid_prefix_for_own_id",
			key:      "avatars/42_abc.png",
			targetID: 42,
			want:     nil,
		},
		{
			name:     "wrong_user_prefix_rejected",
			key:      "avatars/99_abc.png",
			targetID: 42,
			want:     usersDomain.ErrInvalidAvatarKey,
		},
		{
			name:     "outside_avatars_folder_rejected",
			key:      "exams/secret.pdf",
			targetID: 42,
			want:     usersDomain.ErrInvalidAvatarKey,
		},
		{
			name:     "prefix_only_without_underscore_rejected",
			key:      "avatars/42.png",
			targetID: 42,
			want:     usersDomain.ErrInvalidAvatarKey,
		},
		{
			name:     "path_traversal_substring_rejected",
			key:      "../../avatars/42_abc.png",
			targetID: 42,
			want:     usersDomain.ErrInvalidAvatarKey,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := usersDomain.ValidateAvatarKey(tc.key, tc.targetID)
			if !errors.Is(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAuthorizeProfileEdit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		actorID   int64
		targetID  int64
		actorRole authDomain.RoleType
		want      error
	}{
		{
			name:      "self_edit_student_allowed",
			actorID:   42,
			targetID:  42,
			actorRole: authDomain.RoleStudent,
			want:      nil,
		},
		{
			name:      "self_edit_admin_allowed",
			actorID:   1,
			targetID:  1,
			actorRole: authDomain.RoleSystemAdmin,
			want:      nil,
		},
		{
			name:      "admin_override_allowed",
			actorID:   1,
			targetID:  42,
			actorRole: authDomain.RoleSystemAdmin,
			want:      nil,
		},
		{
			name:      "cross_user_methodist_rejected",
			actorID:   2,
			targetID:  42,
			actorRole: authDomain.RoleMethodist,
			want:      usersDomain.ErrProfileEditForbidden,
		},
		{
			name:      "cross_user_teacher_rejected",
			actorID:   3,
			targetID:  42,
			actorRole: authDomain.RoleTeacher,
			want:      usersDomain.ErrProfileEditForbidden,
		},
		{
			name:      "cross_user_secretary_rejected",
			actorID:   4,
			targetID:  42,
			actorRole: authDomain.RoleAcademicSecretary,
			want:      usersDomain.ErrProfileEditForbidden,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := usersDomain.AuthorizeProfileEdit(tc.actorID, tc.targetID, tc.actorRole)
			if !errors.Is(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAuthorizeUserDelete(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		actorID        int64
		targetID       int64
		targetRole     authDomain.RoleType
		adminHeadcount int
		want           error
	}{
		{
			name:           "self_delete_student_rejected",
			actorID:        42,
			targetID:       42,
			targetRole:     authDomain.RoleStudent,
			adminHeadcount: 5,
			want:           usersDomain.ErrCannotDeleteSelf,
		},
		{
			name:           "self_delete_admin_rejected_unconditionally",
			actorID:        1,
			targetID:       1,
			targetRole:     authDomain.RoleSystemAdmin,
			adminHeadcount: 10,
			want:           usersDomain.ErrCannotDeleteSelf,
		},
		{
			name:           "cross_delete_student_allowed",
			actorID:        1,
			targetID:       42,
			targetRole:     authDomain.RoleStudent,
			adminHeadcount: 5,
			want:           nil,
		},
		{
			name:           "last_admin_protected",
			actorID:        2,
			targetID:       42,
			targetRole:     authDomain.RoleSystemAdmin,
			adminHeadcount: 1,
			want:           usersDomain.ErrLastAdminProtected,
		},
		{
			name:           "last_admin_protected_zero_count",
			actorID:        2,
			targetID:       42,
			targetRole:     authDomain.RoleSystemAdmin,
			adminHeadcount: 0,
			want:           usersDomain.ErrLastAdminProtected,
		},
		{
			name:           "admin_delete_non_admin_safe",
			actorID:        1,
			targetID:       42,
			targetRole:     authDomain.RoleMethodist,
			adminHeadcount: 1,
			want:           nil,
		},
		{
			name:           "non_last_admin_deletable",
			actorID:        2,
			targetID:       42,
			targetRole:     authDomain.RoleSystemAdmin,
			adminHeadcount: 2,
			want:           nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := usersDomain.AuthorizeUserDelete(tc.actorID, tc.targetID, tc.targetRole, tc.adminHeadcount)
			if !errors.Is(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
