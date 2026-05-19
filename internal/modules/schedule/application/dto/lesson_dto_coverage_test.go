package dto

// v0.153.8 Phase 6 backfill — closes ToSemesterOutput +
// ToScheduleChangeOutput (both at 0%). Pure entity → DTO mappers,
// no I/O. No production change.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

func TestToSemesterOutput(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	sem := &entities.Semester{
		ID:             3,
		AcademicYearID: 7,
		Name:           "Осенний семестр 2026/27",
		Number:         1,
		StartDate:      start,
		EndDate:        end,
		IsActive:       true,
	}

	got := ToSemesterOutput(sem)
	assert.Equal(t, int64(3), got.ID)
	assert.Equal(t, int64(7), got.AcademicYearID)
	assert.Equal(t, "Осенний семестр 2026/27", got.Name)
	assert.Equal(t, 1, got.Number)
	assert.True(t, got.StartDate.Equal(start))
	assert.True(t, got.EndDate.Equal(end))
	assert.True(t, got.IsActive)
}

func TestToScheduleChangeOutput_MinimalFields(t *testing.T) {
	original := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 10, 14, 18, 0, 0, 0, time.UTC)
	change := &entities.ScheduleChange{
		ID:           101,
		LessonID:     42,
		ChangeType:   domain.ChangeTypeCancelled,
		OriginalDate: original,
		CreatedBy:    7,
		CreatedAt:    createdAt,
	}

	got := ToScheduleChangeOutput(change)
	assert.Equal(t, int64(101), got.ID)
	assert.Equal(t, int64(42), got.LessonID)
	assert.Equal(t, string(domain.ChangeTypeCancelled), got.ChangeType)
	assert.True(t, got.OriginalDate.Equal(original))
	assert.True(t, got.CreatedAt.Equal(createdAt))
	assert.Equal(t, int64(7), got.CreatedBy)
	assert.Nil(t, got.NewDate)
	assert.Nil(t, got.NewClassroomID)
	assert.Nil(t, got.NewTeacherID)
	assert.Nil(t, got.Reason)
}

func TestToScheduleChangeOutput_AllOptionalFieldsPopulated(t *testing.T) {
	original := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	newDate := original.Add(24 * time.Hour)
	newClassroom := int64(202)
	newTeacher := int64(303)
	reason := "Перенос по болезни преподавателя"
	change := &entities.ScheduleChange{
		ID:             102,
		LessonID:       42,
		ChangeType:     domain.ChangeTypeMoved,
		OriginalDate:   original,
		NewDate:        &newDate,
		NewClassroomID: &newClassroom,
		NewTeacherID:   &newTeacher,
		Reason:         &reason,
		CreatedBy:      8,
		CreatedAt:      original,
	}

	got := ToScheduleChangeOutput(change)
	require.NotNil(t, got.NewDate)
	require.NotNil(t, got.NewClassroomID)
	require.NotNil(t, got.NewTeacherID)
	require.NotNil(t, got.Reason)
	assert.True(t, got.NewDate.Equal(newDate))
	assert.Equal(t, int64(202), *got.NewClassroomID)
	assert.Equal(t, int64(303), *got.NewTeacherID)
	assert.Equal(t, "Перенос по болезни преподавателя", *got.Reason)
}
