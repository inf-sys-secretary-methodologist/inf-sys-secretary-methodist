package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

type fakeListPort struct {
	called bool
	gotIn  curUsecases.ListCurriculaInput
	out    curUsecases.CurriculaPage
	err    error
}

func (f *fakeListPort) Execute(_ context.Context, in curUsecases.ListCurriculaInput) (curUsecases.CurriculaPage, error) {
	f.called = true
	f.gotIn = in
	return f.out, f.err
}

func setupListRouter(list handlers.ListCurriculaPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewCurriculumHandler(
		&fakeCreatePort{}, stubGetPort{}, list, stubUpdatePort{},
		stubSubmitPort{}, stubApprovePort{}, stubRejectPort{},
	)
	if role != "" || userID != 0 {
		r.Use(func(c *gin.Context) {
			if userID != 0 {
				c.Set("user_id", userID)
			}
			if role != "" {
				c.Set("role", role)
			}
			c.Next()
		})
	}
	r.GET("/api/curriculum", h.List)
	return r
}

func doList(t *testing.T, r *gin.Engine, query string) *httptest.ResponseRecorder {
	t.Helper()
	path := "/api/curriculum"
	if query != "" {
		path += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCurriculumHandler_List_HappyPath_NoFilters(t *testing.T) {
	c1 := builtCurriculum(t, 1)
	c2 := builtCurriculum(t, 2)
	list := &fakeListPort{out: curUsecases.CurriculaPage{
		Items: []*entities.Curriculum{c1, c2},
		Total: 2,
	}}
	r := setupListRouter(list, "academic_secretary", 7)

	rec := doList(t, r, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, list.called)
	assert.Equal(t, 0, list.gotIn.Limit, "no limit query → use case applies default")
	assert.Equal(t, 0, list.gotIn.Offset)
	assert.Nil(t, list.gotIn.Status)
	assert.Nil(t, list.gotIn.Year)
	assert.Empty(t, list.gotIn.Specialty)
	assert.Nil(t, list.gotIn.CreatedBy)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data.Items, 2)
	assert.Equal(t, 2, resp.Data.Total)
}

func TestCurriculumHandler_List_FilterParsing(t *testing.T) {
	list := &fakeListPort{}
	r := setupListRouter(list, "academic_secretary", 7)

	rec := doList(t, r,
		"status=pending_approval&year=2026&specialty=Информатика&created_by=42&limit=25&offset=100")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, list.gotIn.Status)
	assert.Equal(t, entities.StatusPendingApproval, *list.gotIn.Status)
	require.NotNil(t, list.gotIn.Year)
	assert.Equal(t, 2026, *list.gotIn.Year)
	assert.Equal(t, "Информатика", list.gotIn.Specialty)
	require.NotNil(t, list.gotIn.CreatedBy)
	assert.Equal(t, int64(42), *list.gotIn.CreatedBy)
	assert.Equal(t, 25, list.gotIn.Limit)
	assert.Equal(t, 100, list.gotIn.Offset)
}

func TestCurriculumHandler_List_RejectsInvalidStatus(t *testing.T) {
	list := &fakeListPort{}
	r := setupListRouter(list, "academic_secretary", 7)

	rec := doList(t, r, "status=rejected")
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.False(t, list.called, "use case must not be invoked on bad status filter")
}

func TestCurriculumHandler_List_RejectsInvalidYear(t *testing.T) {
	cases := []string{
		"year=abc",
		"year=1999", // out of [2000, 2100]
		"year=2101",
		"year=-2026",
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			list := &fakeListPort{}
			r := setupListRouter(list, "academic_secretary", 7)

			rec := doList(t, r, q)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, list.called)
		})
	}
}

func TestCurriculumHandler_List_RejectsInvalidCreatedBy(t *testing.T) {
	cases := []string{
		"created_by=abc",
		"created_by=0",
		"created_by=-1",
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			list := &fakeListPort{}
			r := setupListRouter(list, "academic_secretary", 7)

			rec := doList(t, r, q)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, list.called)
		})
	}
}

func TestCurriculumHandler_List_RejectsInvalidPagination(t *testing.T) {
	cases := []string{
		"limit=abc",
		"offset=xyz",
		"limit=-5",
		"offset=-1",
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			list := &fakeListPort{}
			r := setupListRouter(list, "academic_secretary", 7)

			rec := doList(t, r, q)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, list.called)
		})
	}
}

func TestCurriculumHandler_List_RejectsStudent(t *testing.T) {
	list := &fakeListPort{}
	r := setupListRouter(list, "student", 42)

	rec := doList(t, r, "")
	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	assert.False(t, list.called)
}

func TestCurriculumHandler_List_MissingContextReturns401(t *testing.T) {
	list := &fakeListPort{}
	r := setupListRouter(list, "", 0)

	rec := doList(t, r, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_List_TransportErrorReturns500(t *testing.T) {
	list := &fakeListPort{err: errors.New("conn refused")}
	r := setupListRouter(list, "academic_secretary", 7)

	rec := doList(t, r, "")
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}
