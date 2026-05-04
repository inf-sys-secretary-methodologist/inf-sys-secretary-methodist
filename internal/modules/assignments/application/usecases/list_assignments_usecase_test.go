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
)

var errFakeAssignmentRepoFault = errors.New("db down")

func TestListAssignmentsUseCase_Execute(t *testing.T) {
	caller := func(userID int64, unrestricted bool) usecases.CallerScope {
		return usecases.CallerScope{UserID: userID, Unrestricted: unrestricted}
	}

	tests := []struct {
		name         string
		input        usecases.ListAssignmentsInput
		repoErr      error
		repoResult   repositories.AssignmentListResult
		wantErr      error
		wantFilter   repositories.AssignmentListFilter
		wantItems    int
		wantTotalOut int
	}{
		{
			name: "teacher caller forces TeacherID filter",
			input: usecases.ListAssignmentsInput{
				Caller: caller(authorTeacherID, false),
			},
			wantFilter: repositories.AssignmentListFilter{
				TeacherID: ptrInt64(authorTeacherID),
				Limit:     usecases.DefaultListLimit,
				Offset:    0,
			},
		},
		{
			name: "methodist (unrestricted) sees all — no TeacherID filter",
			input: usecases.ListAssignmentsInput{
				Caller: caller(otherTeacherID, true),
			},
			wantFilter: repositories.AssignmentListFilter{
				TeacherID: nil,
				Limit:     usecases.DefaultListLimit,
				Offset:    0,
			},
		},
		{
			name: "subject filter passes through",
			input: usecases.ListAssignmentsInput{
				Caller:  caller(authorTeacherID, false),
				Subject: "Algorithms",
			},
			wantFilter: repositories.AssignmentListFilter{
				TeacherID: ptrInt64(authorTeacherID),
				Subject:   "Algorithms",
				Limit:     usecases.DefaultListLimit,
			},
		},
		{
			name: "group_name filter passes through",
			input: usecases.ListAssignmentsInput{
				Caller:    caller(authorTeacherID, false),
				GroupName: "ИС-21",
			},
			wantFilter: repositories.AssignmentListFilter{
				TeacherID: ptrInt64(authorTeacherID),
				GroupName: "ИС-21",
				Limit:     usecases.DefaultListLimit,
			},
		},
		{
			name: "limit clamped to maximum",
			input: usecases.ListAssignmentsInput{
				Caller: caller(authorTeacherID, true),
				Limit:  10000,
			},
			wantFilter: repositories.AssignmentListFilter{
				Limit: usecases.MaxListLimit,
			},
		},
		{
			name: "negative offset clamped to zero",
			input: usecases.ListAssignmentsInput{
				Caller: caller(authorTeacherID, true),
				Offset: -50,
			},
			wantFilter: repositories.AssignmentListFilter{
				Limit:  usecases.DefaultListLimit,
				Offset: 0,
			},
		},
		{
			name: "explicit limit honoured when within bounds",
			input: usecases.ListAssignmentsInput{
				Caller: caller(authorTeacherID, true),
				Limit:  25,
				Offset: 10,
			},
			wantFilter: repositories.AssignmentListFilter{
				Limit:  25,
				Offset: 10,
			},
		},
		{
			// errors.Is preserves the sentinel through the use-case
			// "list assignments:" wrap; a regression that replaces %w
			// with %v would break the chain and the assertion below.
			name: "repository error is wrapped (sentinel preserved)",
			input: usecases.ListAssignmentsInput{
				Caller: caller(authorTeacherID, true),
			},
			repoErr: errFakeAssignmentRepoFault,
			wantErr: errFakeAssignmentRepoFault,
		},
		{
			name: "items and total propagate to caller",
			input: usecases.ListAssignmentsInput{
				Caller: caller(authorTeacherID, true),
			},
			repoResult: repositories.AssignmentListResult{
				Items: []*entities.Assignment{
					makeAssignmentForList(t, 1, authorTeacherID),
					makeAssignmentForList(t, 2, authorTeacherID),
				},
				Total: 42,
			},
			wantFilter: repositories.AssignmentListFilter{
				Limit: usecases.DefaultListLimit,
			},
			wantItems:    2,
			wantTotalOut: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeAssignmentRepo()
			repo.listErr = tc.repoErr
			repo.listResult = tc.repoResult

			uc := usecases.NewListAssignmentsUseCase(repo)
			out, err := uc.Execute(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected wrap of %v, got %v", tc.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantItems, len(out.Items))
			assert.Equal(t, tc.wantTotalOut, out.Total)

			require.NotNil(t, repo.lastListFilter, "repo.List must be called")
			gotFilter := *repo.lastListFilter
			if tc.wantFilter.TeacherID == nil {
				assert.Nil(t, gotFilter.TeacherID, "TeacherID should be nil")
			} else {
				require.NotNil(t, gotFilter.TeacherID)
				assert.Equal(t, *tc.wantFilter.TeacherID, *gotFilter.TeacherID)
			}
			assert.Equal(t, tc.wantFilter.Subject, gotFilter.Subject)
			assert.Equal(t, tc.wantFilter.GroupName, gotFilter.GroupName)
			assert.Equal(t, tc.wantFilter.Limit, gotFilter.Limit)
			assert.Equal(t, tc.wantFilter.Offset, gotFilter.Offset)
		})
	}
}

func makeAssignmentForList(t *testing.T, id, teacherID int64) *entities.Assignment {
	t.Helper()
	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "Task X", TeacherID: teacherID, GroupName: "ИС-21",
		Subject: "Algorithms", MaxScore: 100, Now: time.Now(),
	})
	require.NoError(t, err)
	a.ID = id
	return a
}

func ptrInt64(v int64) *int64 { return &v }
