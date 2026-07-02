package handlers

// v0.153.10 Phase 6 backfill — covers LessonHandler read endpoints
// (List/GetTimetable/ListChanges/ListClassrooms/ListStudentGroups)
// + handleError sentinel mapping branches. Builds a LessonUseCase
// against minimal in-memory fakes for the 4 repository surfaces.
// No production change.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// ===== Minimal in-memory fakes for the 4 schedule repos =====

type lessonRepoFake struct {
	lessons []*entities.Lesson
	listErr error
	byID    *entities.Lesson
	byIDErr error
}

func (r *lessonRepoFake) Create(_ context.Context, _ *entities.Lesson) error { return nil }
func (r *lessonRepoFake) Save(_ context.Context, _ *entities.Lesson) error   { return nil }
func (r *lessonRepoFake) GetByID(_ context.Context, _ int64) (*entities.Lesson, error) {
	return r.byID, r.byIDErr
}
func (r *lessonRepoFake) Delete(_ context.Context, _ int64) error { return nil }
func (r *lessonRepoFake) List(_ context.Context, _ usecases.LessonFilter, _, _ int) ([]*entities.Lesson, error) {
	return r.lessons, r.listErr
}
func (r *lessonRepoFake) Count(_ context.Context, _ usecases.LessonFilter) (int64, error) {
	return int64(len(r.lessons)), nil
}
func (r *lessonRepoFake) GetTimetable(_ context.Context, _ usecases.LessonFilter) ([]*entities.Lesson, error) {
	return r.lessons, r.listErr
}

type classroomRepoFake struct {
	rooms []*entities.Classroom
}

func (r *classroomRepoFake) GetByID(_ context.Context, _ int64) (*entities.Classroom, error) {
	return nil, nil
}
func (r *classroomRepoFake) List(_ context.Context, _ usecases.ClassroomFilter, _, _ int) ([]*entities.Classroom, error) {
	return r.rooms, nil
}
func (r *classroomRepoFake) Count(_ context.Context, _ usecases.ClassroomFilter) (int64, error) {
	return int64(len(r.rooms)), nil
}

type referenceRepoFake struct {
	groups []*entities.StudentGroup
}

func (r *referenceRepoFake) ListStudentGroups(_ context.Context, _, _ int) ([]*entities.StudentGroup, error) {
	return r.groups, nil
}
func (r *referenceRepoFake) ListDisciplines(_ context.Context, _, _ int) ([]*entities.Discipline, error) {
	return nil, nil
}
func (r *referenceRepoFake) ListSemesters(_ context.Context, _ bool) ([]*entities.Semester, error) {
	return nil, nil
}
func (r *referenceRepoFake) ListLessonTypes(_ context.Context) ([]*entities.LessonType, error) {
	return nil, nil
}
func (r *referenceRepoFake) GetActiveSemester(_ context.Context) (*entities.Semester, error) {
	return nil, nil
}

type changeRepoFake struct {
	changes []*entities.ScheduleChange
	err     error
}

func (r *changeRepoFake) Create(_ context.Context, _ *entities.ScheduleChange) error { return nil }
func (r *changeRepoFake) GetByLessonID(_ context.Context, _ int64) ([]*entities.ScheduleChange, error) {
	return r.changes, r.err
}
func (r *changeRepoFake) GetByDateRange(_ context.Context, _, _ time.Time) ([]*entities.ScheduleChange, error) {
	return r.changes, nil
}

// buildLessonHandler constructs LessonHandler wired against in-memory
// fakes so the read endpoints can fire end-to-end без real DB.
func buildLessonHandler(t *testing.T, lessonRepo *lessonRepoFake, changeRepo *changeRepoFake, classroomRepo *classroomRepoFake, refRepo *referenceRepoFake) *LessonHandler {
	t.Helper()
	uc := usecases.NewLessonUseCase(lessonRepo, classroomRepo, refRepo, changeRepo, nil)
	return NewLessonHandler(uc)
}

// performGet routes a GET request through the handler-mounted engine.
func performGet(t *testing.T, h *LessonHandler, path string, route string, mount func(r *gin.Engine, h *LessonHandler)) *httptest.ResponseRecorder {
	t.Helper()
	r := gin.New()
	mount(r, h)
	_ = route
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

// ===== LessonHandler.List =====

func TestLessonHandler_List_HappyPath(t *testing.T) {
	lessonRepo := &lessonRepoFake{}
	h := buildLessonHandler(t, lessonRepo, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons", "/schedule/lessons", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons", h.List)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLessonHandler_List_InvalidQuery(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons?limit=abc", "/schedule/lessons", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons", h.List)
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_List_UsecaseError(t *testing.T) {
	lessonRepo := &lessonRepoFake{listErr: fmt.Errorf("db down")}
	h := buildLessonHandler(t, lessonRepo, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons", "/schedule/lessons", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons", h.List)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== LessonHandler.GetTimetable =====

func TestLessonHandler_GetTimetable_HappyPath(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons/timetable", "/schedule/lessons/timetable", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons/timetable", h.GetTimetable)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLessonHandler_GetTimetable_InvalidQuery(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons/timetable?day_of_week=abc", "/schedule/lessons/timetable", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons/timetable", h.GetTimetable)
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_GetTimetable_UsecaseError(t *testing.T) {
	lessonRepo := &lessonRepoFake{listErr: fmt.Errorf("db down")}
	h := buildLessonHandler(t, lessonRepo, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons/timetable", "/schedule/lessons/timetable", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons/timetable", h.GetTimetable)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== LessonHandler.ListChanges =====

func TestLessonHandler_ListChanges_HappyPath(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/changes?lesson_id=42", "/schedule/changes", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/changes", h.ListChanges)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLessonHandler_ListChanges_MissingLessonID(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/changes", "/schedule/changes", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/changes", h.ListChanges)
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_ListChanges_InvalidLessonID(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/changes?lesson_id=abc", "/schedule/changes", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/changes", h.ListChanges)
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_ListChanges_UsecaseError(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{err: fmt.Errorf("db down")}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/changes?lesson_id=42", "/schedule/changes", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/changes", h.ListChanges)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== LessonHandler.ListClassrooms =====

func TestLessonHandler_ListClassrooms_HappyPath(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/classrooms", "/schedule/classrooms", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/classrooms", h.ListClassrooms)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLessonHandler_ListClassrooms_WithAllFilters(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/classrooms?building=A&type=lecture&min_capacity=30&is_available=true&limit=50&offset=10", "/schedule/classrooms", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/classrooms", h.ListClassrooms)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== LessonHandler.ListStudentGroups =====

func TestLessonHandler_ListStudentGroups_HappyPath(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/groups", "/schedule/groups", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/groups", h.ListStudentGroups)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== LessonHandler.handleError exhaustive sentinel mapping =====

func TestLessonHandler_HandleError_NotFound(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.handleError(c, usecases.ErrLessonNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLessonHandler_HandleError_Forbidden(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.handleError(c, usecases.ErrForbidden)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_HandleError_InvalidInput(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.handleError(c, usecases.ErrInvalidInput)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_HandleError_UnknownError(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.handleError(c, fmt.Errorf("random error"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Sanity guard так that require is imported (used implicitly via t.Helper assertions).
var _ = require.NotNil
