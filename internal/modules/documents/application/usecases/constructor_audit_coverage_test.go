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

// TestUseCaseConstructors_NilRepoPanic pins the contract: every workflow
// use-case constructor (Archive / Resubmit / AssignExecutor / MarkExecuted /
// Register / StartRouting / SignVisa / Submit / Approve / Reject) MUST
// panic with a name-specific sentinel when constructed with a nil repo,
// so a misconfigured DI never silently boots с a nil-deref at first
// Execute call.
//
// Closes the 60%-coverage gap on all ten New*UseCase constructors —
// happy-path already covered by per-usecase tests; this adds the panic
// branch via assert.PanicsWithValue.
func TestUseCaseConstructors_NilRepoPanic(t *testing.T) {
	audit := &fakeAuditSink{}
	clock := fixedClock(time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC))

	cases := []struct {
		name      string
		construct func()
		msg       string
	}{
		{
			name:      "Archive",
			construct: func() { _ = usecases.NewArchiveDocumentUseCase(nil, audit, clock) },
			msg:       "documents: NewArchiveDocumentUseCase requires non-nil repo",
		},
		{
			name:      "Resubmit",
			construct: func() { _ = usecases.NewResubmitDocumentUseCase(nil, audit, clock) },
			msg:       "documents: NewResubmitDocumentUseCase requires non-nil repo",
		},
		{
			name:      "AssignExecutor",
			construct: func() { _ = usecases.NewAssignExecutorUseCase(nil, audit, clock) },
			msg:       "documents: NewAssignExecutorUseCase requires non-nil repo",
		},
		{
			name:      "MarkExecuted",
			construct: func() { _ = usecases.NewMarkExecutedUseCase(nil, audit, clock) },
			msg:       "documents: NewMarkExecutedUseCase requires non-nil repo",
		},
		{
			name:      "RegisterDocument",
			construct: func() { _ = usecases.NewRegisterDocumentUseCase(nil, audit, clock) },
			msg:       "documents: NewRegisterDocumentUseCase requires non-nil repo",
		},
		{
			name:      "StartRouting",
			construct: func() { _ = usecases.NewStartRoutingUseCase(nil, audit, clock) },
			msg:       "documents: NewStartRoutingUseCase requires non-nil repo",
		},
		{
			name:      "SignVisa",
			construct: func() { _ = usecases.NewSignVisaUseCase(nil, audit, clock) },
			msg:       "documents: NewSignVisaUseCase requires non-nil repo",
		},
		{
			name:      "Submit",
			construct: func() { _ = usecases.NewSubmitDocumentUseCase(nil, audit, clock) },
			msg:       "documents: NewSubmitDocumentUseCase requires non-nil repo",
		},
		{
			name:      "Approve",
			construct: func() { _ = usecases.NewApproveDocumentUseCase(nil, audit, clock) },
			msg:       "documents: NewApproveDocumentUseCase requires non-nil repo",
		},
		{
			name:      "Reject",
			construct: func() { _ = usecases.NewRejectDocumentUseCase(nil, audit, clock) },
			msg:       "documents: NewRejectDocumentUseCase requires non-nil repo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.PanicsWithValue(t, tc.msg, tc.construct)
		})
	}
}

// TestUseCaseConstructors_NilClockFallsBackToTimeNow covers the
// `if clock == nil { clock = time.Now }` branch in every New*UseCase
// constructor — the second uncovered statement that kept these funcs
// at 60% per-func coverage. We don't assert the clock is exactly
// time.Now (it's a function value), only that construction succeeds
// без panic and returns a non-nil pointer.
func TestUseCaseConstructors_NilClockFallsBackToTimeNow(t *testing.T) {
	repo := newFakeRepo()
	audit := &fakeAuditSink{}

	t.Run("Archive", func(t *testing.T) {
		require.NotNil(t, usecases.NewArchiveDocumentUseCase(repo, audit, nil))
	})
	t.Run("Resubmit", func(t *testing.T) {
		require.NotNil(t, usecases.NewResubmitDocumentUseCase(repo, audit, nil))
	})
	t.Run("AssignExecutor", func(t *testing.T) {
		require.NotNil(t, usecases.NewAssignExecutorUseCase(repo, audit, nil))
	})
	t.Run("MarkExecuted", func(t *testing.T) {
		require.NotNil(t, usecases.NewMarkExecutedUseCase(repo, audit, nil))
	})
	t.Run("RegisterDocument", func(t *testing.T) {
		require.NotNil(t, usecases.NewRegisterDocumentUseCase(repo, audit, nil))
	})
	t.Run("StartRouting", func(t *testing.T) {
		require.NotNil(t, usecases.NewStartRoutingUseCase(repo, audit, nil))
	})
	t.Run("SignVisa", func(t *testing.T) {
		require.NotNil(t, usecases.NewSignVisaUseCase(repo, audit, nil))
	})
	t.Run("Submit", func(t *testing.T) {
		require.NotNil(t, usecases.NewSubmitDocumentUseCase(repo, audit, nil))
	})
	t.Run("Approve", func(t *testing.T) {
		require.NotNil(t, usecases.NewApproveDocumentUseCase(repo, audit, nil))
	})
	t.Run("Reject", func(t *testing.T) {
		require.NotNil(t, usecases.NewRejectDocumentUseCase(repo, audit, nil))
	})
}

// TestEmitAudit_NilSinkIsSilent pins the audit_sink.go:40 contract: when
// the AuditSink port is nil the emitAudit helper must early-return
// без panic. Closes 66.7% → 100% per-func gap. We trigger emitAudit
// indirectly via SubmitDocumentUseCase.Execute happy path (the success
// branch emits "document.submitted").
func TestEmitAudit_NilSinkIsSilent(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	actorID := int64(42)
	doc := draftDoc(1, actorID)
	repo := newFakeRepo(doc)

	uc := usecases.NewSubmitDocumentUseCase(repo, nil /* nil AuditSink */, fixedClock(now))

	got, err := uc.Execute(
		context.Background(),
		actorID,
		entities.RoleMethodist,
		usecases.SubmitDocumentInput{ID: 1},
	)
	require.NoError(t, err, "nil AuditSink must not break the success path")
	require.NotNil(t, got)
	assert.Equal(t, entities.DocumentStatusApproval, got.Status)
}

// TestEmitAudit_NilSinkSurvivesDenialBranch — same as above but for the
// denial code path (entity rejects the transition). emitAudit is also
// called inside the if-error branches; nil-sink early-return must hold
// for those too.
func TestEmitAudit_NilSinkSurvivesDenialBranch(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	actorID := int64(42)
	// Approve a doc that's not on approval — entity will reject.
	doc := draftDoc(1, actorID)
	repo := newFakeRepo(doc)

	uc := usecases.NewApproveDocumentUseCase(repo, nil, fixedClock(now))

	_, err := uc.Execute(context.Background(), actorID, usecases.ApproveDocumentInput{ID: 1})
	require.Error(t, err, "approve on draft must fail")
	// The point is: nil sink + denial path did not panic.
}
