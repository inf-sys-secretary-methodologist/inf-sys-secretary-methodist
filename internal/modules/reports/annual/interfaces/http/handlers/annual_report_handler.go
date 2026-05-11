// Package handlers exposes the annual methodist report HTTP surface.
package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/application/usecases"
)

// GenerateAnnualReportPort is the narrow port the handler depends on —
// concrete *usecases.AnnualReportUseCase satisfies it structurally.
type GenerateAnnualReportPort interface {
	Generate(ctx context.Context, in usecases.GenerateAnnualReportInput) ([]byte, error)
}

// AnnualReportHandler routes the annual report endpoint.
type AnnualReportHandler struct {
	generate GenerateAnnualReportPort
}

// NewAnnualReportHandler wires the handler. generate must be non-nil.
func NewAnnualReportHandler(generate GenerateAnnualReportPort) *AnnualReportHandler {
	if generate == nil {
		panic("reports/annual: NewAnnualReportHandler requires non-nil generate port")
	}
	return &AnnualReportHandler{generate: generate}
}

// Generate handles `GET /api/reports/annual?year=YYYY` and streams DOCX
// bytes back. Implementation deferred to GREEN.
func (h *AnnualReportHandler) Generate(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
