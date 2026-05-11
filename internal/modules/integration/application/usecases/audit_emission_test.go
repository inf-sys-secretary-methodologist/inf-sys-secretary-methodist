package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// recordingAuditSink captures every LogAuditEvent invocation so the
// emission contract (action / resource / fields shape) can be pinned
// independently of the platform AuditLogger side effects. Mirror к
// the assignments/messaging fakeAuditSink pattern.
type recordingAuditSink struct {
	events []recordedAuditEvent
}

type recordedAuditEvent struct {
	action   string
	resource string
	fields   map[string]any
}

func (r *recordingAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	r.events = append(r.events, recordedAuditEvent{action: action, resource: resource, fields: fields})
}

// --- SyncUseCase emissions ---

func TestSyncUseCase_AuditEmission_StartSync_AlreadyRunning_NoEmission(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	uc.mu.Lock()
	uc.running[entities.SyncEntityEmployee] = true
	uc.mu.Unlock()

	_, err := uc.StartSync(context.Background(), &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
		Force:      false,
	})
	require.Error(t, err)
	require.Empty(t, sink.events, "already-running rejection must not emit any audit event")
}

func TestSyncUseCase_AuditEmission_UnsupportedEntityType_FailsButEmitsLifecyclePair(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	_, err := uc.StartSync(context.Background(), &dto.StartSyncRequest{
		EntityType: entities.SyncEntityType("unknown"),
		Direction:  entities.SyncDirectionImport,
	})
	require.Error(t, err)

	// One started + one failed event — the failure path must still
	// record that the attempt happened (forensic invariant: every
	// sync attempt leaves a trail even when it never produced rows).
	require.Len(t, sink.events, 2)
	require.Equal(t, "integration.sync_started", sink.events[0].action)
	require.Equal(t, "integration_sync", sink.events[0].resource)
	require.Equal(t, "unknown", sink.events[0].fields["entity_type"])

	require.Equal(t, "integration.sync_failed", sink.events[1].action)
	require.Equal(t, "integration_sync", sink.events[1].resource)
	require.Contains(t, sink.events[1].fields["error"], "unsupported entity type")
}

func TestSyncUseCase_AuditEmission_CancelSync_Success(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	// Seed a running sync log so CancelSync proceeds.
	log := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	log.Start()
	require.NoError(t, syncLogRepo.Create(context.Background(), log))

	err := uc.CancelSync(context.Background(), log.ID)
	require.NoError(t, err)

	require.Len(t, sink.events, 1)
	require.Equal(t, "integration.sync_canceled", sink.events[0].action)
	require.Equal(t, "integration_sync", sink.events[0].resource)
	require.Equal(t, log.ID, sink.events[0].fields["sync_log_id"])
}

func TestSyncUseCase_AuditEmission_CancelSync_NotRunning_NoEmission(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	// Seed a completed (not-running) sync log.
	log := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	log.Start()
	log.Complete()
	require.NoError(t, syncLogRepo.Create(context.Background(), log))

	err := uc.CancelSync(context.Background(), log.ID)
	require.Error(t, err)
	require.Empty(t, sink.events, "cancel attempt on non-running sync must not emit")
}

// --- ConflictUseCase emissions ---

func TestConflictUseCase_AuditEmission_Resolve_Success(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	conflict := createTestConflict(repo, 5, entities.SyncEntityEmployee, "emp-1")

	err := uc.Resolve(context.Background(), conflict.ID, int64(42), &dto.ResolveConflictRequest{
		Resolution: entities.ConflictResolutionUseLocal,
		Notes:      "manual",
	})
	require.NoError(t, err)

	require.Len(t, sink.events, 1)
	ev := sink.events[0]
	require.Equal(t, "integration.conflict_resolved", ev.action)
	require.Equal(t, "integration_conflict", ev.resource)
	require.Equal(t, int64(42), ev.fields["actor_user_id"])
	require.Equal(t, conflict.ID, ev.fields["conflict_id"])
	require.Equal(t, string(entities.ConflictResolutionUseLocal), ev.fields["resolution"])
}

func TestConflictUseCase_AuditEmission_Resolve_AlreadyResolved_NoEmission(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	conflict := createTestConflict(repo, 5, entities.SyncEntityEmployee, "emp-1")
	// Pre-resolve directly through the repo so the use case sees a non-pending state.
	require.NoError(t, repo.Resolve(context.Background(), conflict.ID, entities.ConflictResolutionUseLocal, 1, ""))

	err := uc.Resolve(context.Background(), conflict.ID, int64(42), &dto.ResolveConflictRequest{
		Resolution: entities.ConflictResolutionUseLocal,
	})
	require.Error(t, err)
	require.Empty(t, sink.events, "already-resolved conflict must not emit a fresh event")
}

func TestConflictUseCase_AuditEmission_BulkResolve_Success(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	c1 := createTestConflict(repo, 5, entities.SyncEntityEmployee, "emp-1")
	c2 := createTestConflict(repo, 5, entities.SyncEntityEmployee, "emp-2")
	c3 := createTestConflict(repo, 5, entities.SyncEntityEmployee, "emp-3")

	err := uc.BulkResolve(context.Background(), int64(42), &dto.BulkResolveRequest{
		IDs:        []int64{c1.ID, c2.ID, c3.ID},
		Resolution: entities.ConflictResolutionUseExternal,
	})
	require.NoError(t, err)

	require.Len(t, sink.events, 1, "bulk operation emits one summary event, not one per conflict")
	ev := sink.events[0]
	require.Equal(t, "integration.conflict_bulk_resolved", ev.action)
	require.Equal(t, "integration_conflict", ev.resource)
	require.Equal(t, int64(42), ev.fields["actor_user_id"])
	require.Equal(t, 3, ev.fields["conflict_count"])
	require.Equal(t, string(entities.ConflictResolutionUseExternal), ev.fields["resolution"])
}

func TestConflictUseCase_AuditEmission_BulkResolve_EmptyIDs_NoEmission(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)
	sink := &recordingAuditSink{}
	uc.WithAuditSink(sink)

	err := uc.BulkResolve(context.Background(), int64(42), &dto.BulkResolveRequest{
		IDs:        []int64{},
		Resolution: entities.ConflictResolutionUseLocal,
	})
	require.Error(t, err)
	require.Empty(t, sink.events, "empty ID list must not emit a bulk event")
}

// --- Nil-sink silent paths (backward-compat invariant) ---

func TestSyncUseCase_AuditEmission_NilSinkSilent(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	// No WithAuditSink — sink stays nil.

	uc.mu.Lock()
	uc.running[entities.SyncEntityEmployee] = true
	uc.mu.Unlock()

	_, _ = uc.StartSync(context.Background(), &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
		Force:      false,
	})
	// Must not panic. No assertion on side-effects — backward-compat
	// invariant is that the legacy nil-sink path does not crash.
}

func TestConflictUseCase_AuditEmission_NilSinkSilent(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)
	// No WithAuditSink — sink stays nil.

	conflict := createTestConflict(repo, 5, entities.SyncEntityEmployee, "emp-1")
	err := uc.Resolve(context.Background(), conflict.ID, int64(42), &dto.ResolveConflictRequest{
		Resolution: entities.ConflictResolutionUseLocal,
	})
	require.NoError(t, err)
}
