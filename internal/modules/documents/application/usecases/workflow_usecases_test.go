package usecases_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// --- test doubles ---

// fakeWorkflowRepo is the narrow port the v0.148.0 workflow use cases
// require: GetByID + Update. Captures the last persisted doc so tests
// can assert the audit fields the entity set.
type fakeWorkflowRepo struct {
	mu       sync.Mutex
	stored   map[int64]*entities.Document
	getErr   error
	updErr   error
	updated  *entities.Document
	getCalls int
	updCalls int
}

func newFakeRepo(docs ...*entities.Document) *fakeWorkflowRepo {
	r := &fakeWorkflowRepo{stored: map[int64]*entities.Document{}}
	for _, d := range docs {
		r.stored[d.ID] = d
	}
	return r
}

func (r *fakeWorkflowRepo) GetByID(_ context.Context, id int64) (*entities.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getCalls++
	if r.getErr != nil {
		return nil, r.getErr
	}
	if d, ok := r.stored[id]; ok {
		return d, nil
	}
	return nil, usecases.ErrDocumentNotFound
}

func (r *fakeWorkflowRepo) Update(_ context.Context, d *entities.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updCalls++
	if r.updErr != nil {
		return r.updErr
	}
	r.updated = d
	r.stored[d.ID] = d
	return nil
}

type auditRecord struct {
	Action   string
	Resource string
	Fields   map[string]any
}

type fakeAuditSink struct {
	mu      sync.Mutex
	records []auditRecord
}

func (s *fakeAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, auditRecord{Action: action, Resource: resource, Fields: fields})
}

func (s *fakeAuditSink) Last() auditRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) == 0 {
		return auditRecord{}
	}
	return s.records[len(s.records)-1]
}

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func draftDoc(id, authorID int64) *entities.Document {
	d := entities.NewDocument("План занятий 2026", 1, authorID)
	d.ID = id
	d.Status = entities.DocumentStatusDraft
	return d
}

func docAtStatus(id, authorID int64, status entities.DocumentStatus) *entities.Document {
	d := draftDoc(id, authorID)
	d.Status = status
	return d
}

// --- Submit use case ---

// TestSubmitDocumentUseCase pins the Submit transition с author OR
// edit-role authorization gate + audit emit.
//
// Issue: #227
func TestSubmitDocumentUseCase(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
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
		{name: "author submits own draft", startDoc: draftDoc(1, authorID), actorID: authorID, role: entities.RoleTeacher, wantAudit: "document.submitted"},
		{name: "methodist submits any draft", startDoc: draftDoc(1, authorID), actorID: 99, role: entities.RoleMethodist, wantAudit: "document.submitted"},
		{name: "secretary submits any draft", startDoc: draftDoc(1, authorID), actorID: 99, role: entities.RoleAcademicSecretary, wantAudit: "document.submitted"},
		{name: "admin submits any draft", startDoc: draftDoc(1, authorID), actorID: 99, role: entities.RoleSystemAdmin, wantAudit: "document.submitted"},
		{name: "teacher-non-author rejected", startDoc: draftDoc(1, authorID), actorID: 99, role: entities.RoleTeacher, wantErr: usecases.ErrDocumentForbidden, wantAudit: "document.submit_denied", wantDeny: "forbidden"},
		{name: "student rejected (defense-in-depth)", startDoc: draftDoc(1, authorID), actorID: 99, role: entities.RoleStudent, wantErr: usecases.ErrDocumentForbidden, wantAudit: "document.submit_denied", wantDeny: "forbidden"},
		{name: "not-found", startDoc: nil, actorID: authorID, role: entities.RoleTeacher, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.submit_denied", wantDeny: "not_found"},
		{name: "non-draft rejected", startDoc: docAtStatus(1, authorID, entities.DocumentStatusApproved), actorID: authorID, role: entities.RoleTeacher, wantErr: entities.ErrCannotSubmit, wantAudit: "document.submit_denied", wantDeny: "not_draft"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var repo *fakeWorkflowRepo
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			} else {
				repo = newFakeRepo()
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewSubmitDocumentUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), tc.actorID, tc.role, usecases.SubmitDocumentInput{ID: 1})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				if tc.wantAudit != "" {
					rec := audit.Last()
					assert.Equal(t, tc.wantAudit, rec.Action)
					if tc.wantDeny != "" {
						assert.Equal(t, tc.wantDeny, rec.Fields["reason"])
					}
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, entities.DocumentStatusApproval, got.Status)
			assert.NotNil(t, got.SubmittedBy)
			assert.Equal(t, tc.actorID, *got.SubmittedBy)
			assert.NotNil(t, got.SubmittedAt)
			assert.Equal(t, "document.submitted", audit.Last().Action)
		})
	}
}

