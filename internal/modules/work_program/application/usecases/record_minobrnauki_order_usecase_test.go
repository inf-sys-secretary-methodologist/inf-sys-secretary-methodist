package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// fakeRecordOrderRepo is a minimal recordMinobrnaukiOrderRepo double.
type fakeRecordOrderRepo struct {
	saveCalls     int
	saved         *entities.MinobrnaukiOrder
	savedAffected []int64
	saveErr       error
	idAssigned    int64
}

func (f *fakeRecordOrderRepo) Save(_ context.Context, o *entities.MinobrnaukiOrder, affected []int64) error {
	f.saveCalls++
	f.saved = o
	f.savedAffected = affected
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.idAssigned > 0 {
		o.SetID(f.idAssigned)
	}
	return nil
}

type revisionTriggerCall struct {
	actorID, orderID int64
	orderNumber      string
	affected         []int64
}

type fakeRevisionTrigger struct {
	calls []revisionTriggerCall
}

func (f *fakeRevisionTrigger) Execute(_ context.Context, actorID, orderID int64, orderNumber string, affected []int64) (TriggerOrderRevisionsResult, error) {
	f.calls = append(f.calls, revisionTriggerCall{actorID, orderID, orderNumber, affected})
	return TriggerOrderRevisionsResult{}, nil
}

func validRecordOrderInput() RecordMinobrnaukiOrderInput {
	return RecordMinobrnaukiOrderInput{
		OrderNumber:            "№ 1078 от 12.05.2026",
		Title:                  "Об изменении ФГОС 09.03.01",
		PublishedAt:            time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		ChangeScope:            "major",
		Summary:                "Обновлён перечень компетенций",
		AffectedWorkProgramIDs: []int64{11, 22},
	}
}

func TestNewRecordMinobrnaukiOrderUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewRecordMinobrnaukiOrderUseCase(nil, ...) did not panic")
		}
	}()
	NewRecordMinobrnaukiOrderUseCase(nil, &recordingAuditSink{})
}

func TestRecordMinobrnaukiOrderUseCase_HappyPath_Methodist(t *testing.T) {
	repo := &fakeRecordOrderRepo{idAssigned: 100}
	audit := &recordingAuditSink{}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, audit)

	order, err := uc.Execute(context.Background(), 42, "methodist", validRecordOrderInput())
	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, int64(100), order.ID())
	assert.Equal(t, int64(42), order.UploadedBy(), "UploadedBy derives from actorID")
	assert.Equal(t, domain.MinobrnaukiOrderChangeScopeMajor, order.ChangeScope())
	assert.Equal(t, "№ 1078 от 12.05.2026", order.OrderNumber())

	require.Equal(t, 1, repo.saveCalls)
	assert.Equal(t, []int64{11, 22}, repo.savedAffected, "affected WP ids passed through to repo")

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "minobrnauki_order.recorded", ev.Action)
	assert.Equal(t, int64(42), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(100), ev.Fields["minobrnauki_order_id"])
}

func TestRecordMinobrnaukiOrderUseCase_AllowedRoles(t *testing.T) {
	for _, role := range []string{"methodist", "academic_secretary", "system_admin"} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeRecordOrderRepo{idAssigned: 1}
			uc := NewRecordMinobrnaukiOrderUseCase(repo, &recordingAuditSink{})
			_, err := uc.Execute(context.Background(), 1, role, validRecordOrderInput())
			require.NoError(t, err, "role %q must be allowed to record an order", role)
			assert.Equal(t, 1, repo.saveCalls)
		})
	}
}

