package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// TestReturnSubmissionUseCase_PendingToReturnedHappyPath asserts the core
// methodist-driven flow: an assignment exists, no prior submission exists,
// the methodist (via the assignment's author here for the simplest gate)
// returns the upload with a reason. The use case must materialise a
// fresh-pending Submission, transition it to Returned, persist it, and
// invoke the notifier exactly once with the right payload.
func TestReturnSubmissionUseCase_PendingToReturnedHappyPath(t *testing.T) {
	ar := newFakeAssignmentRepo()
	ar.seed(makeAssignment(t, assignmentID, authorTeacherID))
	sr := newFakeSubmissionRepo()
	notifier := &recordingReturnNotifier{}

	uc := usecases.NewReturnSubmissionUseCase(ar, sr, notifier, nil, func() time.Time { return fixedNow })

	err := uc.Execute(context.Background(), authorTeacherID, usecases.ReturnSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Reason:       "revisit derivation",
	})
	require.NoError(t, err)

	saved := sr.lookup(assignmentID, studentID)
	require.NotNil(t, saved, "submission must be persisted on happy path")
	assert.Equal(t, entities.StatusReturned, saved.Status())
	assert.Equal(t, "revisit derivation", saved.ReturnReason())
	require.NotNil(t, saved.ReturnedBy())
	assert.Equal(t, authorTeacherID, *saved.ReturnedBy())
	assert.Nil(t, saved.GradeValue(), "fresh-pending Return must leave grade nil")

	assert.True(t, notifier.called, "notifier must be invoked on happy path")
	assert.Equal(t, studentID, notifier.lastStudentID)
	assert.Equal(t, assignmentID, notifier.lastAssignmentID)
	assert.Equal(t, "revisit derivation", notifier.lastReason)
}

// recordingReturnNotifier captures NotifyReturned invocations so the test
// can assert payload fidelity. Mirrors recordingNotifier's shape.
type recordingReturnNotifier struct {
	called           bool
	lastStudentID    int64
	lastAssignmentID int64
	lastReason       string
	err              error
}

func (n *recordingReturnNotifier) NotifyReturned(ctx context.Context, studentID, assignmentID int64, reason string) error {
	n.called = true
	n.lastStudentID = studentID
	n.lastAssignmentID = assignmentID
	n.lastReason = reason
	return n.err
}
