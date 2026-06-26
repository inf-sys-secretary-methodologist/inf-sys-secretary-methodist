// Package handlers exposes the annual methodist report HTTP surface.
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/application/usecases"
)

const (
	minYear  = 2000
	maxYear  = 2100
	mimeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

// allowedRoles are the only role values that may pull the annual report.
// Methodist owns the curriculum + assignments lifecycle; system_admin =
// privileged superset for audit / debugging access. Academic secretary
// is read-only observer (no decision-maker authority) и excluded.
var allowedRoles = map[string]bool{
	"methodist":    true,
	"system_admin": true,
}

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

// Generate handles `GET /api/reports/annual?year=YYYY` and streams the
// rendered DOCX as a file download.
//
// Response codes:
//   - 200 + DOCX body for methodist / system_admin with valid year
//   - 401 if user_id missing from auth context
//   - 403 if role missing / not in allowedRoles
//   - 422 if year missing / non-numeric / outside [2000, 2100]
//   - 500 if the use case fails (aggregate / render path)
//
// @Summary Generate annual methodist report
// @Description Returns a DOCX file aggregating curricula / grades / hours
// @Description and document activity for the given calendar year.
// @Tags reports
// @Security BearerAuth
// @Produce application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Param year query int true "Calendar year (2000-2100)"
// @Success 200 {file} binary "DOCX bytes"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/reports/annual [get]
func (h *AnnualReportHandler) Generate(c *gin.Context) {
	actorID, ok := actorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	roleRaw, exists := c.Get("role")
	roleStr, _ := roleRaw.(string)
	if !exists || !allowedRoles[roleStr] {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	year, ok := parseYear(c.Query("year"))
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "year must be an integer in [2000, 2100]"})
		return
	}

	docxBytes, err := h.generate.Generate(c.Request.Context(), usecases.GenerateAnnualReportInput{
		Year:    year,
		ActorID: actorID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate annual report"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="annual_report_%d.docx"`, year))
	c.Data(http.StatusOK, mimeDOCX, docxBytes)
}

func actorIDFromContext(c *gin.Context) (int64, bool) {
	raw, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := raw.(int64)
	if !ok || id == 0 {
		return 0, false
	}
	return id, true
}

func parseYear(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	year, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if year < minYear || year > maxYear {
		return 0, false
	}
	return year, true
}
