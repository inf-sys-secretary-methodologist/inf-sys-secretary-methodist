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

// TestReturnSubmissionUseCase_GradedToReturnedAuditsPreviousGrade asserts the
// audit-content invariant for the Graded → Returned transition. When the
// methodist returns a submission that was previously graded, the resulting
// "assignment.returned" audit event MUST capture the prior grade value and
// feedback BEFORE Submission.Return clears them. Forensic compliance: a
// reviewer reading the audit log post-incident needs to see what was undone.
func TestReturnSubmissionUseCase_GradedToReturnedAuditsPreviousGrade(t *testing.T) {
	ar := newFakeAssignmentRepo()
	ar.seed(makeAssignment(t, assignmentID, authorTeacherID))
	sr := newFakeSubmissionRepo()

	// Pre-grade the submission so Return clears it.
	score, _ := entities.NewScore(85)
	pre := entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour))
	require.NoError(t, pre.Grade(score, "good", authorTeacherID, fixedNow.Add(-time.Hour)))
	sr.seed(pre)

	notifier := &recordingReturnNotifier{}
	audit := &recordingAuditSink{}

	uc := usecases.NewReturnSubmissionUseCase(ar, sr, notifier, audit, func() time.Time { return fixedNow })

	err := uc.Execute(context.Background(), authorTeacherID, usecases.ReturnSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Reason:       "redo derivation step",
	})
	require.NoError(t, err)

	// Find the assignment.returned event.
	var ret *recordedAuditEvent
	for i := range audit.events {
		if audit.events[i].Action == "assignment.returned" {
			ret = &audit.events[i]
			break
		}
	}
	require.NotNil(t, ret, "expected assignment.returned audit event")
	assert.Equal(t, "assignment", ret.Resource)
	assert.Equal(t, int64(assignmentID), ret.Fields["assignment_id"])
	assert.Equal(t, int64(studentID), ret.Fields["student_id"])
	assert.Equal(t, "redo derivation step", ret.Fields["reason"])
	assert.Equal(t, 85, ret.Fields["previous_grade"], "previous_grade must be captured before Return clears it")
	assert.Equal(t, "good", ret.Fields["previous_feedback"], "previous_feedback must be captured before Return clears it")
	assert.Equal(t, authorTeacherID, ret.Fields["actor_user_id"])
}

// TestReturnSubmissionUseCase_PendingToReturnedAuditsWithoutPreviousGrade
// asserts the negative half of the same invariant: when there is no prior
// grade (first-touch return on a fresh-pending submission), the audit
// fields map MUST NOT include previous_grade / previous_feedback keys.
// Avoids empty-NULL noise in the audit stream — keys present only when
// they carry meaning.
func TestReturnSubmissionUseCase_PendingToReturnedAuditsWithoutPreviousGrade(t *testing.T) {
	ar := newFakeAssignmentRepo()
	ar.seed(makeAssignment(t, assignmentID, authorTeacherID))
	sr := newFakeSubmissionRepo() // no seed — first-touch path

	audit := &recordingAuditSink{}
	uc := usecases.NewReturnSubmissionUseCase(ar, sr, &recordingReturnNotifier{}, audit, func() time.Time { return fixedNow })

	err := uc.Execute(context.Background(), authorTeacherID, usecases.ReturnSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Reason:       "wrong file",
	})
	require.NoError(t, err)

	var ret *recordedAuditEvent
	for i := range audit.events {
		if audit.events[i].Action == "assignment.returned" {
			ret = &audit.events[i]
			break
		}
	}
	require.NotNil(t, ret)
	_, hasPrev := ret.Fields["previous_grade"]
	assert.False(t, hasPrev, "previous_grade key must be absent when there was no prior grade (avoid empty-NULL noise)")
	_, hasPrevFb := ret.Fields["previous_feedback"]
	assert.False(t, hasPrevFb, "previous_feedback key must be absent when there was no prior grade")
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

// recordingAuditSink is a test double for usecases.AuditSink that captures
// every emitted audit event with a defensive copy of its fields map (the
// use case may reuse the map after the call). Tests then locate events
// by Action and assert on Resource / Fields content.
type recordingAuditSink struct {
	events []recordedAuditEvent
}

type recordedAuditEvent struct {
	Action   string
	Resource string
	Fields   map[string]any
}

func (r *recordingAuditSink) LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any) {
	cp := make(map[string]any, len(fields))
	for k, v := range fields {
		cp[k] = v
	}
	r.events = append(r.events, recordedAuditEvent{Action: action, Resource: resource, Fields: cp})
}
