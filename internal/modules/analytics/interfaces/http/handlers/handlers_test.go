package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

func init() { gin.SetMode(gin.TestMode) }

func setupRouter() *gin.Engine { return gin.New() }

// setupAnalyticsRouter returns a router pre-loaded with system_admin auth
// context, mirroring what RequireNonStudent middleware would set in
// production. Tests that exercise scope assembly (teacher/methodist/...)
// build their own router with stubAuth instead.
func setupAnalyticsRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("role", "system_admin")
		c.Next()
	})
	return r
}

func withAuth(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) { c.Set("user_id", userID); c.Next() }
}

func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ===== Mock Repositories =====

type mockAnalyticsRepo struct{ mock.Mock }

func (m *mockAnalyticsRepo) GetAtRiskStudents(ctx context.Context, scope *entities.TeacherScope, limit, offset int) ([]entities.StudentRiskScore, int64, error) {
	args := m.Called(ctx, scope, limit, offset)
	return args.Get(0).([]entities.StudentRiskScore), args.Get(1).(int64), args.Error(2)
}
func (m *mockAnalyticsRepo) GetStudentRisk(ctx context.Context, studentID int64) (*entities.StudentRiskScore, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StudentRiskScore), args.Error(1)
}
func (m *mockAnalyticsRepo) GetGroupSummary(ctx context.Context, groupName string) (*entities.GroupAnalyticsSummary, error) {
	args := m.Called(ctx, groupName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.GroupAnalyticsSummary), args.Error(1)
}
func (m *mockAnalyticsRepo) GetAllGroupsSummary(ctx context.Context, scope *entities.TeacherScope) ([]entities.GroupAnalyticsSummary, error) {
	args := m.Called(ctx, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.GroupAnalyticsSummary), args.Error(1)
}
func (m *mockAnalyticsRepo) GetStudentsByRiskLevel(ctx context.Context, scope *entities.TeacherScope, riskLevel entities.RiskLevel, limit, offset int) ([]entities.StudentRiskScore, int64, error) {
	args := m.Called(ctx, scope, riskLevel, limit, offset)
	return args.Get(0).([]entities.StudentRiskScore), args.Get(1).(int64), args.Error(2)
}
func (m *mockAnalyticsRepo) GetMonthlyAttendanceTrend(ctx context.Context, months int) ([]entities.MonthlyAttendanceTrend, error) {
	args := m.Called(ctx, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.MonthlyAttendanceTrend), args.Error(1)
}
func (m *mockAnalyticsRepo) GetRiskWeightConfig(ctx context.Context) (*entities.RiskWeightConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RiskWeightConfig), args.Error(1)
}
func (m *mockAnalyticsRepo) UpdateRiskWeightConfig(ctx context.Context, cfg *entities.RiskWeightConfig) error {
	return m.Called(ctx, cfg).Error(0)
}
func (m *mockAnalyticsRepo) SaveRiskHistory(ctx context.Context, entry *entities.RiskHistoryEntry) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *mockAnalyticsRepo) GetStudentRiskHistory(ctx context.Context, studentID int64, limit int) ([]entities.RiskHistoryEntry, error) {
	args := m.Called(ctx, studentID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.RiskHistoryEntry), args.Error(1)
}

type mockAttendanceRepo struct{ mock.Mock }

