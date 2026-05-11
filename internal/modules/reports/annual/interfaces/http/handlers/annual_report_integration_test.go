package handlers_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	assignmentEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	assignmentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	curriculumEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	curriculumRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	documentEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	documentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/infrastructure/docxgen"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/interfaces/http/handlers"
)

// --- aggregate fakes for the real usecase --------------------------------

type fakeCurriculumAgg struct {
	rows []curriculumRepos.CurriculumYearSpecialtyAgg
}

func (f *fakeCurriculumAgg) AggregateByYearSpecialty(_ context.Context, _ int) ([]curriculumRepos.CurriculumYearSpecialtyAgg, error) {
	return f.rows, nil
}

type fakeAssignmentAgg struct {
	rows []assignmentRepos.AssignmentGradeDistributionAgg
}

func (f *fakeAssignmentAgg) AggregateGradeDistribution(_ context.Context, _, _ time.Time) ([]assignmentRepos.AssignmentGradeDistributionAgg, error) {
	return f.rows, nil
}

type fakeHoursAgg struct {
	rows []curriculumRepos.DisciplineItemHoursAgg
}

func (f *fakeHoursAgg) AggregateHoursByYear(_ context.Context, _ int) ([]curriculumRepos.DisciplineItemHoursAgg, error) {
	return f.rows, nil
}

type fakeDocsAgg struct {
	rows []documentRepos.DocumentActivityByTypeAgg
}

func (f *fakeDocsAgg) AggregateActivityByType(_ context.Context, _, _ time.Time) ([]documentRepos.DocumentActivityByTypeAgg, error) {
	return f.rows, nil
}

type recordingAudit struct {
	calls []map[string]any
}

func (r *recordingAudit) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	r.calls = append(r.calls, map[string]any{"action": action, "resource": resource, "fields": fields})
}

// TestAnnualReport_Integration_HappyPath_ProducesValidDocx wires the
// FULL chain: gin router → handler → real AnnualReportUseCase → fake
// aggregate repos (canned rows) → REAL docxgen.Renderer → DOCX response
// bytes, served over httptest.NewServer for a true HTTP roundtrip.
//
// Asserts:
//  1. status 200 + correct DOCX content-type + filename header
//  2. response body parses as a valid OOXML zip package
//  3. word/document.xml is present, contains every section header +
//     every rendered aggregate cell value, AND XML metacharacters in
//     fixture data (& < > ") are emitted as escaped entities — never
//     verbatim — proving the escape-safe pipeline through paragraph
//     builder + substitution helper + gin response writer
//  4. audit event report.annual_generated fired exactly once with the
//     expected actor / year fields
//
// Closes Tier 3.1 deferred from v0.129.0 reviewer round
// (manual Word verification dependency → automated post-response zip
// validation).
func TestAnnualReport_Integration_HappyPath_ProducesValidDocx(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Fixtures intentionally carry XML metacharacters (& < > ") in one
	// section each so the test fails if any layer (renderer paragraph
	// builder, substitution helper, gin Data writer) regresses on escape
	// safety. The assertions below check that the escaped forms (&amp; /
	// &lt; / &gt;) appear in word/document.xml — never the raw metachars
	// inside the rendered text runs.
	const (
		year         = 2026
		actorID      = int64(42)
		specialty    = "09.02.07 Информационные системы"
		subject      = `Алгоритмы & <структуры "данных">` // <- XML metachars
		curriculumTl = "Учебный план ИС-21"
		docType      = "Приказ"
	)

	curricula := &fakeCurriculumAgg{rows: []curriculumRepos.CurriculumYearSpecialtyAgg{
		{Specialty: specialty, Status: curriculumEntities.StatusApproved, Count: 3},
	}}
	grades := &fakeAssignmentAgg{rows: []assignmentRepos.AssignmentGradeDistributionAgg{
		{Subject: subject, Status: assignmentEntities.StatusGraded, Count: 12},
	}}
	hours := &fakeHoursAgg{rows: []curriculumRepos.DisciplineItemHoursAgg{
		{CurriculumID: 1, CurriculumTitle: curriculumTl, Lectures: 36, Practice: 18, Lab: 12, SelfStudy: 24},
	}}
	docs := &fakeDocsAgg{rows: []documentRepos.DocumentActivityByTypeAgg{
		{TypeName: docType, Status: documentEntities.DocumentStatusApproved, Count: 7},
	}}
	audit := &recordingAudit{}

	uc := usecases.NewAnnualReportUseCase(curricula, grades, hours, docs, docxgen.NewRenderer(), audit)
	h := handlers.NewAnnualReportHandler(uc)

	r := gin.New()
	r.Use(withAuth(actorID, "methodist"))
	r.GET("/api/reports/annual", h.Generate)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/reports/annual?year=2026")
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", resp.Header.Get("Content-Type"))
	require.Equal(t, `attachment; filename="annual_report_2026.docx"`, resp.Header.Get("Content-Disposition"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body, "response body must not be empty")

	// (2) DOCX = zip package — parse it.
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	require.NoError(t, err, "response body must parse as a zip archive")

	// (3) word/document.xml must be inside and contain section markup +
	//     rendered aggregate cell values (escape-safe rendering preserved).
	docXML := readZipEntry(t, zr, "word/document.xml")
	require.NotEmpty(t, docXML, "word/document.xml must be present and non-empty")

	require.Contains(t, docXML, "Годовой отчёт методиста за 2026 год")
	require.Contains(t, docXML, "1. Учебные планы")
	require.Contains(t, docXML, "2. Распределение оценок")
	require.Contains(t, docXML, "3. Часы по дисциплинам")
	require.Contains(t, docXML, "4. Документооборот")

	require.Contains(t, docXML, specialty)
	require.Contains(t, docXML, curriculumTl)
	require.Contains(t, docXML, docType)

	// Escape-safe pipeline check: raw subject contains & < > " which must
	// be emitted as XML entities (&amp; / &lt; / &gt; / &#34;), and the
	// raw metacharacters must NOT survive verbatim inside a <w:t> run.
	const escapedSubject = `Алгоритмы &amp; &lt;структуры &#34;данных&#34;&gt;`
	require.Contains(t, docXML, escapedSubject,
		"subject text must be XML-escaped before reaching word/document.xml")
	require.NotContains(t, docXML, subject,
		"raw &/</>/\" must never leak verbatim into the rendered DOCX body")

	// (4) audit event fired once with expected fields.
	require.Len(t, audit.calls, 1, "report.annual_generated must fire exactly once on success")
	require.Equal(t, "report.annual_generated", audit.calls[0]["action"])
	require.Equal(t, "report", audit.calls[0]["resource"])
	fields, ok := audit.calls[0]["fields"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, year, fields["year"])
	require.Equal(t, actorID, fields["actor_user_id"])
}

// readZipEntry pulls the contents of name from zr or fails the test.
func readZipEntry(t *testing.T, zr *zip.Reader, name string) string {
	t.Helper()
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		require.NoError(t, err)
		defer func() { _ = rc.Close() }()
		var buf bytes.Buffer
		_, err = io.Copy(&buf, rc)
		require.NoError(t, err)
		return buf.String()
	}
	t.Fatalf("zip entry %q not found; entries: %s", name, zipEntryNames(zr))
	return ""
}

func zipEntryNames(zr *zip.Reader) string {
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return strings.Join(names, ", ")
}
