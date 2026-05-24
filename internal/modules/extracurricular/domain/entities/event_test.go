package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// validEventParams returns a syntactically valid NewExtracurricularEventParams
// instance to use as baseline в invariant-violation tests. Tests mutate
// one field at a time to isolate each invariant (mirror к
// validSectionParams).
func validEventParams(now time.Time) NewExtracurricularEventParams {
	return NewExtracurricularEventParams{
		Title:          "Концерт ко Дню учителя",
		Description:    "Праздничный концерт силами студенческого театра",
		Category:       CategoryCultural,
		TargetAudience: TargetAudienceAll,
		Location:       "Актовый зал, корпус 2",
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
		MaxCapacity:    nil,
		OrganizerID:    42,
		Now:            now,
	}
}

func TestNewExtracurricularEvent_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e, err := NewExtracurricularEvent(validEventParams(now))
	if err != nil {
		t.Fatalf("NewExtracurricularEvent returned unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("NewExtracurricularEvent returned nil entity with no error")
	}
	if got, want := e.Title(), "Концерт ко Дню учителя"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := e.Category(), CategoryCultural; got != want {
		t.Errorf("Category() = %q, want %q", got, want)
	}
	if got, want := e.TargetAudience(), TargetAudienceAll; got != want {
		t.Errorf("TargetAudience() = %q, want %q", got, want)
	}
	if got, want := e.Status(), StatusDraft; got != want {
		t.Errorf("Status() = %q, want %q (fresh events default to draft)", got, want)
	}
	if got, want := e.OrganizerID(), int64(42); got != want {
		t.Errorf("OrganizerID() = %d, want %d", got, want)
	}
	if got, want := e.Version(), 0; got != want {
		t.Errorf("Version() = %d, want %d for fresh event", got, want)
	}
	if e.MaxCapacity() != nil {
		t.Errorf("MaxCapacity() = %v, want nil for unlimited", e.MaxCapacity())
	}
	if !e.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", e.CreatedAt(), now)
	}
	if !e.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", e.UpdatedAt(), now)
	}
	if len(e.Participants()) != 0 {
		t.Errorf("Participants() = %v, want empty for fresh event", e.Participants())
	}
}

func TestNewExtracurricularEvent_TrimsTextFields(t *testing.T) {
	now := time.Now()
	p := validEventParams(now)
	p.Title = "  Концерт ко Дню учителя  "
	p.Description = "  Праздничный концерт  "
	p.Location = "  Актовый зал  "
	e, err := NewExtracurricularEvent(p)
	if err != nil {
		t.Fatalf("NewExtracurricularEvent returned unexpected error: %v", err)
	}
	if got, want := e.Title(), "Концерт ко Дню учителя"; got != want {
		t.Errorf("Title() = %q, want trimmed canonical form %q", got, want)
	}
	if got, want := e.Description(), "Праздничный концерт"; got != want {
		t.Errorf("Description() = %q, want trimmed %q", got, want)
	}
	if got, want := e.Location(), "Актовый зал"; got != want {
		t.Errorf("Location() = %q, want trimmed %q", got, want)
	}
}

func TestNewExtracurricularEvent_AcceptsBlankOptionalFields(t *testing.T) {
	now := time.Now()
	p := validEventParams(now)
	p.Description = ""
	p.Location = ""
	e, err := NewExtracurricularEvent(p)
	if err != nil {
		t.Fatalf("NewExtracurricularEvent returned unexpected error: %v", err)
	}
	if got := e.Description(); got != "" {
		t.Errorf("Description() = %q, want empty for blank input", got)
	}
	if got := e.Location(); got != "" {
		t.Errorf("Location() = %q, want empty for blank input", got)
	}
}

func TestNewExtracurricularEvent_InvariantViolations(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		mutate  func(*NewExtracurricularEventParams)
		wantErr error
	}{
		{
			name:    "empty title",
			mutate:  func(p *NewExtracurricularEventParams) { p.Title = "   " },
			wantErr: ErrInvalidEvent,
		},
		{
			name: "oversize title",
			mutate: func(p *NewExtracurricularEventParams) {
				p.Title = strings.Repeat("ы", maxEventTitleLen+1)
			},
			wantErr: ErrInvalidEvent,
		},
		{
			name: "oversize description",
			mutate: func(p *NewExtracurricularEventParams) {
				p.Description = strings.Repeat("а", maxEventDescriptionLen+1)
			},
			wantErr: ErrInvalidEvent,
		},
		{
			name: "oversize location",
			mutate: func(p *NewExtracurricularEventParams) {
				p.Location = strings.Repeat("л", maxEventLocationLen+1)
			},
			wantErr: ErrInvalidEvent,
		},
		{
			name:    "non-positive organizer_id",
			mutate:  func(p *NewExtracurricularEventParams) { p.OrganizerID = 0 },
			wantErr: ErrInvalidEvent,
		},
		{
			name:    "negative organizer_id",
			mutate:  func(p *NewExtracurricularEventParams) { p.OrganizerID = -1 },
			wantErr: ErrInvalidEvent,
		},
		{
			name: "start_at equal end_at",
			mutate: func(p *NewExtracurricularEventParams) {
				t0 := now.Add(48 * time.Hour)
				p.StartAt = t0
				p.EndAt = t0
			},
			wantErr: ErrInvalidEvent,
		},
		{
			name: "start_at after end_at",
			mutate: func(p *NewExtracurricularEventParams) {
				p.StartAt = now.Add(50 * time.Hour)
				p.EndAt = now.Add(48 * time.Hour)
			},
			wantErr: ErrInvalidEvent,
		},
		{
			name: "negative max_capacity",
			mutate: func(p *NewExtracurricularEventParams) {
				neg := -1
				p.MaxCapacity = &neg
			},
			wantErr: ErrInvalidEvent,
		},
		// invalid_category + invalid_target_audience deferred к Pair 2
		// (VO IsValid() restrictive impl). Pair 1 GREEN still calls
		// IsValid() on both fields, but the stub returns always-true so
		// Pair 1 cannot fail those cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validEventParams(now)
			tt.mutate(&p)
			e, err := NewExtracurricularEvent(p)
			if err == nil {
				t.Fatalf("NewExtracurricularEvent expected error, got nil event=%+v", e)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewExtracurricularEvent error = %v, want errors.Is(%v)", err, tt.wantErr)
			}
			if e != nil {
				t.Errorf("NewExtracurricularEvent returned non-nil entity on invariant violation: %+v", e)
			}
		})
	}
}

func TestNewExtracurricularEvent_AcceptsZeroMaxCapacity(t *testing.T) {
	// max_capacity == 0 is semantically distinct from nil — a zero-cap
	// event is valid construction-wise (e.g. for "view-only" event
	// pages without registration). Registration attempts will fail с
	// ErrEventFull. Mirrors the CHECK constraint chk_extracurricular_capacity_nonneg
	// which accepts ≥ 0.
	now := time.Now()
	p := validEventParams(now)
	zero := 0
	p.MaxCapacity = &zero
	e, err := NewExtracurricularEvent(p)
	if err != nil {
		t.Fatalf("NewExtracurricularEvent returned unexpected error: %v", err)
	}
	if e.MaxCapacity() == nil || *e.MaxCapacity() != 0 {
		t.Errorf("MaxCapacity() = %v, want pointer to 0", e.MaxCapacity())
	}
}
