package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// xlsxMIME is the OOXML spreadsheet content type used for export downloads.
const xlsxMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

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

// NewStudentDebtTransferHandler wires the handler. Both ports are required;
// a nil port panics (failure-closed DI).
func NewStudentDebtTransferHandler(importDebts ImportDebtsPort, exportDebts ExportDebtsPort) *StudentDebtTransferHandler {
	if importDebts == nil || exportDebts == nil {
		panic("student_debts: NewStudentDebtTransferHandler requires non-nil ports")
	}
	return &StudentDebtTransferHandler{importDebts: importDebts, exportDebts: exportDebts}
}

// RegisterStudentDebtTransferRoutes mounts the transfer endpoints under
// /student-debts. The caller applies authentication first. Static segments
// (/import, /export) take routing priority over /:id.
func RegisterStudentDebtTransferRoutes(rg *gin.RouterGroup, h *StudentDebtTransferHandler) {
	g := rg.Group("/student-debts")
	g.POST("/import", h.Import)
	g.GET("/export", h.Export)
}

// Import handles POST /student-debts/import — a multipart "file" upload that
// is streamed into the import use case (EDIT_ROLES only, enforced there).
// Per-row problems travel back inside ImportResult.Errors (still 200); a
// forbidden actor is 403; a malformed document is 400.
func (h *StudentDebtTransferHandler) Import(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("a multipart \"file\" field is required"))
		return
	}
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("the uploaded file could not be read"))
		return
	}
	defer func() { _ = src.Close() }()

	result, err := h.importDebts.Execute(c.Request.Context(), actorID, role, src)
	if err != nil {
		// Authorization is role-based and pre-parse — a true 403. Any other
		// failure is a malformed/unparseable document — a 400, not a 500.
		if errors.Is(err, entities.ErrDebtAccessForbidden) {
			c.JSON(http.StatusForbidden, response.Forbidden("not authorized to import the debt registry"))
			return
		}
		c.JSON(http.StatusBadRequest, response.BadRequest("the uploaded document could not be parsed"))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapImportResult(result)))
}

// Export handles GET /student-debts/export — streams the role-scoped registry
// as an xlsx attachment. The list filter (group/status/semester) is honored;
// pagination is irrelevant (the export serializes the whole matching set).
// A denial is 403; a repo/serialization failure is 500.
func (h *StudentDebtTransferHandler) Export(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	// parseListFilter clamps Limit/Offset, but the export repo ignores them —
	// only group/status/semester matter here.
	data, err := h.exportDebts.Execute(c.Request.Context(), actorID, role, parseListFilter(c))
	if err != nil {
		if errors.Is(err, entities.ErrDebtAccessForbidden) {
			c.JSON(http.StatusForbidden, response.Forbidden("not authorized to export the debt registry"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.InternalError("internal error"))
		return
	}

	c.Header("Content-Disposition", `attachment; filename="student-debts.xlsx"`)
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Data(http.StatusOK, xlsxMIME, data)
}
