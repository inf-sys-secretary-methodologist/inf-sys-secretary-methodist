package entities

import (
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestNewTeachingLoad(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		semesterID   int64
		groupID      int64
		disciplineID int64
		teacherID    int64
		lessonTypeID int64
		pairs        int
		weekType     domain.WeekType
		wantErr      error
	}{
		{name: "valid all-weeks", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 2, weekType: domain.WeekTypeAll, wantErr: nil},
		{name: "valid odd", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 1, weekType: domain.WeekTypeOdd, wantErr: nil},
		{name: "zero semester", semesterID: 0, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 2, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadReference},
		{name: "zero group", semesterID: 1, groupID: 0, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 2, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadReference},
		{name: "negative teacher", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: -1, lessonTypeID: 5, pairs: 2, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadReference},
		{name: "zero discipline", semesterID: 1, groupID: 2, disciplineID: 0, teacherID: 4, lessonTypeID: 5, pairs: 2, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadReference},
		{name: "zero lesson type", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 0, pairs: 2, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadReference},
		{name: "zero pairs", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 0, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadPairs},
		{name: "negative pairs", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: -3, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadPairs},
		{name: "too many pairs", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 999, weekType: domain.WeekTypeAll, wantErr: ErrInvalidLoadPairs},
		{name: "bad week type", semesterID: 1, groupID: 2, disciplineID: 3, teacherID: 4, lessonTypeID: 5, pairs: 2, weekType: domain.WeekType("weekly"), wantErr: ErrInvalidLoadWeekType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load, err := NewTeachingLoad(tt.semesterID, tt.groupID, tt.disciplineID, tt.teacherID, tt.lessonTypeID, tt.pairs, tt.weekType, now)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if load != nil {
					t.Fatalf("expected nil load on error, got %+v", load)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if load == nil {
				t.Fatal("expected load, got nil")
			}
			if load.PairsPerWeek != tt.pairs || load.WeekType != tt.weekType || load.SemesterID != tt.semesterID {
				t.Fatalf("fields not set: got %+v", load)
			}
			if !load.CreatedAt.Equal(now) || !load.UpdatedAt.Equal(now) {
				t.Fatalf("timestamps not set to now")
			}
		})
	}
}
