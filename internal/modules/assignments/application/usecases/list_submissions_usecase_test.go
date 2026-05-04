package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

var errFakeSubmissionRepoFault = errors.New("db down")

func TestListSubmissionsUseCase_Execute(t *testing.T) {
	statusPending := entities.StatusPending
	statusGraded := entities.StatusGraded

	twoSubs := []views.SubmissionView{
		{ID: 1, AssignmentID: assignmentID, StudentID: 7, StudentName: "Иван Петров", Status: entities.StatusPending, CreatedAt: time.Now()},
		{ID: 2, AssignmentID: assignmentID, StudentID: 8, StudentName: "Анна Смирнова", Status: entities.StatusGraded, CreatedAt: time.Now()},
	}

	tests := []struct {
		name           string
		caller         usecases.CallerScope
		askID          int64
		seedID         int64
		statusFilter   *entities.SubmissionStatus
		repoSubs       []views.SubmissionView
		repoErr        error
		assignmentErr  error
		wantErr        error
		wantStatus     *entities.SubmissionStatus
		wantCount      int
	}{
		{
			name:       "teacher lists own assignment submissions",
			caller:     usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:      assignmentID,
			seedID:     assignmentID,
			repoSubs:   twoSubs,
			wantCount:  2,
		},
		{
			name:    "teacher denied on foreign assignment",
			caller:  usecases.CallerScope{UserID: otherTeacherID, Unrestricted: false},
			askID:   assignmentID,
			seedID:  assignmentID,
			wantErr: entities.ErrAssignmentScopeForbidden,
		},
		{
			name:      "methodist lists any assignment submissions",
			caller:    usecases.CallerScope{UserID: int64(2), Unrestricted: true},
			askID:     assignmentID,
			seedID:    assignmentID,
			repoSubs:  twoSubs,
			wantCount: 2,
		},
		{
			name:    "assignment not found surfaces sentinel",
			caller:  usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:   999,
			seedID:  assignmentID,
			wantErr: repositories.ErrAssignmentNotFound,
		},
		{
			name:         "status=pending filter passes through",
			caller:       usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:        assignmentID,
			seedID:       assignmentID,
			statusFilter: &statusPending,
			repoSubs:     twoSubs[:1],
			wantStatus:   &statusPending,
			wantCount:    1,
		},
		{
			name:         "status=graded filter passes through",
			caller:       usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:        assignmentID,
			seedID:       assignmentID,
			statusFilter: &statusGraded,
			repoSubs:     twoSubs[1:],
			wantStatus:   &statusGraded,
			wantCount:    1,
		},
		{
			name:      "nil status returns all submissions",
			caller:    usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:     assignmentID,
			seedID:    assignmentID,
			repoSubs:  twoSubs,
			wantCount: 2,
		},
		{
			// Verify the sentinel survives the use-case wrap via errors.Is
			// — a future change replacing %w with %v would break the chain
			// and silently let this test pass with assert.Contains.
			name:    "submission repo error wrapped (sentinel preserved)",
			caller:  usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:   assignmentID,
			seedID:  assignmentID,
			repoErr: errFakeSubmissionRepoFault,
			wantErr: errFakeSubmissionRepoFault,
		},
		{
			name:      "empty submissions returned as empty slice",
			caller:    usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:     assignmentID,
			seedID:    assignmentID,
			repoSubs:  nil,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ar := newFakeAssignmentRepo()
			ar.seed(makeAssignment(t, tc.seedID, authorTeacherID))

			sr := newFakeSubmissionRepo()
			sr.listResult = tc.repoSubs
			sr.listErr = tc.repoErr

			uc := usecases.NewListSubmissionsUseCase(ar, sr)
			out, err := uc.Execute(context.Background(), usecases.ListSubmissionsInput{
				Caller:       tc.caller,
				AssignmentID: tc.askID,
				Status:       tc.statusFilter,
			})

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected wrap of %v, got %v", tc.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantCount, len(out))

			if tc.wantStatus == nil {
				assert.Nil(t, sr.lastListStatus)
			} else {
				require.NotNil(t, sr.lastListStatus, "status filter must be passed to repo")
				assert.Equal(t, *tc.wantStatus, *sr.lastListStatus)
			}
			assert.Equal(t, tc.askID, sr.lastListAssignmentID)
		})
	}
}
