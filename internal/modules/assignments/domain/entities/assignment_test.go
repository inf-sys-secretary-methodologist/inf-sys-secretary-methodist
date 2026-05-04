package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

func TestNewAssignment_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	due := now.Add(7 * 24 * time.Hour)

	base := entities.NewAssignmentParams{
		Title:       "Lab 1: linked lists",
		Description: "Implement a doubly-linked list",
		TeacherID:   42,
		GroupName:   "ИС-21",
		Subject:     "Algorithms",
		MaxScore:    100,
		DueDate:     &due,
		Now:         now,
	}

	tests := []struct {
		name    string
		mutate  func(p *entities.NewAssignmentParams)
		wantErr error
	}{
		{name: "valid params construct successfully", mutate: func(p *entities.NewAssignmentParams) {}},
		{name: "no due date is acceptable", mutate: func(p *entities.NewAssignmentParams) { p.DueDate = nil }},

		{name: "empty title is rejected",
			mutate:  func(p *entities.NewAssignmentParams) { p.Title = "" },
			wantErr: entities.ErrInvalidAssignment},
		{name: "whitespace-only title is rejected",
			mutate:  func(p *entities.NewAssignmentParams) { p.Title = "   \t  " },
			wantErr: entities.ErrInvalidAssignment},
		{name: "empty group_name is rejected",
			mutate:  func(p *entities.NewAssignmentParams) { p.GroupName = "" },
			wantErr: entities.ErrInvalidAssignment},
		{name: "whitespace-only group_name is rejected",
			mutate:  func(p *entities.NewAssignmentParams) { p.GroupName = "  " },
			wantErr: entities.ErrInvalidAssignment},
		{name: "max_score zero is rejected",
			mutate:  func(p *entities.NewAssignmentParams) { p.MaxScore = 0 },
			wantErr: entities.ErrInvalidAssignment},
		{name: "max_score negative is rejected",
			mutate:  func(p *entities.NewAssignmentParams) { p.MaxScore = -10 },
			wantErr: entities.ErrInvalidAssignment},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := base
			tc.mutate(&p)
			a, err := entities.NewAssignment(p)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected error wrapping %v, got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, a)
			assert.Equal(t, p.Title, a.Title())
			assert.Equal(t, p.Description, a.Description())
			assert.Equal(t, p.TeacherID, a.TeacherID())
			assert.Equal(t, p.GroupName, a.GroupName())
			assert.Equal(t, p.Subject, a.Subject())
			assert.Equal(t, p.MaxScore, a.MaxScore())
			assert.Equal(t, p.DueDate, a.DueDate())
			assert.Equal(t, p.Now, a.CreatedAt())
			assert.Equal(t, p.Now, a.UpdatedAt())
		})
	}
}

func TestNewAssignment_StoresTrimmedTitleAndGroupName(t *testing.T) {
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)

	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title:     "  L1  ",
		GroupName: "  ИС-21  ",
		TeacherID: 42,
		Subject:   "Algo",
		MaxScore:  100,
		Now:       now,
	})

	require.NoError(t, err)
	require.NotNil(t, a)
	// Domain doc says invariant is "trimmed-non-empty". Storing the raw
	// input means the entity violates its own invariant — the value
	// passes validation but is not the canonical form.
	assert.Equal(t, "L1", a.Title(), "title must be stored canonicalised (trimmed)")
	assert.Equal(t, "ИС-21", a.GroupName(), "group_name must be stored canonicalised (trimmed)")
}

func TestAssignment_AuthorizeGrader(t *testing.T) {
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "L1", TeacherID: 42, GroupName: "ИС-21",
		Subject: "Algo", MaxScore: 100, Now: now,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		userID  int64
		wantErr error
	}{
		{name: "author can grade", userID: 42},
		{name: "different teacher is forbidden", userID: 7, wantErr: entities.ErrAssignmentScopeForbidden},
		{name: "zero user id is forbidden", userID: 0, wantErr: entities.ErrAssignmentScopeForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := a.AuthorizeGrader(tc.userID)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestAssignment_NewSubmissionScore(t *testing.T) {
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "L1", TeacherID: 42, GroupName: "ИС-21",
		Subject: "Algo", MaxScore: 100, Now: now,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		value   int
		wantErr error
		wantVal int
	}{
		{name: "value within range constructs Score", value: 85, wantVal: 85},
		{name: "value equal to max constructs Score", value: 100, wantVal: 100},
		{name: "zero value is acceptable", value: 0, wantVal: 0},
		{name: "value over max is rejected", value: 150, wantErr: entities.ErrInvalidScore},
		{name: "negative value is rejected", value: -1, wantErr: entities.ErrInvalidScore},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score, err := a.NewSubmissionScore(tc.value)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected %v, got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantVal, score.Value())
		})
	}
}
