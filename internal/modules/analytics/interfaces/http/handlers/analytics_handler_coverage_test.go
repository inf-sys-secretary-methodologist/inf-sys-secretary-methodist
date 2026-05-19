package handlers

// v0.153.9 Phase 6 backfill — closes 4 funcs 0% in analytics_handler.go:
// GetStudentRiskHistory, GetRiskWeightConfig, UpdateRiskWeightConfig,
// ExportAtRiskStudents. Mirrors existing mock/router pattern from
// handlers_test.go. No production change.

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// ===== GetStudentRiskHistory =====

func TestAnalyticsHandler_GetStudentRiskHistory_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	// system_admin role → buildScope returns nil scope → no GetStudentRisk
	// scope-check is bypassed; only GetStudentRiskHistory hit.
	ar.On("GetStudentRiskHistory", mock.Anything, int64(7), 90).
		Return([]entities.RiskHistoryEntry{{StudentID: 7, RiskScore: 75.0, RiskLevel: "high"}}, nil)

	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk/history", h.GetStudentRiskHistory)
	w := performRequest(router, http.MethodGet, "/students/7/risk/history", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetStudentRiskHistory_InvalidID(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk/history", h.GetStudentRiskHistory)
	w := performRequest(router, http.MethodGet, "/students/abc/risk/history", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_GetStudentRiskHistory_LimitParam(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentRiskHistory", mock.Anything, int64(7), 30).
		Return([]entities.RiskHistoryEntry{}, nil)
	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk/history", h.GetStudentRiskHistory)
	w := performRequest(router, http.MethodGet, "/students/7/risk/history?limit=30", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetStudentRiskHistory_UsecaseError(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentRiskHistory", mock.Anything, int64(7), 90).
		Return(nil, fmt.Errorf("db error"))
	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk/history", h.GetStudentRiskHistory)
	w := performRequest(router, http.MethodGet, "/students/7/risk/history", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== GetRiskWeightConfig =====

func TestAnalyticsHandler_GetRiskWeightConfig_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetRiskWeightConfig", mock.Anything).
		Return(&entities.RiskWeightConfig{
			AttendanceWeight:      0.4,
			GradeWeight:           0.3,
			SubmissionWeight:      0.2,
			InactivityWeight:      0.1,
			HighRiskThreshold:     70,
			CriticalRiskThreshold: 85,
		}, nil)
	router := setupAnalyticsRouter()
	router.GET("/risk-config", h.GetRiskWeightConfig)
	w := performRequest(router, http.MethodGet, "/risk-config", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetRiskWeightConfig_Error(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetRiskWeightConfig", mock.Anything).
		Return(nil, fmt.Errorf("db error"))
	router := setupAnalyticsRouter()
	router.GET("/risk-config", h.GetRiskWeightConfig)
	w := performRequest(router, http.MethodGet, "/risk-config", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== UpdateRiskWeightConfig =====

func TestAnalyticsHandler_UpdateRiskWeightConfig_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("UpdateRiskWeightConfig", mock.Anything, mock.Anything).Return(nil)

	router := setupAnalyticsRouter()
	router.PUT("/risk-config", h.UpdateRiskWeightConfig)
	body := dto.UpdateRiskWeightConfigRequest{
		AttendanceWeight:      0.4,
		GradeWeight:           0.3,
		SubmissionWeight:      0.2,
		InactivityWeight:      0.1,
		HighRiskThreshold:     70,
		CriticalRiskThreshold: 85,
	}
	w := performRequest(router, http.MethodPut, "/risk-config", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_UpdateRiskWeightConfig_Unauthorized(t *testing.T) {
	// No user_id in ctx — bypasses setupAnalyticsRouter; bare router.
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupRouter()
	router.PUT("/risk-config", h.UpdateRiskWeightConfig)
	body := dto.UpdateRiskWeightConfigRequest{AttendanceWeight: 1.0}
	w := performRequest(router, http.MethodPut, "/risk-config", body)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnalyticsHandler_UpdateRiskWeightConfig_InvalidBody(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.PUT("/risk-config", h.UpdateRiskWeightConfig)
	// not-JSON body
	w := performRequest(router, http.MethodPut, "/risk-config", "not-json")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_UpdateRiskWeightConfig_WeightsSumError(t *testing.T) {
	// Usecase returns error when weights don't sum to 1.0 → handler 500.
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.PUT("/risk-config", h.UpdateRiskWeightConfig)
	body := dto.UpdateRiskWeightConfigRequest{
		AttendanceWeight:  0.5,
		GradeWeight:       0.5,
		SubmissionWeight:  0.5,
		InactivityWeight:  0.5, // sum = 2.0 ≠ 1.0
		HighRiskThreshold: 70, CriticalRiskThreshold: 85,
	}
	w := performRequest(router, http.MethodPut, "/risk-config", body)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== ExportAtRiskStudents =====

func TestAnalyticsHandler_ExportAtRiskStudents_CSV(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	groupName := "ИС-21"
	attendance := 75.0
	gradeAvg := 4.2
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, 20, 0).
		Return([]entities.StudentRiskScore{
			{StudentID: 1, StudentName: "Иван Петров", GroupName: &groupName, RiskScore: 78, RiskLevel: "high", AttendanceRate: &attendance, GradeAverage: &gradeAvg},
			{StudentID: 2, StudentName: "Анна Смирнова", RiskScore: 65, RiskLevel: "medium"},
		}, int64(2), nil)

	router := setupAnalyticsRouter()
	router.GET("/export", h.ExportAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/export?format=csv", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "csv")
}

func TestAnalyticsHandler_ExportAtRiskStudents_XLSX(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	groupName := "ИС-21"
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, 20, 0).
		Return([]entities.StudentRiskScore{
			{StudentID: 1, StudentName: "Иван", GroupName: &groupName, RiskScore: 78, RiskLevel: "high"},
		}, int64(1), nil)

	router := setupAnalyticsRouter()
	router.GET("/export", h.ExportAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/export?format=xlsx", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_ExportAtRiskStudents_DefaultFormat(t *testing.T) {
	// No format query → defaults to csv branch.
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, 20, 0).
		Return([]entities.StudentRiskScore{}, int64(0), nil)
	router := setupAnalyticsRouter()
	router.GET("/export", h.ExportAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/export", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_ExportAtRiskStudents_UsecaseError(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, 20, 0).
		Return([]entities.StudentRiskScore{}, int64(0), fmt.Errorf("db error"))
	router := setupAnalyticsRouter()
	router.GET("/export", h.ExportAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/export?format=csv", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
