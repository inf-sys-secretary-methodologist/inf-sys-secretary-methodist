package http_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	docHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
)

// --- test repo shared with workflow_usecases_test.go pattern ---

type fakeWorkflowRepo struct {
	mu     sync.Mutex
	stored map[int64]*entities.Document
}

func newRepo(docs ...*entities.Document) *fakeWorkflowRepo {
	r := &fakeWorkflowRepo{stored: map[int64]*entities.Document{}}
	for _, d := range docs {
		r.stored[d.ID] = d
	}
	return r
}

func (r *fakeWorkflowRepo) GetByID(_ context.Context, id int64) (*entities.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.stored[id]; ok {
		return d, nil
	}
	return nil, usecases.ErrDocumentNotFound
}

func (r *fakeWorkflowRepo) Update(_ context.Context, d *entities.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stored[d.ID] = d
	return nil
}

type noopAudit struct{}

func (noopAudit) LogAuditEvent(_ context.Context, _, _ string, _ map[string]any) {}

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func draftDoc(id, authorID int64) *entities.Document {
	d := entities.NewDocument("План занятий 2026", 1, authorID)
	d.ID = id
	return d
}

func docAtStatus(id, authorID int64, s entities.DocumentStatus) *entities.Document {
	d := draftDoc(id, authorID)
	d.Status = s
	return d
}

// adminRoleGate mirrors the production RequireRole(AcademicSecretary,
// SystemAdmin) middleware — only those two roles pass. Anything else
// → 403. Catches the test-vs-production drift class flagged by reviewer
// per `feedback_route_extraction_for_security_test`.
func adminRoleGate() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		role, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if role != "academic_secretary" && role != "system_admin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

// mountWorkflow mounts the three workflow endpoints onto a fresh
// gin.Engine using the production-shaped split: submit on a non-student
// group, approve/reject on an admin-gated /admin/documents sub-group.
// Mirror к main.go:1820-1832 routing layout so a regression in the
// admin gate surfaces here, not only in manual smoke.
func mountWorkflow(t *testing.T, userID int64, role string, repo *fakeWorkflowRepo) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	submitUC := usecases.NewSubmitDocumentUseCase(repo, noopAudit{}, fixedClock(now))
	approveUC := usecases.NewApproveDocumentUseCase(repo, noopAudit{}, fixedClock(now))
	rejectUC := usecases.NewRejectDocumentUseCase(repo, noopAudit{}, fixedClock(now))
	registerUC := usecases.NewRegisterDocumentUseCase(repo, noopAudit{}, fixedClock(now))
	startRoutingUC := usecases.NewStartRoutingUseCase(repo, noopAudit{}, fixedClock(now))
	signVisaUC := usecases.NewSignVisaUseCase(repo, noopAudit{}, fixedClock(now))
	assignExecutorUC := usecases.NewAssignExecutorUseCase(repo, noopAudit{}, fixedClock(now))
	markExecutedUC := usecases.NewMarkExecutedUseCase(repo, noopAudit{}, fixedClock(now))
	h := docHttp.NewWorkflowHandler(submitUC, approveUC, rejectUC, registerUC, startRoutingUC, signVisaUC, assignExecutorUC, markExecutedUC)

	r := gin.New()
	api := r.Group("/api")
	api.Use(withAuth(userID, role))

	submitGroup := api.Group("/documents")
	docHttp.RegisterSubmitRoute(submitGroup, h)

	adminGroup := api.Group("/admin/documents")
	adminGroup.Use(adminRoleGate())
	docHttp.RegisterAdminWorkflowRoutes(adminGroup, h)
	return r
}

// --- Submit endpoint integration ---

func TestSubmit_HappyPath(t *testing.T) {
	repo := newRepo(draftDoc(1, 42))
	r := mountWorkflow(t, 42, "teacher", repo)
	w := performRequest(r, "POST", "/api/documents/1/submit", nil)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusApproval, stored.Status)
	require.NotNil(t, stored.SubmittedBy)
	assert.Equal(t, int64(42), *stored.SubmittedBy)
}

