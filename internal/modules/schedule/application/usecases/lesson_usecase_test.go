package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// ===================== Create =====================

func TestLessonUseCase_Create_HappyPath(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	input := validCreateLessonInput()
	lesson, err := uc.Create(ctx, 1, input)

	require.NoError(t, err)
	assert.NotZero(t, lesson.ID)
	assert.Equal(t, int64(1), lesson.SemesterID)
	assert.Equal(t, int64(2), lesson.DisciplineID)
	assert.Equal(t, domain.Monday, lesson.DayOfWeek)
	assert.Equal(t, "09:00", lesson.TimeStart)
	assert.Equal(t, "10:30", lesson.TimeEnd)
	assert.Equal(t, domain.WeekTypeAll, lesson.WeekType)
}

func TestLessonUseCase_Create_InvalidDayOfWeek(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	input := validCreateLessonInput()
	input.DayOfWeek = domain.DayOfWeek(0) // invalid

	_, err := uc.Create(ctx, 1, input)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestLessonUseCase_Create_InvalidTimeRange(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	input := validCreateLessonInput()
	input.TimeStart = "10:30"
	input.TimeEnd = "09:00" // end before start

	_, err := uc.Create(ctx, 1, input)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestLessonUseCase_Create_InvalidWeekType(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	input := validCreateLessonInput()
	input.WeekType = domain.WeekType("invalid")

	_, err := uc.Create(ctx, 1, input)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestLessonUseCase_Create_WithNotes(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	input := validCreateLessonInput()
	notes := "Laboratory session"
	input.Notes = &notes

	lesson, err := uc.Create(ctx, 1, input)
	require.NoError(t, err)
	require.NotNil(t, lesson.Notes)
	assert.Equal(t, "Laboratory session", *lesson.Notes)
}

// ===================== GetByID =====================

func TestLessonUseCase_GetByID_Found(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, validCreateLessonInput())
	require.NoError(t, err)

	found, err := uc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestLessonUseCase_GetByID_NotFound(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	_, err := uc.GetByID(ctx, 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

// ===================== Update =====================

func TestLessonUseCase_Update_HappyPath(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, validCreateLessonInput())
	require.NoError(t, err)

	newClassroom := int64(99)
	newNotes := "Updated notes"
	updated, err := uc.Update(ctx, 1, created.ID, UpdateLessonInputForUC{
		ClassroomID: &newClassroom,
		Notes:       &newNotes,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(99), updated.ClassroomID)
	require.NotNil(t, updated.Notes)
	assert.Equal(t, "Updated notes", *updated.Notes)
}

func TestLessonUseCase_Update_NotFound(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	newClassroom := int64(99)
	_, err := uc.Update(ctx, 1, 999, UpdateLessonInputForUC{
		ClassroomID: &newClassroom,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

func TestLessonUseCase_Update_InvalidDayOfWeek(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, validCreateLessonInput())
	require.NoError(t, err)

	invalidDay := domain.DayOfWeek(0)
	_, err = uc.Update(ctx, 1, created.ID, UpdateLessonInputForUC{
		DayOfWeek: &invalidDay,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

// ===================== Delete =====================

func TestLessonUseCase_Delete_HappyPath(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	created, err := uc.Create(ctx, 1, validCreateLessonInput())
	require.NoError(t, err)

	err = uc.Delete(ctx, 1, created.ID)
	require.NoError(t, err)

	_, err = uc.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

func TestLessonUseCase_Delete_NotFound(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

// ===================== List =====================

func TestLessonUseCase_List(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	_, _ = uc.Create(ctx, 1, validCreateLessonInput())
	_, _ = uc.Create(ctx, 1, validCreateLessonInput())

	lessons, err := uc.List(ctx, repositories.LessonFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, lessons, 2)
}

// ===================== Count =====================

func TestLessonUseCase_Count(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	_, _ = uc.Create(ctx, 1, validCreateLessonInput())
	_, _ = uc.Create(ctx, 1, validCreateLessonInput())

	count, err := uc.Count(ctx, repositories.LessonFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// ===================== GetTimetable =====================

func TestLessonUseCase_GetTimetable(t *testing.T) {
	uc, _ := setupLessonUseCase()
	ctx := context.Background()

	_, _ = uc.Create(ctx, 1, validCreateLessonInput())

	lessons, err := uc.GetTimetable(ctx, repositories.LessonFilter{})
	require.NoError(t, err)
	assert.Len(t, lessons, 1)
}

// ===================== CreateChange =====================

func TestLessonUseCase_CreateChange_HappyPath(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	lesson, _ := uc.Create(ctx, 1, validCreateLessonInput())

	change, err := uc.CreateChange(ctx, 1, CreateChangeInputForUC{
		LessonID:     lesson.ID,
		ChangeType:   domain.ChangeTypeCancelled,
		OriginalDate: time.Now(),
	})
	require.NoError(t, err)
	assert.NotZero(t, change.ID)
	assert.Equal(t, lesson.ID, change.LessonID)
	assert.Equal(t, domain.ChangeTypeCancelled, change.ChangeType)
	assert.Equal(t, int64(1), change.CreatedBy)
}

func TestLessonUseCase_CreateChange_InvalidChangeType(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	lesson, _ := uc.Create(ctx, 1, validCreateLessonInput())

	_, err := uc.CreateChange(ctx, 1, CreateChangeInputForUC{
		LessonID:     lesson.ID,
		ChangeType:   domain.ChangeType("invalid"),
		OriginalDate: time.Now(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

// ===================== ListChanges =====================

func TestLessonUseCase_ListChanges(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	lesson, _ := uc.Create(ctx, 1, validCreateLessonInput())
	_, _ = uc.CreateChange(ctx, 1, CreateChangeInputForUC{
		LessonID:     lesson.ID,
		ChangeType:   domain.ChangeTypeCancelled,
		OriginalDate: time.Now(),
	})

	changes, err := uc.ListChanges(ctx, lesson.ID)
	require.NoError(t, err)
	assert.Len(t, changes, 1)
}

// ===================== Reference data delegates =====================

func TestLessonUseCase_ListClassrooms(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	classrooms, err := uc.ListClassrooms(ctx, repositories.ClassroomFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, classrooms)
}

func TestLessonUseCase_ListStudentGroups(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	groups, err := uc.ListStudentGroups(ctx, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, groups)
}

func TestLessonUseCase_ListDisciplines(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	disciplines, err := uc.ListDisciplines(ctx, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, disciplines)
}

func TestLessonUseCase_ListSemesters(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	semesters, err := uc.ListSemesters(ctx, false)
	require.NoError(t, err)
	assert.Empty(t, semesters)
}

func TestLessonUseCase_ListLessonTypes(t *testing.T) {
	uc, _, _, _, _ := setupLessonUseCaseAll()
	ctx := context.Background()

	types, err := uc.ListLessonTypes(ctx)
	require.NoError(t, err)
	assert.Empty(t, types)
}
