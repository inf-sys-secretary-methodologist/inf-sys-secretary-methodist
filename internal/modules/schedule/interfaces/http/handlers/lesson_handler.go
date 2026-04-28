// Package handlers contains HTTP request handlers for the schedule module.
package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// LessonHandler handles HTTP requests for lesson endpoints.
type LessonHandler struct {
	lessonUseCase *usecases.LessonUseCase
}

// NewLessonHandler creates a new LessonHandler.
func NewLessonHandler(lessonUseCase *usecases.LessonUseCase) *LessonHandler {
	return &LessonHandler{lessonUseCase: lessonUseCase}
}

// getUserID extracts user ID from context.
func (h *LessonHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return 0, false
	}
	return userID.(int64), true
}

// getIDParam extracts ID parameter from URL.
func (h *LessonHandler) getIDParam(c *gin.Context, param string) (int64, bool) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, false
	}
	return id, true
}

// canModifySchedule checks if user has permission to create/update/delete lessons.
// Only system_admin and academic_secretary have full schedule access.
func (h *LessonHandler) canModifySchedule(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	roleStr, ok := role.(string)
	if !ok {
		return false
	}
	return roleStr == "system_admin" || roleStr == "academic_secretary"
}

// requireScheduleWrite checks permission and returns 403 if denied.
func (h *LessonHandler) requireScheduleWrite(c *gin.Context) bool {
	if !h.canModifySchedule(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions for schedule modification"})
		return false
	}
	return true
}

