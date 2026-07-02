package entities

import (
	"errors"
	"testing"
	"time"
)

func TestNewLessonSlot(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		number    int
		timeStart string
		timeEnd   string
		wantErr   error
	}{
		{name: "valid slot", number: 1, timeStart: "08:30", timeEnd: "10:00", wantErr: nil},
		{name: "valid late slot", number: 6, timeStart: "17:40", timeEnd: "19:10", wantErr: nil},
		{name: "zero number", number: 0, timeStart: "08:30", timeEnd: "10:00", wantErr: ErrInvalidSlotNumber},
		{name: "negative number", number: -2, timeStart: "08:30", timeEnd: "10:00", wantErr: ErrInvalidSlotNumber},
		{name: "bad start format", number: 1, timeStart: "8:30", timeEnd: "10:00", wantErr: ErrInvalidSlotTimeFormat},
		{name: "non-time start", number: 1, timeStart: "morning", timeEnd: "10:00", wantErr: ErrInvalidSlotTimeFormat},
		{name: "bad end format", number: 1, timeStart: "08:30", timeEnd: "25:00", wantErr: ErrInvalidSlotTimeFormat},
		{name: "end before start", number: 1, timeStart: "10:00", timeEnd: "08:30", wantErr: ErrInvalidSlotTimeRange},
		{name: "end equals start", number: 1, timeStart: "08:30", timeEnd: "08:30", wantErr: ErrInvalidSlotTimeRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, err := NewLessonSlot(tt.number, tt.timeStart, tt.timeEnd, now)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if slot != nil {
					t.Fatalf("expected nil slot on error, got %+v", slot)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if slot == nil {
				t.Fatal("expected slot, got nil")
			}
			if slot.Number != tt.number || slot.TimeStart != tt.timeStart || slot.TimeEnd != tt.timeEnd {
				t.Fatalf("fields not set: got %+v", slot)
			}
			if !slot.CreatedAt.Equal(now) || !slot.UpdatedAt.Equal(now) {
				t.Fatalf("timestamps not set to now: got created=%v updated=%v", slot.CreatedAt, slot.UpdatedAt)
			}
		})
	}
}
