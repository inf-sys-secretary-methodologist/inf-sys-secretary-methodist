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

// TestResubmitSubmissionUseCase_ReturnedToPendingHappyPath asserts the core
// student-driven flow that closes the academic loop opened by v0.111.0:
// a returned submission exists; the owning student resubmits it for a
// fresh grading cycle; the use case transitions the submission back to
// Pending, persists it, notifies the teacher (the assignment author),
// and emits assignment.resubmitted with the prior return_reason captured
// BEFORE Resubmit clears it.
//
// The previous_return_reason capture mirrors the previous_grade /
// previous_feedback invariant on the Return side — forensic compliance
// requires the audit trail to record what the transition undid.
func TestResubmitSubmissionUseCase_ReturnedToPendingHappyPath(t *testing.T) {
	ar := newFakeAssignmentRepo()
	ar.seed(makeAssignment(t, assignmentID, authorTeacherID))

	// Seed a returned submission via the legitimate Return path so the
	// entity invariants hold (status=returned, return triple set, grade
	// triple cleared by Return). Building it via &entities.Submission{}
	// would bypass that and is forbidden anyway.
	pre := entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour))
	require.NoError(t, pre.Return("missing details", authorTeacherID, fixedNow.Add(-time.Hour)))

	sr := newFakeSubmissionRepo()
	sr.seed(pre)

	notifier := &recordingResubmitNotifier{}
	audit := &recordingAuditSink{}

	uc := usecases.NewResubmitSubmissionUseCase(ar, sr, notifier, audit, func() time.Time { return fixedNow })

	err := uc.Execute(context.Background(), studentID, usecases.ResubmitSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID,
	})
	require.NoError(t, err)

	saved := sr.lookup(assignmentID, studentID)
	require.NotNil(t, saved, "submission must remain persisted after Resubmit")
	assert.Equal(t, entities.StatusPending, saved.Status())
	assert.Equal(t, "", saved.ReturnReason(), "return_reason cleared by Resubmit")
	assert.Nil(t, saved.ReturnedBy(), "returned_by cleared by Resubmit")
	assert.Nil(t, saved.ReturnedAt(), "returned_at cleared by Resubmit")
	assert.Equal(t, fixedNow, saved.UpdatedAt())

	// Notifier targets the teacher (assignment author) — they need to know
	// a fresh attempt is awaiting grading.
	assert.True(t, notifier.called, "notifier must be invoked on happy path")
	assert.Equal(t, authorTeacherID, notifier.lastTeacherID)
	assert.Equal(t, int64(assignmentID), notifier.lastAssignmentID)

	// assignment.resubmitted captures actor + ids + previous_return_reason.
	var ret *recordedAuditEvent
	for i := range audit.events {
		if audit.events[i].Action == "assignment.resubmitted" {
			ret = &audit.events[i]
			break
		}
	}
	require.NotNil(t, ret, "expected assignment.resubmitted audit event")
	assert.Equal(t, "assignment", ret.Resource)
	assert.Equal(t, int64(assignmentID), ret.Fields["assignment_id"])
	assert.Equal(t, int64(studentID), ret.Fields["student_id"])
	assert.Equal(t, "missing details", ret.Fields["previous_return_reason"],
		"previous_return_reason must be captured before Resubmit clears it")
	assert.Equal(t, studentID, ret.Fields["actor_user_id"])
}

// recordingResubmitNotifier captures NotifyResubmitted invocations so the
// test can assert payload fidelity (teacher gets the ping, with the right
// ids). Mirrors recordingReturnNotifier's shape on the return side.
type recordingResubmitNotifier struct {
	called           bool
	lastTeacherID    int64
	lastAssignmentID int64
	err              error
}

func (n *recordingResubmitNotifier) NotifyResubmitted(ctx context.Context, teacherID, assignmentID int64) error {
	n.called = true
	n.lastTeacherID = teacherID
	n.lastAssignmentID = assignmentID
	return n.err
}
