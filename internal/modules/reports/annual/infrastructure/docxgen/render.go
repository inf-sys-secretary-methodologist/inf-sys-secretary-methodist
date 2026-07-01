// Package docxgen renders the annual methodist report to DOCX bytes by
// synthesizing a minimal Office Open XML package programmatically and
// applying section markup via the internal/shared/docx substitution
// helper. Zero third-party deps per ADR-2.
package docxgen

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	assignmentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	curriculumRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	documentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/docx"
)

const emptyMarker = "Нет данных за период"

// Renderer is the stateless DOCX synthesizer. Satisfies the
// AnnualReportRenderer port in the usecases package structurally.
type Renderer struct{}

// NewRenderer constructs a Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderAnnualReport assembles the annual report DOCX bytes for the
// given year and pre-aggregated section data. Output is deterministic
// (no clock dependency) so tests can assert byte equality and so the
// same input from production yields identical artifacts run-to-run.
func (r *Renderer) RenderAnnualReport(
	year int,
	curricula []curriculumRepos.CurriculumYearSpecialtyAgg,
	grades []assignmentRepos.AssignmentGradeDistributionAgg,
	hours []curriculumRepos.DisciplineItemHoursAgg,
	activity []documentRepos.DocumentActivityByTypeAgg,
) ([]byte, error) {
	template, err := buildBaseTemplate()
	if err != nil {
		return nil, fmt.Errorf("docxgen: build template: %w", err)
	}

	replacements := map[string]string{
		"{{year}}":              strconv.Itoa(year),
		"{{curricula_section}}": renderCurriculaSection(curricula),
		"{{grades_section}}":    renderGradesSection(grades),
		"{{hours_section}}":     renderHoursSection(hours),
		"{{docs_section}}":      renderDocsSection(activity),
	}

	out, err := docx.Substitute(template, replacements)
	if err != nil {
		return nil, fmt.Errorf("docxgen: substitute: %w", err)
	}
	return out, nil
}

// buildBaseTemplate emits a 3-entry DOCX package (Content Types + global
// relationships + word/document.xml with placeholders) sufficient for
// Word / LibreOffice to open. Modified time on entries is zero so the
// build is reproducible.
func buildBaseTemplate() ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	if err := writeZipEntry(w, "[Content_Types].xml", contentTypesXML); err != nil {
		return nil, err
	}
	if err := writeZipEntry(w, "_rels/.rels", packageRelsXML); err != nil {
		return nil, err
	}
	if err := writeZipEntry(w, "word/document.xml", documentXML); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("docxgen: close zip: %w", err)
	}
	return buf.Bytes(), nil
}

func writeZipEntry(w *zip.Writer, name, body string) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	f, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("docxgen: create entry %s: %w", name, err)
	}
	if _, err := f.Write([]byte(body)); err != nil {
		return fmt.Errorf("docxgen: write entry %s: %w", name, err)
	}
	return nil
}

func renderCurriculaSection(rows []curriculumRepos.CurriculumYearSpecialtyAgg) string {
	if len(rows) == 0 {
		return paragraph(emptyMarker)
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(paragraph(fmt.Sprintf("%s — %s: %d",
			r.Specialty, string(r.Status), r.Count)))
	}
	return b.String()
}

func renderGradesSection(rows []assignmentRepos.AssignmentGradeDistributionAgg) string {
	if len(rows) == 0 {
		return paragraph(emptyMarker)
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(paragraph(fmt.Sprintf("%s — %s: %d",
			r.Subject, string(r.Status), r.Count)))
	}
	return b.String()
}

func renderHoursSection(rows []curriculumRepos.DisciplineItemHoursAgg) string {
	if len(rows) == 0 {
		return paragraph(emptyMarker)
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(paragraph(fmt.Sprintf(
			"%s — лекции: %d, практика: %d, лаб: %d, самостоят: %d",
			r.CurriculumTitle, r.Lectures, r.Practice, r.Lab, r.SelfStudy)))
	}
	return b.String()
}

func renderDocsSection(rows []documentRepos.DocumentActivityByTypeAgg) string {
	if len(rows) == 0 {
		return paragraph(emptyMarker)
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(paragraph(fmt.Sprintf("%s — %s: %d",
			r.TypeName, string(r.Status), r.Count)))
	}
	return b.String()
}

// paragraph wraps escaped text into a Word paragraph. XML metacharacters
// in the data are escaped here so neither callers nor placeholder
// substitution can inject markup.
func paragraph(text string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(text))
	return `<w:p><w:r><w:t xml:space="preserve">` + buf.String() + `</w:t></w:r></w:p>`
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const packageRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const documentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:r><w:t xml:space="preserve">Годовой отчёт методиста за {{year}} год</w:t></w:r></w:p>
<w:p><w:r><w:t xml:space="preserve">1. Учебные планы</w:t></w:r></w:p>
{{curricula_section}}
<w:p><w:r><w:t xml:space="preserve">2. Распределение оценок</w:t></w:r></w:p>
{{grades_section}}
<w:p><w:r><w:t xml:space="preserve">3. Часы по дисциплинам</w:t></w:r></w:p>
{{hours_section}}
<w:p><w:r><w:t xml:space="preserve">4. Документооборот</w:t></w:r></w:p>
{{docs_section}}
</w:body>
</w:document>`
