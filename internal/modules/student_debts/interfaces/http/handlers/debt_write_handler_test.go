package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/interfaces/http/handlers"
)

type fakeSchedule struct {
	debt     *entities.StudentDebt
	err      error
	gotIn    sdUsecases.ScheduleResitInput
	gotActor int64
	gotRole  string
}

func (f *fakeSchedule) Execute(_ context.Context, actorID int64, role string, in sdUsecases.ScheduleResitInput) (*entities.StudentDebt, error) {
	f.gotActor, f.gotRole, f.gotIn = actorID, role, in
	return f.debt, f.err
}

type fakeRecord struct {
	debt     *entities.StudentDebt
	err      error
	gotIn    sdUsecases.RecordResitResultInput
	gotActor int64
	gotRole  string
}

func (f *fakeRecord) Execute(_ context.Context, actorID int64, role string, in sdUsecases.RecordResitResultInput) (*entities.StudentDebt, error) {
	f.gotActor, f.gotRole, f.gotIn = actorID, role, in
	return f.debt, f.err
}

func newWriteRouter(s *fakeSchedule, rec *fakeRecord, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	for _, h := range mw {
		api.Use(h)
	}
	h := handlers.NewStudentDebtWriteHandler(s, rec)
	handlers.RegisterStudentDebtWriteRoutes(api, h)
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

// --- ScheduleResit ----------------------------------------------------------

func TestStudentDebtWriteHandler_ScheduleResit_HappyPath(t *testing.T) {
	s := &fakeSchedule{debt: sampleDebt(t)}
	r := newWriteRouter(s, &fakeRecord{}, withAuth(7, "methodist"))

	body := map[string]any{"scheduled_date": "2026-07-01T09:00:00Z", "examiner": "Петров П.П."}
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(7), s.gotActor)
	assert.Equal(t, int64(55), s.gotIn.DebtID)
	assert.Equal(t, "Петров П.П.", s.gotIn.Examiner)
	assert.Equal(t, 2026, s.gotIn.ScheduledDate.Year())
	assert.Equal(t, 7, int(s.gotIn.ScheduledDate.Month()))
}

func TestStudentDebtWriteHandler_ScheduleResit_MissingExaminerIs400(t *testing.T) {
	r := newWriteRouter(&fakeSchedule{}, &fakeRecord{}, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit",
		map[string]any{"scheduled_date": "2026-07-01T09:00:00Z"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentDebtWriteHandler_ScheduleResit_BadDateIs400(t *testing.T) {
	r := newWriteRouter(&fakeSchedule{}, &fakeRecord{}, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit",
		map[string]any{"scheduled_date": "01.07.2026", "examiner": "Петров П.П."})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentDebtWriteHandler_ScheduleResit_BadIDIs400(t *testing.T) {
	r := newWriteRouter(&fakeSchedule{}, &fakeRecord{}, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/abc/resit",
		map[string]any{"scheduled_date": "2026-07-01T09:00:00Z", "examiner": "Петров П.П."})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentDebtWriteHandler_ScheduleResit_DebtClosedIs409(t *testing.T) {
	s := &fakeSchedule{err: entities.ErrDebtClosed}
	r := newWriteRouter(s, &fakeRecord{}, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit",
		map[string]any{"scheduled_date": "2026-07-01T09:00:00Z", "examiner": "Петров П.П."})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestStudentDebtWriteHandler_ScheduleResit_ForbiddenIs403(t *testing.T) {
	// Write denial is role-based and happens before any existence check —
	// a true 403, not an IDOR-collapse 404.
	s := &fakeSchedule{err: entities.ErrDebtAccessForbidden}
	r := newWriteRouter(s, &fakeRecord{}, withAuth(9, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit",
		map[string]any{"scheduled_date": "2026-07-01T09:00:00Z", "examiner": "Петров П.П."})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudentDebtWriteHandler_ScheduleResit_NotFoundIs404(t *testing.T) {
	s := &fakeSchedule{err: repositories.ErrStudentDebtNotFound}
	r := newWriteRouter(s, &fakeRecord{}, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit",
		map[string]any{"scheduled_date": "2026-07-01T09:00:00Z", "examiner": "Петров П.П."})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- RecordResitResult ------------------------------------------------------

func TestStudentDebtWriteHandler_RecordResult_HappyPath(t *testing.T) {
	rec := &fakeRecord{debt: sampleDebt(t)}
	r := newWriteRouter(&fakeSchedule{}, rec, withAuth(7, "methodist"))

	grade := 5
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/attempts/1/result",
		map[string]any{"result": "passed", "grade": grade})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(55), rec.gotIn.DebtID)
	assert.Equal(t, 1, rec.gotIn.AttemptNo)
	assert.Equal(t, entities.ResitResultPassed, rec.gotIn.Result)
	require.NotNil(t, rec.gotIn.Grade)
	assert.Equal(t, 5, *rec.gotIn.Grade)
}

func TestStudentDebtWriteHandler_RecordResult_NoScheduledResitIs409(t *testing.T) {
	rec := &fakeRecord{err: entities.ErrNoScheduledResit}
	r := newWriteRouter(&fakeSchedule{}, rec, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/attempts/1/result",
		map[string]any{"result": "failed"})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestStudentDebtWriteHandler_RecordResult_InvalidRecordIs422(t *testing.T) {
	rec := &fakeRecord{err: entities.ErrInvalidResitRecord}
	r := newWriteRouter(&fakeSchedule{}, rec, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/attempts/1/result",
		map[string]any{"result": "passed", "grade": 999})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestStudentDebtWriteHandler_RecordResult_BadAttemptNoIs400(t *testing.T) {
	r := newWriteRouter(&fakeSchedule{}, &fakeRecord{}, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/attempts/abc/result",
		map[string]any{"result": "passed"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentDebtWriteHandler_MissingAuthIs401(t *testing.T) {
	r := newWriteRouter(&fakeSchedule{}, &fakeRecord{})
	w := doJSON(t, r, http.MethodPost, "/api/v1/student-debts/55/resit",
		map[string]any{"scheduled_date": "2026-07-01T09:00:00Z", "examiner": "Петров П.П."})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewStudentDebtWriteHandler_PanicsOnNilPort(t *testing.T) {
	s, rec := &fakeSchedule{}, &fakeRecord{}
	cases := map[string]func(){
		"nil schedule": func() { handlers.NewStudentDebtWriteHandler(nil, rec) },
		"nil record":   func() { handlers.NewStudentDebtWriteHandler(s, nil) },
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
