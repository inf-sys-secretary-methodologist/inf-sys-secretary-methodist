package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// Import1CDebtsPort syncs the debt registry from 1С. Unlike the xlsx import it
// takes no upload stream — its source is the 1С OData API, fetched server-side.
type Import1CDebtsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string) (sdUsecases.ImportResult, error)
}

// StudentDebt1CImportHandler serves the 1С debt-sync endpoint. It is a thin
// sibling of the transfer handler (rather than a third port on it) so the
// upload-based import/export surface stays untouched.
type StudentDebt1CImportHandler struct {
	import1C Import1CDebtsPort
}

// NewStudentDebt1CImportHandler wires the handler. The port is required; a nil
// port panics (failure-closed DI).
func NewStudentDebt1CImportHandler(import1C Import1CDebtsPort) *StudentDebt1CImportHandler {
	if import1C == nil {
		panic("student_debts: NewStudentDebt1CImportHandler requires a non-nil port")
	}
	return &StudentDebt1CImportHandler{import1C: import1C}
}

// RegisterStudentDebt1CImportRoutes mounts POST /student-debts/import-1c. The
// caller applies authentication first; the static segment takes routing
// priority over /:id.
func RegisterStudentDebt1CImportRoutes(rg *gin.RouterGroup, h *StudentDebt1CImportHandler) {
	g := rg.Group("/student-debts")
	g.POST("/import-1c", h.Import1C)
}

// Import1C handles POST /student-debts/import-1c — triggers a server-side pull
// of the 1С academic-debt catalog into the registry (EDIT_ROLES only, enforced
// in the use case). Per-row problems travel back inside ImportResult.Errors
// (still 200); a forbidden actor is 403; a 1С transport/parse failure is 502.
func (h *StudentDebt1CImportHandler) Import1C(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	result, err := h.import1C.Execute(c.Request.Context(), actorID, role)
	if err != nil {
		// Authorization is role-based and pre-fetch — a true 403. Any other
		// failure is a 1С transport/parse problem — an upstream-dependency
		// error (502), not a client or server fault.
		if errors.Is(err, entities.ErrDebtAccessForbidden) {
			c.JSON(http.StatusForbidden, response.Forbidden("not authorized to import the debt registry from 1С"))
			return
		}
		c.JSON(http.StatusBadGateway, response.ErrorResponse("UPSTREAM_1C_ERROR", "the 1С debt catalog could not be fetched"))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapImportResult(result)))
}
