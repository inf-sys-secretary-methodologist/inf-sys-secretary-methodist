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
)

func newReturnedSubmission(t *testing.T, aid, studentID int64, now time.Time) *entities.Submission {
	t.Helper()
	s := entities.NewSubmission(aid, studentID, now)
	require.NoError(t, s.Return("please add citations", 99, now))
	return s
}

func newAssignmentForTest(t *testing.T, id, teacherID int64, now time.Time) *entities.Assignment {
	t.Helper()
	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title:     "Lab 1",
		TeacherID: teacherID,
		GroupName: "БСБО-01-22",
		Subject:   "Math",
		MaxScore:  100,
		Now:       now,
	})
	require.NoError(t, err)
	a.ID = id
	return a
}

func TestGetMyAssignmentDetailUseCase_Execute(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	const studentID = int64(7)
	const aid = int64(10)

	t.Run("happy path returns combined view", func(t *testing.T) {
		ar := newFakeAssignmentRepo()
		ar.seed(newAssignmentForTest(t, aid, 42, now))
		sr := newFakeSubmissionRepo()
		sr.seed(newReturnedSubmission(t, aid, studentID, now))

		uc := usecases.NewGetMyAssignmentDetailUseCase(ar, sr)
		view, err := uc.Execute(context.Background(), usecases.GetMyAssignmentDetailInput{
			AssignmentID: aid,
			StudentID:    studentID,
		})
		require.NoError(t, err)
		assert.Equal(t, aid, view.AssignmentID)
		assert.Equal(t, "Lab 1", view.Title)
		assert.Equal(t, studentID, view.StudentID)
		assert.Equal(t, entities.StatusReturned, view.Status)
		assert.Equal(t, "please add citations", view.ReturnReason)
	})

	t.Run("assignment not found surfaces sentinel", func(t *testing.T) {
		ar := newFakeAssignmentRepo()
		sr := newFakeSubmissionRepo()
		uc := usecases.NewGetMyAssignmentDetailUseCase(ar, sr)
		_, err := uc.Execute(context.Background(), usecases.GetMyAssignmentDetailInput{
			AssignmentID: aid,
			StudentID:    studentID,
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, usecases.ErrAssignmentNotFound))
	})

	t.Run("submission not found surfaces sentinel", func(t *testing.T) {
		ar := newFakeAssignmentRepo()
		ar.seed(newAssignmentForTest(t, aid, 42, now))
		sr := newFakeSubmissionRepo()
		uc := usecases.NewGetMyAssignmentDetailUseCase(ar, sr)
		_, err := uc.Execute(context.Background(), usecases.GetMyAssignmentDetailInput{
			AssignmentID: aid,
			StudentID:    studentID,
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, usecases.ErrSubmissionNotFound))
	})

	t.Run("non-positive student id is rejected", func(t *testing.T) {
		ar := newFakeAssignmentRepo()
		sr := newFakeSubmissionRepo()
		uc := usecases.NewGetMyAssignmentDetailUseCase(ar, sr)
		_, err := uc.Execute(context.Background(), usecases.GetMyAssignmentDetailInput{
			AssignmentID: aid,
			StudentID:    0,
		})
		require.Error(t, err)
	})

	t.Run("AuthorizeReader rejects a row whose StudentID does not match the caller", func(t *testing.T) {
		// Pins the use-case-level wiring of AuthorizeReader: if a future
		// refactor drops the call, this test breaks. The keyed lookup at
		// the SQL layer makes this state unreachable through the normal
		// repo path, but defense-in-depth means a foreign-row response
		// (e.g. from a buggy repo, an alternate caller, or a future
		// CLI invoker) must still be rejected with ErrSubmissionOwnerOnly,
		// not silently leaked.
		ar := newFakeAssignmentRepo()
		ar.seed(newAssignmentForTest(t, aid, 42, now))
		foreignSubmission := newReturnedSubmission(t, aid, studentID, now) // StudentID = 7
		sr := newFakeSubmissionRepo()
		sr.forceReturn = foreignSubmission // any caller gets the StudentID=7 row

		uc := usecases.NewGetMyAssignmentDetailUseCase(ar, sr)
		_, err := uc.Execute(context.Background(), usecases.GetMyAssignmentDetailInput{
			AssignmentID: aid,
			StudentID:    99, // caller is NOT the owner
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, entities.ErrSubmissionOwnerOnly),
			"AuthorizeReader must reject foreign callers — got %v", err)
	})
}

func TestNewGetMyAssignmentDetailUseCase_PanicsOnNilDeps(t *testing.T) {
	ar := newFakeAssignmentRepo()
	sr := newFakeSubmissionRepo()

	assert.Panics(t, func() { usecases.NewGetMyAssignmentDetailUseCase(nil, sr) })
	assert.Panics(t, func() { usecases.NewGetMyAssignmentDetailUseCase(ar, nil) })
}
