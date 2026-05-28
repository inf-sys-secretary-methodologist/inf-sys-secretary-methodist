package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// Outbound ports + DTOs for LLM-backed draft generation (ADR-7). Like
// the repository/audit ports, these live in the application layer per
// the DIP gate — the use case depends only on these narrow interfaces;
// the concrete adapters (OpenAI-compatible HTTP generator, curriculum
// discipline lookup, Redis rate limiter) are wired in main.go so the
// module never imports ai/ or curriculum/ directly.

// DraftRequest is the discipline context handed to a DraftGenerator.
// Built by GenerateDraftUseCase from the РПД aggregate plus discipline
// metadata sourced from curriculum (so generated topics fit the real
// hour budget and the prompt names the right discipline / control form).
type DraftRequest struct {
	DisciplineName     string
	SpecialtyCode      string
	ApplicableFromYear int
	HoursLecture       int
	HoursPractice      int
	HoursLab           int
	HoursSelfStudy     int
	ControlForm        string
	Annotation         string
}

// CompetenceDraft / TopicDraft / ReferenceDraft are the generator's
// untyped output rows. The use case maps them into domain entities via
// the aggregate's constructors, so a malformed row (bad enum, empty
// title) fails the whole generation rather than bypassing an invariant.
type CompetenceDraft struct {
	Code        string
	Type        string // pk | ok | uk
	Description string
}

// TopicDraft is one generated тема (lecture/practice/lab/self_study).
type TopicDraft struct {
	Kind  string // lecture | practice | lab | self_study
	Title string
	Hours int
}

// ReferenceDraft is one generated литература/источник row.
type ReferenceDraft struct {
	Kind     string // main | additional | electronic
	Citation string
}

// DraftResult is the structured content a DraftGenerator produces for
// an empty draft РПД. Annotation is intentionally NOT generated — the
// author supplies it at creation (PR 8e); generation fills the
// structured collections only.
type DraftResult struct {
	Goals       []string
	Competences []CompetenceDraft
	Topics      []TopicDraft
	References  []ReferenceDraft
}

// DisciplineInfo is curriculum-sourced metadata used to ground the
// generation prompt: the discipline name, its hour budget, and the
// control form.
type DisciplineInfo struct {
	Name           string
	HoursLecture   int
	HoursPractice  int
	HoursLab       int
	HoursSelfStudy int
	ControlForm    string
}

// DraftGenerator is the outbound port for LLM-backed draft generation.
// The concrete adapter (OpenAI-compatible HTTP client with configurable
// base_url/key/model — OpenRouter by default) is injected in main.go;
// tests substitute a deterministic fake.
type DraftGenerator interface {
	GenerateDraft(ctx context.Context, req DraftRequest) (DraftResult, error)
}

// DisciplineInfoProvider supplies discipline metadata for a discipline
// id, sourced from the curriculum module via an adapter wired in
// main.go (no cross-module import inside work_program).
type DisciplineInfoProvider interface {
	GetDisciplineInfo(ctx context.Context, disciplineID int64) (DisciplineInfo, error)
}

// GenerationRateLimiter guards the cost of LLM calls: Allow reports
// whether userID may run another generation now (5/hour/user per
// ADR-7). The concrete adapter is wired in main.go.
type GenerationRateLimiter interface {
	Allow(ctx context.Context, userID int64) (bool, error)
}

// generateDraftRepo is the narrow persistence port the generation use
// case needs: load the aggregate, then write the filled draft back.
type generateDraftRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}
