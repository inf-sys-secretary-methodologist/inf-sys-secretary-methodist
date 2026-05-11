// Package docxgen renders the annual methodist report to DOCX bytes by
// synthesizing a minimal Office Open XML package programmatically and
// applying section markup via the internal/shared/docx substitution
// helper. Zero third-party deps per ADR-2.
package docxgen

import (
	"errors"

	assignmentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	curriculumRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	documentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// Renderer is the stateless DOCX synthesizer. Satisfies the
// AnnualReportRenderer port in the usecases package structurally.
type Renderer struct{}

// NewRenderer constructs a Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderAnnualReport assembles the annual report DOCX bytes for the
// given year and pre-aggregated section data. Implementation deferred
// to GREEN.
func (r *Renderer) RenderAnnualReport(
	_ int,
	_ []curriculumRepos.CurriculumYearSpecialtyAgg,
	_ []assignmentRepos.AssignmentGradeDistributionAgg,
	_ []curriculumRepos.DisciplineItemHoursAgg,
	_ []documentRepos.DocumentActivityByTypeAgg,
) ([]byte, error) {
	return nil, errors.New("docxgen: render annual report not implemented")
}
