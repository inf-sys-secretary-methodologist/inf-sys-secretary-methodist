package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
)

func TestGetAssignmentUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		caller  usecases.CallerScope
		askID   int64
		seedID  int64
		wantErr error
	}{
		{
			name:   "teacher reads own assignment",
			caller: usecases.CallerScope{UserID: authorTeacherID, Unrestricted: false},
			askID:  assignmentID,
			seedID: assignmentID,
		},
		{
			name:    "teacher denied on foreign assignment",
			caller:  usecases.CallerScope{UserID: otherTeacherID, Unrestricted: false},
			askID:   assignmentID,
			seedID:  assignmentID,
			wantErr: entities.ErrAssignmentScopeForbidden,
		},
		{
			name:   "methodist reads any assignment",
			caller: usecases.CallerScope{UserID: otherTeacherID, Unrestricted: true},
			askID:  assignmentID,
			seedID: assignmentID,
		},
		{
			name:   "admin reads any assignment",
			caller: usecases.CallerScope{UserID: int64(1), Unrestricted: true},
			askID:  assignmentID,
			seedID: assignmentID,
		},
		{
			name:    "not found surfaces sentinel",
			caller:  usecases.CallerScope{UserID: authorTeacherID, Unrestricted: true},
			askID:   999,
			seedID:  assignmentID,
			wantErr: repositories.ErrAssignmentNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeAssignmentRepo()
			repo.seed(makeAssignment(t, tc.seedID, authorTeacherID))

			uc := usecases.NewGetAssignmentUseCase(repo)
			got, err := uc.Execute(context.Background(), usecases.GetAssignmentInput{
				Caller:       tc.caller,
				AssignmentID: tc.askID,
			})

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected %v, got %v", tc.wantErr, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.askID, got.ID)
		})
	}
}
