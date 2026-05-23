package domain_test

import (
	"errors"
	"testing"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

func TestAuthorizeFileAccess(t *testing.T) {
	const uploaderID = int64(10)
	const otherID = int64(99)

	file := &entities.FileMetadata{ID: 1, UploadedBy: uploaderID}

	cases := []struct {
		name      string
		actorID   int64
		actorRole authDomain.RoleType
		action    domain.FileAction
		wantErr   bool
	}{
		{"uploader reads own file", uploaderID, authDomain.RoleStudent, domain.FileActionRead, false},
		{"uploader attaches own file", uploaderID, authDomain.RoleStudent, domain.FileActionAttach, false},
		{"uploader creates version of own file", uploaderID, authDomain.RoleStudent, domain.FileActionCreateVersion, false},
		{"uploader deletes own file", uploaderID, authDomain.RoleStudent, domain.FileActionDelete, false},

		{"other student reads — denied", otherID, authDomain.RoleStudent, domain.FileActionRead, true},
		{"other student attaches — denied", otherID, authDomain.RoleStudent, domain.FileActionAttach, true},
		{"other student creates version — denied", otherID, authDomain.RoleStudent, domain.FileActionCreateVersion, true},
		{"other student deletes — denied", otherID, authDomain.RoleStudent, domain.FileActionDelete, true},

		{"system_admin reads other's file — allowed", otherID, authDomain.RoleSystemAdmin, domain.FileActionRead, false},
		{"system_admin attaches other's file — denied (no admin write override)", otherID, authDomain.RoleSystemAdmin, domain.FileActionAttach, true},
		{"system_admin creates version в чужом файле — denied", otherID, authDomain.RoleSystemAdmin, domain.FileActionCreateVersion, true},
		{"system_admin deletes other's file — denied", otherID, authDomain.RoleSystemAdmin, domain.FileActionDelete, true},

		{"methodist reads other's file — denied (no role override for read)", otherID, authDomain.RoleMethodist, domain.FileActionRead, true},
		{"academic_secretary reads other's file — denied", otherID, authDomain.RoleAcademicSecretary, domain.FileActionRead, true},
		{"teacher reads other's file — denied", otherID, authDomain.RoleTeacher, domain.FileActionRead, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.AuthorizeFileAccess(tc.actorID, tc.actorRole, file, tc.action)
			if tc.wantErr {
				if !errors.Is(err, domain.ErrFileAccessDenied) {
					t.Errorf("AuthorizeFileAccess() = %v, want ErrFileAccessDenied", err)
				}
				return
			}
			if err != nil {
				t.Errorf("AuthorizeFileAccess() = %v, want nil", err)
			}
		})
	}
}

func TestAuthorizeFileAccess_NilFile(t *testing.T) {
	err := domain.AuthorizeFileAccess(1, authDomain.RoleSystemAdmin, nil, domain.FileActionRead)
	if !errors.Is(err, domain.ErrFileAccessDenied) {
		t.Errorf("AuthorizeFileAccess(nil file) = %v, want ErrFileAccessDenied", err)
	}
}

func TestAuthorizeFileAccess_UnknownAction(t *testing.T) {
	file := &entities.FileMetadata{ID: 1, UploadedBy: 10}
	err := domain.AuthorizeFileAccess(10, authDomain.RoleStudent, file, domain.FileAction("unknown"))
	if !errors.Is(err, domain.ErrFileAccessDenied) {
		t.Errorf("AuthorizeFileAccess(unknown action) = %v, want ErrFileAccessDenied (default-deny)", err)
	}
}
