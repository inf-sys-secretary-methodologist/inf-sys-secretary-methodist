package http_test

// v0.153.9 Phase 6 backfill — closes WorkflowHandler.Register (0% → covered).
// Mirrors existing workflow_handler_test.go fixtures (mountWorkflow,
// fakeWorkflowRepo, docAtStatus, performRequest, withAuth).
// No production change.

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

func TestRegister_HappyPath(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproved))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]string{"number": "01-15/2026"}
	w := performRequest(r, "POST", "/api/admin/documents/1/register", body)

	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	stored := repo.stored[1]
	assert.Equal(t, entities.DocumentStatusRegistered, stored.Status)
	require.NotNil(t, stored.RegistrationNumber)
	assert.Equal(t, "01-15/2026", *stored.RegistrationNumber)
}

func TestRegister_NotFound(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 7, "system_admin", repo)
	body := map[string]string{"number": "01-15/2026"}
	w := performRequest(r, "POST", "/api/admin/documents/999/register", body)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegister_InvalidNumber_Returns422(t *testing.T) {
	// Number too short → entity rejects with ErrInvalidRegistrationNumber.
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproved))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	body := map[string]string{"number": "ab"}
	w := performRequest(r, "POST", "/api/admin/documents/1/register", body)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_ConflictForNonApprovedStatus(t *testing.T) {
	repo := newRepo(draftDoc(1, 42))
	r := mountWorkflow(t, 7, "system_admin", repo)
	body := map[string]string{"number": "01-15/2026"}
	w := performRequest(r, "POST", "/api/admin/documents/1/register", body)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_InvalidJSONBody(t *testing.T) {
	repo := newRepo(docAtStatus(1, 42, entities.DocumentStatusApproved))
	r := mountWorkflow(t, 7, "academic_secretary", repo)
	// String passed as body — fails to unmarshal к struct → 422.
	w := performRequest(r, "POST", "/api/admin/documents/1/register", "not-json")
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_InvalidDocID(t *testing.T) {
	repo := newRepo()
	r := mountWorkflow(t, 7, "system_admin", repo)
	body := map[string]string{"number": "01-15/2026"}
	w := performRequest(r, "POST", "/api/admin/documents/abc/register", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
