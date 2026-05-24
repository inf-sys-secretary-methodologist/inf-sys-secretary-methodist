package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestUpdateBasics_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e, err := NewExtracurricularEvent(validEventParams(now))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	later := now.Add(time.Hour)
	newStart := now.Add(72 * time.Hour)
	newEnd := now.Add(76 * time.Hour)
	cap50 := 50
	p := UpdateEventBasicsParams{
		Title:          "Обновлённый концерт",
		Description:    "Новое описание",
		Category:       CategorySports,
		TargetAudience: TargetAudienceStudents,
		Location:       "Спортзал",
		StartAt:        newStart,
		EndAt:          newEnd,
		MaxCapacity:    &cap50,
		Now:            later,
	}
	if err := e.UpdateBasics(p); err != nil {
		t.Fatalf("UpdateBasics: %v", err)
	}
	if got, want := e.Title(), "Обновлённый концерт"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := e.Category(), CategorySports; got != want {
		t.Errorf("Category() = %q, want %q", got, want)
	}
	if got, want := e.TargetAudience(), TargetAudienceStudents; got != want {
		t.Errorf("TargetAudience() = %q, want %q", got, want)
	}
	if got := e.MaxCapacity(); got == nil || *got != 50 {
		t.Errorf("MaxCapacity() = %v, want 50", got)
	}
	if !e.UpdatedAt().Equal(later) {
		t.Errorf("UpdatedAt() = %v, want %v", e.UpdatedAt(), later)
	}
}

func TestUpdateBasics_RejectsClosedStatuses(t *testing.T) {
	now := time.Now()
	closed := []struct {
		name       string
		transition func(*ExtracurricularEvent) error
	}{
		{name: "canceled", transition: func(e *ExtracurricularEvent) error { return e.Cancel(time.Now()) }},
		{name: "completed", transition: func(e *ExtracurricularEvent) error {
			if err := e.Publish(time.Now()); err != nil {
				return err
			}
			return e.Complete(time.Now())
		}},
	}
	for _, tt := range closed {
		t.Run(tt.name, func(t *testing.T) {
			p := validEventParams(now)
			e, err := NewExtracurricularEvent(p)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}
			if err := tt.transition(e); err != nil {
				t.Fatalf("transition: %v", err)
			}
			updateP := UpdateEventBasicsParams{
				Title:          "should fail",
				Category:       CategoryCultural,
				TargetAudience: TargetAudienceAll,
				StartAt:        now.Add(48 * time.Hour),
				EndAt:          now.Add(50 * time.Hour),
				Now:            now,
			}
			err = e.UpdateBasics(updateP)
			if err == nil {
				t.Fatal("UpdateBasics on closed status returned nil, want ErrCannotEditEvent")
			}
			if !errors.Is(err, ErrCannotEditEvent) {
				t.Errorf("UpdateBasics error = %v, want errors.Is(%v)", err, ErrCannotEditEvent)
			}
		})
	}
}

func TestUpdateBasics_RejectsInvariantViolations(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		mutate func(*UpdateEventBasicsParams)
	}{
		{name: "empty title", mutate: func(p *UpdateEventBasicsParams) { p.Title = "" }},
		{name: "oversize title", mutate: func(p *UpdateEventBasicsParams) {
			p.Title = strings.Repeat("ы", maxEventTitleLen+1)
		}},
		{name: "oversize description", mutate: func(p *UpdateEventBasicsParams) {
			p.Description = strings.Repeat("а", maxEventDescriptionLen+1)
		}},
		{name: "oversize location", mutate: func(p *UpdateEventBasicsParams) {
			p.Location = strings.Repeat("л", maxEventLocationLen+1)
		}},
		{name: "invalid category", mutate: func(p *UpdateEventBasicsParams) { p.Category = Category("bogus") }},
		{name: "invalid audience", mutate: func(p *UpdateEventBasicsParams) { p.TargetAudience = TargetAudience("bogus") }},
		{name: "start_at after end_at", mutate: func(p *UpdateEventBasicsParams) {
			p.StartAt = now.Add(50 * time.Hour)
			p.EndAt = now.Add(48 * time.Hour)
		}},
		{name: "negative max_capacity", mutate: func(p *UpdateEventBasicsParams) {
			neg := -1
			p.MaxCapacity = &neg
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := validEventParams(now)
			e, err := NewExtracurricularEvent(ep)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}
			updateP := UpdateEventBasicsParams{
				Title:          ep.Title,
				Description:    ep.Description,
				Category:       ep.Category,
				TargetAudience: ep.TargetAudience,
				Location:       ep.Location,
				StartAt:        ep.StartAt,
				EndAt:          ep.EndAt,
				MaxCapacity:    ep.MaxCapacity,
				Now:            now,
			}
			tt.mutate(&updateP)
			err = e.UpdateBasics(updateP)
			if err == nil {
				t.Fatal("UpdateBasics returned nil, want ErrInvalidEvent")
			}
			if !errors.Is(err, ErrInvalidEvent) {
				t.Errorf("UpdateBasics error = %v, want errors.Is(%v)", err, ErrInvalidEvent)
			}
		})
	}
}

