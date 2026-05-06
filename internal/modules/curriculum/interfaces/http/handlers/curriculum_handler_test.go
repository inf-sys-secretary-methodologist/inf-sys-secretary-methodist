package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeCreatePort is the test double for the Create use case port.
type fakeCreatePort struct {
	called   bool
	gotActor int64
	gotInput curUsecases.CreateCurriculumInput
	out      *entities.Curriculum
	err      error
}

func (f *fakeCreatePort) Execute(_ context.Context, actorID int64, in curUsecases.CreateCurriculumInput) (*entities.Curriculum, error) {
	f.called = true
	f.gotActor = actorID
	f.gotInput = in
	return f.out, f.err
}

// stub ports for the not-under-test endpoints — non-nil so the
// constructor's failure-closed check passes without exercising them.
type stubGetPort struct{}

func (stubGetPort) Execute(context.Context, int64) (*entities.Curriculum, error) {
	return nil, errors.New("stub: not implemented")
}

type stubListPort struct{}

func (stubListPort) Execute(context.Context, curUsecases.ListCurriculaInput) (curUsecases.CurriculaPage, error) {
	return curUsecases.CurriculaPage{}, errors.New("stub: not implemented")
}

type stubUpdatePort struct{}

func (stubUpdatePort) Execute(context.Context, int64, bool, curUsecases.UpdateCurriculumInput) (*entities.Curriculum, error) {
	return nil, errors.New("stub: not implemented")
}

