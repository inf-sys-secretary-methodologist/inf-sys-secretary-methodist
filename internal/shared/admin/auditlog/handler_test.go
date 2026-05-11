package auditlog_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/auditlog"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

func init() { gin.SetMode(gin.TestMode) }

// stubAuthMiddleware mirrors the production JWTMiddleware contract by
// writing the same context keys (`user_id`, `role`) the handler reads
// through authContext-style lookups. Pinned per memory
// `feedback_handler_context_key_must_match_middleware` — never
// substitute with ad-hoc keys.
func stubAuthMiddleware(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// requireSystemAdmin mirrors production
// authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin)) so the
// integration test exercises the same gate the production router
// wires under adminGroup.
func requireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "system_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// recordingReader captures the AuditLogFilter so handler tests can
// assert end-to-end that query parsing → use case → reader plumbing
// preserves filter shape without parse-then-stringify drift.
type recordingReader struct {
	captured logging.AuditLogFilter
	items    []*logging.AuditLog
	total    int
	err      error
}

func (r *recordingReader) List(_ context.Context, filter logging.AuditLogFilter) (logging.AuditLogListResult, error) {
	r.captured = filter
	if r.err != nil {
		return logging.AuditLogListResult{}, r.err
	}
	return logging.AuditLogListResult{Items: r.items, Total: r.total}, nil
}

// newTestRouter builds a real gin.Engine wired with the same
// middleware chain (`adminGroup := admin / RequireRole(system_admin)`)
// the production main.go uses. Tests exercise the route through this
// router so middleware-handler integration is pinned.
func newTestRouter(t *testing.T, reader logging.AuditLogReader, userID int64, role string) http.Handler {
	t.Helper()
	r := gin.New()
	uc := auditlog.NewAdminAuditLogUseCase(reader)
	h := auditlog.NewAdminAuditLogHandler(uc)

	adminGroup := r.Group("/api/admin")
	adminGroup.Use(stubAuthMiddleware(userID, role))
	adminGroup.Use(requireSystemAdmin())
	adminGroup.GET("/audit-logs", h.List)
	return r
}

func doGET(t *testing.T, router http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAdminAuditLogHandler_List_RoleGate(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		wantCode int
	}{
		{"system_admin allowed", "system_admin", http.StatusOK},
		{"methodist forbidden", "methodist", http.StatusForbidden},
		{"academic_secretary forbidden", "academic_secretary", http.StatusForbidden},
		{"teacher forbidden", "teacher", http.StatusForbidden},
		{"student forbidden", "student", http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := &recordingReader{}
			router := newTestRouter(t, reader, 1, tc.role)
			rec := doGET(t, router, "/api/admin/audit-logs")
			require.Equal(t, tc.wantCode, rec.Code)
		})
	}
}

func TestAdminAuditLogHandler_List_QueryParsing(t *testing.T) {
	createdAt := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	fromStr := "2026-05-01T00:00:00Z"
	toStr := "2026-05-31T00:00:00Z"
	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)

	reader := &recordingReader{
		items: []*logging.AuditLog{{ID: 1, CreatedAt: createdAt, Action: "a", Resource: "r"}},
		total: 1,
	}
	router := newTestRouter(t, reader, 999, "system_admin")

	target := "/api/admin/audit-logs?" + strings.Join([]string{
		"action=curriculum.approved",
		"resource=curriculum",
		"user_id=42",
		"from=" + fromStr,
		"to=" + toStr,
		"limit=25",
		"offset=50",
	}, "&")

	rec := doGET(t, router, target)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, "curriculum.approved", reader.captured.Action)
	require.Equal(t, "curriculum", reader.captured.Resource)
	require.NotNil(t, reader.captured.UserID)
	require.Equal(t, int64(42), *reader.captured.UserID)
	require.NotNil(t, reader.captured.From)
	require.Equal(t, from, *reader.captured.From)
	require.NotNil(t, reader.captured.To)
	require.Equal(t, to, *reader.captured.To)
	require.Equal(t, 25, reader.captured.Limit)
	require.Equal(t, 50, reader.captured.Offset)
}

func TestAdminAuditLogHandler_List_QueryParsing_InvalidInputs(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"user_id not a number", "user_id=abc"},
		{"user_id negative", "user_id=-1"},
		{"from not RFC3339", "from=2026-05-01"},
		{"to not RFC3339", "to=yesterday"},
		{"limit not a number", "limit=ten"},
		{"limit negative", "limit=-5"},
		{"offset not a number", "offset=ten"},
		{"offset negative", "offset=-1"},
		{"from after to (sentinel range)", "from=2026-06-01T00:00:00Z&to=2026-05-01T00:00:00Z"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := &recordingReader{}
			router := newTestRouter(t, reader, 1, "system_admin")
			rec := doGET(t, router, "/api/admin/audit-logs?"+tc.query)
			require.Equal(t, http.StatusBadRequest, rec.Code, "body=%s", rec.Body.String())
		})
	}
}

