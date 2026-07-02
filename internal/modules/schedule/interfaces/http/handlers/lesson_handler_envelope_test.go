package handlers

// Envelope contract tests: every lesson/timetable/changes endpoint must wrap
// its payload in the standard {success, data, meta} response envelope so the
// frontend (scheduleLessonsApi/scheduleChangesApi read response.data) receives
// data in the shape it expects — list/timetable/changes as a bare array,
// get-by-id as an object. Previously these endpoints returned naked payloads,
// which broke the timetable readers.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// envelope mirrors the shared response.Response success shape for assertions.
type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    struct {
		Pagination *struct {
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	} `json:"meta"`
}

func decodeEnvelope(t *testing.T, body string) envelope {
	t.Helper()
	var e envelope
	require.NoError(t, json.Unmarshal([]byte(body), &e), "body must be valid JSON envelope: %s", body)
	assert.True(t, e.Success, "envelope.success must be true")
	return e
}

func assertDataIsArray(t *testing.T, e envelope) {
	t.Helper()
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(e.Data)), "["),
		"envelope.data must be a JSON array, got: %s", string(e.Data))
}

func assertDataIsObject(t *testing.T, e envelope) {
	t.Helper()
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(e.Data)), "{"),
		"envelope.data must be a JSON object, got: %s", string(e.Data))
}

func TestLessonHandler_List_Envelope(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons", "/schedule/lessons", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons", h.List)
	})
	require.Equal(t, http.StatusOK, w.Code)
	assertDataIsArray(t, decodeEnvelope(t, w.Body.String()))
}

func TestLessonHandler_List_PaginationMeta(t *testing.T) {
	now := time.Now()
	lesson := entities.NewLesson(1, 1, 1, 1, 1, 1, domain.Monday, "09:00", "10:30", domain.WeekTypeAll, now, now, now)
	h := buildLessonHandler(t, &lessonRepoFake{lessons: []*entities.Lesson{lesson}}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons", "/schedule/lessons", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons", h.List)
	})
	require.Equal(t, http.StatusOK, w.Code)

	e := decodeEnvelope(t, w.Body.String())
	require.NotNil(t, e.Meta.Pagination, "list response must carry pagination meta")
	// default limit is 100, offset 0, one lesson total.
	assert.Equal(t, 1, e.Meta.Pagination.Page, "first page")
	assert.Equal(t, 100, e.Meta.Pagination.PerPage)
	assert.Equal(t, 1, e.Meta.Pagination.Total)
	assert.Equal(t, 1, e.Meta.Pagination.TotalPages)
}

func TestLessonHandler_GetTimetable_Envelope(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons/timetable", "/schedule/lessons/timetable", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons/timetable", h.GetTimetable)
	})
	require.Equal(t, http.StatusOK, w.Code)
	assertDataIsArray(t, decodeEnvelope(t, w.Body.String()))
}

func TestLessonHandler_ListChanges_Envelope(t *testing.T) {
	h := buildLessonHandler(t, &lessonRepoFake{}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/changes?lesson_id=42", "/schedule/changes", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/changes", h.ListChanges)
	})
	require.Equal(t, http.StatusOK, w.Code)
	assertDataIsArray(t, decodeEnvelope(t, w.Body.String()))
}

func TestLessonHandler_GetByID_Envelope(t *testing.T) {
	now := time.Now()
	lesson := entities.NewLesson(1, 1, 1, 1, 1, 1, domain.Monday, "09:00", "10:30", domain.WeekTypeAll, now, now, now)
	h := buildLessonHandler(t, &lessonRepoFake{byID: lesson}, &changeRepoFake{}, &classroomRepoFake{}, &referenceRepoFake{})
	w := performGet(t, h, "/schedule/lessons/7", "/schedule/lessons/:id", func(r *gin.Engine, h *LessonHandler) {
		r.GET("/schedule/lessons/:id", h.GetByID)
	})
	require.Equal(t, http.StatusOK, w.Code)
	assertDataIsObject(t, decodeEnvelope(t, w.Body.String()))
}
