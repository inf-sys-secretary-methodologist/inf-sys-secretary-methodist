package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// ===== Fake revision ports =====

type fakeCreateRevision struct {
	result   *entities.WorkProgram
	err      error
	called   bool
	gotIn    wpUsecases.CreateRevisionInput
	gotActor int64
	gotRole  string
}

func (f *fakeCreateRevision) Execute(_ context.Context, actorID int64, role string, in wpUsecases.CreateRevisionInput) (*entities.WorkProgram, error) {
	f.called, f.gotIn, f.gotActor, f.gotRole = true, in, actorID, role
	return f.result, f.err
}

type fakeSubmitRevision struct {
	result *entities.WorkProgram
	err    error
	called bool
	gotIn  wpUsecases.SubmitRevisionInput
}

func (f *fakeSubmitRevision) Execute(_ context.Context, _ int64, _ string, in wpUsecases.SubmitRevisionInput) (*entities.WorkProgram, error) {
	f.called, f.gotIn = true, in
	return f.result, f.err
}

type fakeApproveRevision struct {
	result *entities.WorkProgram
	err    error
	called bool
	gotIn  wpUsecases.ApproveRevisionInput
}

func (f *fakeApproveRevision) Execute(_ context.Context, _ int64, _ string, in wpUsecases.ApproveRevisionInput) (*entities.WorkProgram, error) {
	f.called, f.gotIn = true, in
	return f.result, f.err
}

type fakeRejectRevision struct {
	result *entities.WorkProgram
	err    error
	called bool
	gotIn  wpUsecases.RejectRevisionInput
}

func (f *fakeRejectRevision) Execute(_ context.Context, _ int64, _ string, in wpUsecases.RejectRevisionInput) (*entities.WorkProgram, error) {
	f.called, f.gotIn = true, in
	return f.result, f.err
}

func newRevisionRouter(
	fc *fakeCreateRevision, fs *fakeSubmitRevision, fa *fakeApproveRevision, fr *fakeRejectRevision,
	mw ...gin.HandlerFunc,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if fc == nil {
		fc = &fakeCreateRevision{}
	}
	if fs == nil {
		fs = &fakeSubmitRevision{}
	}
	if fa == nil {
		fa = &fakeApproveRevision{}
	}
	if fr == nil {
		fr = &fakeRejectRevision{}
	}
	h := NewRevisionHandler(fc, fs, fa, fr)
	api := r.Group("/api/v1")
	for _, m := range mw {
		api.Use(m)
	}
	RegisterRevisionRoutes(api, h)
	return r
}

func validCreateRevisionBody() CreateRevisionRequest {
	return CreateRevisionRequest{
		ChangeType:    "literature",
		ChangeSummary: "Обновлён список литературы по приказу Минобрнауки",
	}
}

// ===== Create =====

func TestRevisionHandler_Create_HappyPath(t *testing.T) {
	fc := &fakeCreateRevision{result: sampleWP(t)}
	r := newRevisionRouter(fc, nil, nil, nil, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions", validCreateRevisionBody())

	assert.Equal(t, http.StatusCreated, w.Code)
	require.True(t, fc.called)
	assert.Equal(t, int64(42), fc.gotActor, "author derives from JWT")
	assert.Equal(t, int64(99), fc.gotIn.WorkProgramID, "wp id from path")
	assert.Equal(t, "literature", fc.gotIn.ChangeType)
}

func TestRevisionHandler_Create_Unauthorized(t *testing.T) {
	r := newRevisionRouter(nil, nil, nil, nil) // no withAuth
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions", validCreateRevisionBody())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRevisionHandler_Create_BadBody(t *testing.T) {
	fc := &fakeCreateRevision{}
	r := newRevisionRouter(fc, nil, nil, nil, withAuth(42, "teacher"))
	body := validCreateRevisionBody()
	body.ChangeSummary = "" // violates binding:"required"
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fc.called)
}