func TestAdminAuditLogHandler_List_DefaultLimitWhenOmitted(t *testing.T) {
	reader := &recordingReader{}
	router := newTestRouter(t, reader, 1, "system_admin")
	rec := doGET(t, router, "/api/admin/audit-logs")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, auditlog.DefaultLimit, reader.captured.Limit)
}

func TestAdminAuditLogHandler_List_LimitClampedAtMax(t *testing.T) {
	reader := &recordingReader{}
	router := newTestRouter(t, reader, 1, "system_admin")
	rec := doGET(t, router, "/api/admin/audit-logs?limit=99999")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, auditlog.MaxLimit, reader.captured.Limit)
}

func TestAdminAuditLogHandler_List_ResponseShape(t *testing.T) {
	actorID := int64(42)
	actorIP := "10.0.0.5"
	corrID := "req-7c4f"
	createdAt := time.Date(2026, 5, 10, 12, 30, 0, 0, time.UTC)

	reader := &recordingReader{
		items: []*logging.AuditLog{
			{
				ID:            11,
				CreatedAt:     createdAt,
				Action:        "curriculum.approved",
				Resource:      "curriculum",
				ActorUserID:   &actorID,
				ActorIP:       &actorIP,
				CorrelationID: &corrID,
				Fields:        map[string]any{"curriculum_id": float64(7)},
			},
			{
				ID:        12,
				CreatedAt: createdAt.Add(time.Minute),
				Action:    "auth.logout",
				Resource:  "session",
				Fields:    map[string]any{},
			},
		},
		total: 42,
	}
	router := newTestRouter(t, reader, 1, "system_admin")
	rec := doGET(t, router, "/api/admin/audit-logs?limit=10&offset=20")
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())

	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			ID            int64          `json:"id"`
			CreatedAt     string         `json:"created_at"`
			Action        string         `json:"action"`
			Resource      string         `json:"resource"`
			ActorUserID   *int64         `json:"actor_user_id"`
			ActorIP       *string        `json:"actor_ip"`
			CorrelationID *string        `json:"correlation_id"`
			Fields        map[string]any `json:"fields"`
		} `json:"data"`
		Meta struct {
			Pagination *struct {
				Page       int `json:"page"`
				PerPage    int `json:"per_page"`
				Total      int `json:"total"`
				TotalPages int `json:"total_pages"`
			} `json:"pagination"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	require.Len(t, resp.Data, 2)

	require.Equal(t, int64(11), resp.Data[0].ID)
	require.Equal(t, "curriculum.approved", resp.Data[0].Action)
	require.NotNil(t, resp.Data[0].ActorUserID)
	require.Equal(t, int64(42), *resp.Data[0].ActorUserID)
	require.NotNil(t, resp.Data[0].ActorIP)
	require.Equal(t, "10.0.0.5", *resp.Data[0].ActorIP)
	require.NotNil(t, resp.Data[0].CorrelationID)
	require.Equal(t, "req-7c4f", *resp.Data[0].CorrelationID)
	require.Equal(t, float64(7), resp.Data[0].Fields["curriculum_id"])
	// RFC3339 round-trip
	parsed, err := time.Parse(time.RFC3339, resp.Data[0].CreatedAt)
	require.NoError(t, err)
	require.Equal(t, createdAt.UTC(), parsed.UTC())

	require.Nil(t, resp.Data[1].ActorUserID)
	require.Nil(t, resp.Data[1].ActorIP)
	require.Nil(t, resp.Data[1].CorrelationID)
	require.Empty(t, resp.Data[1].Fields)

	require.NotNil(t, resp.Meta.Pagination)
	require.Equal(t, 10, resp.Meta.Pagination.PerPage)
	require.Equal(t, 42, resp.Meta.Pagination.Total)
	// offset 20 / per_page 10 → page 3
	require.Equal(t, 3, resp.Meta.Pagination.Page)
	// 42 / 10 ceiling = 5
	require.Equal(t, 5, resp.Meta.Pagination.TotalPages)
}

func TestAdminAuditLogHandler_List_UseCaseError500(t *testing.T) {
	reader := &recordingReader{err: errInternal}
	router := newTestRouter(t, reader, 1, "system_admin")
	rec := doGET(t, router, "/api/admin/audit-logs")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAdminAuditLogHandler_List_LargeOffsetParses(t *testing.T) {
	reader := &recordingReader{}
	router := newTestRouter(t, reader, 1, "system_admin")
	bigOffset := strconv.Itoa(123456)
	rec := doGET(t, router, "/api/admin/audit-logs?offset="+bigOffset)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 123456, reader.captured.Offset)
}

// errInternal is a deterministic error returned from the reader to
// exercise the handler's 500 mapping.
var errInternal = errReaderInternal()

func errReaderInternal() error {
	return errInternalSentinel
}

var errInternalSentinel = errors.New("reader internal")
