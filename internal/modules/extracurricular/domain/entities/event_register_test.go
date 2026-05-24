package entities

import (
	"errors"
	"testing"
	"time"
)

// publishedEvent returns an event already в `published` status — the
// only state that accepts registration per ADR-2. Tests build на этом
// helper to focus on Register/Unregister invariants.
func publishedEvent(t *testing.T, now time.Time, maxCapacity *int) *ExtracurricularEvent {
	t.Helper()
	p := validEventParams(now)
	p.MaxCapacity = maxCapacity
	e, err := NewExtracurricularEvent(p)
	if err != nil {
		t.Fatalf("setup NewExtracurricularEvent: %v", err)
	}
	if err := e.Publish(now); err != nil {
		t.Fatalf("setup Publish: %v", err)
	}
	return e
}

func TestRegister_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e := publishedEvent(t, now, nil)
	regAt := now.Add(time.Hour)
	if err := e.Register(101, regAt); err != nil {
		t.Fatalf("Register returned unexpected error: %v", err)
	}
	parts := e.Participants()
	if len(parts) != 1 {
		t.Fatalf("Participants() len = %d, want 1", len(parts))
	}
	if parts[0].UserID != 101 {
		t.Errorf("Participants[0].UserID = %d, want 101", parts[0].UserID)
	}
	if !parts[0].RegisteredAt.Equal(regAt) {
		t.Errorf("Participants[0].RegisteredAt = %v, want %v", parts[0].RegisteredAt, regAt)
	}
	if !e.HasParticipant(101) {
		t.Error("HasParticipant(101) = false after Register, want true")
	}
}

func TestRegister_RejectsDoubleRegistration(t *testing.T) {
	now := time.Now()
	e := publishedEvent(t, now, nil)
	if err := e.Register(101, now); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := e.Register(101, now.Add(time.Hour))
	if err == nil {
		t.Fatal("second Register returned nil, want ErrParticipantExists")
	}
	if !errors.Is(err, ErrParticipantExists) {
		t.Errorf("Register error = %v, want errors.Is(%v)", err, ErrParticipantExists)
	}
	if len(e.Participants()) != 1 {
		t.Errorf("Participants() len = %d, want 1 (no duplicate)", len(e.Participants()))
	}
}

func TestRegister_RespectsCapacity(t *testing.T) {
	now := time.Now()
	cap2 := 2
	e := publishedEvent(t, now, &cap2)
	if err := e.Register(101, now); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	if err := e.Register(102, now); err != nil {
		t.Fatalf("second Register: %v", err)
	}
	// Third registration must fail — cap reached.
	err := e.Register(103, now)
	if err == nil {
		t.Fatal("third Register returned nil, want ErrEventFull")
	}
	if !errors.Is(err, ErrEventFull) {
		t.Errorf("Register error = %v, want errors.Is(%v)", err, ErrEventFull)
	}
	if len(e.Participants()) != 2 {
		t.Errorf("Participants() len = %d, want 2 (capacity respected)", len(e.Participants()))
	}
}

func TestRegister_ZeroCapacityRejectsAll(t *testing.T) {
	// max_capacity == 0 is valid construction, blocks all registrations.
	now := time.Now()
	zero := 0
	e := publishedEvent(t, now, &zero)
	err := e.Register(101, now)
	if err == nil {
		t.Fatal("Register on zero-cap event returned nil, want ErrEventFull")
	}
	if !errors.Is(err, ErrEventFull) {
		t.Errorf("Register error = %v, want errors.Is(%v)", err, ErrEventFull)
	}
}

func TestRegister_UnlimitedCapacity(t *testing.T) {
	// max_capacity == nil → unlimited (no cap check).
	now := time.Now()
	e := publishedEvent(t, now, nil)
	for i := int64(1); i <= 100; i++ {
		if err := e.Register(i, now); err != nil {
			t.Fatalf("Register user %d: %v", i, err)
		}
	}
	if len(e.Participants()) != 100 {
		t.Errorf("Participants() len = %d, want 100", len(e.Participants()))
	}
}

func TestRegister_RejectsClosedStatuses(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		toStatus  func(*ExtracurricularEvent) error
		wantErrIs error
	}{
		{
			name:      "draft — not visible, not registrable",
			toStatus:  func(e *ExtracurricularEvent) error { return nil }, // stay draft
			wantErrIs: ErrEventNotOpenForRegistration,
		},
		{
			name:      "canceled — terminal, not registrable",
			toStatus:  func(e *ExtracurricularEvent) error { return e.Cancel(time.Now()) },
			wantErrIs: ErrEventNotOpenForRegistration,
		},
		{
			name:      "completed — archived, not registrable",
			toStatus:  func(e *ExtracurricularEvent) error { return e.Complete(time.Now()) },
			wantErrIs: ErrEventNotOpenForRegistration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validEventParams(now)
			e, err := NewExtracurricularEvent(p)
			if err != nil {
				t.Fatalf("setup NewExtracurricularEvent: %v", err)
			}
			// Some closed statuses require Publish first (canceled,
			// completed); draft test stays at construction status.
			if tt.name != "draft — not visible, not registrable" {
				if err := e.Publish(now); err != nil {
					t.Fatalf("setup Publish: %v", err)
				}
			}
			if err := tt.toStatus(e); err != nil {
				t.Fatalf("setup transition: %v", err)
			}
			err = e.Register(101, now.Add(time.Hour))
			if err == nil {
				t.Fatal("Register on closed status returned nil, want sentinel")
			}
			if !errors.Is(err, tt.wantErrIs) {
				t.Errorf("Register error = %v, want errors.Is(%v)", err, tt.wantErrIs)
			}
		})
	}
}

