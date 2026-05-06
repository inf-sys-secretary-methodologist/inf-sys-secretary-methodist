package usecases_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

const (
	authorTeacherID = int64(42)
	otherTeacherID  = int64(99)
	studentID       = int64(7)
	assignmentID    = int64(10)
)

var fixedNow = time.Date(2026, 5, 4, 12, 30, 0, 0, time.UTC)

func TestSaveGradeUseCase_Execute(t *testing.T) {
	gradedSubFixture := func() *entities.Submission {
		s := entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-1*time.Hour))
		score, _ := entities.NewScore(50)
		_ = s.Grade(score, "previous", authorTeacherID, fixedNow.Add(-30*time.Minute))
		return s
	}

	tests := []struct {
		name           string
		teacherID      int64
		input          usecases.SaveGradeInput
		seedSubmission *entities.Submission
		notifyErr      error
		wantErr        error
		wantNotified   bool
		wantSavedValue *int
		wantSavedFb    string
	}{
		{
			name:      "happy path persists graded submission and notifies student",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID,
				Value: 85, Feedback: "great work",
			},
			wantNotified:   true,
			wantSavedValue: ptrInt(85),
			wantSavedFb:    "great work",
		},
		{
			// Reviewer note: the "submission already exists in pending"
			// path was previously uncovered. Grade must mutate the
			// existing entity in place rather than create a fresh one.
			name:      "existing pending submission is graded in place",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID,
				Value: 70, Feedback: "ok",
			},
			seedSubmission: entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour)),
			wantNotified:   true,
			wantSavedValue: ptrInt(70),
			wantSavedFb:    "ok",
		},
		{
			name:      "assignment not found surfaces repository sentinel",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: 999, StudentID: studentID, Value: 50,
			},
			wantErr: repositories.ErrAssignmentNotFound,
		},
		{
			name:      "non-author teacher is forbidden",
			teacherID: otherTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID, Value: 50,
			},
			wantErr: entities.ErrAssignmentScopeForbidden,
		},
		{
			name:      "score above max rejected",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID, Value: 150,
			},
			wantErr: entities.ErrInvalidScore,
		},
		{
			name:      "negative score rejected",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID, Value: -5,
			},
			wantErr: entities.ErrInvalidScore,
		},
		{
			name:      "already-graded submission rejected",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID, Value: 80,
			},
			seedSubmission: gradedSubFixture(),
			wantErr:        entities.ErrAlreadyGraded,
		},
		{
			name:      "notification failure does not fail the grading",
			teacherID: authorTeacherID,
			input: usecases.SaveGradeInput{
				AssignmentID: assignmentID, StudentID: studentID,
				Value: 60, Feedback: "ok",
			},
			notifyErr:      errors.New("smtp down"),
			wantNotified:   true,
			wantSavedValue: ptrInt(60),
			wantSavedFb:    "ok",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ar := newFakeAssignmentRepo()
			ar.seed(makeAssignment(t, assignmentID, authorTeacherID))

			sr := newFakeSubmissionRepo()
			if tc.seedSubmission != nil {
				sr.seed(tc.seedSubmission)
			}

			notifier := &recordingNotifier{err: tc.notifyErr}

			uc := usecases.NewSaveGradeUseCase(ar, sr, notifier, nil, func() time.Time { return fixedNow })

			err := uc.Execute(context.Background(), tc.teacherID, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected error wrapping %v, got %v", tc.wantErr, err)
				assert.Equal(t, 0, sr.saveCalls,
					"submission Save must not be called when use case fails")
				assert.False(t, notifier.called, "notifier should not be called when use case fails")
				return
			}

			require.NoError(t, err)
			saved := sr.lookup(tc.input.AssignmentID, tc.input.StudentID)
			require.NotNil(t, saved, "submission must be persisted on happy path")
			require.NotNil(t, saved.GradeValue())
			assert.Equal(t, *tc.wantSavedValue, *saved.GradeValue())
			assert.Equal(t, tc.wantSavedFb, saved.Feedback())
			assert.True(t, saved.IsGraded())

			assert.Equal(t, tc.wantNotified, notifier.called)
			if tc.wantNotified {
				assert.Equal(t, tc.input.StudentID, notifier.lastStudentID)
				assert.Equal(t, tc.input.AssignmentID, notifier.lastAssignmentID)
				assert.Equal(t, tc.input.Value, notifier.lastScore)
				assert.Equal(t, 100, notifier.lastMax)
			}
		})
	}
}

