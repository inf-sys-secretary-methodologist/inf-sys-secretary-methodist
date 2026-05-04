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

// TestResubmitSubmissionUseCase_ForeignStudentDenied is honest backfill
// coverage of the security-relevant denial path: a non-owning actor must
// be rejected by AuthorizeResubmitter, the failure must leave a forensic
// trail (assignment.resubmit_denied audit), and no side effects must
// land — no Save, no notifier call. The wiring for this already shipped
// in T7 alongside the happy path (defense-in-depth is unnatural to
// separate from the main flow), so this is NOT a RED→GREEN cycle —
// it's coverage of the existing branch.
func TestResubmitSubmissionUseCase_ForeignStudentDenied(t *testing.T) {
	const foreignStudentID int64 = 999

	ar := newFakeAssignmentRepo()
	ar.seed(makeAssignment(t, assignmentID, authorTeacherID))

	// Seed a returned submission owned by studentID — the foreign actor
	// will try to resubmit it and must be rejected.
	pre := entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour))
	require.NoError(t, pre.Return("missing details", authorTeacherID, fixedNow.Add(-time.Hour)))
	sr := newFakeSubmissionRepo()
	sr.seed(pre)

	notifier := &recordingResubmitNotifier{}
	audit := &recordingAuditSink{}

	uc := usecases.NewResubmitSubmissionUseCase(ar, sr, notifier, audit, func() time.Time { return fixedNow })

	err := uc.Execute(context.Background(), foreignStudentID, usecases.ResubmitSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID, // belongs to studentID, not foreignStudentID
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, entities.ErrSubmissionOwnerOnly),
		"expected ErrSubmissionOwnerOnly, got %v", err)

	// State invariants on denial path: persistence and notifier must be
	// completely silent. AuthorizeResubmitter fires before any mutation.
	assert.Equal(t, 0, sr.saveCalls, "Save must not be called on denied path")
	assert.False(t, notifier.called, "notifier must not be invoked on denied path")
	assert.Equal(t, entities.StatusReturned, sr.lookup(assignmentID, studentID).Status(),
		"submission status must remain unchanged on denial")

	// Forensic trail: a refused attempt must leave an audit event so on-call
	// can see denied access patterns. Mirrors return_denied / grade_denied.
	var denied *recordedAuditEvent
	for i := range audit.events {
		if audit.events[i].Action == "assignment.resubmit_denied" {
			denied = &audit.events[i]
			break
		}
	}
	require.NotNil(t, denied, "expected assignment.resubmit_denied audit event")
	assert.Equal(t, "assignment", denied.Resource)
	assert.Equal(t, int64(assignmentID), denied.Fields["assignment_id"])
	assert.Equal(t, int64(studentID), denied.Fields["student_id"])
	assert.Equal(t, foreignStudentID, denied.Fields["actor_user_id"])
	assert.Equal(t, "not_owner", denied.Fields["reason"])

	// And — critically — assignment.resubmitted MUST NOT have fired. A
	// denied attempt cannot leak a success event onto the audit stream.
	for _, e := range audit.events {
		assert.NotEqual(t, "assignment.resubmitted", e.Action,
			"resubmitted success event must not fire on denied path")
	}
}

// TestResubmitSubmissionUseCase_ErrorMatrix sweeps the failure modes of
// Execute as a single table — pinning each branch's error wrapping AND
// the side-effect contract on that branch (which calls fire, which audit
// events appear or stay silent). All cases pass at HEAD; this is honest
// backfill of code already written, not a TDD cycle.
//
// The happy path is intentionally NOT in this table — it lives in its
// own test above to keep the matrix focused on the negative space and
// the best-effort notifier semantics.
func TestResubmitSubmissionUseCase_ErrorMatrix(t *testing.T) {
	pendingSubFixture := func() *entities.Submission {
		return entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour))
	}
	returnedSubFixture := func() *entities.Submission {
		s := entities.NewSubmission(assignmentID, studentID, fixedNow.Add(-2*time.Hour))
		require.NoError(t, s.Return("missing details", authorTeacherID, fixedNow.Add(-time.Hour)))
		return s
	}

	tests := []struct {
		name             string
		assignmentExists bool
		seed             *entities.Submission
		notifierErr      error
		notifierNil      bool
		wantErrIs        error // nil means happy
		wantSaveCalls    int
		wantNotifyCalled bool
		wantSuccessAudit bool
		wantFailureAudit bool
	}{
		{
			name:             "assignment not found surfaces repository sentinel",
			assignmentExists: false,
			wantErrIs:        repositories.ErrAssignmentNotFound,
		},
		{
			name:             "submission not found surfaces repository sentinel",
			assignmentExists: true,
			seed:             nil,
			wantErrIs:        repositories.ErrSubmissionNotFound,
		},
		{
			name:             "pending submission rejected by entity invariant",
			assignmentExists: true,
			seed:             pendingSubFixture(),
			wantErrIs:        entities.ErrNotReturned,
		},
		{
			name:             "notifier failure does not abort (best-effort semantics)",
			assignmentExists: true,
			seed:             returnedSubFixture(),
			notifierErr:      errors.New("smtp down"),
			wantSaveCalls:    1,
			wantNotifyCalled: true,
			wantSuccessAudit: true,
			wantFailureAudit: true,
		},
		{
			name:             "nil notifier — happy path skips notify silently",
			assignmentExists: true,
			seed:             returnedSubFixture(),
			notifierNil:      true,
			wantSaveCalls:    1,
			wantSuccessAudit: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ar := newFakeAssignmentRepo()
			if tc.assignmentExists {
				ar.seed(makeAssignment(t, assignmentID, authorTeacherID))
			}
			sr := newFakeSubmissionRepo()
			if tc.seed != nil {
				sr.seed(tc.seed)
			}

			recNotifier := &recordingResubmitNotifier{err: tc.notifierErr}
			var notifier usecases.ResubmitSubmissionNotifier
			if !tc.notifierNil {
				notifier = recNotifier
			}
			audit := &recordingAuditSink{}

			uc := usecases.NewResubmitSubmissionUseCase(ar, sr, notifier, audit, func() time.Time { return fixedNow })

			err := uc.Execute(context.Background(), studentID, usecases.ResubmitSubmissionInput{
				AssignmentID: assignmentID,
				StudentID:    studentID,
			})

			if tc.wantErrIs != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErrIs),
					"expected error wrapping %v, got %v", tc.wantErrIs, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.wantSaveCalls, sr.saveCalls,
				"Save call count mismatch")
			if !tc.notifierNil {
				assert.Equal(t, tc.wantNotifyCalled, recNotifier.called,
					"notifier invocation mismatch")
			}

			hasSuccess, hasFailure := false, false
			for _, e := range audit.events {
				switch e.Action {
				case "assignment.resubmitted":
					hasSuccess = true
				case "assignment.resubmit_notify_failed":
					hasFailure = true
				}
			}
			assert.Equal(t, tc.wantSuccessAudit, hasSuccess,
				"assignment.resubmitted audit presence")
			assert.Equal(t, tc.wantFailureAudit, hasFailure,
				"assignment.resubmit_notify_failed audit presence")
		})
	}
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