func (m *mockAttendanceRepo) CreateLesson(ctx context.Context, lesson *entities.Lesson) error {
	return m.Called(ctx, lesson).Error(0)
}
func (m *mockAttendanceRepo) GetLessonByID(ctx context.Context, id int64) (*entities.Lesson, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Lesson), args.Error(1)
}
func (m *mockAttendanceRepo) GetLessonsByGroup(ctx context.Context, groupName string) ([]entities.Lesson, error) {
	args := m.Called(ctx, groupName)
	return args.Get(0).([]entities.Lesson), args.Error(1)
}
func (m *mockAttendanceRepo) GetLessonsByTeacher(ctx context.Context, teacherID int64) ([]entities.Lesson, error) {
	args := m.Called(ctx, teacherID)
	return args.Get(0).([]entities.Lesson), args.Error(1)
}
func (m *mockAttendanceRepo) MarkAttendance(ctx context.Context, record *entities.AttendanceRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *mockAttendanceRepo) BulkMarkAttendance(ctx context.Context, records []entities.AttendanceRecord) error {
	return m.Called(ctx, records).Error(0)
}
func (m *mockAttendanceRepo) GetAttendanceByLesson(ctx context.Context, lessonID int64, date string) ([]entities.AttendanceRecord, error) {
	args := m.Called(ctx, lessonID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.AttendanceRecord), args.Error(1)
}
func (m *mockAttendanceRepo) GetAttendanceByStudent(ctx context.Context, studentID int64, fromDate, toDate string) ([]entities.AttendanceRecord, error) {
	args := m.Called(ctx, studentID, fromDate, toDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.AttendanceRecord), args.Error(1)
}
func (m *mockAttendanceRepo) UpdateAttendance(ctx context.Context, record *entities.AttendanceRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *mockAttendanceRepo) GetStudentAttendanceStats(ctx context.Context, studentID int64) (*entities.AttendanceStats, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AttendanceStats), args.Error(1)
}

type mockGradeRepo struct{ mock.Mock }

func (m *mockGradeRepo) CreateGrade(ctx context.Context, grade *entities.Grade) error {
	return m.Called(ctx, grade).Error(0)
}
func (m *mockGradeRepo) GetGradesByStudent(ctx context.Context, studentID int64) ([]entities.Grade, error) {
	args := m.Called(ctx, studentID)
	return args.Get(0).([]entities.Grade), args.Error(1)
}
func (m *mockGradeRepo) GetGradesBySubject(ctx context.Context, studentID int64, subject string) ([]entities.Grade, error) {
	args := m.Called(ctx, studentID, subject)
	return args.Get(0).([]entities.Grade), args.Error(1)
}
func (m *mockGradeRepo) UpdateGrade(ctx context.Context, grade *entities.Grade) error {
	return m.Called(ctx, grade).Error(0)
}
func (m *mockGradeRepo) DeleteGrade(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockGradeRepo) GetStudentGradeStats(ctx context.Context, studentID int64) (*entities.GradeStats, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.GradeStats), args.Error(1)
}

func newUC(ar *mockAnalyticsRepo, atr *mockAttendanceRepo, gr *mockGradeRepo) *usecases.AnalyticsUseCase {
	return usecases.NewAnalyticsUseCase(ar, atr, gr, nil)
}

// noopScopeRepo is a TeacherScopeRepository that always returns no
// groups. It exists so handler tests can satisfy the non-nil scopeRepo
// invariant without exercising scope assembly. Tests that DO exercise
// scope (scope_handler_test.go) use mockTeacherScopeRepo instead.
type noopScopeRepo struct{}

func (noopScopeRepo) ListGroupNames(_ context.Context, _ int64) ([]string, error) {
	return nil, nil
}

// ===== AnalyticsHandler Tests =====

func TestNewAnalyticsHandler(t *testing.T)  { assert.NotNil(t, NewAnalyticsHandler(nil, noopScopeRepo{})) }
func TestNewAttendanceHandler(t *testing.T) { assert.NotNil(t, NewAttendanceHandler(nil)) }