func TestSubmit_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 42, "teacher", repo)
	w := performRequest(r, "POST", "/api/documents/999/submit", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubmit_ForbiddenForNonAuthorTeacher(t *testing.T) {
	repo := newRepo(draftDoc(1, 42))
	r := mountWorkflow(t, 99, "teacher", repo)
	w := performRequest(r, "POST", "/api/documents/1/submit", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSubmit_ConflictForNonDraft(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproved))
	r := mountWorkflow(t, 42, "teacher", repo)
	w := performRequest(r, "POST", "/api/documents/1/submit", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- Approve endpoint integration ---

func TestApprove_HappyPath(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproval))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/approve", nil)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	assert.Equal(t, entities.DocumentStatusApproved, repo.stored[1].Status)
}

func TestApprove_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 7, "system_admin", repo)
	w := performRequest(r, "POST", "/api/admin/documents/999/approve", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestApprove_ConflictForNonApprovalStatus(t *testing.T) {
	repo := newRepo(draftDoc(1, 42))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/approve", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- Reject endpoint integration ---

func TestReject_HappyPath(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproval))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]string{"reason": "Шаблон 2023 устарел — обновите за неделю"}
	w := performRequest(r, "POST", "/api/admin/documents/1/reject", body)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusRejected, stored.Status)
	require.NotNil(t, stored.RejectedReason)
}

func TestReject_BadReasonShort(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproval))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]string{"reason": "коротко"}
	w := performRequest(r, "POST", "/api/admin/documents/1/reject", body)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestReject_MissingReason(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproval))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]string{}
	w := performRequest(r, "POST", "/api/admin/documents/1/reject", body)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestReject_ConflictForNonApprovalStatus(t *testing.T) {
	repo := newRepo(draftDoc(1, 42))
	r := mountWorkflow(t, 7, "system_admin", repo)
	body := map[string]string{"reason": "Корректное обоснование отказа"}
	w := performRequest(r, "POST", "/api/admin/documents/1/reject", body)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestAdminGate_MethodistBlockedFromApprove pins the route-level
// admin gate: methodist может submit, но НЕ approve/reject. Catches
// a regression class where the admin middleware гайляется
// от /admin/documents в main.go. Defense-in-depth per
// `feedback_route_extraction_for_security_test`.
func TestAdminGate_MethodistBlockedFromApprove(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproval))
	r := mountWorkflow(t, 99, "methodist", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/approve", nil)
	assert.Equal(t, http.StatusForbidden, w.Code, "methodist must not approve via admin gate")
}

// TestAdminGate_TeacherBlockedFromReject — same defense as above для
// the reject path.
func TestAdminGate_TeacherBlockedFromReject(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproval))
	r := mountWorkflow(t, 42, "teacher", repo)
	body := map[string]string{"reason": "Корректное обоснование отказа"}
	w := performRequest(r, "POST", "/api/admin/documents/1/reject", body)
	assert.Equal(t, http.StatusForbidden, w.Code, "teacher must not reject via admin gate")
}

// --- StartRouting endpoint integration (#231) ---

func TestStartRouting_HappyPath(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRegistered))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/start-routing", nil)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusRouting, stored.Status)
	require.NotNil(t, stored.RoutedBy)
	assert.Equal(t, int64(7), *stored.RoutedBy)
}

func TestStartRouting_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 7, "system_admin", repo)
	w := performRequest(r, "POST", "/api/admin/documents/999/start-routing", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStartRouting_ConflictForNonRegistered(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproved))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/start-routing", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestAdminGate_MethodistBlockedFromStartRouting pins T1-C: methodist
// must not pass production admin gate для routing.
func TestAdminGate_MethodistBlockedFromStartRouting(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRegistered))
	r := mountWorkflow(t, 99, "methodist", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/start-routing", nil)
	assert.Equal(t, http.StatusForbidden, w.Code, "methodist must not start routing via admin gate")
}

// --- SignVisa endpoint integration (#231) ---

func TestSignVisa_HappyPath(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRouting))
	r := mountWorkflow(t, 9, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/sign-visa", nil)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusExecution, stored.Status)
	require.NotNil(t, stored.VisaSignedBy)
	assert.Equal(t, int64(9), *stored.VisaSignedBy)
}

func TestSignVisa_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 9, "system_admin", repo)
	w := performRequest(r, "POST", "/api/admin/documents/999/sign-visa", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSignVisa_ConflictForNonRouting(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRegistered))
	r := mountWorkflow(t, 9, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/sign-visa", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestAdminGate_TeacherBlockedFromSignVisa pins T1-C: teacher must
// not pass production admin gate для visa signing.
func TestAdminGate_TeacherBlockedFromSignVisa(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRouting))
	r := mountWorkflow(t, 42, "teacher", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/sign-visa", nil)
	assert.Equal(t, http.StatusForbidden, w.Code, "teacher must not sign visa via admin gate")
}

