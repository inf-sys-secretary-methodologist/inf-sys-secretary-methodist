package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToLessonOutput(t *testing.T) {
	now := time.Now()
	notes := "Important lecture"
	color := "#3B82F6"
	code := "CS101"
	classroomName := "Main Hall"
	classroomType := "lecture"

	lesson := &entities.Lesson{
		ID:           1,
		SemesterID:   10,
		DisciplineID: 20,
		LessonTypeID: 30,
		TeacherID:    40,
		GroupID:       50,
		ClassroomID:  60,
		DayOfWeek:    domain.Monday,
		TimeStart:    "09:00",
		TimeEnd:      "10:30",
		WeekType:     domain.WeekTypeAll,
		DateStart:    now,
		DateEnd:      now.Add(120 * 24 * time.Hour),
		Notes:        &notes,
		IsCancelled:  false,
		CreatedAt:    now,
		UpdatedAt:    now,
		Discipline: &entities.Discipline{
			ID:   20,
			Name: "Computer Science",
			Code: &code,
		},
		LessonType: &entities.LessonType{
			ID:        30,
			Name:      "Lecture",
			ShortName: "Lec",
			Color:     &color,
		},
		Classroom: &entities.Classroom{
			ID:          60,
			Building:    "A",
			Number:      "101",
			Name:        &classroomName,
			Capacity:    200,
			Type:        &classroomType,
			IsAvailable: true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Group: &entities.StudentGroup{
			ID:   50,
			Name: "CS-21",
		},
		Teacher: &entities.TeacherInfo{
			ID:    40,
			Name:  "Prof. Smith",
			Email: "smith@example.com",
		},
	}

	output := ToLessonOutput(lesson)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.SemesterID)
	assert.Equal(t, int64(20), output.DisciplineID)
	assert.Equal(t, int64(30), output.LessonTypeID)
	assert.Equal(t, int64(40), output.TeacherID)
	assert.Equal(t, int64(50), output.GroupID)
	assert.Equal(t, int64(60), output.ClassroomID)
	assert.Equal(t, 1, output.DayOfWeek)
	assert.Equal(t, "09:00", output.TimeStart)
	assert.Equal(t, "10:30", output.TimeEnd)
	assert.Equal(t, "all", output.WeekType)
	assert.Equal(t, now, output.DateStart)
	assert.Equal(t, &notes, output.Notes)
	assert.False(t, output.IsCancelled)

	// associations
	require.NotNil(t, output.Discipline)
	assert.Equal(t, int64(20), output.Discipline.ID)
	assert.Equal(t, "Computer Science", output.Discipline.Name)
	assert.Equal(t, &code, output.Discipline.Code)

	require.NotNil(t, output.LessonType)
	assert.Equal(t, int64(30), output.LessonType.ID)
	assert.Equal(t, "Lecture", output.LessonType.Name)
	assert.Equal(t, "Lec", output.LessonType.ShortName)
	assert.Equal(t, &color, output.LessonType.Color)

	require.NotNil(t, output.Classroom)
	assert.Equal(t, int64(60), output.Classroom.ID)
	assert.Equal(t, "A", output.Classroom.Building)
	assert.Equal(t, "101", output.Classroom.Number)
	assert.Equal(t, &classroomName, output.Classroom.Name)
	assert.Equal(t, 200, output.Classroom.Capacity)

	require.NotNil(t, output.Group)
	assert.Equal(t, int64(50), output.Group.ID)
	assert.Equal(t, "CS-21", output.Group.Name)

	require.NotNil(t, output.Teacher)
	assert.Equal(t, int64(40), output.Teacher.ID)
	assert.Equal(t, "Prof. Smith", output.Teacher.Name)
	assert.Equal(t, "smith@example.com", output.Teacher.Email)
}

func TestLessonFilterInput_ToFilter(t *testing.T) {
	semesterID := int64(1)
	groupID := int64(2)
	teacherID := int64(3)
	dayOfWeek := 1
	weekType := "odd"

	input := &LessonFilterInput{
		SemesterID: &semesterID,
		GroupID:    &groupID,
		TeacherID:  &teacherID,
		DayOfWeek:  &dayOfWeek,
		WeekType:   &weekType,
		Limit:      50,
		Offset:     10,
	}

	filter := input.ToFilter()

	assert.Equal(t, &semesterID, filter.SemesterID)
	assert.Equal(t, &groupID, filter.GroupID)
	assert.Equal(t, &teacherID, filter.TeacherID)
	assert.Nil(t, filter.ClassroomID)
	assert.Nil(t, filter.DisciplineID)

	require.NotNil(t, filter.DayOfWeek)
	assert.Equal(t, domain.Monday, *filter.DayOfWeek)

	require.NotNil(t, filter.WeekType)
	assert.Equal(t, domain.WeekTypeOdd, *filter.WeekType)
}

func TestToLessonOutput_NilAssociations(t *testing.T) {
	now := time.Now()
	lesson := &entities.Lesson{
		ID:           1,
		SemesterID:   10,
		DisciplineID: 20,
		LessonTypeID: 30,
		TeacherID:    40,
		GroupID:       50,
		ClassroomID:  60,
		DayOfWeek:    domain.Tuesday,
		TimeStart:    "14:00",
		TimeEnd:      "15:30",
		WeekType:     domain.WeekTypeEven,
		DateStart:    now,
		DateEnd:      now.Add(120 * 24 * time.Hour),
		IsCancelled:  false,
		CreatedAt:    now,
		UpdatedAt:    now,
		// All associations nil
	}

	output := ToLessonOutput(lesson)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Nil(t, output.Discipline)
	assert.Nil(t, output.LessonType)
	assert.Nil(t, output.Classroom)
	assert.Nil(t, output.Group)
	assert.Nil(t, output.Teacher)
	assert.Nil(t, output.Notes)
}