// builtCurriculum returns a freshly-constructed draft for happy-path
// fakes to return — its content matches the canonical request body
// used in TestCurriculumHandler_Create_HappyPath.
func builtCurriculum(t *testing.T, id int64) *entities.Curriculum {
	t.Helper()
	c, err := entities.NewCurriculum(entities.NewCurriculumParams{
		Title:       "ИВТ-2026",
		Code:        "09.03.04-2026",
		Specialty:   "Информатика",
		Year:        2026,
		Description: "desc",
		CreatedBy:   42,
		Now:         time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	c.ID = id
	return c
}

// setupCreateRouter wires the handler with the supplied Create port
// and stubs for the rest. role+user_id are injected by middleware
// when withAuth=true; passing role="" disables the role injection
// (simulating an unauthenticated request) but still sets user_id so
// the 403-vs-401 distinction shows up cleanly.
func setupCreateRouter(create handlers.CreateCurriculumPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewCurriculumHandler(
		create, stubGetPort{}, stubListPort{}, stubUpdatePort{},
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
	r.POST("/api/curriculum", h.Create)
	return r
}

func doCreate(t *testing.T, r *gin.Engine, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if s, ok := body.(string); ok {
		req := httptest.NewRequest(http.MethodPost, "/api/curriculum", strings.NewReader(s))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(http.MethodPost, "/api/curriculum", &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestNewCurriculumHandler_PanicsOnNilPort(t *testing.T) {
	cases := []struct {
		name    string
		create  handlers.CreateCurriculumPort
		get     handlers.GetCurriculumPort
		list    handlers.ListCurriculaPort
		update  handlers.UpdateCurriculumPort
		submit  handlers.SubmitForApprovalPort
		approve handlers.ApproveCurriculumPort
		reject  handlers.RejectCurriculumPort
	}{
		{"create nil", nil, stubGetPort{}, stubListPort{}, stubUpdatePort{}, stubSubmitPort{}, stubApprovePort{}, stubRejectPort{}},
		{"get nil", &fakeCreatePort{}, nil, stubListPort{}, stubUpdatePort{}, stubSubmitPort{}, stubApprovePort{}, stubRejectPort{}},
		{"list nil", &fakeCreatePort{}, stubGetPort{}, nil, stubUpdatePort{}, stubSubmitPort{}, stubApprovePort{}, stubRejectPort{}},
		{"update nil", &fakeCreatePort{}, stubGetPort{}, stubListPort{}, nil, stubSubmitPort{}, stubApprovePort{}, stubRejectPort{}},
		{"submit nil", &fakeCreatePort{}, stubGetPort{}, stubListPort{}, stubUpdatePort{}, nil, stubApprovePort{}, stubRejectPort{}},
		{"approve nil", &fakeCreatePort{}, stubGetPort{}, stubListPort{}, stubUpdatePort{}, stubSubmitPort{}, nil, stubRejectPort{}},
		{"reject nil", &fakeCreatePort{}, stubGetPort{}, stubListPort{}, stubUpdatePort{}, stubSubmitPort{}, stubApprovePort{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("NewCurriculumHandler with %s did not panic", tc.name)
				}
			}()
			handlers.NewCurriculumHandler(tc.create, tc.get, tc.list, tc.update,
				tc.submit, tc.approve, tc.reject)
		})
	}
}

func TestCurriculumHandler_Create_HappyPath_Methodist(t *testing.T) {
	create := &fakeCreatePort{out: builtCurriculum(t, 42)}
	r := setupCreateRouter(create, "methodist", 7)

	body := map[string]any{
		"title":       "ИВТ-2026",
		"code":        "09.03.04-2026",
		"specialty":   "Информатика",
		"year":        2026,
		"description": "desc",
	}
	rec := doCreate(t, r, body)
	assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	require.True(t, create.called)
	assert.Equal(t, int64(7), create.gotActor)
	assert.Equal(t, "ИВТ-2026", create.gotInput.Title)
	assert.Equal(t, "09.03.04-2026", create.gotInput.Code)
	assert.Equal(t, 2026, create.gotInput.Year)

	var resp struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.EqualValues(t, 42, resp.Data["id"])
	assert.Equal(t, "ИВТ-2026", resp.Data["title"])
	assert.Equal(t, "09.03.04-2026", resp.Data["code"])
	assert.Equal(t, "draft", resp.Data["status"])
}

func TestCurriculumHandler_Create_HappyPath_Admin(t *testing.T) {
	create := &fakeCreatePort{out: builtCurriculum(t, 42)}
	r := setupCreateRouter(create, "system_admin", 99)

	body := map[string]any{
		"title": "T", "code": "C", "specialty": "S", "year": 2026,
	}
	rec := doCreate(t, r, body)
	assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, int64(99), create.gotActor)
}

func TestCurriculumHandler_Create_RejectsNonWriteRoles(t *testing.T) {
	cases := []struct {
		name string
		role string
	}{
		{"teacher → 403", "teacher"},
		{"academic_secretary → 403", "academic_secretary"},
		{"student → 403 (despite group middleware)", "student"},
		{"unknown role → 403", "unknown_thing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			create := &fakeCreatePort{}
			r := setupCreateRouter(create, tc.role, 7)

			body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
			rec := doCreate(t, r, body)
			assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			assert.False(t, create.called, "use case must not be invoked on role denial")
		})
	}
}

func TestCurriculumHandler_Create_MissingUserContextReturns401(t *testing.T) {
	create := &fakeCreatePort{}
	r := setupCreateRouter(create, "", 0) // no auth middleware

	body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
	rec := doCreate(t, r, body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
	assert.False(t, create.called)
}

func TestCurriculumHandler_Create_MalformedBodyReturns400(t *testing.T) {
	create := &fakeCreatePort{}
	r := setupCreateRouter(create, "methodist", 7)

	rec := doCreate(t, r, "not-json-at-all")
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.False(t, create.called)
}

func TestCurriculumHandler_Create_DomainErrorMappings(t *testing.T) {
	cases := []struct {
		name  string
		ucErr error
		want  int
	}{
		{"invariant violation → 422", entities.ErrInvalidCurriculum, http.StatusUnprocessableEntity},
		{"code conflict → 409", repositories.ErrCurriculumCodeExists, http.StatusConflict},
		{"unknown error → 500", errors.New("unexpected"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			create := &fakeCreatePort{err: tc.ucErr}
			r := setupCreateRouter(create, "methodist", 7)

			body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
			rec := doCreate(t, r, body)
			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
		})
	}
}
