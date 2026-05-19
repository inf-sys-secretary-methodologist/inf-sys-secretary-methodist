package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// Each workflow Execute path has an identical "if repo.Update fails,
// propagate without emitting success audit" branch. Existing tests
// cover the happy + denial paths for all use cases, but the Update-
// error branch was only pinned for Reject (workflow_usecases_test.go).
// This file closes that gap for Submit / Approve / Archive / Resubmit /
// MarkExecuted / Register / StartRouting — bringing each Execute func
// from 92-94% per-func toward 100% by exercising the same transport-
// failure invariant: error returned, success audit absent.
//
// Closes Phase 6 #196 branch gaps на documents/application/usecases.

const updateErrSentinel = "transport down"

func assertNoSuccessAudit(t *testing.T, audit *fakeAuditSink, successAction string) {
	t.Helper()
	for _, rec := range audit.records {
		assert.NotEqual(t, successAction, rec.Action, "success audit must not fire on transport failure")
	}
}

func TestSubmitDocumentUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := draftDoc(1, 42)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewSubmitDocumentUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 42, entities.RoleMethodist, usecases.SubmitDocumentInput{ID: 1})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.submitted")
}

func TestApproveDocumentUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusApproval)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewApproveDocumentUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 7, usecases.ApproveDocumentInput{ID: 1})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.approved")
}

func TestArchiveDocumentUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusExecuted)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewArchiveDocumentUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 7, usecases.ArchiveDocumentInput{ID: 1})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.archived")
}

func TestResubmitDocumentUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusRejected)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewResubmitDocumentUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 42, entities.RoleMethodist, usecases.ResubmitDocumentInput{ID: 1})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.resubmitted")
}

func TestMarkExecutedUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusExecution)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewMarkExecutedUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 7, usecases.MarkExecutedInput{ID: 1})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.executed")
}

func TestRegisterDocumentUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusApproved)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewRegisterDocumentUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 7, usecases.RegisterDocumentInput{ID: 1, Number: "REG-2026-001"})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.registered")
}

func TestStartRoutingUseCase_RepoUpdateError(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC)
	doc := docAtStatus(1, 42, entities.DocumentStatusRegistered)
	repo := newFakeRepo(doc)
	repo.updErr = errors.New(updateErrSentinel)
	audit := &fakeAuditSink{}
	uc := usecases.NewStartRoutingUseCase(repo, audit, fixedClock(now))

	_, err := uc.Execute(context.Background(), 7, usecases.StartRoutingInput{ID: 1})
	require.Error(t, err)
	assertNoSuccessAudit(t, audit, "document.routed")
}