// --- Approve use case ---

func TestApproveDocumentUseCase(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	adminID := int64(7)

	cases := []struct {
		name      string
		startDoc  *entities.Document
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproval), wantAudit: "document.approved"},
		{name: "not-found", startDoc: nil, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.approve_denied", wantDeny: "not_found"},
		{name: "non-approval status rejected", startDoc: draftDoc(1, 42), wantErr: entities.ErrCannotApprove, wantAudit: "document.approve_denied", wantDeny: "not_approval"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewApproveDocumentUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), adminID, usecases.ApproveDocumentInput{ID: 1})

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
			assert.Equal(t, entities.DocumentStatusApproved, got.Status)
			assert.NotNil(t, got.ApprovedBy)
			assert.Equal(t, adminID, *got.ApprovedBy)
			assert.Equal(t, "document.approved", audit.Last().Action)
		})
	}
}

// --- Reject use case ---

func TestRejectDocumentUseCase(t *testing.T) {
	now := time.Date(2026, 5, 16, 14, 0, 0, 0, time.UTC)
	adminID := int64(7)
	validReason := "Шаблон 2023 устарел — обновите за неделю"
	shortReason := "коротко"

	cases := []struct {
		name      string
		startDoc  *entities.Document
		reason    string
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproval), reason: validReason, wantAudit: "document.rejected"},
		{name: "not-found", startDoc: nil, reason: validReason, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.reject_denied", wantDeny: "not_found"},
		{name: "non-approval status rejected", startDoc: draftDoc(1, 42), reason: validReason, wantErr: entities.ErrCannotReject, wantAudit: "document.reject_denied", wantDeny: "not_approval"},
		{name: "short reason rejected", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproval), reason: shortReason, wantErr: entities.ErrRejectionReasonInvalid, wantAudit: "document.reject_denied", wantDeny: "invalid_reason"},
		{name: "empty reason rejected", startDoc: docAtStatus(1, 42, entities.DocumentStatusApproval), reason: "", wantErr: entities.ErrRejectionReasonInvalid, wantAudit: "document.reject_denied", wantDeny: "invalid_reason"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewRejectDocumentUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), adminID, usecases.RejectDocumentInput{ID: 1, Reason: tc.reason})

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
			assert.Equal(t, entities.DocumentStatusRejected, got.Status)
			assert.NotNil(t, got.RejectedBy)
			assert.Equal(t, adminID, *got.RejectedBy)
			assert.NotNil(t, got.RejectedReason)
			assert.Equal(t, "document.rejected", audit.Last().Action)
		})
	}
}

// TestRejectDocumentUseCase_RepoUpdateError pins что transport failure
// после successful domain transition does не emit success audit (
// audit log = policy decisions, not infra). Mirror к curriculum.
func TestRejectDocumentUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 16, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusApproval)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New("db down")
	audit := &fakeAuditSink{}
	uc := usecases.NewRejectDocumentUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 7, usecases.RejectDocumentInput{ID: 1, Reason: "Корректный обоснованный отказ"})
	require.Error(t, err)
	for _, rec := range audit.records {
		assert.NotEqual(t, "document.rejected", rec.Action, "success audit must not fire on transport failure")
	}
}
