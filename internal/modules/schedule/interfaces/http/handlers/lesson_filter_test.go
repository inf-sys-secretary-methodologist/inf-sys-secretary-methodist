package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestLessonFilterFromInput(t *testing.T) {
	semesterID := int64(1)
	groupID := int64(2)
	teacherID := int64(3)
	dayOfWeek := 1
	weekType := "odd"

	input := dto.LessonFilterInput{
		SemesterID: &semesterID,
		GroupID:    &groupID,
		TeacherID:  &teacherID,
		DayOfWeek:  &dayOfWeek,
		WeekType:   &weekType,
		Limit:      50,
		Offset:     10,
	}

	filter := lessonFilterFromInput(input)

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