// TestStartRouting_EnvelopeContract pins T1-B: backend returns
// {"data": doc} envelope (response.Success), not raw doc. Frontend
// hook unwraps .data — regression here breaks the UI silently.
func TestStartRouting_EnvelopeContract(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRegistered))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/start-routing", nil)

	require.Equal(t, http.StatusOK, w.Code)
	body := parseResponseBody(w)
	_, ok := body["data"]
	assert.True(t, ok, "response must wrap doc in .data envelope; got keys=%v", keysOf(body))
}

// TestSignVisa_EnvelopeContract — same T1-B defense for visa path.
func TestSignVisa_EnvelopeContract(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRouting))
	r := mountWorkflow(t, 9, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/sign-visa", nil)

	require.Equal(t, http.StatusOK, w.Code)
	body := parseResponseBody(w)
	_, ok := body["data"]
	assert.True(t, ok, "response must wrap doc in .data envelope; got keys=%v", keysOf(body))
}

func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// --- AssignExecutor endpoint integration (#232) ---

func TestAssignExecutor_HappyPath_NilDue(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]interface{}{"executor_id": 13}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusExecution, stored.Status, "AssignExecutor must NOT change status")
	require.NotNil(t, stored.ExecutorAssignedTo)
	assert.Equal(t, int64(13), *stored.ExecutorAssignedTo)
}

func TestAssignExecutor_HappyPath_WithDueDate(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]interface{}{"executor_id": 13, "due_date": "2026-05-24"}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	require.NotNil(t, repo.stored[1].ExecutorDueDate)
}

func TestAssignExecutor_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 7, "system_admin", repo)
	body := map[string]interface{}{"executor_id": 13}
	w := performRequest(r, "POST", "/api/admin/documents/999/assign-executor", body)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssignExecutor_ConflictForNonExecution(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRouting))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]interface{}{"executor_id": 13}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAssignExecutor_InvalidExecutorID(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]interface{}{"executor_id": 0}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAssignExecutor_InvalidDueDateFormat(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]interface{}{"executor_id": 13, "due_date": "not-a-date"}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// TestAdminGate_MethodistBlockedFromAssignExecutor pins T1-C.
func TestAdminGate_MethodistBlockedFromAssignExecutor(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 99, "methodist", repo)
	body := map[string]interface{}{"executor_id": 13}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)
	assert.Equal(t, http.StatusForbidden, w.Code, "methodist must not assign executor via admin gate")
}

func TestAssignExecutor_EnvelopeContract(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]interface{}{"executor_id": 13}
	w := performRequest(r, "POST", "/api/admin/documents/1/assign-executor", body)

	require.Equal(t, http.StatusOK, w.Code)
	respBody := parseResponseBody(w)
	_, ok := respBody["data"]
	assert.True(t, ok, "response must wrap doc в .data envelope; got keys=%v", keysOf(respBody))
}

// --- MarkExecuted endpoint integration (#232) ---

func TestMarkExecuted_HappyPath(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/mark-executed", nil)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusExecuted, stored.Status)
	require.NotNil(t, stored.ExecutedBy)
	assert.Equal(t, int64(7), *stored.ExecutedBy)
}

func TestMarkExecuted_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 7, "system_admin", repo)
	w := performRequest(r, "POST", "/api/admin/documents/999/mark-executed", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMarkExecuted_ConflictForNonExecution(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusRouting))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/mark-executed", nil)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestAdminGate_TeacherBlockedFromMarkExecuted pins T1-C.
func TestAdminGate_TeacherBlockedFromMarkExecuted(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 42, "teacher", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/mark-executed", nil)
	assert.Equal(t, http.StatusForbidden, w.Code, "teacher must not mark executed via admin gate")
}

func TestMarkExecuted_EnvelopeContract(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusExecution))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	w := performRequest(r, "POST", "/api/admin/documents/1/mark-executed", nil)

	require.Equal(t, http.StatusOK, w.Code)
	respBody := parseResponseBody(w)
	_, ok := respBody["data"]
	assert.True(t, ok, "response must wrap doc в .data envelope; got keys=%v", keysOf(respBody))
}
