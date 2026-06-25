package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// TestHandle_ServesStudentDebtsCatalog pins the academic-debts catalog route
// added in PR7a (#431): the mock must serve ODataStudentDebt rows so the
// integration client's GetAllStudentDebts can be exercised end-to-end against
// the mock-1c harness.
func TestHandle_ServesStudentDebtsCatalog(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Catalog_АкадемическиеЗадолженности", nil)
	rec := httptest.NewRecorder()

	handle(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out odataList[entities.ODataStudentDebt]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.NotEmpty(t, out.Value, "debt catalog must return sample rows")

	first := out.Value[0]
	assert.NotEmpty(t, first.RefKey, "debt must carry a 1С GUID source reference")
	assert.NotEmpty(t, first.GroupName, "debt must reference a student group")
	assert.NotEmpty(t, first.Discipline)
	assert.NotEmpty(t, first.ControlForm)
}

// TestHandle_StudentDebtsPaginationGuard mirrors the employee/student guard:
// a $skip beyond the first page yields an empty set so the client's paginator
// terminates.
func TestHandle_StudentDebtsPaginationGuard(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Catalog_АкадемическиеЗадолженности?$skip=100", nil)
	rec := httptest.NewRecorder()

	handle(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out odataList[entities.ODataStudentDebt]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Empty(t, out.Value)
}
