package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupEventRouter(handler *EventHandler) *gin.Engine {
	r := gin.New()
	r.POST("/events", handler.Create)
	r.PUT("/events/:id", handler.Update)
	r.DELETE("/events/:id", handler.Delete)
	r.GET("/events/:id", handler.GetByID)
	r.GET("/events", handler.List)
	r.GET("/events/range", handler.GetByDateRange)
	r.GET("/events/upcoming", handler.GetUpcoming)
	r.POST("/events/:id/cancel", handler.Cancel)
	r.POST("/events/:id/reschedule", handler.Reschedule)
	r.POST("/events/:id/participants", handler.AddParticipants)
	r.DELETE("/events/:id/participants/:user_id", handler.RemoveParticipant)
	r.POST("/events/:id/respond", handler.UpdateParticipantStatus)
	r.GET("/events/invitations", handler.GetPendingInvitations)
	return r
}

func TestEventHandler_Create_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_Create_InvalidJSON(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.POST("/events", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Create(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_Update_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/events/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_Update_InvalidID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.PUT("/events/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Update(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/events/abc", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_Update_InvalidJSON(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.PUT("/events/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Update(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/events/1", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_Delete_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/events/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_Delete_InvalidID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.DELETE("/events/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Delete(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/events/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_GetByDateRange_MissingParams(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/range", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_GetByDateRange_InvalidStartDate(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/range?start=bad&end=2024-01-01T00:00:00Z", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_GetByDateRange_InvalidEndDate(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/range?start=2024-01-01T00:00:00Z&end=bad", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_GetUpcoming_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/upcoming", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_Cancel_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/1/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_Cancel_InvalidID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.POST("/events/:id/cancel", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Cancel(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/abc/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_Reschedule_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/1/reschedule", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_Reschedule_InvalidID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.POST("/events/:id/reschedule", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Reschedule(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/abc/reschedule", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_Reschedule_InvalidJSON(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.POST("/events/:id/reschedule", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Reschedule(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/1/reschedule", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_AddParticipants_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/1/participants", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_RemoveParticipant_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/events/1/participants/2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_RemoveParticipant_InvalidEventID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.DELETE("/events/:id/participants/:user_id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.RemoveParticipant(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/events/abc/participants/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_RemoveParticipant_InvalidParticipantID(t *testing.T) {
	handler := NewEventHandler(nil)
	r := gin.New()
	r.DELETE("/events/:id/participants/:user_id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.RemoveParticipant(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/events/1/participants/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventHandler_UpdateParticipantStatus_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/events/1/respond", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEventHandler_GetPendingInvitations_Unauthorized(t *testing.T) {
	handler := NewEventHandler(nil)
	r := setupEventRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events/invitations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
