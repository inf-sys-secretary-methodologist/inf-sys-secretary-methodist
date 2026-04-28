package entities

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestNewLesson_Deterministic(t *testing.T) {
	fixedTime := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	dateStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	dateEnd := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	lesson := NewLesson(1, 2, 3, 4, 5, 6, domain.Monday, "09:00", "10:30", domain.WeekTypeAll, dateStart, dateEnd, fixedTime)
	if lesson == nil {
		t.Fatal("NewLesson returned nil")
	}
	if lesson.SemesterID != 1 {
		t.Errorf("SemesterID = %d, want 1", lesson.SemesterID)
	}
	if lesson.DayOfWeek != domain.Monday {
		t.Errorf("DayOfWeek = %v, want Monday", lesson.DayOfWeek)
	}
	if lesson.TimeStart != "09:00" {
		t.Errorf("TimeStart = %v, want 09:00", lesson.TimeStart)
	}
	if !lesson.DateStart.Equal(dateStart) {
		t.Errorf("DateStart = %v, want %v", lesson.DateStart, dateStart)
	}
	if !lesson.DateEnd.Equal(dateEnd) {
		t.Errorf("DateEnd = %v, want %v", lesson.DateEnd, dateEnd)
	}
	if !lesson.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want %v", lesson.CreatedAt, fixedTime)
	}
	if !lesson.UpdatedAt.Equal(fixedTime) {
		t.Errorf("UpdatedAt = %v, want %v", lesson.UpdatedAt, fixedTime)
	}
	if lesson.IsCancelled {
		t.Error("new lesson should not be cancelled")
	}
}

func TestNewScheduleChange_Deterministic(t *testing.T) {
	fixedTime := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	origDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

	change := NewScheduleChange(42, domain.ChangeTypeCancelled, origDate, 7, fixedTime)
	if change == nil {
		t.Fatal("NewScheduleChange returned nil")
	}
	if change.LessonID != 42 {
		t.Errorf("LessonID = %d, want 42", change.LessonID)
	}
	if !change.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want %v", change.CreatedAt, fixedTime)
	}
}

func TestLesson_Validate(t *testing.T) {
	validLesson := func() *Lesson {
		now := time.Now()
		return NewLesson(1, 1, 1, 1, 1, 1, domain.Monday, "09:00", "10:30", domain.WeekTypeAll,
			time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			now,
		)
	}

	tests := []struct {
		name    string
		modify  func(*Lesson)
		wantErr error
	}{
		{
			name:    "valid lesson",
			modify:  func(_ *Lesson) {},
			wantErr: nil,
		},
		{
			name:    "invalid day of week 0",
			modify:  func(l *Lesson) { l.DayOfWeek = domain.DayOfWeek(0) },
			wantErr: ErrInvalidDayOfWeek,
		},
		{
			name:    "invalid day of week 8",
			modify:  func(l *Lesson) { l.DayOfWeek = domain.DayOfWeek(8) },
			wantErr: ErrInvalidDayOfWeek,
		},
		{
			name:    "invalid week type",
			modify:  func(l *Lesson) { l.WeekType = domain.WeekType("invalid") },
			wantErr: ErrInvalidWeekType,
		},
		{
			name:    "end time before start time",
			modify:  func(l *Lesson) { l.TimeStart = "10:30"; l.TimeEnd = "09:00" },
			wantErr: ErrInvalidTimeRange,
		},
		{
			name:    "equal start and end time",
			modify:  func(l *Lesson) { l.TimeStart = "09:00"; l.TimeEnd = "09:00" },
			wantErr: ErrInvalidTimeRange,
		},
		{
			name: "date_end before date_start",
			modify: func(l *Lesson) {
				l.DateStart = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
				l.DateEnd = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
			},
			wantErr: ErrInvalidDateRange,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := validLesson()
			tt.modify(l)
			err := l.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() expected error %v, got nil", tt.wantErr)
				} else if err != tt.wantErr {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}