func TestUpdateBasics_RejectsCapacityBelowParticipants(t *testing.T) {
	// Reducing max_capacity below current participant count would
	// violate the aggregate invariant `len(participants) <= maxCapacity`.
	// UpdateBasics must reject such reductions per DDD always-valid rule.
	now := time.Now()
	e := publishedEvent(t, now, nil)
	for _, uid := range []int64{101, 102, 103} {
		if err := e.Register(uid, now); err != nil {
			t.Fatalf("setup Register: %v", err)
		}
	}
	smallCap := 2
	updateP := UpdateEventBasicsParams{
		Title:          "title",
		Category:       CategoryCultural,
		TargetAudience: TargetAudienceAll,
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
		MaxCapacity:    &smallCap,
		Now:            now,
	}
	err := e.UpdateBasics(updateP)
	if err == nil {
		t.Fatal("UpdateBasics with capacity below participants returned nil, want ErrInvalidEvent")
	}
	if !errors.Is(err, ErrInvalidEvent) {
		t.Errorf("UpdateBasics error = %v, want errors.Is(%v)", err, ErrInvalidEvent)
	}
}

func TestAuthorizeEventCreate(t *testing.T) {
	tests := []struct {
		role    string
		isAdmin bool
		wantErr bool
	}{
		{role: "system_admin", isAdmin: true, wantErr: false},
		{role: "methodist", isAdmin: false, wantErr: false},
		{role: "academic_secretary", isAdmin: false, wantErr: false},
		{role: "teacher", isAdmin: false, wantErr: true},
		{role: "student", isAdmin: false, wantErr: true},
		{role: "", isAdmin: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			err := AuthorizeEventCreate(tt.role, tt.isAdmin)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("AuthorizeEventCreate(%q, isAdmin=%v) err=%v, wantErr=%v", tt.role, tt.isAdmin, err, tt.wantErr)
			}
			if gotErr && !errors.Is(err, ErrEventScopeForbidden) {
				t.Errorf("AuthorizeEventCreate err = %v, want errors.Is(%v)", err, ErrEventScopeForbidden)
			}
		})
	}
}

func TestAuthorizeEventEdit(t *testing.T) {
	const organizerID int64 = 42
	tests := []struct {
		name      string
		actorID   int64
		actorRole string
		isAdmin   bool
		wantErr   bool
	}{
		{name: "admin allowed for any event", actorID: 999, actorRole: "system_admin", isAdmin: true, wantErr: false},
		{name: "organizer self-edit (methodist)", actorID: organizerID, actorRole: "methodist", isAdmin: false, wantErr: false},
		{name: "organizer self-edit (secretary)", actorID: organizerID, actorRole: "academic_secretary", isAdmin: false, wantErr: false},
		{name: "other methodist denied", actorID: 100, actorRole: "methodist", isAdmin: false, wantErr: true},
		{name: "other secretary denied", actorID: 100, actorRole: "academic_secretary", isAdmin: false, wantErr: true},
		{name: "teacher denied even for own", actorID: organizerID, actorRole: "teacher", isAdmin: false, wantErr: true},
		{name: "student denied even for own", actorID: organizerID, actorRole: "student", isAdmin: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AuthorizeEventEdit(tt.actorID, organizerID, tt.actorRole, tt.isAdmin)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("AuthorizeEventEdit err=%v, wantErr=%v", err, tt.wantErr)
			}
			if gotErr && !errors.Is(err, ErrEventScopeForbidden) {
				t.Errorf("AuthorizeEventEdit err = %v, want errors.Is(%v)", err, ErrEventScopeForbidden)
			}
		})
	}
}

func TestCanViewEvent(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		audience TargetAudience
		want     bool
	}{
		// `all` visible to everyone
		{name: "all/student", role: "student", audience: TargetAudienceAll, want: true},
		{name: "all/teacher", role: "teacher", audience: TargetAudienceAll, want: true},
		{name: "all/secretary", role: "academic_secretary", audience: TargetAudienceAll, want: true},
		// `students` audience
		{name: "students/student", role: "student", audience: TargetAudienceStudents, want: true},
		{name: "students/teacher", role: "teacher", audience: TargetAudienceStudents, want: false},
		// `teachers` audience
		{name: "teachers/teacher", role: "teacher", audience: TargetAudienceTeachers, want: true},
		{name: "teachers/student", role: "student", audience: TargetAudienceTeachers, want: false},
		// `staff` includes methodist + secretary + admin
		{name: "staff/methodist", role: "methodist", audience: TargetAudienceStaff, want: true},
		{name: "staff/secretary", role: "academic_secretary", audience: TargetAudienceStaff, want: true},
		{name: "staff/teacher", role: "teacher", audience: TargetAudienceStaff, want: false},
		{name: "staff/student", role: "student", audience: TargetAudienceStaff, want: false},
		// unknown role rejected
		{name: "unknown role/all", role: "ghost", audience: TargetAudienceAll, want: true},
		{name: "unknown role/students", role: "ghost", audience: TargetAudienceStudents, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanViewEvent(tt.role, tt.audience)
			if got != tt.want {
				t.Errorf("CanViewEvent(%q, %q) = %v, want %v", tt.role, tt.audience, got, tt.want)
			}
		})
	}
}
