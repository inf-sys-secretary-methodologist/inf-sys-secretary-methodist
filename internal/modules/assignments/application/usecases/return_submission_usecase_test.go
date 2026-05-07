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

// auditActionReturned is the canonical audit event action emitted by
// ReturnSubmissionUseCase. Used by multiple test cases that scan
// audit log events for the matching action — extracted to constant
// so a future rename touches one site.
const auditActionReturned = "assignment.returned"

// TestReturnSubmissionUseCase_PendingToReturnedHappyPath asserts the core
// methodist-driven flow: an assignment exists, no prior submission exists,
// the methodist (via the assignment's author here for the simplest gate)
// returns the upload with a reason. The use case must materialize a
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
		if audit.events[i].Action == auditActionReturned {
			ret = &audit.events[i]
			break
		}
	}
	require.NotNil(t, ret, "expected assignment.returned audit event")
	assert.Equal(t, "assignment", ret.Resource)
	assert.Equal(t, assignmentID, ret.Fields["assignment_id"])
	assert.Equal(t, studentID, ret.Fields["student_id"])
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
		if audit.events[i].Action == auditActionReturned {
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

// TestReturnSubmissionUseCase_AuthzAndErrorPaths sweeps the authorisation
// and error-mapping matrix for Execute as a single table. Each row pins a
// distinct failure mode plus the side-effect contract the use case must
// honor on that failure: never persist, never notify, and emit an audit
// event ONLY for the authz-denied path (matching SaveGrade's policy that
// post-authz validation failures stay silent on the audit channel).
//
// The happy path is intentionally NOT in this table — it lives in its
// own test above to keep the matrix focused on the negative space.
func TestReturnSubmissionUseCase_AuthzAndErrorPaths(t *testing.T) {
	// returnedSubFixture builds a Submission that has already transitioned
	// to Returned, so that Execute will load it and Return will reject it
	// with ErrAlreadyReturned. Constructed via the legitimate Return path
	// (not a struct literal) to keep the entity invariants honest.
	returnedSubFixture := func() *entities.Submission {
		s := entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour))
		require.NoError(t, s.Return("first return", authorTeacherID, fixedNow.Add(-time.Hour)))
		return s
	}

	tests := []struct {
		name             string
		actor            int64
		input            usecases.ReturnSubmissionInput
		seed             *entities.Submission
		wantErrIs        error
		wantSaveCalls    int
		wantNotifyCalled bool
		wantAuditAction  string // "" means no audit action expected
		wantAuditExtra   map[string]any
	}{
		{
			name:  "assignment not found surfaces repository sentinel",
			actor: authorTeacherID,
			input: usecases.ReturnSubmissionInput{
				AssignmentID: 999, StudentID: studentID, Reason: "x",
			},
			wantErrIs: repositories.ErrAssignmentNotFound,
		},
		{
			name:  "non-author actor is forbidden and audit return_denied fires",
			actor: otherTeacherID,
			input: usecases.ReturnSubmissionInput{
				AssignmentID: assignmentID, StudentID: studentID, Reason: "x",
			},
			wantErrIs:       entities.ErrAssignmentScopeForbidden,
			wantAuditAction: "assignment.return_denied",
			wantAuditExtra:  map[string]any{"reason": "not_author"},
		},
		{
			name:  "already-returned submission rejected",
			actor: authorTeacherID,
			input: usecases.ReturnSubmissionInput{
				AssignmentID: assignmentID, StudentID: studentID, Reason: "x",
			},
			seed:      returnedSubFixture(),
			wantErrIs: entities.ErrAlreadyReturned,
		},
		{
			name:  "empty reason rejected by entity invariant",
			actor: authorTeacherID,
			input: usecases.ReturnSubmissionInput{
				AssignmentID: assignmentID, StudentID: studentID, Reason: "",
			},
			wantErrIs: entities.ErrInvalidReturn,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ar := newFakeAssignmentRepo()
			ar.seed(makeAssignment(t, assignmentID, authorTeacherID))
			sr := newFakeSubmissionRepo()
			if tc.seed != nil {
				sr.seed(tc.seed)
			}
			notifier := &recordingReturnNotifier{}
			audit := &recordingAuditSink{}

			uc := usecases.NewReturnSubmissionUseCase(ar, sr, notifier, audit, func() time.Time { return fixedNow })

			err := uc.Execute(context.Background(), tc.actor, tc.input)

			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.wantErrIs),
				"expected error wrapping %v, got %v", tc.wantErrIs, err)
			assert.Equal(t, tc.wantSaveCalls, sr.saveCalls,
				"submission Save count mismatch on error path")
			assert.Equal(t, tc.wantNotifyCalled, notifier.called,
				"notifier should not be invoked on error paths")

			if tc.wantAuditAction != "" {
				var found *recordedAuditEvent
				for i := range audit.events {
					if audit.events[i].Action == tc.wantAuditAction {
						found = &audit.events[i]
						break
					}
				}
				require.NotNil(t, found, "expected audit event %q", tc.wantAuditAction)
				assert.Equal(t, "assignment", found.Resource)
				assert.Equal(t, tc.actor, found.Fields["actor_user_id"])
				for k, v := range tc.wantAuditExtra {
					assert.Equal(t, v, found.Fields[k], "audit field %q mismatch", k)
				}
			} else {
				// No audit expected — assert audit log is silent on this
				// error path. Post-authz validation failures must not leak
				// onto the audit stream (consistent with SaveGrade).
				assert.Empty(t, audit.events, "no audit event should fire on this error path")
			}
		})
	}
}