func TestRecordMinobrnaukiOrderUseCase_DeniedRoles(t *testing.T) {
	for _, role := range []string{"teacher", "student", ""} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeRecordOrderRepo{}
			audit := &recordingAuditSink{}
			uc := NewRecordMinobrnaukiOrderUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 1, role, validRecordOrderInput())
			assert.True(t, errors.Is(err, domain.ErrMinobrnaukiOrderScopeForbidden),
				"role %q must be denied with scope-forbidden sentinel, got %v", role, err)
			assert.Zero(t, repo.saveCalls, "repo.Save must not be called on denied role")

			require.Len(t, audit.events, 1)
			assert.Equal(t, "minobrnauki_order.record_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
		})
	}
}

func TestRecordMinobrnaukiOrderUseCase_InvalidInput_AuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeRecordOrderRepo{}
	audit := &recordingAuditSink{}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, audit)

	in := validRecordOrderInput()
	in.OrderNumber = "   " // violates NewMinobrnaukiOrder invariant

	_, err := uc.Execute(context.Background(), 42, "methodist", in)
	assert.True(t, errors.Is(err, domain.ErrInvalidMinobrnaukiOrder),
		"expected ErrInvalidMinobrnaukiOrder, got %v", err)
	assert.Zero(t, repo.saveCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "minobrnauki_order.record_denied", audit.events[0].Action)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

func TestRecordMinobrnaukiOrderUseCase_InvalidChangeScope_ReturnsSentinel(t *testing.T) {
	repo := &fakeRecordOrderRepo{}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, &recordingAuditSink{})

	in := validRecordOrderInput()
	in.ChangeScope = "bogus"

	_, err := uc.Execute(context.Background(), 42, "methodist", in)
	assert.True(t, errors.Is(err, domain.ErrInvalidMinobrnaukiOrder),
		"invalid change_scope must surface ErrInvalidMinobrnaukiOrder, got %v", err)
	assert.Zero(t, repo.saveCalls)
}

func TestRecordMinobrnaukiOrderUseCase_TransportErrorPropagatesWithoutSuccessAudit(t *testing.T) {
	repo := &fakeRecordOrderRepo{saveErr: errors.New("conn refused")}
	audit := &recordingAuditSink{}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist", validRecordOrderInput())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
	assert.Empty(t, audit.events, "transport errors must not produce a recorded/denied audit event")
}

func TestRecordMinobrnaukiOrderUseCase_FiresRevisionTriggerAfterSave(t *testing.T) {
	repo := &fakeRecordOrderRepo{idAssigned: 100}
	trig := &fakeRevisionTrigger{}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, &recordingAuditSink{}).WithRevisionTrigger(trig)

	_, err := uc.Execute(context.Background(), 42, "methodist", validRecordOrderInput())
	require.NoError(t, err)

	require.Len(t, trig.calls, 1, "the revision trigger fires once after a successful record")
	assert.Equal(t, int64(42), trig.calls[0].actorID, "actor threads through to the trigger")
	assert.Equal(t, int64(100), trig.calls[0].orderID, "trigger receives the persisted order id")
	assert.Equal(t, "№ 1078 от 12.05.2026", trig.calls[0].orderNumber)
	assert.Equal(t, []int64{11, 22}, trig.calls[0].affected, "affected ids forwarded to the trigger")
}

func TestRecordMinobrnaukiOrderUseCase_DeniedRole_DoesNotFireTrigger(t *testing.T) {
	repo := &fakeRecordOrderRepo{}
	trig := &fakeRevisionTrigger{}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, &recordingAuditSink{}).WithRevisionTrigger(trig)

	_, err := uc.Execute(context.Background(), 1, "teacher", validRecordOrderInput())
	require.Error(t, err)
	assert.Empty(t, trig.calls, "a denied record never fires the revision trigger")
}

func TestRecordMinobrnaukiOrderUseCase_NilSinkTolerated(t *testing.T) {
	repo := &fakeRecordOrderRepo{idAssigned: 1}
	uc := NewRecordMinobrnaukiOrderUseCase(repo, nil)

	order, err := uc.Execute(context.Background(), 42, "methodist", validRecordOrderInput())
	require.NoError(t, err)
	assert.Equal(t, int64(1), order.ID())
}
