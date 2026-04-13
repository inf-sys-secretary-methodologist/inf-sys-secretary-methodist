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

func TestDashboardHandler_Export_InvalidJSON(t *testing.T) {
	handler := NewDashboardHandler(nil)
	r := gin.New()
	r.POST("/export", handler.Export)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/export", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestDashboardHandler_Export_InvalidFormat(t *testing.T) {
	handler := NewDashboardHandler(nil)
	r := gin.New()
	r.POST("/export", handler.Export)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/export", strings.NewReader(`{"format":"csv"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "format must be")
}

func TestDashboardHandler_Export_ValidPDF(t *testing.T) {
	handler := NewDashboardHandler(nil)
	r := gin.New()
	r.POST("/export", handler.Export)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/export", strings.NewReader(`{"format":"pdf"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "dashboard-export.pdf")
}

func TestDashboardHandler_Export_ValidXLSX(t *testing.T) {
	handler := NewDashboardHandler(nil)
	r := gin.New()
	r.POST("/export", handler.Export)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/export", strings.NewReader(`{"format":"xlsx"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "dashboard-export.xlsx")
}
