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

// TestAssignExecutorUseCase pins the v0.151.0 AssignExecutorUseCase
// contract: validate executor → load → entity.AssignExecutor →
// repo.Update + AuditSink emit on each outcome. Mirror к
// StartRoutingUseCase pattern с extra invalid_executor branch.
//
// Issue: #232
func TestAssignExecutorUseCase(t *testing.T) {
	now := time.Date(2026, 5, 17, 11, 0, 0, 0, time.UTC)
	due := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	actorID := int64(7)
	executorID := int64(13)

	cases := []struct {
		name       string
		startDoc   *entities.Document
		executorID int64
		dueDate    *time.Time
		wantErr    error
		wantAudit  string
		wantDeny   string
	}{
		{name: "happy path nil due", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecution), executorID: executorID, dueDate: nil, wantAudit: "document.executor_assigned"},
		{name: "happy path с due", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecution), executorID: executorID, dueDate: &due, wantAudit: "document.executor_assigned"},
		{name: "zero executor id rejected", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecution), executorID: 0, wantErr: usecases.ErrInvalidExecutor, wantAudit: "document.assign_executor_denied", wantDeny: "invalid_executor"},
		{name: "negative executor id rejected", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecution), executorID: -5, wantErr: usecases.ErrInvalidExecutor, wantAudit: "document.assign_executor_denied", wantDeny: "invalid_executor"},
		{name: "not-found", startDoc: nil, executorID: executorID, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.assign_executor_denied", wantDeny: "not_found"},
		{name: "non-execution status rejected", startDoc: draftDoc(1, 42), executorID: executorID, wantErr: entities.ErrCannotAssignExecutor, wantAudit: "document.assign_executor_denied", wantDeny: "not_execution"},
		{name: "executed status rejected (already past)", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecuted), executorID: executorID, wantErr: entities.ErrCannotAssignExecutor, wantAudit: "document.assign_executor_denied", wantDeny: "not_execution"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewAssignExecutorUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), actorID, usecases.AssignExecutorInput{ID: 1, ExecutorID: tc.executorID, DueDate: tc.dueDate})

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
			assert.Equal(t, entities.DocumentStatusExecution, got.Status, "AssignExecutor must NOT change status")
			require.NotNil(t, got.ExecutorAssignedTo)
			assert.Equal(t, executorID, *got.ExecutorAssignedTo)
			assert.Equal(t, "document.executor_assigned", audit.Last().Action)
			assert.EqualValues(t, executorID, audit.Last().Fields["executor_user_id"])
		})
	}
}

// TestMarkExecutedUseCase pins the v0.151.0 MarkExecutedUseCase contract:
// load → entity.MarkExecuted → repo.Update + AuditSink emit.
//
// Issue: #232
func TestMarkExecutedUseCase(t *testing.T) {
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	actorID := int64(7)

	cases := []struct {
		name      string
		startDoc  *entities.Document
		wantErr   error
		wantAudit string
		wantDeny  string
	}{
		{name: "happy path", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecution), wantAudit: "document.executed"},
		{name: "not-found", startDoc: nil, wantErr: usecases.ErrDocumentNotFound, wantAudit: "document.mark_executed_denied", wantDeny: "not_found"},
		{name: "non-execution status rejected", startDoc: draftDoc(1, 42), wantErr: entities.ErrCannotMarkExecuted, wantAudit: "document.mark_executed_denied", wantDeny: "not_execution"},
		{name: "executed status rejected (already)", startDoc: docAtStatus(1, 42, entities.DocumentStatusExecuted), wantErr: entities.ErrCannotMarkExecuted, wantAudit: "document.mark_executed_denied", wantDeny: "not_execution"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tc.startDoc != nil {
				repo = newFakeRepo(tc.startDoc)
			}
			audit := &fakeAuditSink{}
			uc := usecases.NewMarkExecutedUseCase(repo, audit, fixedClock(now))
			got, err := uc.Execute(context.Background(), actorID, usecases.MarkExecutedInput{ID: 1})

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
			assert.Equal(t, entities.DocumentStatusExecuted, got.Status)
			require.NotNil(t, got.ExecutedBy)
			assert.Equal(t, actorID, *got.ExecutedBy)
			assert.Equal(t, "document.executed", audit.Last().Action)
		})
	}
}