func TestRegister_RejectsNonPositiveUserID(t *testing.T) {
	now := time.Now()
	e := publishedEvent(t, now, nil)
	tests := []int64{0, -1, -100}
	for _, uid := range tests {
		err := e.Register(uid, now)
		if err == nil {
			t.Errorf("Register(uid=%d) returned nil, want ErrInvalidEvent", uid)
			continue
		}
		if !errors.Is(err, ErrInvalidEvent) {
			t.Errorf("Register(uid=%d) error = %v, want errors.Is(%v)", uid, err, ErrInvalidEvent)
		}
	}
}

func TestUnregister_HappyPath(t *testing.T) {
	now := time.Now()
	e := publishedEvent(t, now, nil)
	if err := e.Register(101, now); err != nil {
		t.Fatalf("setup Register: %v", err)
	}
	if err := e.Unregister(101); err != nil {
		t.Fatalf("Unregister returned unexpected error: %v", err)
	}
	if e.HasParticipant(101) {
		t.Error("HasParticipant(101) = true after Unregister, want false")
	}
	if len(e.Participants()) != 0 {
		t.Errorf("Participants() len = %d after Unregister, want 0", len(e.Participants()))
	}
}

func TestUnregister_NotFound(t *testing.T) {
	now := time.Now()
	e := publishedEvent(t, now, nil)
	err := e.Unregister(101)
	if err == nil {
		t.Fatal("Unregister of un-registered user returned nil, want ErrParticipantNotFound")
	}
	if !errors.Is(err, ErrParticipantNotFound) {
		t.Errorf("Unregister error = %v, want errors.Is(%v)", err, ErrParticipantNotFound)
	}
}

func TestUnregister_PreservesOtherParticipants(t *testing.T) {
	now := time.Now()
	e := publishedEvent(t, now, nil)
	for _, uid := range []int64{101, 102, 103} {
		if err := e.Register(uid, now); err != nil {
			t.Fatalf("setup Register(%d): %v", uid, err)
		}
	}
	if err := e.Unregister(102); err != nil {
		t.Fatalf("Unregister(102): %v", err)
	}
	parts := e.Participants()
	if len(parts) != 2 {
		t.Fatalf("Participants() len = %d, want 2", len(parts))
	}
	ids := map[int64]bool{}
	for _, pt := range parts {
		ids[pt.UserID] = true
	}
	if !ids[101] || !ids[103] || ids[102] {
		t.Errorf("Participants after Unregister(102) = %v, want {101,103}", ids)
	}
}

func TestPublish_FromDraftOnly(t *testing.T) {
	now := time.Now()
	// draft → published OK
	{
		p := validEventParams(now)
		e, err := NewExtracurricularEvent(p)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		later := now.Add(time.Hour)
		if err := e.Publish(later); err != nil {
			t.Fatalf("Publish from draft: %v", err)
		}
		if e.Status() != StatusPublished {
			t.Errorf("Status() = %q, want %q", e.Status(), StatusPublished)
		}
		if !e.UpdatedAt().Equal(later) {
			t.Errorf("UpdatedAt() = %v, want %v", e.UpdatedAt(), later)
		}
	}
	// published → published rejected (idempotent error not silent)
	{
		e := publishedEvent(t, now, nil)
		if err := e.Publish(now); err == nil {
			t.Error("Publish on already-published returned nil, want error")
		}
	}
	// canceled → published rejected
	{
		e := publishedEvent(t, now, nil)
		if err := e.Cancel(now); err != nil {
			t.Fatalf("setup Cancel: %v", err)
		}
		if err := e.Publish(now); err == nil {
			t.Error("Publish on canceled returned nil, want error")
		}
	}
}

func TestCancel_FromActiveOnly(t *testing.T) {
	now := time.Now()
	// draft → canceled OK
	{
		p := validEventParams(now)
		e, err := NewExtracurricularEvent(p)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := e.Cancel(now); err != nil {
			t.Fatalf("Cancel from draft: %v", err)
		}
		if e.Status() != StatusCanceled {
			t.Errorf("Status() = %q, want %q", e.Status(), StatusCanceled)
		}
	}
	// published → canceled OK
	{
		e := publishedEvent(t, now, nil)
		if err := e.Cancel(now); err != nil {
			t.Fatalf("Cancel from published: %v", err)
		}
		if e.Status() != StatusCanceled {
			t.Errorf("Status() = %q, want %q", e.Status(), StatusCanceled)
		}
	}
	// canceled → canceled rejected
	{
		e := publishedEvent(t, now, nil)
		if err := e.Cancel(now); err != nil {
			t.Fatalf("setup Cancel: %v", err)
		}
		if err := e.Cancel(now); err == nil {
			t.Error("Cancel on already-canceled returned nil, want error")
		}
	}
}

func TestComplete_FromPublishedOnly(t *testing.T) {
	now := time.Now()
	// published → completed OK
	{
		e := publishedEvent(t, now, nil)
		if err := e.Complete(now); err != nil {
			t.Fatalf("Complete from published: %v", err)
		}
		if e.Status() != StatusCompleted {
			t.Errorf("Status() = %q, want %q", e.Status(), StatusCompleted)
		}
	}
	// draft → completed rejected
	{
		p := validEventParams(now)
		e, err := NewExtracurricularEvent(p)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := e.Complete(now); err == nil {
			t.Error("Complete from draft returned nil, want error")
		}
	}
	// canceled → completed rejected
	{
		e := publishedEvent(t, now, nil)
		if err := e.Cancel(now); err != nil {
			t.Fatalf("setup Cancel: %v", err)
		}
		if err := e.Complete(now); err == nil {
			t.Error("Complete from canceled returned nil, want error")
		}
	}
}
