package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/interfaces/http/handlers"
)

// --- fakes ------------------------------------------------------------------

type fakeGet struct {
	debt     *entities.StudentDebt
	err      error
	gotID    int64
	gotActor int64
	gotRole  string
}

func (f *fakeGet) Execute(_ context.Context, actorID int64, role string, id int64) (*entities.StudentDebt, error) {
	f.gotActor, f.gotRole, f.gotID = actorID, role, id
	return f.debt, f.err
}

type fakeList struct {
	res       repositories.StudentDebtListResult
	err       error
	gotFilter repositories.StudentDebtListFilter
	gotActor  int64
	gotRole   string
}

func (f *fakeList) Execute(_ context.Context, actorID int64, role string, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	f.gotActor, f.gotRole, f.gotFilter = actorID, role, filter
	return f.res, f.err
}

type fakeListMy struct {
	res       repositories.StudentDebtListResult
	err       error
	gotActor  int64
	gotFilter repositories.StudentDebtListFilter
}

func (f *fakeListMy) Execute(_ context.Context, actorID int64, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	f.gotActor, f.gotFilter = actorID, filter
	return f.res, f.err
}

type fakeStats struct {
	res      repositories.StudentDebtStats
	err      error
	gotActor int64
	gotRole  string
}

func (f *fakeStats) Execute(_ context.Context, actorID int64, role string, _ repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
	f.gotActor, f.gotRole = actorID, role
	return f.res, f.err
}

// --- harness ----------------------------------------------------------------

func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func newRouter(g *fakeGet, l *fakeList, m *fakeListMy, s *fakeStats, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	for _, h := range mw {
		api.Use(h)
	}
	handler := handlers.NewStudentDebtHandler(g, l, m, s)
	handlers.RegisterStudentDebtRoutes(api, handler)
	return r
}

func doGET(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sampleDebt(t *testing.T) *entities.StudentDebt {
	t.Helper()
	d, err := entities.NewStudentDebt("Иванов Иван", "ИВТ-21", "Базы данных", 3, entities.ControlFormExam)
	require.NoError(t, err)
	d.ID = 55
	require.NoError(t, d.ScheduleResit(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC), "Петров П.П.", time.Now()))
	return d
}

// --- List -------------------------------------------------------------------

func TestStudentDebtHandler_List_HappyPath(t *testing.T) {
	l := &fakeList{res: repositories.StudentDebtListResult{
		Items: []repositories.StudentDebtListItem{{ID: 1, StudentFullName: "Иванов Иван", Status: entities.DebtStatusOpen}},
		Total: 1,
	}}
	r := newRouter(&fakeGet{}, l, &fakeListMy{}, &fakeStats{}, withAuth(7, "methodist"))

	w := doGET(t, r, "/api/v1/student-debts?group_name=%D0%98%D0%92%D0%A2-21&status=open&semester=3&limit=20&offset=0")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(7), l.gotActor)
	assert.Equal(t, "methodist", l.gotRole)
	assert.Equal(t, "ИВТ-21", l.gotFilter.GroupName)
	require.NotNil(t, l.gotFilter.Status)
	assert.Equal(t, entities.DebtStatusOpen, *l.gotFilter.Status)
	require.NotNil(t, l.gotFilter.Semester)
	assert.Equal(t, 3, *l.gotFilter.Semester)
	assert.Equal(t, 20, l.gotFilter.Limit)
}

func TestStudentDebtHandler_List_PaginationClamp(t *testing.T) {
	// The repo uses filter.Limit raw as SQL LIMIT $6, so a zero/negative
	// limit would mean LIMIT 0 (empty page) and an unbounded limit would let
	// a caller pull the whole registry in one query. The handler clamps both.
	cases := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"zero limit defaults", "?limit=0", 50, 0},
		{"missing limit defaults", "", 50, 0},
		{"over-cap clamps", "?limit=9999", 200, 0},
		{"negative offset floored", "?limit=10&offset=-5", 10, 0},
		{"in-range passthrough", "?limit=30&offset=60", 30, 60},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := &fakeList{}
			r := newRouter(&fakeGet{}, l, &fakeListMy{}, &fakeStats{}, withAuth(7, "methodist"))
			w := doGET(t, r, "/api/v1/student-debts"+tc.query)
			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tc.wantLimit, l.gotFilter.Limit, "limit clamp")
			assert.Equal(t, tc.wantOffset, l.gotFilter.Offset, "offset floor")
		})
	}
}

func TestStudentDebtHandler_Routing_MyDoesNotHitGet(t *testing.T) {
	// /my and /stats are static segments registered alongside /:id; if the
	// route tree ever lost their priority, /my would fall into Get where
	// parsePositiveID("my") fails with 400. Pin that /my reaches My.
	g := &fakeGet{err: repositories.ErrStudentDebtNotFound} // would 404 if hit
	m := &fakeListMy{res: repositories.StudentDebtListResult{Total: 0}}
	r := newRouter(g, &fakeList{}, m, &fakeStats{}, withAuth(42, "student"))

	w := doGET(t, r, "/api/v1/student-debts/my")
	assert.Equal(t, http.StatusOK, w.Code, "/my must route to My, not /:id")
	assert.Equal(t, int64(42), m.gotActor)
	assert.Equal(t, int64(0), g.gotID, "Get must not be invoked for /my")
}