// handleError handles use case errors.
func (h *LessonHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrLessonNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
	case errors.Is(err, usecases.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, usecases.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// Create handles lesson creation.
func (h *LessonHandler) Create(c *gin.Context) {
	if !h.requireScheduleWrite(c) {
		return
	}
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.CreateLessonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dateStart, err := time.Parse("2006-01-02", input.DateStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_start format, expected YYYY-MM-DD"})
		return
	}
	dateEnd, err := time.Parse("2006-01-02", input.DateEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_end format, expected YYYY-MM-DD"})
		return
	}

	ucInput := usecases.CreateLessonInputForUC{
		SemesterID:   input.SemesterID,
		DisciplineID: input.DisciplineID,
		LessonTypeID: input.LessonTypeID,
		TeacherID:    input.TeacherID,
		GroupID:      input.GroupID,
		ClassroomID:  input.ClassroomID,
		DayOfWeek:    domain.DayOfWeek(input.DayOfWeek),
		TimeStart:    input.TimeStart,
		TimeEnd:      input.TimeEnd,
		WeekType:     domain.WeekType(input.WeekType),
		DateStart:    dateStart,
		DateEnd:      dateEnd,
		Notes:        input.Notes,
	}

	lesson, err := h.lessonUseCase.Create(c.Request.Context(), userID, ucInput)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToLessonOutput(lesson))
}

// List lists lessons with filters.
func (h *LessonHandler) List(c *gin.Context) {
	var input dto.LessonFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Limit == 0 {
		input.Limit = 100
	}

	filter := input.ToFilter()

	lessons, err := h.lessonUseCase.List(c.Request.Context(), filter, input.Limit, input.Offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	total, err := h.lessonUseCase.Count(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := dto.LessonListOutput{
		Lessons: make([]dto.LessonOutput, 0, len(lessons)),
		Total:   total,
		Limit:   input.Limit,
		Offset:  input.Offset,
	}
	for _, lesson := range lessons {
		output.Lessons = append(output.Lessons, dto.ToLessonOutput(lesson))
	}

	c.JSON(http.StatusOK, output)
}

// GetTimetable returns the timetable (flat array, no pagination).
func (h *LessonHandler) GetTimetable(c *gin.Context) {
	var input dto.LessonFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := input.ToFilter()

	lessons, err := h.lessonUseCase.GetTimetable(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.LessonOutput, 0, len(lessons))
	for _, lesson := range lessons {
		output = append(output, dto.ToLessonOutput(lesson))
	}

	c.JSON(http.StatusOK, output)
}

// GetByID retrieves a lesson by ID.
func (h *LessonHandler) GetByID(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	lesson, err := h.lessonUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToLessonOutput(lesson))
}

// Update updates a lesson.
func (h *LessonHandler) Update(c *gin.Context) {
	if !h.requireScheduleWrite(c) {
		return
	}
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.UpdateLessonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ucInput := usecases.UpdateLessonInputForUC{
		SemesterID:   input.SemesterID,
		DisciplineID: input.DisciplineID,
		LessonTypeID: input.LessonTypeID,
		TeacherID:    input.TeacherID,
		GroupID:      input.GroupID,
		ClassroomID:  input.ClassroomID,
		Notes:        input.Notes,
	}

	if input.DayOfWeek != nil {
		d := domain.DayOfWeek(*input.DayOfWeek)
		ucInput.DayOfWeek = &d
	}
	if input.TimeStart != nil {
		ucInput.TimeStart = input.TimeStart
	}
	if input.TimeEnd != nil {
		ucInput.TimeEnd = input.TimeEnd
	}
	if input.WeekType != nil {
		w := domain.WeekType(*input.WeekType)
		ucInput.WeekType = &w
	}

	lesson, err := h.lessonUseCase.Update(c.Request.Context(), userID, id, ucInput)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToLessonOutput(lesson))
}

// Delete deletes a lesson.
func (h *LessonHandler) Delete(c *gin.Context) {
	if !h.requireScheduleWrite(c) {
		return
	}
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.lessonUseCase.Delete(c.Request.Context(), userID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateChange creates a schedule change.
func (h *LessonHandler) CreateChange(c *gin.Context) {
	if !h.requireScheduleWrite(c) {
		return
	}
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.CreateChangeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ucInput := usecases.CreateChangeInputForUC{
		LessonID:       input.LessonID,
		ChangeType:     domain.ChangeType(input.ChangeType),
		OriginalDate:   input.OriginalDate,
		NewDate:        input.NewDate,
		NewClassroomID: input.NewClassroomID,
		NewTeacherID:   input.NewTeacherID,
		Reason:         input.Reason,
	}

	change, err := h.lessonUseCase.CreateChange(c.Request.Context(), userID, ucInput)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToScheduleChangeOutput(change))
}

// ListChanges lists schedule changes for a lesson.
func (h *LessonHandler) ListChanges(c *gin.Context) {
	lessonIDStr := c.Query("lesson_id")
	if lessonIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lesson_id is required"})
		return
	}

	lessonID, err := strconv.ParseInt(lessonIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson_id"})
		return
	}

	changes, err := h.lessonUseCase.ListChanges(c.Request.Context(), lessonID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.ScheduleChangeOutput, 0, len(changes))
	for _, change := range changes {
		output = append(output, dto.ToScheduleChangeOutput(change))
	}

	c.JSON(http.StatusOK, gin.H{"changes": output})
}

// ListClassrooms lists classrooms.
func (h *LessonHandler) ListClassrooms(c *gin.Context) {
	var filter repositories.ClassroomFilter

	if building := c.Query("building"); building != "" {
		filter.Building = &building
	}
	if classroomType := c.Query("type"); classroomType != "" {
		filter.Type = &classroomType
	}
	if minCapStr := c.Query("min_capacity"); minCapStr != "" {
		if minCap, err := strconv.Atoi(minCapStr); err == nil {
			filter.MinCapacity = &minCap
		}
	}
	if avail := c.Query("is_available"); avail != "" {
		isAvail := avail == "true"
		filter.IsAvailable = &isAvail
	}

	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	classrooms, err := h.lessonUseCase.ListClassrooms(c.Request.Context(), filter, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.ClassroomOutput, 0, len(classrooms))
	for _, cr := range classrooms {
		output = append(output, dto.ToClassroomOutput(cr))
	}

	c.JSON(http.StatusOK, gin.H{"classrooms": output})
}

// ListStudentGroups lists student groups.
func (h *LessonHandler) ListStudentGroups(c *gin.Context) {
	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	groups, err := h.lessonUseCase.ListStudentGroups(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.StudentGroupOutput, 0, len(groups))
	for _, g := range groups {
		output = append(output, dto.ToStudentGroupOutput(g))
	}

	c.JSON(http.StatusOK, gin.H{"student_groups": output})
}

// ListDisciplines lists disciplines.
func (h *LessonHandler) ListDisciplines(c *gin.Context) {
	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	disciplines, err := h.lessonUseCase.ListDisciplines(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.DisciplineOutput, 0, len(disciplines))
	for _, d := range disciplines {
		output = append(output, dto.ToDisciplineOutput(d))
	}

	c.JSON(http.StatusOK, gin.H{"disciplines": output})
}

// ListSemesters lists semesters.
func (h *LessonHandler) ListSemesters(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"

	semesters, err := h.lessonUseCase.ListSemesters(c.Request.Context(), activeOnly)
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.SemesterOutput, 0, len(semesters))
	for _, s := range semesters {
		output = append(output, dto.ToSemesterOutput(s))
	}

	c.JSON(http.StatusOK, gin.H{"semesters": output})
}

// ListLessonTypes lists lesson types.
func (h *LessonHandler) ListLessonTypes(c *gin.Context) {
	types, err := h.lessonUseCase.ListLessonTypes(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	output := make([]dto.LessonTypeOutput, 0, len(types))
	for _, lt := range types {
		output = append(output, dto.ToLessonTypeOutput(lt))
	}

	c.JSON(http.StatusOK, gin.H{"lesson_types": output})
}