// --- fakes -----------------------------------------------------------------

func makeAssignment(t *testing.T, id, teacherID int64) *entities.Assignment {
	t.Helper()
	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "L1", TeacherID: teacherID, GroupName: "ИС-21",
		Subject: "Algo", MaxScore: 100, Now: fixedNow.Add(-24 * time.Hour),
	})
	require.NoError(t, err)
	a.ID = id
	return a
}

type fakeAssignmentRepo struct {
	byID map[int64]*entities.Assignment

	// List support — used by ListAssignmentsUseCase tests.
	listResult     repositories.AssignmentListResult
	listErr        error
	lastListFilter *repositories.AssignmentListFilter
}

func newFakeAssignmentRepo() *fakeAssignmentRepo {
	return &fakeAssignmentRepo{byID: map[int64]*entities.Assignment{}}
}
func (r *fakeAssignmentRepo) seed(a *entities.Assignment) { r.byID[a.ID] = a }
func (r *fakeAssignmentRepo) GetByID(ctx context.Context, id int64) (*entities.Assignment, error) {
	if a, ok := r.byID[id]; ok {
		return a, nil
	}
	return nil, repositories.ErrAssignmentNotFound
}
func (r *fakeAssignmentRepo) List(ctx context.Context, filter repositories.AssignmentListFilter) (repositories.AssignmentListResult, error) {
	f := filter
	r.lastListFilter = &f
	if r.listErr != nil {
		return repositories.AssignmentListResult{}, r.listErr
	}
	return r.listResult, nil
}

type fakeSubmissionRepo struct {
	byKey      map[string]*entities.Submission
	saveCalls  int
	lastSaveAt string

	// ListByAssignment support — used by ListSubmissionsUseCase tests.
	listResult           []views.SubmissionView
	listErr              error
	lastListAssignmentID int64
	lastListStatus       *entities.SubmissionStatus

	// forceReturn, when non-nil, is returned from GetByAssignmentAndStudent
	// regardless of the (aid, sid) key — used to exercise downstream
	// ownership invariants on a row whose StudentID does not match the
	// caller, a state otherwise unreachable through the keyed lookup.
	forceReturn *entities.Submission
}

func newFakeSubmissionRepo() *fakeSubmissionRepo {
	return &fakeSubmissionRepo{byKey: map[string]*entities.Submission{}}
}
func subKey(aid, sid int64) string { return fmt.Sprintf("%d:%d", aid, sid) }
func (r *fakeSubmissionRepo) seed(s *entities.Submission) {
	r.byKey[subKey(s.AssignmentID, s.StudentID)] = s
}
func (r *fakeSubmissionRepo) GetByAssignmentAndStudent(ctx context.Context, aid, sid int64) (*entities.Submission, error) {
	if r.forceReturn != nil {
		return r.forceReturn, nil
	}
	if s, ok := r.byKey[subKey(aid, sid)]; ok {
		return s, nil
	}
	return nil, repositories.ErrSubmissionNotFound
}
func (r *fakeSubmissionRepo) Save(ctx context.Context, s *entities.Submission) error {
	r.byKey[subKey(s.AssignmentID, s.StudentID)] = s
	r.saveCalls++
	r.lastSaveAt = subKey(s.AssignmentID, s.StudentID)
	return nil
}
func (r *fakeSubmissionRepo) ListByAssignment(ctx context.Context, assignmentID int64, status *entities.SubmissionStatus) ([]views.SubmissionView, error) {
	r.lastListAssignmentID = assignmentID
	r.lastListStatus = status
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listResult, nil
}
func (r *fakeSubmissionRepo) ListByStudent(ctx context.Context, studentID int64, status *entities.SubmissionStatus) ([]views.StudentAssignmentView, error) {
	return nil, nil
}
func (r *fakeSubmissionRepo) lookup(aid, sid int64) *entities.Submission {
	return r.byKey[subKey(aid, sid)]
}

type recordingNotifier struct {
	called           bool
	lastStudentID    int64
	lastAssignmentID int64
	lastScore        int
	lastMax          int
	err              error
}

func (n *recordingNotifier) NotifyGraded(ctx context.Context, studentID, assignmentID int64, score, maxScore int) error {
	n.called = true
	n.lastStudentID = studentID
	n.lastAssignmentID = assignmentID
	n.lastScore = score
	n.lastMax = maxScore
	return n.err
}

func ptrInt(v int) *int { return &v }