func TestStudentDebtHandler_List_ForbiddenIs403(t *testing.T) {
	l := &fakeList{err: entities.ErrDebtAccessForbidden}
	r := newRouter(&fakeGet{}, l, &fakeListMy{}, &fakeStats{}, withAuth(5, "student"))

	w := doGET(t, r, "/api/v1/student-debts")
	assert.Equal(t, http.StatusForbidden, w.Code, "list denial is role-based, a true 403 (no resource to hide)")
}

func TestStudentDebtHandler_List_MissingAuthIs401(t *testing.T) {
	r := newRouter(&fakeGet{}, &fakeList{}, &fakeListMy{}, &fakeStats{}) // no auth middleware
	w := doGET(t, r, "/api/v1/student-debts")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Get --------------------------------------------------------------------

func TestStudentDebtHandler_Get_HappyPath(t *testing.T) {
	g := &fakeGet{debt: sampleDebt(t)}
	r := newRouter(g, &fakeList{}, &fakeListMy{}, &fakeStats{}, withAuth(7, "methodist"))

	w := doGET(t, r, "/api/v1/student-debts/55")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(55), g.gotID)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
}

func TestStudentDebtHandler_Get_BadIDIs400(t *testing.T) {
	r := newRouter(&fakeGet{}, &fakeList{}, &fakeListMy{}, &fakeStats{}, withAuth(7, "methodist"))
	w := doGET(t, r, "/api/v1/student-debts/abc")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentDebtHandler_Get_NotFoundIs404(t *testing.T) {
	g := &fakeGet{err: repositories.ErrStudentDebtNotFound}
	r := newRouter(g, &fakeList{}, &fakeListMy{}, &fakeStats{}, withAuth(7, "methodist"))
	w := doGET(t, r, "/api/v1/student-debts/55")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStudentDebtHandler_Get_ForbiddenHiddenAs404ForNonManager(t *testing.T) {
	g := &fakeGet{err: entities.ErrDebtAccessForbidden}
	r := newRouter(g, &fakeList{}, &fakeListMy{}, &fakeStats{}, withAuth(9, "teacher"))
	w := doGET(t, r, "/api/v1/student-debts/55")
	assert.Equal(t, http.StatusNotFound, w.Code, "non-manager scope denial collapses to 404 (IDOR)")
}

func TestStudentDebtHandler_Get_ForbiddenStays403ForManager(t *testing.T) {
	g := &fakeGet{err: entities.ErrDebtAccessForbidden}
	r := newRouter(g, &fakeList{}, &fakeListMy{}, &fakeStats{}, withAuth(1, "system_admin"))
	w := doGET(t, r, "/api/v1/student-debts/55")
	assert.Equal(t, http.StatusForbidden, w.Code, "a manager already knows the resource exists; a true 403")
}

// --- My ---------------------------------------------------------------------

func TestStudentDebtHandler_My_HappyPath(t *testing.T) {
	m := &fakeListMy{res: repositories.StudentDebtListResult{
		Items: []repositories.StudentDebtListItem{{ID: 1}},
		Total: 1,
	}}
	r := newRouter(&fakeGet{}, &fakeList{}, m, &fakeStats{}, withAuth(42, "student"))

	w := doGET(t, r, "/api/v1/student-debts/my?semester=3")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(42), m.gotActor)
	require.NotNil(t, m.gotFilter.Semester)
	assert.Equal(t, 3, *m.gotFilter.Semester)
}

// --- Stats ------------------------------------------------------------------

func TestStudentDebtHandler_Stats_HappyPath(t *testing.T) {
	s := &fakeStats{res: repositories.StudentDebtStats{Total: 9, Open: 4}}
	r := newRouter(&fakeGet{}, &fakeList{}, &fakeListMy{}, s, withAuth(7, "methodist"))

	w := doGET(t, r, "/api/v1/student-debts/stats")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(7), s.gotActor)
}

func TestStudentDebtHandler_Stats_ForbiddenIs403(t *testing.T) {
	s := &fakeStats{err: entities.ErrDebtAccessForbidden}
	r := newRouter(&fakeGet{}, &fakeList{}, &fakeListMy{}, s, withAuth(5, "student"))
	w := doGET(t, r, "/api/v1/student-debts/stats")
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNewStudentDebtHandler_PanicsOnNilPort(t *testing.T) {
	g, l, m, s := &fakeGet{}, &fakeList{}, &fakeListMy{}, &fakeStats{}
	cases := map[string]func(){
		"nil get":   func() { handlers.NewStudentDebtHandler(nil, l, m, s) },
		"nil list":  func() { handlers.NewStudentDebtHandler(g, nil, m, s) },
		"nil my":    func() { handlers.NewStudentDebtHandler(g, l, nil, s) },
		"nil stats": func() { handlers.NewStudentDebtHandler(g, l, m, nil) },
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic on nil port")
				}
			}()
			build()
		})
	}
}
