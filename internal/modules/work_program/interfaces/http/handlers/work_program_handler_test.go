package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// ===== Fake usecase ports =====

type fakeCreate struct {
	result   *entities.WorkProgram
	err      error
	called   bool
	gotIn    wpUsecases.CreateWorkProgramInput
	gotActor int64
	gotRole  string
}

func (f *fakeCreate) Execute(_ context.Context, actorID int64, role string, in wpUsecases.CreateWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	f.gotIn = in
	f.gotActor = actorID
	f.gotRole = role
	return f.result, f.err
}

type fakeGet struct {
	result *entities.WorkProgram
	err    error
	called bool
}

func (f *fakeGet) Execute(_ context.Context, _ int64, _ string, _ wpUsecases.GetWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	return f.result, f.err
}

type fakeList struct {
	result wpUsecases.ListWorkProgramsResult
	err    error
	called bool
	gotIn  wpUsecases.ListWorkProgramsInput
}

func (f *fakeList) Execute(_ context.Context, _ int64, _ string, in wpUsecases.ListWorkProgramsInput) (wpUsecases.ListWorkProgramsResult, error) {
	f.called = true
	f.gotIn = in
	return f.result, f.err
}

type fakeSubmit struct {
	result   *entities.WorkProgram
	err      error
	called   bool
	gotID    int64
	gotActor int64
	gotRole  string
}

func (f *fakeSubmit) Execute(_ context.Context, actorID int64, role string, in wpUsecases.SubmitWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	f.gotID = in.ID
	f.gotActor = actorID
	f.gotRole = role
	return f.result, f.err
}

type fakeApprove struct {
	result   *entities.WorkProgram
	err      error
	called   bool
	gotID    int64
	gotActor int64
}

func (f *fakeApprove) Execute(_ context.Context, actorID int64, _ string, in wpUsecases.ApproveWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	f.gotID = in.ID
	f.gotActor = actorID
	return f.result, f.err
}

type fakeReject struct {
	result    *entities.WorkProgram
	err       error
	called    bool
	gotID     int64
	gotReason string
}

func (f *fakeReject) Execute(_ context.Context, _ int64, _ string, in wpUsecases.RejectWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	f.gotID = in.ID
	f.gotReason = in.Reason
	return f.result, f.err
}

type fakeDiscard struct {
	result *entities.WorkProgram
	err    error
	called bool
	gotID  int64
}

func (f *fakeDiscard) Execute(_ context.Context, _ int64, _ string, in wpUsecases.DiscardDraftWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	f.gotID = in.ID
	return f.result, f.err
}

// withAuth pre-sets user_id + role в the gin context — mirrors what the
// RequireAuth middleware does in production. Pinning the exact context
// keys (`user_id`, `role`) catches drift per
// feedback_handler_context_key_must_match_middleware.
func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func sampleWP(t *testing.T) *entities.WorkProgram {
	t.Helper()
	ts := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	return entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID:                 99,
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Аннотация",
		Status:             domain.StatusDraft,
		AuthorID:           42,
		Version:            0,
		CreatedAt:          ts,
		UpdatedAt:          ts,
	})
}

// newRouter builds a router for the read+create tests. The four
// transition ports are filled with no-op fakes (the ctor requires all
// seven non-nil) so existing read+create tests need not name them.
func newRouter(fc *fakeCreate, fg *fakeGet, fl *fakeList, mw ...gin.HandlerFunc) *gin.Engine {
	return newRouterFull(fc, fg, fl, &fakeSubmit{}, &fakeApprove{}, &fakeReject{}, &fakeDiscard{}, mw...)
}

// newRouterFull wires all seven ports — transition tests use it directly
// to inject the fake under test.
func newRouterFull(
	fc *fakeCreate, fg *fakeGet, fl *fakeList,
	fsub *fakeSubmit, fapp *fakeApprove, frej *fakeReject, fdis *fakeDiscard,
	mw ...gin.HandlerFunc,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWorkProgramHandler(fc, fg, fl, fsub, fapp, frej, fdis)
	api := r.Group("/api/v1")
	for _, m := range mw {
		api.Use(m)
	}
	RegisterWorkProgramRoutes(api, h)
	return r
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validCreateBody() CreateWorkProgramRequest {
	return CreateWorkProgramRequest{
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Аннотация",
	}
}

// ===== Create =====

func TestWorkProgramHandler_Create_HappyPath(t *testing.T) {
	fc := &fakeCreate{result: sampleWP(t)}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, fc.called)
	// Author + role derive from JWT context, not request body.
	assert.Equal(t, int64(42), fc.gotActor)
	assert.Equal(t, "teacher", fc.gotRole)
	assert.Equal(t, int64(7), fc.gotIn.DisciplineID)
}