func TestAnalyticsHandler_GetAtRiskStudents_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, 20, 0).Return([]entities.StudentRiskScore{}, int64(0), nil)
	router := setupAnalyticsRouter()
	router.GET("/at-risk", h.GetAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/at-risk", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetAtRiskStudents_WithParams(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, 10, 10).Return([]entities.StudentRiskScore{}, int64(0), nil)
	router := setupAnalyticsRouter()
	router.GET("/at-risk", h.GetAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/at-risk?page=2&page_size=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetAtRiskStudents_Error(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAtRiskStudents", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentRiskScore{}, int64(0), fmt.Errorf("err"))
	router := setupAnalyticsRouter()
	router.GET("/at-risk", h.GetAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/at-risk", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetStudentRisk_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentRisk", mock.Anything, int64(1)).Return(&entities.StudentRiskScore{StudentID: 1}, nil)
	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk", h.GetStudentRisk)
	w := performRequest(router, http.MethodGet, "/students/1/risk", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetStudentRisk_InvalidID(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk", h.GetStudentRisk)
	w := performRequest(router, http.MethodGet, "/students/abc/risk", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_GetStudentRisk_NotFound(t *testing.T) {
	// Usecase wraps errors, so exact match won't work at handler level
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentRisk", mock.Anything, int64(1)).Return(nil, fmt.Errorf("student not found"))
	router := setupAnalyticsRouter()
	router.GET("/students/:id/risk", h.GetStudentRisk)
	w := performRequest(router, http.MethodGet, "/students/1/risk", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetGroupSummary_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetGroupSummary", mock.Anything, "G1").Return(&entities.GroupAnalyticsSummary{GroupName: "G1"}, nil)
	router := setupAnalyticsRouter()
	router.GET("/groups/:name/summary", h.GetGroupSummary)
	w := performRequest(router, http.MethodGet, "/groups/G1/summary", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetGroupSummary_NotFound(t *testing.T) {
	// The usecase wraps errors, so "group not found" becomes "failed to get group summary: group not found"
	// The handler only matches exact "group not found", so this becomes a 500
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetGroupSummary", mock.Anything, "X").Return(nil, fmt.Errorf("group not found"))
	router := setupAnalyticsRouter()
	router.GET("/groups/:name/summary", h.GetGroupSummary)
	w := performRequest(router, http.MethodGet, "/groups/X/summary", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetGroupSummary_Error(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetGroupSummary", mock.Anything, "X").Return(nil, fmt.Errorf("db error"))
	router := setupAnalyticsRouter()
	router.GET("/groups/:name/summary", h.GetGroupSummary)
	w := performRequest(router, http.MethodGet, "/groups/X/summary", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetAllGroupsSummary_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAllGroupsSummary", mock.Anything, mock.Anything).Return([]entities.GroupAnalyticsSummary{}, nil)
	router := setupAnalyticsRouter()
	router.GET("/groups/summary", h.GetAllGroupsSummary)
	w := performRequest(router, http.MethodGet, "/groups/summary", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetAllGroupsSummary_Error(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetAllGroupsSummary", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))
	router := setupAnalyticsRouter()
	router.GET("/groups/summary", h.GetAllGroupsSummary)
	w := performRequest(router, http.MethodGet, "/groups/summary", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetStudentsByRiskLevel_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentsByRiskLevel", mock.Anything, mock.Anything, mock.Anything, 20, 0).Return([]entities.StudentRiskScore{}, int64(0), nil)
	router := setupAnalyticsRouter()
	router.GET("/risk-level/:level", h.GetStudentsByRiskLevel)
	w := performRequest(router, http.MethodGet, "/risk-level/high", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetStudentsByRiskLevel_InvalidLevel(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.GET("/risk-level/:level", h.GetStudentsByRiskLevel)
	w := performRequest(router, http.MethodGet, "/risk-level/invalid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_GetStudentsByRiskLevel_WithParams(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentsByRiskLevel", mock.Anything, mock.Anything, mock.Anything, 10, 10).Return([]entities.StudentRiskScore{}, int64(0), nil)
	router := setupAnalyticsRouter()
	router.GET("/risk-level/:level", h.GetStudentsByRiskLevel)
	w := performRequest(router, http.MethodGet, "/risk-level/low?page=2&page_size=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetStudentsByRiskLevel_Error(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetStudentsByRiskLevel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentRiskScore{}, int64(0), fmt.Errorf("err"))
	router := setupAnalyticsRouter()
	router.GET("/risk-level/:level", h.GetStudentsByRiskLevel)
	w := performRequest(router, http.MethodGet, "/risk-level/high", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetAttendanceTrend_Success(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetMonthlyAttendanceTrend", mock.Anything, 6).Return([]entities.MonthlyAttendanceTrend{}, nil)
	router := setupAnalyticsRouter()
	router.GET("/attendance-trend", h.GetAttendanceTrend)
	w := performRequest(router, http.MethodGet, "/attendance-trend", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetAttendanceTrend_WithMonths(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetMonthlyAttendanceTrend", mock.Anything, 12).Return([]entities.MonthlyAttendanceTrend{}, nil)
	router := setupAnalyticsRouter()
	router.GET("/attendance-trend", h.GetAttendanceTrend)
	w := performRequest(router, http.MethodGet, "/attendance-trend?months=12", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetAttendanceTrend_Error(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), noopScopeRepo{})
	ar.On("GetMonthlyAttendanceTrend", mock.Anything, 6).Return(nil, fmt.Errorf("err"))
	router := setupAnalyticsRouter()
	router.GET("/attendance-trend", h.GetAttendanceTrend)
	w := performRequest(router, http.MethodGet, "/attendance-trend", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== AttendanceHandler Tests =====

func TestAttendanceHandler_MarkAttendance_NoAuth(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/mark", h.MarkAttendance)
	w := performRequest(router, http.MethodPost, "/mark", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAttendanceHandler_MarkAttendance_InvalidJSON(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/mark", withAuth(1), h.MarkAttendance)
	req := httptest.NewRequest(http.MethodPost, "/mark", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_BulkMarkAttendance_NoAuth(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/bulk", h.BulkMarkAttendance)
	w := performRequest(router, http.MethodPost, "/bulk", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAttendanceHandler_BulkMarkAttendance_InvalidJSON(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/bulk", withAuth(1), h.BulkMarkAttendance)
	req := httptest.NewRequest(http.MethodPost, "/bulk", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_GetLessonAttendance_InvalidID(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.GET("/lesson/:id/date/:date", h.GetLessonAttendance)
	w := performRequest(router, http.MethodGet, "/lesson/abc/date/2024-01-01", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_CreateLesson_NoAuth(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/lessons", h.CreateLesson)
	w := performRequest(router, http.MethodPost, "/lessons", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAttendanceHandler_CreateLesson_InvalidJSON(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/lessons", withAuth(1), h.CreateLesson)
	req := httptest.NewRequest(http.MethodPost, "/lessons", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_handleError_LessonNotFound(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("lesson not found")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAttendanceHandler_handleError_StudentNotFound(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("student not found")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAttendanceHandler_handleError_InvalidDateFormat(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("invalid lesson date format")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_handleError_Generic(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("other")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAttendanceHandler_getUserID_InvalidType(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/mark", func(c *gin.Context) { c.Set("user_id", "bad"); c.Next() }, h.MarkAttendance)
	w := performRequest(router, http.MethodPost, "/mark", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAttendanceHandler_MarkAttendance_ValidationError(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/mark", withAuth(1), h.MarkAttendance)
	// Empty required fields
	w := performRequest(router, http.MethodPost, "/mark", map[string]interface{}{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_MarkAttendance_Success(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	g1 := "G1"
	atr.On("GetLessonByID", mock.Anything, int64(1)).Return(&entities.Lesson{ID: 1, GroupName: &g1}, nil)
	atr.On("MarkAttendance", mock.Anything, mock.Anything).Return(nil)

	router := setupRouter()
	router.POST("/mark", withAuth(1), h.MarkAttendance)
	w := performRequest(router, http.MethodPost, "/mark", map[string]interface{}{
		"lesson_id":   1,
		"student_id":  1,
		"lesson_date": "2024-01-15",
		"status":      "present",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAttendanceHandler_MarkAttendance_Error(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	atr.On("MarkAttendance", mock.Anything, mock.Anything).Return(fmt.Errorf("db error"))

	router := setupRouter()
	router.POST("/mark", withAuth(1), h.MarkAttendance)
	w := performRequest(router, http.MethodPost, "/mark", map[string]interface{}{
		"lesson_id":   1,
		"student_id":  1,
		"lesson_date": "2024-01-15",
		"status":      "present",
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAttendanceHandler_BulkMarkAttendance_ValidationError(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/bulk", withAuth(1), h.BulkMarkAttendance)
	w := performRequest(router, http.MethodPost, "/bulk", map[string]interface{}{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_BulkMarkAttendance_Success(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	atr.On("BulkMarkAttendance", mock.Anything, mock.Anything).Return(nil)

	router := setupRouter()
	router.POST("/bulk", withAuth(1), h.BulkMarkAttendance)
	w := performRequest(router, http.MethodPost, "/bulk", map[string]interface{}{
		"lesson_id":   1,
		"lesson_date": "2024-01-15",
		"records": []map[string]interface{}{
			{"student_id": 1, "status": "present"},
		},
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAttendanceHandler_BulkMarkAttendance_Error(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	atr.On("BulkMarkAttendance", mock.Anything, mock.Anything).Return(fmt.Errorf("db error"))

	router := setupRouter()
	router.POST("/bulk", withAuth(1), h.BulkMarkAttendance)
	w := performRequest(router, http.MethodPost, "/bulk", map[string]interface{}{
		"lesson_id":   1,
		"lesson_date": "2024-01-15",
		"records": []map[string]interface{}{
			{"student_id": 1, "status": "present"},
		},
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAttendanceHandler_GetLessonAttendance_Success(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	g1c := "G1"
	atr.On("GetLessonByID", mock.Anything, int64(1)).Return(&entities.Lesson{ID: 1, GroupName: &g1c}, nil)
	atr.On("GetAttendanceByLesson", mock.Anything, int64(1), "2024-01-15").Return([]entities.AttendanceRecord{}, nil)

	router := setupRouter()
	router.GET("/lesson/:id/date/:date", h.GetLessonAttendance)
	w := performRequest(router, http.MethodGet, "/lesson/1/date/2024-01-15", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAttendanceHandler_GetLessonAttendance_Error(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	atr.On("GetAttendanceByLesson", mock.Anything, int64(1), "2024-01-15").Return(nil, fmt.Errorf("db error"))

	router := setupRouter()
	router.GET("/lesson/:id/date/:date", h.GetLessonAttendance)
	w := performRequest(router, http.MethodGet, "/lesson/1/date/2024-01-15", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAttendanceHandler_CreateLesson_ValidationError(t *testing.T) {
	h := NewAttendanceHandler(nil)
	router := setupRouter()
	router.POST("/lessons", withAuth(1), h.CreateLesson)
	w := performRequest(router, http.MethodPost, "/lessons", map[string]interface{}{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttendanceHandler_CreateLesson_Success(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	atr.On("CreateLesson", mock.Anything, mock.Anything).Return(nil)

	router := setupRouter()
	router.POST("/lessons", withAuth(1), h.CreateLesson)
	w := performRequest(router, http.MethodPost, "/lessons", map[string]interface{}{
		"name":    "Math Lesson",
		"subject": "Math",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAttendanceHandler_CreateLesson_Error(t *testing.T) {
	atr := new(mockAttendanceRepo)
	uc := newUC(new(mockAnalyticsRepo), atr, new(mockGradeRepo))
	h := NewAttendanceHandler(uc)

	atr.On("CreateLesson", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

	router := setupRouter()
	router.POST("/lessons", withAuth(1), h.CreateLesson)
	w := performRequest(router, http.MethodPost, "/lessons", map[string]interface{}{
		"name":    "Math Lesson",
		"subject": "Math",
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_handleError_StudentNotFound(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("student not found")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAnalyticsHandler_handleError_GroupNotFound(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("group not found")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAnalyticsHandler_handleError_Generic(t *testing.T) {
	h := NewAnalyticsHandler(nil, noopScopeRepo{})
	router := setupAnalyticsRouter()
	router.GET("/t", func(c *gin.Context) { h.handleError(c, fmt.Errorf("other")) })
	w := performRequest(router, http.MethodGet, "/t", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