// TestReturnSubmissionUseCase_NotifierFailureDoesNotAbort pins the
// best-effort-notifier semantics: a notifier error MUST NOT abort the
// return. The state transition is the system of record — a transient
// SMTP outage cannot block the methodist's workflow. The use case must
// still persist the Returned submission and emit BOTH the failure-audit
// event (on-call visibility) AND the success-audit event (the actual
// state transition record).
func TestReturnSubmissionUseCase_NotifierFailureDoesNotAbort(t *testing.T) {
	ar := newFakeAssignmentRepo()
	ar.seed(makeAssignment(t, assignmentID, authorTeacherID))
	sr := newFakeSubmissionRepo()
	notifier := &recordingReturnNotifier{err: errors.New("smtp down")}
	audit := &recordingAuditSink{}

	uc := usecases.NewReturnSubmissionUseCase(ar, sr, notifier, audit, func() time.Time { return fixedNow })

	err := uc.Execute(context.Background(), authorTeacherID, usecases.ReturnSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Reason:       "revisit",
	})
	require.NoError(t, err, "notifier failure must NOT abort the return")

	saved := sr.lookup(assignmentID, studentID)
	require.NotNil(t, saved, "submission must be persisted even when notifier fails")
	assert.Equal(t, entities.StatusReturned, saved.Status())
	assert.Equal(t, 1, sr.saveCalls)
	assert.True(t, notifier.called)

	// Both events must be present: failure audit (best-effort visibility for on-call)
	// AND success audit (the state transition is the system of record).
	var failed, returned *recordedAuditEvent
	for i := range audit.events {
		switch audit.events[i].Action {
		case "assignment.return_notify_failed":
			failed = &audit.events[i]
		case auditActionReturned:
			returned = &audit.events[i]
		}
	}
	require.NotNil(t, failed, "expected assignment.return_notify_failed audit event")
	require.NotNil(t, returned, "expected assignment.returned audit event")

	assert.Equal(t, "smtp down", failed.Fields["error"], "audit must capture notifier error message")
	assert.Equal(t, assignmentID, failed.Fields["assignment_id"])
	assert.Equal(t, studentID, failed.Fields["student_id"])
	assert.Equal(t, authorTeacherID, failed.Fields["actor_user_id"])

	assert.Equal(t, "revisit", returned.Fields["reason"])
	assert.Equal(t, authorTeacherID, returned.Fields["actor_user_id"])
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