func TestWorkProgramHandler_Create_Unauthenticated(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{}) // no withAuth
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Create_InvalidBodyMaps400(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))
	// Missing required title / specialty_code.
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", CreateWorkProgramRequest{
		DisciplineID: 7,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkProgramHandler_Create_ForbiddenMaps403(t *testing.T) {
	// Create is a collection POST — a role-based denial is a true 403
	// (no resource id to enumerate, so no IDOR collapse here).
	fc := &fakeCreate{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestWorkProgramHandler_Create_InvalidWorkProgramMaps422(t *testing.T) {
	fc := &fakeCreate{err: domain.ErrInvalidWorkProgram}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestWorkProgramHandler_Create_IdentityExistsMaps409(t *testing.T) {
	fc := &fakeCreate{err: repositories.ErrWorkProgramIdentityExists}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ===== Get =====

func TestWorkProgramHandler_Get_HappyPath(t *testing.T) {
	fg := &fakeGet{result: sampleWP(t)}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fg.called)
}

func TestWorkProgramHandler_Get_Unauthenticated(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{})
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Get_InvalidIDMaps400(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkProgramHandler_Get_NotFoundMaps404(t *testing.T) {
	fg := &fakeGet{err: repositories.ErrWorkProgramNotFound}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/404", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// IDOR mitigation: a non-admin denied by scope must see 404, not 403,
// so resource ids cannot be enumerated by privilege class (OWASP).
func TestWorkProgramHandler_Get_ForbiddenHiddenAs404ForNonAdmin(t *testing.T) {
	fg := &fakeGet{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Admins are entitled to know the resource exists — they keep the 403
// signal (they can see all РПД per ADR-5, so this is defensive).
func TestWorkProgramHandler_Get_ForbiddenStays403ForAdmin(t *testing.T) {
	fg := &fakeGet{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(1, "system_admin"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ===== List =====

func TestWorkProgramHandler_List_HappyPath(t *testing.T) {
	fl := &fakeList{result: wpUsecases.ListWorkProgramsResult{
		Items: []repositories.ListItem{
			{ID: 1, Title: "РПД A", Status: domain.StatusApproved},
			{ID: 2, Title: "РПД B", Status: domain.StatusDraft},
		},
		Total: 2,
	}}
	r := newRouter(&fakeCreate{}, &fakeGet{}, fl, withAuth(42, "methodist"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fl.called)
}

func TestWorkProgramHandler_List_Unauthenticated(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{})
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_List_ParsesQueryFilters(t *testing.T) {
	fl := &fakeList{}
	r := newRouter(&fakeCreate{}, &fakeGet{}, fl, withAuth(42, "methodist"))
	w := doJSON(t, r, http.MethodGet,
		"/api/v1/work-programs?status=approved&discipline_id=7&specialty_code=09.03.01&applicable_from_year=2026&author_id=42&limit=10&offset=20",
		nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, fl.gotIn.Status)
	assert.Equal(t, "approved", *fl.gotIn.Status)
	require.NotNil(t, fl.gotIn.DisciplineID)
	assert.Equal(t, int64(7), *fl.gotIn.DisciplineID)
	assert.Equal(t, "09.03.01", fl.gotIn.SpecialtyCode)
	require.NotNil(t, fl.gotIn.ApplicableFromYear)
	assert.Equal(t, 2026, *fl.gotIn.ApplicableFromYear)
	require.NotNil(t, fl.gotIn.AuthorID)
	assert.Equal(t, int64(42), *fl.gotIn.AuthorID)
	assert.Equal(t, 10, fl.gotIn.Limit)
	assert.Equal(t, 20, fl.gotIn.Offset)
}

func TestWorkProgramHandler_List_OmittedFiltersStayNil(t *testing.T) {
	fl := &fakeList{}
	r := newRouter(&fakeCreate{}, &fakeGet{}, fl, withAuth(42, "methodist"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs", nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, fl.gotIn.Status)
	assert.Nil(t, fl.gotIn.DisciplineID)
	assert.Nil(t, fl.gotIn.ApplicableFromYear)
	assert.Nil(t, fl.gotIn.AuthorID)
	assert.Equal(t, "", fl.gotIn.SpecialtyCode)
}

// Unknown-role denial on a collection endpoint is a true 403 (no
// specific resource id to enumerate), so no IDOR collapse here.
func TestWorkProgramHandler_List_ForbiddenMaps403(t *testing.T) {
	fl := &fakeList{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(&fakeCreate{}, &fakeGet{}, fl, withAuth(7, "guest"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Malformed numeric query params are silently dropped (lenient contract,
// consistent with the extracurricular handler) — unparseable filters
// stay nil and bad limit/offset zero out (use case then applies its own
// defaults). Pins this intent so a later "stricter parsing" change is a
// deliberate, test-visible decision rather than an accident.
func TestWorkProgramHandler_List_MalformedNumericParamsDropped(t *testing.T) {
	fl := &fakeList{}
	r := newRouter(&fakeCreate{}, &fakeGet{}, fl, withAuth(42, "methodist"))
	w := doJSON(t, r, http.MethodGet,
		"/api/v1/work-programs?discipline_id=abc&author_id=xyz&applicable_from_year=nope&limit=foo&offset=bar",
		nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, fl.gotIn.DisciplineID)
	assert.Nil(t, fl.gotIn.AuthorID)
	assert.Nil(t, fl.gotIn.ApplicableFromYear)
	assert.Equal(t, 0, fl.gotIn.Limit)
	assert.Equal(t, 0, fl.gotIn.Offset)
}

// ===== transition-test router helpers =====

func submitRouter(fs *fakeSubmit, mw ...gin.HandlerFunc) *gin.Engine {
	return newRouterFull(&fakeCreate{}, &fakeGet{}, &fakeList{}, fs, &fakeApprove{}, &fakeReject{}, &fakeDiscard{}, mw...)
}

// ===== Submit =====

func TestWorkProgramHandler_Submit_HappyPath(t *testing.T) {
	fs := &fakeSubmit{result: sampleWP(t)}
	r := submitRouter(fs, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/submit", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fs.called)
	assert.Equal(t, int64(99), fs.gotID)
	assert.Equal(t, int64(42), fs.gotActor)
	assert.Equal(t, "teacher", fs.gotRole)
}

func TestWorkProgramHandler_Submit_Unauthenticated(t *testing.T) {
	r := submitRouter(&fakeSubmit{})
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/submit", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Submit_InvalidIDMaps400(t *testing.T) {
	r := submitRouter(&fakeSubmit{}, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/abc/submit", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkProgramHandler_Submit_NotFoundMaps404(t *testing.T) {
	fs := &fakeSubmit{err: repositories.ErrWorkProgramNotFound}
	r := submitRouter(fs, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/404/submit", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkProgramHandler_Submit_InvalidTransitionMaps422(t *testing.T) {
	fs := &fakeSubmit{err: domain.ErrInvalidStatusTransition}
	r := submitRouter(fs, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/submit", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// Resource-scoped transition → non-admin scope denial collapses to 404.
func TestWorkProgramHandler_Submit_ForbiddenHiddenAs404ForNonAdmin(t *testing.T) {
	fs := &fakeSubmit{err: domain.ErrWorkProgramScopeForbidden}
	r := submitRouter(fs, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/submit", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkProgramHandler_Submit_VersionConflictMaps409(t *testing.T) {
	fs := &fakeSubmit{err: repositories.ErrWorkProgramVersionConflict}
	r := submitRouter(fs, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/submit", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ===== Approve =====

func approveRouter(fa *fakeApprove, mw ...gin.HandlerFunc) *gin.Engine {
	return newRouterFull(&fakeCreate{}, &fakeGet{}, &fakeList{}, &fakeSubmit{}, fa, &fakeReject{}, &fakeDiscard{}, mw...)
}

func TestWorkProgramHandler_Approve_HappyPath(t *testing.T) {
	fa := &fakeApprove{result: sampleWP(t)}
	r := approveRouter(fa, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/approve", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fa.called)
	assert.Equal(t, int64(99), fa.gotID)
	// Approver identity derives from the JWT actor, not the request body.
	assert.Equal(t, int64(5), fa.gotActor)
}

func TestWorkProgramHandler_Approve_Unauthenticated(t *testing.T) {
	r := approveRouter(&fakeApprove{})
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/approve", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Approve_InvalidIDMaps400(t *testing.T) {
	r := approveRouter(&fakeApprove{}, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/abc/approve", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// A non-approver (teacher) hitting approve is scope-denied → collapsed to
// 404 for the non-admin (IDOR), not 403.
func TestWorkProgramHandler_Approve_ForbiddenHiddenAs404ForNonAdmin(t *testing.T) {
	fa := &fakeApprove{err: domain.ErrWorkProgramScopeForbidden}
	r := approveRouter(fa, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/approve", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkProgramHandler_Approve_InvalidTransitionMaps422(t *testing.T) {
	fa := &fakeApprove{err: domain.ErrInvalidStatusTransition}
	r := approveRouter(fa, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/approve", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== Reject =====

func rejectRouter(fr *fakeReject, mw ...gin.HandlerFunc) *gin.Engine {
	return newRouterFull(&fakeCreate{}, &fakeGet{}, &fakeList{}, &fakeSubmit{}, &fakeApprove{}, fr, &fakeDiscard{}, mw...)
}

func TestWorkProgramHandler_Reject_HappyPath(t *testing.T) {
	fr := &fakeReject{result: sampleWP(t)}
	r := rejectRouter(fr, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/reject",
		RejectWorkProgramRequest{Reason: "часы не соответствуют учебному плану"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fr.called)
	assert.Equal(t, int64(99), fr.gotID)
	assert.Equal(t, "часы не соответствуют учебному плану", fr.gotReason)
}

func TestWorkProgramHandler_Reject_Unauthenticated(t *testing.T) {
	r := rejectRouter(&fakeReject{})
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/reject",
		RejectWorkProgramRequest{Reason: "x"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Reject_InvalidIDMaps400(t *testing.T) {
	r := rejectRouter(&fakeReject{}, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/abc/reject",
		RejectWorkProgramRequest{Reason: "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Empty reason is rejected at the binding layer (binding:"required").
func TestWorkProgramHandler_Reject_MissingReasonMaps400(t *testing.T) {
	fr := &fakeReject{}
	r := rejectRouter(fr, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/reject",
		RejectWorkProgramRequest{Reason: ""})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fr.called) // never reaches the use case
}

// A whitespace-only reason passes binding but the domain trims + rejects
// it → ErrRejectReasonRequired → 422.
func TestWorkProgramHandler_Reject_ReasonRequiredMaps422(t *testing.T) {
	fr := &fakeReject{err: domain.ErrRejectReasonRequired}
	r := rejectRouter(fr, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/reject",
		RejectWorkProgramRequest{Reason: "   "})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestWorkProgramHandler_Reject_ForbiddenHiddenAs404ForNonAdmin(t *testing.T) {
	fr := &fakeReject{err: domain.ErrWorkProgramScopeForbidden}
	r := rejectRouter(fr, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/reject",
		RejectWorkProgramRequest{Reason: "x"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===== Discard =====

func discardRouter(fd *fakeDiscard, mw ...gin.HandlerFunc) *gin.Engine {
	return newRouterFull(&fakeCreate{}, &fakeGet{}, &fakeList{}, &fakeSubmit{}, &fakeApprove{}, &fakeReject{}, fd, mw...)
}

func TestWorkProgramHandler_Discard_HappyPath(t *testing.T) {
	fd := &fakeDiscard{result: sampleWP(t)}
	r := discardRouter(fd, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/discard", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fd.called)
	assert.Equal(t, int64(99), fd.gotID)
}

func TestWorkProgramHandler_Discard_Unauthenticated(t *testing.T) {
	r := discardRouter(&fakeDiscard{})
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/discard", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Discard_InvalidIDMaps400(t *testing.T) {
	r := discardRouter(&fakeDiscard{}, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/abc/discard", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkProgramHandler_Discard_ForbiddenHiddenAs404ForNonAdmin(t *testing.T) {
	fd := &fakeDiscard{err: domain.ErrWorkProgramScopeForbidden}
	r := discardRouter(fd, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/discard", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkProgramHandler_Discard_InvalidTransitionMaps422(t *testing.T) {
	fd := &fakeDiscard{err: domain.ErrInvalidStatusTransition}
	r := discardRouter(fd, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/discard", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== VersionConflict symmetry =====
//
// Submit pins VersionConflict→409 above; Approve/Reject/Discard each also
// call repo.Update and can surface ErrWorkProgramVersionConflict. Pin the
// 409 mapping on all three so the optimistic-lock contract is symmetric
// across every mutating transition (the shared mapper already handles it —
// these guard against a future per-endpoint divergence).

func TestWorkProgramHandler_Approve_VersionConflictMaps409(t *testing.T) {
	fa := &fakeApprove{err: repositories.ErrWorkProgramVersionConflict}
	r := approveRouter(fa, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/approve", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestWorkProgramHandler_Reject_VersionConflictMaps409(t *testing.T) {
	fr := &fakeReject{err: repositories.ErrWorkProgramVersionConflict}
	r := rejectRouter(fr, withAuth(5, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/reject",
		RejectWorkProgramRequest{Reason: "x"})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestWorkProgramHandler_Discard_VersionConflictMaps409(t *testing.T) {
	fd := &fakeDiscard{err: repositories.ErrWorkProgramVersionConflict}
	r := discardRouter(fd, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/discard", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}
