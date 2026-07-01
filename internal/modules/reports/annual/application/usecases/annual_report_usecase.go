package usecases

import (
	"context"
	"fmt"
	"time"

	assignmentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	curriculumRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	documentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// curriculumAggregateRepo is the narrow port for AggregateByYearSpecialty.
// CurriculumRepositoryPG satisfies it structurally.
type curriculumAggregateRepo interface {
	AggregateByYearSpecialty(ctx context.Context, year int) ([]curriculumRepos.CurriculumYearSpecialtyAgg, error)
}

// assignmentAggregateRepo is the narrow port for AggregateGradeDistribution.
type assignmentAggregateRepo interface {
	AggregateGradeDistribution(ctx context.Context, from, to time.Time) ([]assignmentRepos.AssignmentGradeDistributionAgg, error)
}

// disciplineItemAggregateRepo is the narrow port for AggregateHoursByYear.
type disciplineItemAggregateRepo interface {
	AggregateHoursByYear(ctx context.Context, year int) ([]curriculumRepos.DisciplineItemHoursAgg, error)
}

// documentAggregateRepo is the narrow port for AggregateActivityByType.
type documentAggregateRepo interface {
	AggregateActivityByType(ctx context.Context, from, to time.Time) ([]documentRepos.DocumentActivityByTypeAgg, error)
}

// AnnualReportRenderer is the port for DOCX rendering. Concrete impl
// in infrastructure/docxgen. Defined в consumer package per Clean
// Architecture DIP — renderer is a side-effect adapter to the usecase.
type AnnualReportRenderer interface {
	RenderAnnualReport(
		year int,
		curricula []curriculumRepos.CurriculumYearSpecialtyAgg,
		grades []assignmentRepos.AssignmentGradeDistributionAgg,
		hours []curriculumRepos.DisciplineItemHoursAgg,
		activity []documentRepos.DocumentActivityByTypeAgg,
	) ([]byte, error)
}

// GenerateAnnualReportInput is the public DTO. Year — calendar year
// (ADR-4); ActorID — authenticated caller used for audit forensic trail.
type GenerateAnnualReportInput struct {
	Year    int
	ActorID int64
}

// AnnualReportUseCase orchestrates the 4 aggregate fetches + DOCX render
// + audit emit. Read-only: no writes to any data store.
type AnnualReportUseCase struct {
	curriculumRepo curriculumAggregateRepo
	assignmentRepo assignmentAggregateRepo
	itemRepo       disciplineItemAggregateRepo
	documentRepo   documentAggregateRepo
	renderer       AnnualReportRenderer
	audit          AuditSink
}

// NewAnnualReportUseCase wires the use case. All 5 producer-side
// dependencies are required (non-nil); audit is optional.
func NewAnnualReportUseCase(
	curriculumRepo curriculumAggregateRepo,
	assignmentRepo assignmentAggregateRepo,
	itemRepo disciplineItemAggregateRepo,
	documentRepo documentAggregateRepo,
	renderer AnnualReportRenderer,
	audit AuditSink,
) *AnnualReportUseCase {
	if curriculumRepo == nil {
		panic("annual_report: NewAnnualReportUseCase requires non-nil curriculumRepo")
	}
	if assignmentRepo == nil {
		panic("annual_report: NewAnnualReportUseCase requires non-nil assignmentRepo")
	}
	if itemRepo == nil {
		panic("annual_report: NewAnnualReportUseCase requires non-nil itemRepo")
	}
	if documentRepo == nil {
		panic("annual_report: NewAnnualReportUseCase requires non-nil documentRepo")
	}
	if renderer == nil {
		panic("annual_report: NewAnnualReportUseCase requires non-nil renderer")
	}
	return &AnnualReportUseCase{
		curriculumRepo: curriculumRepo,
		assignmentRepo: assignmentRepo,
		itemRepo:       itemRepo,
		documentRepo:   documentRepo,
		renderer:       renderer,
		audit:          audit,
	}
}

// Generate runs the orchestration: 4 aggregate fetches → DOCX render →
// audit emit. The audit emit is best-effort and only fires after
// rendering succeeds (forensic trail tracks completed generations only).
// Time window is half-open [Y-01-01, (Y+1)-01-01) UTC per ADR-4.
func (uc *AnnualReportUseCase) Generate(ctx context.Context, in GenerateAnnualReportInput) ([]byte, error) {
	from := time.Date(in.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(in.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	curricula, err := uc.curriculumRepo.AggregateByYearSpecialty(ctx, in.Year)
	if err != nil {
		return nil, fmt.Errorf("annual_report: curricula aggregate: %w", err)
	}

	grades, err := uc.assignmentRepo.AggregateGradeDistribution(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("annual_report: grades aggregate: %w", err)
	}

	hours, err := uc.itemRepo.AggregateHoursByYear(ctx, in.Year)
	if err != nil {
		return nil, fmt.Errorf("annual_report: hours aggregate: %w", err)
	}

	activity, err := uc.documentRepo.AggregateActivityByType(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("annual_report: documents activity aggregate: %w", err)
	}

	docxBytes, err := uc.renderer.RenderAnnualReport(in.Year, curricula, grades, hours, activity)
	if err != nil {
		return nil, fmt.Errorf("annual_report: render: %w", err)
	}

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "report.annual_generated", "report", map[string]any{
			"year":          in.Year,
			"actor_user_id": in.ActorID,
		})
	}
	return docxBytes, nil
}
