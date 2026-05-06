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
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

// fakeStudentScopedRepo is a minimal fake implementing only the read
// methods the My Assignments use case needs. It deliberately does not
// reuse fakeSubmissionRepo (which carries grade/return state for
// mutation tests) — the read-side fake stays small so the test file
// reads top-down.
type fakeStudentScopedRepo struct {
	out           []views.StudentAssignmentView
	err           error
	lastStudentID int64
	lastStatus    *entities.SubmissionStatus
}

func (r *fakeStudentScopedRepo) ListByStudent(ctx context.Context, studentID int64, status *entities.SubmissionStatus) ([]views.StudentAssignmentView, error) {
	r.lastStudentID = studentID
	r.lastStatus = status
	if r.err != nil {
		return nil, r.err
	}
	return r.out, nil
}

func TestListMyAssignmentsUseCase_Execute(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	pending := entities.StatusPending

	rows := []views.StudentAssignmentView{
		{AssignmentID: 10, Title: "Lab 1", Subject: "Math", GroupName: "БСБО-01-22", MaxScore: 100, SubmissionID: 1, StudentID: 7, Status: entities.StatusPending, AssignmentCreatedAt: now, SubmissionCreatedAt: now},
		{AssignmentID: 11, Title: "Lab 2", Subject: "Math", GroupName: "БСБО-01-22", MaxScore: 50, SubmissionID: 2, StudentID: 7, Status: entities.StatusGraded, AssignmentCreatedAt: now, SubmissionCreatedAt: now},
	}

	tests := []struct {
		name          string
		studentID     int64
		statusFilter  *entities.SubmissionStatus
		repoOut       []views.StudentAssignmentView
		repoErr       error
		wantErr       bool
		wantCount     int
		wantStatusFwd *entities.SubmissionStatus
	}{
		{
			name:          "happy path returns rows from repo",
			studentID:     7,
			repoOut:       rows,
			wantCount:     2,
			wantStatusFwd: nil,
		},
		{
			name:          "status filter passes through verbatim",
			studentID:     7,
			statusFilter:  &pending,
			repoOut:       rows[:1],
			wantCount:     1,
			wantStatusFwd: &pending,
		},
		{
			name:      "repo error wraps",
			studentID: 7,
			repoErr:   errors.New("db down"),
			wantErr:   true,
		},
		{
			name:      "non-positive student id is rejected",
			studentID: 0,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeStudentScopedRepo{out: tc.repoOut, err: tc.repoErr}
			uc := usecases.NewListMyAssignmentsUseCase(repo)

			got, err := uc.Execute(context.Background(), usecases.ListMyAssignmentsInput{
				StudentID: tc.studentID,
				Status:    tc.statusFilter,
			})
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tc.wantCount)
			assert.Equal(t, tc.studentID, repo.lastStudentID)
			assert.Equal(t, tc.wantStatusFwd, repo.lastStatus)
		})
	}
}

func TestNewListMyAssignmentsUseCase_PanicsOnNilRepo(t *testing.T) {
	assert.Panics(t, func() {
		usecases.NewListMyAssignmentsUseCase(nil)
	})
}
