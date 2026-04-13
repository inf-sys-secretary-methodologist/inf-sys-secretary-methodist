package http

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

func TestEmailHandler_SendEmail_InvalidJSON(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendEmail_MissingTo(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send",
		strings.NewReader(`{"subject":"test","body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendWelcomeEmail_InvalidJSON(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/welcome", handler.SendWelcomeEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/welcome", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendWelcomeEmail_MissingFields(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/welcome", handler.SendWelcomeEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/welcome",
		strings.NewReader(`{"email":"not-valid"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
