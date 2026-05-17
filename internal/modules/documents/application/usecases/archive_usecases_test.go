package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// TestArchiveDocumentUseCase pins the v0.152.0 ArchiveDocumentUseCase
// contract: load → entity.Archive → repo.Update + AuditSink emit on
// each outcome. Mirror к MarkExecutedUseCase pattern.
//
// Issue: #233
func TestArchiveDocumentUseCase(t *testing.T) {
	now := time.Date(2026, 5, 17, 14, 0, 0, 0, time.UTC)
	actorID := int64(7)

	cases := []struct {
		name      string
		startDoc  *entities.Document
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecuted), wantAudit: "document.archived"},
		{name: "not-found", startDoc: nil, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.archive_denied", wantDeny: "not_found"},
		{name: "non-executed status rejected", startDoc: draftDoc(1, 42), wantErr: entities.ErrCannotArchive, wantAudit: "document.archive_denied", wantDeny: "not_executed"},
		{name: "execution status rejected (must mark first)", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecution), wantErr: entities.ErrCannotArchive, wantAudit: "document.archive_denied", wantDeny: "not_executed"},
		{name: "archived status rejected (already)", startDoc: docAtStatus(1, 42, entities.DocumentStatusArchived), wantErr: entities.ErrCannotArchive, wantAudit: "document.archive_denied", wantDeny: "not_executed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewArchiveDocumentUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), actorID, usecases.ArchiveDocumentInput{ID: 1})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				rec := audit.Last()
				assert.Equal(t, tc.wantAudit, rec.Action)
				if tc.wantDeny != "" {
					assert.Equal(t, tc.wantDeny, rec.Fields["reason"])
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, entities.DocumentStatusArchived, got.Status)
			require.NotNil(t, got.ArchivedBy)
			assert.Equal(t, actorID, *got.ArchivedBy)
			assert.Equal(t, "document.archived", audit.Last().Action)
		})
	}
}

// TestResubmitDocumentUseCase pins the v0.152.0 ResubmitDocumentUseCase
// contract: load → author-or-edit-role gate → entity.Resubmit →
// repo.Update + AuditSink emit. Mirror к SubmitDocumentUseCase auth gate.
//
// Issue: #233
func TestResubmitDocumentUseCase(t *testing.T) {
	now := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)
	authorID := int64(42)

	cases := []struct {
		name      string
		startDoc  *entities.Document
		actorID   int64
		role      entities.UserRole
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "author resubmits own rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusRejected), actorID: authorID, role: entities.RoleTeacher, wantAudit: "document.resubmitted"},
		{name: "methodist resubmits any rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusRejected), actorID: 99, role: entities.RoleMethodist, wantAudit: "document.resubmitted"},
		{name: "secretary resubmits any rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusRejected), actorID: 99, role: entities.RoleAcademicSecretary, wantAudit: "document.resubmitted"},
		{name: "admin resubmits any rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusRejected), actorID: 99, role: entities.RoleSystemAdmin, wantAudit: "document.resubmitted"},
		{name: "teacher-non-author rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusRejected), actorID: 99, role: entities.RoleTeacher, wantErr: usecases.ErrDocumentForbidden, wantAudit: "document.resubmit_denied", wantDeny: "forbidden"},
		{name: "student rejected (defense-in-depth)", startDoc: docAtStatus(1, authorID, entities.DocumentStatusRejected), actorID: 99, role: entities.RoleStudent, wantErr: usecases.ErrDocumentForbidden, wantAudit: "document.resubmit_denied", wantDeny: "forbidden"},
		{name: "not-found", startDoc: nil, actorID: authorID, role: entities.RoleTeacher, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.resubmit_denied", wantDeny: "not_found"},
		{name: "non-rejected status rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusDraft), actorID: authorID, role: entities.RoleTeacher, wantErr: entities.ErrCannotResubmit, wantAudit: "document.resubmit_denied", wantDeny: "not_rejected"},
		{name: "approved status rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusApproved), actorID: 99, role: entities.RoleSystemAdmin, wantErr: entities.ErrCannotResubmit, wantAudit: "document.resubmit_denied", wantDeny: "not_rejected"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewResubmitDocumentUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), tc.actorID, tc.role, usecases.ResubmitDocumentInput{ID: 1})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				rec := audit.Last()
				assert.Equal(t, tc.wantAudit, rec.Action)
				if tc.wantDeny != "" {
					assert.Equal(t, tc.wantDeny, rec.Fields["reason"])
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, entities.DocumentStatusDraft, got.Status)
			assert.Nil(t, got.RejectedBy, "RejectedBy must be cleared on resubmit")
			assert.Nil(t, got.RejectedReason, "RejectedReason must be cleared on resubmit")
			assert.Equal(t, "document.resubmitted", audit.Last().Action)
		})
	}
}
