package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ImportDebtsPort ingests a registry document (xlsx now, 1С later) supplied
// as a multipart upload stream.
type ImportDebtsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, src io.Reader) (sdUsecases.ImportResult, error)
}

// ExportDebtsPort serializes the role-scoped registry into a downloadable
// document.
type ExportDebtsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) ([]byte, error)
}

// StudentDebtTransferHandler serves the bulk import/export endpoints.
type StudentDebtTransferHandler struct {
	importDebts ImportDebtsPort
	exportDebts ExportDebtsPort
}

// NewStudentDebtTransferHandler wires the handler. Both ports are required.
func NewStudentDebtTransferHandler(importDebts ImportDebtsPort, exportDebts ExportDebtsPort) *StudentDebtTransferHandler {
	if importDebts == nil || exportDebts == nil {
		panic("student_debts: NewStudentDebtTransferHandler requires non-nil ports")
	}
	return &StudentDebtTransferHandler{importDebts: importDebts, exportDebts: exportDebts}
}

// RegisterStudentDebtTransferRoutes mounts the transfer endpoints under
// /student-debts. Static segments take routing priority over /:id.
func RegisterStudentDebtTransferRoutes(rg *gin.RouterGroup, h *StudentDebtTransferHandler) {
	g := rg.Group("/student-debts")
	g.POST("/import", h.Import)
	g.GET("/export", h.Export)
}

// Import handles POST /student-debts/import (stub — implemented in GREEN).
func (h *StudentDebtTransferHandler) Import(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response.InternalError("not implemented"))
}

// Export handles GET /student-debts/export (stub — implemented in GREEN).
func (h *StudentDebtTransferHandler) Export(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response.InternalError("not implemented"))
}
