package usecases

import (
	"context"
	"sync"
	"testing"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
)

// v0.160.1 polish Item 6 — pins the forensic invariant that
// role_change / bulk_department_update / bulk_position_update audit
// rows carry `actor_user_id` (matching the post-#283 ADR-4
// status_change / delete_denied shape).
//
// Replays the RED→GREEN cycle that the original Item 6 `feat:`
// commit lacked: was `feat:` with arg-count swap in existing tests
// but no assertion on the new audit fields. This file adds the
// assertions so future refactors cannot silently drop actor_user_id.

// recordingAuditSink captures (action, resource, fields) triples
// emitted via LogAuditEvent. Distinct from testAuditLogger (which
// writes-and-drops к the real logger) so tests can assert on the
// emitted shape.
type recordingAuditSink struct {
	mu     sync.Mutex
	events []recordedAuditEvent
}

type recordedAuditEvent struct {
	action   string
	resource string
	fields   map[string]any
}

func (s *recordingAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make(map[string]any, len(fields))
	for k, v := range fields {
		copied[k] = v
	}
	s.events = append(s.events, recordedAuditEvent{action: action, resource: resource, fields: copied})
}

func (s *recordingAuditSink) find(action string) *recordedAuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.events {
		if s.events[i].action == action {
			return &s.events[i]
		}
	}
	return nil
}

func TestUserUseCase_AuditConsistency_ActorUserID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		action         string
		resource       string
		exercise       func(t *testing.T, uc *UserUseCase, actorID int64)
		wantTargetUser any
		wantExtra      map[string]any
	}{
		{
			name:     "role_change_carries_actor_and_target",
			action:   "role_change",
			resource: "user",
			exercise: func(t *testing.T, uc *UserUseCase, actorID int64) {
				t.Helper()
				if err := uc.UpdateUserRole(context.Background(), actorID, 42, &dto.UpdateUserRoleInput{Role: "teacher"}); err != nil {
					t.Fatalf("UpdateUserRole: %v", err)
				}
			},
			wantTargetUser: int64(42),
			wantExtra: map[string]any{
				"new_role": authDomain.RoleType("teacher"),
			},
		},
		{
			name:     "bulk_department_update_carries_actor_and_targets",
			action:   "bulk_department_update",
			resource: "user_profile",
			exercise: func(t *testing.T, uc *UserUseCase, actorID int64) {
				t.Helper()
				if err := uc.BulkUpdateDepartment(context.Background(), actorID, &dto.BulkUpdateDepartmentInput{
					UserIDs:      []int64{42, 43, 44},
					DepartmentID: nil,
				}); err != nil {
					t.Fatalf("BulkUpdateDepartment: %v", err)
				}
			},
			wantTargetUser: []int64{42, 43, 44},
			wantExtra:      map[string]any{},
		},
		{
			name:     "bulk_position_update_carries_actor_and_targets",
			action:   "bulk_position_update",
			resource: "user_profile",
			exercise: func(t *testing.T, uc *UserUseCase, actorID int64) {
				t.Helper()
				if err := uc.BulkUpdatePosition(context.Background(), actorID, &dto.BulkUpdatePositionInput{
					UserIDs:    []int64{42, 43},
					PositionID: nil,
				}); err != nil {
					t.Fatalf("BulkUpdatePosition: %v", err)
				}
			},
			wantTargetUser: []int64{42, 43},
			wantExtra:      map[string]any{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			userRepo := NewMockUserRepository()
			profileRepo := NewMockUserProfileRepository()
			deptRepo := NewMockDepartmentRepository()
			posRepo := NewMockPositionRepository()
			sink := &recordingAuditSink{}

			// Seed user 42 directly via Save (Create overwrites ID with auto-increment)
			_ = userRepo.Save(context.Background(), &authEntities.User{ID: 42, Email: "t@x", Role: authDomain.RoleStudent})

			uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, sink, nil)
			tc.exercise(t, uc, 7) // actorID = 7

			rec := sink.find(tc.action)
			if rec == nil {
				t.Fatalf("expected audit event %q к be emitted; got %+v", tc.action, sink.events)
			}
			if rec.resource != tc.resource {
				t.Errorf("resource: got %q, want %q", rec.resource, tc.resource)
			}
			if rec.fields["actor_user_id"] != int64(7) {
				t.Errorf("actor_user_id: got %v (%T), want int64(7)", rec.fields["actor_user_id"], rec.fields["actor_user_id"])
			}

			// Target field name + value
			targetKey := "target_user_id"
			if _, isSlice := tc.wantTargetUser.([]int64); isSlice {
				targetKey = "target_user_ids"
			}
			if got := rec.fields[targetKey]; !equalAny(got, tc.wantTargetUser) {
				t.Errorf("%s: got %v, want %v", targetKey, got, tc.wantTargetUser)
			}

			// Defense-in-depth: forbid the pre-v0.160.1 keys leaking back.
			if _, ok := rec.fields["user_id"]; ok {
				t.Errorf("audit row must not carry pre-rename `user_id` key (use target_user_id)")
			}
			if _, ok := rec.fields["user_ids"]; ok {
				t.Errorf("audit row must not carry pre-rename `user_ids` key (use target_user_ids)")
			}

			for k, v := range tc.wantExtra {
				if got := rec.fields[k]; !equalAny(got, v) {
					t.Errorf("extra field %s: got %v, want %v", k, got, v)
				}
			}
		})
	}
}

// equalAny compares two `any` values that may hold scalars or
// []int64 slices (the actor-audit field shapes used here). Avoids
// pulling reflect.DeepEqual для the common path while keeping slice
// equality reliable.
func equalAny(got, want any) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	gotSlice, gotOk := got.([]int64)
	wantSlice, wantOk := want.([]int64)
	if gotOk && wantOk {
		if len(gotSlice) != len(wantSlice) {
			return false
		}
		for i := range gotSlice {
			if gotSlice[i] != wantSlice[i] {
				return false
			}
		}
		return true
	}
	return got == want
}
