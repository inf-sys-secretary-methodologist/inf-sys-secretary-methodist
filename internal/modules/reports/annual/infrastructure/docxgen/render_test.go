package docxgen_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	assignmentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	assignmentEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	curriculumEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	curriculumRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	documentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	documentEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/infrastructure/docxgen"
)

func readDocumentXML(t *testing.T, docxBytes []byte) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	require.NoError(t, err)
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		require.NoError(t, err)
		defer func() { _ = rc.Close() }()
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		return string(b)
	}
	t.Fatalf("word/document.xml not found in output")
	return ""
}

func TestRenderer_RenderAnnualReport_HappyPath(t *testing.T) {
	r := docxgen.NewRenderer()

	curricula := []curriculumRepos.CurriculumYearSpecialtyAgg{
		{Specialty: "Информатика и вычислительная техника", Status: curriculumEntities.StatusApproved, Count: 3},
		{Specialty: "Прикладная информатика", Status: curriculumEntities.StatusDraft, Count: 1},
	}
	grades := []assignmentRepos.AssignmentGradeDistributionAgg{
		{Subject: "Алгоритмы", Status: assignmentEntities.StatusGraded, Count: 12},
	}
	hours := []curriculumRepos.DisciplineItemHoursAgg{
		{CurriculumID: 1, CurriculumTitle: "ИВТ-2026", Lectures: 64, Practice: 32, Lab: 16, SelfStudy: 88},
	}
	activity := []documentRepos.DocumentActivityByTypeAgg{
		{TypeName: "Приказ", Status: documentEntities.DocumentStatusApproved, Count: 5},
	}

	out, err := r.RenderAnnualReport(2026, curricula, grades, hours, activity)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	zr, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err, "output must be a valid ZIP archive")

	entryNames := make(map[string]bool, len(zr.File))
	for _, f := range zr.File {
		entryNames[f.Name] = true
	}
	require.True(t, entryNames["[Content_Types].xml"], "DOCX must contain [Content_Types].xml")
	require.True(t, entryNames["_rels/.rels"], "DOCX must contain _rels/.rels")
	require.True(t, entryNames["word/document.xml"], "DOCX must contain word/document.xml")

	xmlBody := readDocumentXML(t, out)
	require.Contains(t, xmlBody, "2026", "year must appear in title")
	require.Contains(t, xmlBody, "Информатика и вычислительная техника", "curriculum specialty must appear")
	require.Contains(t, xmlBody, "Алгоритмы", "assignment subject must appear")
	require.Contains(t, xmlBody, "ИВТ-2026", "curriculum hours title must appear")
	require.Contains(t, xmlBody, "Приказ", "document type name must appear")
}

func TestRenderer_RenderAnnualReport_AllEmpty_HasFallbackMarkers(t *testing.T) {
	r := docxgen.NewRenderer()

	out, err := r.RenderAnnualReport(2020, nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	xmlBody := readDocumentXML(t, out)
	require.Contains(t, xmlBody, "2020", "year must appear even when all sections empty")
	require.Equal(t, 4, strings.Count(xmlBody, "Нет данных за период"),
		"each of 4 sections must render fallback marker when empty (ADR-11)")
}

func TestRenderer_RenderAnnualReport_DeterministicOutput(t *testing.T) {
	r := docxgen.NewRenderer()

	out1, err := r.RenderAnnualReport(2026, nil, nil, nil, nil)
	require.NoError(t, err)
	out2, err := r.RenderAnnualReport(2026, nil, nil, nil, nil)
	require.NoError(t, err)

	require.Equal(t, out1, out2, "same input must produce byte-identical output (deterministic build)")
}

func TestRenderer_RenderAnnualReport_NoXMLInjectionThroughData(t *testing.T) {
	r := docxgen.NewRenderer()

	curricula := []curriculumRepos.CurriculumYearSpecialtyAgg{
		{Specialty: "<script>boom</script>", Status: curriculumEntities.StatusDraft, Count: 1},
	}

	out, err := r.RenderAnnualReport(2026, curricula, nil, nil, nil)
	require.NoError(t, err)

	xmlBody := readDocumentXML(t, out)
	require.NotContains(t, xmlBody, "<script>", "user data with XML metacharacters must be escaped")
	require.Contains(t, xmlBody, "&lt;script&gt;", "escaped form must be present")
}
