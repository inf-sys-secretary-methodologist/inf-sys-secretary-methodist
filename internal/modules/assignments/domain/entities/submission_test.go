package entities_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

func TestNewSubmission_StartsPending(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	s := entities.NewSubmission(10, 42, now)

	require.NotNil(t, s)
	assert.Equal(t, int64(10), s.AssignmentID)
	assert.Equal(t, int64(42), s.StudentID)
	assert.Equal(t, entities.StatusPending, s.Status())
	assert.False(t, s.IsGraded())
	assert.Nil(t, s.GradeValue())
	assert.Nil(t, s.GradedBy())
	assert.Nil(t, s.GradedAt())
	assert.Equal(t, now, s.CreatedAt())
	assert.Equal(t, now, s.UpdatedAt())
}

func TestSubmission_Grade(t *testing.T) {
	created := time.Date(2026, 5, 4, 9, 0, 0, 0, time.UTC)
	graded := time.Date(2026, 5, 4, 15, 30, 0, 0, time.UTC)
	score, err := entities.NewScore(85)
	require.NoError(t, err)

	tests := []struct {
		name       string
		setupTwice bool
		wantErr    error
		wantStatus entities.SubmissionStatus
		wantValue  *int
	}{
		{
			name:       "grade transitions pending submission to graded",
			setupTwice: false,
			wantStatus: entities.StatusGraded,
			wantValue:  intPtr(85),
		},
		{
			name:       "grading an already-graded submission fails with ErrAlreadyGraded",
			setupTwice: true,
			wantErr:    entities.ErrAlreadyGraded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := entities.NewSubmission(10, 42, created)
			if tc.setupTwice {
				require.NoError(t, s.Grade(score, "ok", 7, graded))
			}

			err := s.Grade(score, "good work", 7, graded)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected error wrapping %v, got %v", tc.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, s.Status())
			assert.True(t, s.IsGraded())
			require.NotNil(t, s.GradeValue())
			assert.Equal(t, *tc.wantValue, *s.GradeValue())
			assert.Equal(t, "good work", s.Feedback())
			require.NotNil(t, s.GradedBy())
			assert.Equal(t, int64(7), *s.GradedBy())
			require.NotNil(t, s.GradedAt())
			assert.Equal(t, graded, *s.GradedAt())
			assert.Equal(t, graded, s.UpdatedAt())
		})
	}
}

func TestSubmissionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    entities.SubmissionStatus
		want bool
	}{
		{name: "pending is valid", s: entities.StatusPending, want: true},
		{name: "graded is valid", s: entities.StatusGraded, want: true},
		{name: "returned is valid", s: entities.StatusReturned, want: true},
		{name: "empty is invalid", s: entities.SubmissionStatus(""), want: false},
		{name: "arbitrary string is invalid", s: entities.SubmissionStatus("draft"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.s.IsValid())
		})
	}
}

func TestSubmission_Return_AlreadyReturnedRejected(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	s := entities.NewSubmission(1, 7, now)
	require.NoError(t, s.Return("not enough detail", 99, now))
	require.Equal(t, entities.StatusReturned, s.Status())

	err := s.Return("still not enough", 99, now)
	require.Error(t, err)
	assert.True(t, errors.Is(err, entities.ErrAlreadyReturned),
		"expected error wrapping ErrAlreadyReturned, got %v", err)
}

func TestSubmission_Return_InvariantValidation(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		reason     string
		returnedBy int64
		wantErrIs  error // nil means must succeed
	}{
		{name: "empty reason rejected", reason: "", returnedBy: 99, wantErrIs: entities.ErrInvalidReturn},
		{name: "whitespace-only reason rejected", reason: "   \t\n", returnedBy: 99, wantErrIs: entities.ErrInvalidReturn},
		{name: "reason at 4096-char boundary accepted", reason: strings.Repeat("a", 4096), returnedBy: 99, wantErrIs: nil},
		{name: "reason exceeds 4096 chars rejected", reason: strings.Repeat("a", 4097), returnedBy: 99, wantErrIs: entities.ErrInvalidReturn},
		{name: "zero returnedBy rejected", reason: "fine", returnedBy: 0, wantErrIs: entities.ErrInvalidReturn},
		{name: "negative returnedBy rejected", reason: "fine", returnedBy: -1, wantErrIs: entities.ErrInvalidReturn},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := entities.NewSubmission(1, 7, now)

			err := s.Return(tc.reason, tc.returnedBy, now)

			if tc.wantErrIs == nil {
				require.NoError(t, err)
				assert.Equal(t, entities.StatusReturned, s.Status())
				return
			}
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.wantErrIs),
				"got %v, want errors.Is(%v) = true", err, tc.wantErrIs)
		})
	}
}

func intPtr(v int) *int { return &v }
