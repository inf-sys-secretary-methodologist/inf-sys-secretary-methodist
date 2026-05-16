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
	h := docHttp.NewWorkflowHandler(submitUC, approveUC, rejectUC, registerUC)

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