func TestRevisionHandler_Create_BadID(t *testing.T) {
	fc := &fakeCreateRevision{}
	r := newRevisionRouter(fc, nil, nil, nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/abc/revisions", validCreateRevisionBody())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fc.called)
}

func TestRevisionHandler_Create_ForbiddenHiddenAsNotFound(t *testing.T) {
	fc := &fakeCreateRevision{err: domain.ErrWorkProgramScopeForbidden}
	r := newRevisionRouter(fc, nil, nil, nil, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions", validCreateRevisionBody())
	assert.Equal(t, http.StatusNotFound, w.Code, "non-admin forbidden collapses to 404 (IDOR)")
}

func TestRevisionHandler_Create_NotPermitted(t *testing.T) {
	fc := &fakeCreateRevision{err: domain.ErrRevisionNotPermitted}
	r := newRevisionRouter(fc, nil, nil, nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions", validCreateRevisionBody())
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== Submit =====

func TestRevisionHandler_Submit_HappyPath(t *testing.T) {
	fs := &fakeSubmitRevision{result: sampleWP(t)}
	r := newRevisionRouter(nil, fs, nil, nil, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/submit", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.True(t, fs.called)
	assert.Equal(t, int64(99), fs.gotIn.WorkProgramID)
	assert.Equal(t, int64(500), fs.gotIn.RevisionID)
}

func TestRevisionHandler_Submit_RevisionNotFound(t *testing.T) {
	fs := &fakeSubmitRevision{err: domain.ErrRevisionNotFound}
	r := newRevisionRouter(nil, fs, nil, nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/submit", nil)
	assert.Equal(t, http.StatusNotFound, w.Code, "ErrRevisionNotFound maps to 404")
}

func TestRevisionHandler_Submit_BadRevisionID(t *testing.T) {
	fs := &fakeSubmitRevision{}
	r := newRevisionRouter(nil, fs, nil, nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/xyz/submit", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fs.called)
}

// ===== Approve =====

func TestRevisionHandler_Approve_HappyPath(t *testing.T) {
	fa := &fakeApproveRevision{result: sampleWP(t)}
	r := newRevisionRouter(nil, nil, fa, nil, withAuth(55, "methodist"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/approve", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.True(t, fa.called)
	assert.Equal(t, int64(500), fa.gotIn.RevisionID)
}

func TestRevisionHandler_Approve_WrongStatus(t *testing.T) {
	fa := &fakeApproveRevision{err: domain.ErrInvalidStatusTransition}
	r := newRevisionRouter(nil, nil, fa, nil, withAuth(55, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/approve", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== Reject =====

func TestRevisionHandler_Reject_HappyPath(t *testing.T) {
	fr := &fakeRejectRevision{result: sampleWP(t)}
	r := newRevisionRouter(nil, nil, nil, fr, withAuth(55, "methodist"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/reject",
		RejectRevisionRequest{Reason: "Не соответствует приказу"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.True(t, fr.called)
	assert.Equal(t, int64(500), fr.gotIn.RevisionID)
	assert.Equal(t, "Не соответствует приказу", fr.gotIn.Reason)
}

func TestRevisionHandler_Reject_BadBody(t *testing.T) {
	fr := &fakeRejectRevision{}
	r := newRevisionRouter(nil, nil, nil, fr, withAuth(55, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/reject",
		RejectRevisionRequest{Reason: ""}) // violates binding:"required"
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fr.called)
}

func TestRevisionHandler_Reject_EmptyReasonFromDomain(t *testing.T) {
	fr := &fakeRejectRevision{err: domain.ErrRejectReasonRequired}
	r := newRevisionRouter(nil, nil, nil, fr, withAuth(55, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/revisions/500/reject",
		RejectRevisionRequest{Reason: "x"})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestNewRevisionHandler_PanicsOnNilPort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil port")
		}
	}()
	NewRevisionHandler(nil, &fakeSubmitRevision{}, &fakeApproveRevision{}, &fakeRejectRevision{})
}

// TestRevisionRoutes_CoexistWithWorkProgramRoutes guards the one genuinely
// new structural risk: both RegisterWorkProgramRoutes and
// RegisterRevisionRoutes mount on the same group and share the :id param at
// the same tree position. gin's router panics on a wildcard-name conflict,
// so this pins that the two registrars coexist — a future param rename on
// one side fails here in CI instead of at server boot.
func TestRevisionRoutes_CoexistWithWorkProgramRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")

	wp := NewWorkProgramHandler(
		&fakeCreate{}, &fakeGet{}, &fakeList{},
		&fakeSubmit{}, &fakeApprove{}, &fakeReject{}, &fakeDiscard{}, &fakeGenerate{},
	)
	rev := NewRevisionHandler(&fakeCreateRevision{}, &fakeSubmitRevision{}, &fakeApproveRevision{}, &fakeRejectRevision{})

	assert.NotPanics(t, func() {
		RegisterWorkProgramRoutes(api, wp)
		RegisterRevisionRoutes(api, rev)
	}, "WP + revision routes must coexist on one group without a gin wildcard conflict")
}
