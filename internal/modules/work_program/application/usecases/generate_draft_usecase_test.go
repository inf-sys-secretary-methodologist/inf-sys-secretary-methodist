package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// --- test doubles ---

type fakeGenerateRepo struct {
	wp          *entities.WorkProgram
	getErr      error
	updateErr   error
	updated     *entities.WorkProgram
	getCalls    int
	updateCalls int
}

func (f *fakeGenerateRepo) GetByID(_ context.Context, _ int64) (*entities.WorkProgram, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.wp, nil
}

func (f *fakeGenerateRepo) Update(_ context.Context, wp *entities.WorkProgram) error {
	f.updateCalls++
	f.updated = wp
	return f.updateErr
}

type fakeDraftGenerator struct {
	result DraftResult
	err    error
	gotReq DraftRequest
	calls  int
}

func (f *fakeDraftGenerator) GenerateDraft(_ context.Context, req DraftRequest) (DraftResult, error) {
	f.calls++
	f.gotReq = req
	if f.err != nil {
		return DraftResult{}, f.err
	}
	return f.result, nil
}

type fakeDisciplineProvider struct {
	info  DisciplineInfo
	err   error
	gotID int64
	calls int
}

func (f *fakeDisciplineProvider) GetDisciplineInfo(_ context.Context, id int64) (DisciplineInfo, error) {
	f.calls++
	f.gotID = id
	if f.err != nil {
		return DisciplineInfo{}, f.err
	}
	return f.info, nil
}

type fakeRateLimiter struct {
	allowed bool
	err     error
	gotUser int64
	calls   int
}

func (f *fakeRateLimiter) Allow(_ context.Context, userID int64) (bool, error) {
	f.calls++
	f.gotUser = userID
	if f.err != nil {
		return false, f.err
	}
	return f.allowed, nil
}

func allowingLimiter() *fakeRateLimiter { return &fakeRateLimiter{allowed: true} }

func sampleResult() DraftResult {
	return DraftResult{
		Goals: []string{"Сформировать навыки проектирования БД", "Изучить SQL"},
		Competences: []CompetenceDraft{
			{Code: "ПК-1", Type: "pk", Description: "Способен проектировать БД"},
			{Code: "УК-2", Type: "uk", Description: "Способен работать с данными"},
		},
		Topics: []TopicDraft{
			{Kind: "lecture", Title: "Реляционная модель", Hours: 4},
			{Kind: "practice", Title: "Нормализация", Hours: 6},
		},
		References: []ReferenceDraft{
			{Kind: "main", Citation: "Дейт К. Введение в системы баз данных"},
		},
		Assessments: []AssessmentDraft{
			{Type: "current", Description: "Контрольная работа по нормализации", MaxScore: 30,
				ExampleQuestions: []string{"Приведите отношение к 3НФ"}},
			{Type: "final", Description: "Экзамен по курсу", MaxScore: 70},
		},
	}
}

func sampleDisciplineInfo() DisciplineInfo {
	return DisciplineInfo{
		Name:           "Базы данных и СУБД",
		HoursLecture:   32,
		HoursPractice:  48,
		HoursLab:       16,
		HoursSelfStudy: 24,
		ControlForm:    "экзамен",
	}
}

// --- tests ---

func TestNewGenerateDraftUseCase_PanicsOnNilDeps(t *testing.T) {
	repo := &fakeGenerateRepo{}
	gen := &fakeDraftGenerator{}
	disc := &fakeDisciplineProvider{}
	lim := allowingLimiter()

	cases := []struct {
		name string
		call func()
	}{
		{"nil repo", func() { NewGenerateDraftUseCase(nil, gen, disc, lim, nil) }},
		{"nil generator", func() { NewGenerateDraftUseCase(repo, nil, disc, lim, nil) }},
		{"nil disciplines", func() { NewGenerateDraftUseCase(repo, gen, nil, lim, nil) }},
		{"nil limiter", func() { NewGenerateDraftUseCase(repo, gen, disc, nil, nil) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("%s: expected panic on nil dependency", tc.name)
				}
			}()
			tc.call()
		})
	}
}

func TestGenerateDraftUseCase_RoleGateDenied(t *testing.T) {
	cases := []struct {
		name string
		role string
	}{
		{"student", "student"},
		{"academic_secretary", "academic_secretary"},
		{"unknown_role", "guest"},
		{"empty_role", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeGenerateRepo{}
			lim := allowingLimiter()
			audit := &recordingAuditSink{}
			uc := NewGenerateDraftUseCase(repo, &fakeDraftGenerator{}, &fakeDisciplineProvider{}, lim, audit)

			_, err := uc.Execute(context.Background(), 7, tc.role, 100)
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
				"role %q must be denied, got %v", tc.role, err)
			assert.Zero(t, repo.getCalls, "denied role must not hit repo")
			assert.Zero(t, lim.calls, "denied role must not consume rate budget")
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.generate_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
			assert.Equal(t, int64(7), audit.events[0].Fields["actor_user_id"])
		})
	}
}

func TestGenerateDraftUseCase_RateLimited(t *testing.T) {
	repo := &fakeGenerateRepo{}
	lim := &fakeRateLimiter{allowed: false}
	audit := &recordingAuditSink{}
	uc := NewGenerateDraftUseCase(repo, &fakeDraftGenerator{}, &fakeDisciplineProvider{}, lim, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", 100)
	assert.True(t, errors.Is(err, domain.ErrGenerationRateLimited), "got %v", err)
	assert.Equal(t, int64(7), lim.gotUser)
	assert.Zero(t, repo.getCalls, "rate-limited request must not hit repo")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.generate_denied", audit.events[0].Action)
	assert.Equal(t, "rate_limited", audit.events[0].Fields["reason"])
}

func TestGenerateDraftUseCase_RateLimiterErrorPropagates(t *testing.T) {
	sentinel := errors.New("redis down")
	repo := &fakeGenerateRepo{}
	lim := &fakeRateLimiter{err: sentinel}
	uc := NewGenerateDraftUseCase(repo, &fakeDraftGenerator{}, &fakeDisciplineProvider{}, lim, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), 7, "teacher", 100)
	assert.ErrorIs(t, err, sentinel)
	assert.Zero(t, repo.getCalls)
}

func TestGenerateDraftUseCase_NotFoundPropagatesWithoutAudit(t *testing.T) {
	repo := &fakeGenerateRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewGenerateDraftUseCase(repo, &fakeDraftGenerator{}, &fakeDisciplineProvider{}, allowingLimiter(), audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", 100)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	assert.Empty(t, audit.events, "not-found must not audit (ID typos / race deletes are common)")
}

func TestGenerateDraftUseCase_OwnershipIDORCollapse(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name    string
		actorID int64
		role    string
	}{
		{"teacher_not_author", 99, "teacher"},
		{"methodist_not_author", 11, "methodist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
			repo := &fakeGenerateRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewGenerateDraftUseCase(repo, &fakeDraftGenerator{}, &fakeDisciplineProvider{}, allowingLimiter(), audit)

			_, err := uc.Execute(context.Background(), tc.actorID, tc.role, 100)
			assert.True(t, errors.Is(err, repositories.ErrWorkProgramNotFound),
				"non-author must collapse to NotFound (IDOR), got %v", err)
			assert.Zero(t, repo.updateCalls, "must not persist on ownership denial")
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.generate_denied", audit.events[0].Action)
			assert.Equal(t, "not_owner", audit.events[0].Fields["reason"])
		})
	}
}

func TestGenerateDraftUseCase_FrozenStatusDenied(t *testing.T) {
	const authorID = int64(7)
	for _, st := range []domain.Status{
		domain.StatusPendingApproval, domain.StatusApproved, domain.StatusArchived,
	} {
		t.Run(string(st), func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, st)
			repo := &fakeGenerateRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewGenerateDraftUseCase(repo, &fakeDraftGenerator{}, &fakeDisciplineProvider{}, allowingLimiter(), audit)

			_, err := uc.Execute(context.Background(), authorID, "teacher", 100)
			assert.True(t, errors.Is(err, domain.ErrCannotEditFrozenStatus), "status %s got %v", st, err)
			assert.Zero(t, repo.updateCalls, "must not persist on frozen status")
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.generate_denied", audit.events[0].Action)
			assert.Equal(t, "frozen_status", audit.events[0].Fields["reason"])
		})
	}
}

func TestGenerateDraftUseCase_HappyPath(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name    string
		actorID int64
		role    string
		status  domain.Status
	}{
		{"author_teacher_draft", authorID, "teacher", domain.StatusDraft},
		{"author_teacher_needs_revision", authorID, "teacher", domain.StatusNeedsRevision},
		{"system_admin_override", 999, "system_admin", domain.StatusDraft},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, tc.status)
			repo := &fakeGenerateRepo{wp: wp}
			gen := &fakeDraftGenerator{result: sampleResult()}
			disc := &fakeDisciplineProvider{info: sampleDisciplineInfo()}
			lim := allowingLimiter()
			audit := &recordingAuditSink{}
			uc := NewGenerateDraftUseCase(repo, gen, disc, lim, audit)

			got, err := uc.Execute(context.Background(), tc.actorID, tc.role, 100)
			require.NoError(t, err)
			require.NotNil(t, got)

			// discipline lookup keyed on the WP's discipline id (fixture = 7)
			assert.Equal(t, int64(7), disc.gotID)
			// request grounded in curriculum info, not the WP title
			assert.Equal(t, "Базы данных и СУБД", gen.gotReq.DisciplineName)
			assert.Equal(t, "09.03.01", gen.gotReq.SpecialtyCode)
			assert.Equal(t, 2026, gen.gotReq.ApplicableFromYear)
			assert.Equal(t, 32, gen.gotReq.HoursLecture)
			assert.Equal(t, 48, gen.gotReq.HoursPractice)
			assert.Equal(t, "экзамен", gen.gotReq.ControlForm)

			goals := got.Goals()
			require.Len(t, goals, 2)
			assert.Equal(t, "Сформировать навыки проектирования БД", goals[0].Text())
			assert.Equal(t, 0, goals[0].OrderIndex())
			assert.Equal(t, 1, goals[1].OrderIndex())

			comps := got.Competences()
			require.Len(t, comps, 2)
			assert.Equal(t, "ПК-1", comps[0].Code())
			assert.Equal(t, domain.CompetenceTypePK, comps[0].Type())
			assert.Equal(t, "УК-2", comps[1].Code(), "competences keep emitted order by slice position")
			assert.Equal(t, domain.CompetenceTypeUK, comps[1].Type())

			topics := got.Topics()
			require.Len(t, topics, 2)
			assert.Equal(t, domain.TopicKindLecture, topics[0].Kind())
			assert.Equal(t, "Реляционная модель", topics[0].Title())
			assert.Equal(t, 4, topics[0].Hours())
			assert.Equal(t, 0, topics[0].OrderIndex())
			assert.Equal(t, domain.TopicKindPractice, topics[1].Kind())
			assert.Equal(t, 1, topics[1].OrderIndex())

			refs := got.References()
			require.Len(t, refs, 1)
			assert.Equal(t, domain.ReferenceKindMain, refs[0].Kind())
			assert.Equal(t, "Дейт К. Введение в системы баз данных", refs[0].Citation())

			asmts := got.Assessments()
			require.Len(t, asmts, 2, "ФОС items must be generated into the draft")
			assert.Equal(t, domain.AssessmentTypeCurrent, asmts[0].Type())
			assert.Equal(t, "Контрольная работа по нормализации", asmts[0].Description())
			assert.Equal(t, 30, asmts[0].MaxScore())
			require.Len(t, asmts[0].ExampleQuestions(), 1)
			assert.Equal(t, "Приведите отношение к 3НФ", asmts[0].ExampleQuestions()[0])
			assert.Equal(t, domain.AssessmentTypeFinal, asmts[1].Type(), "ФОС items keep emitted order by slice position")
			assert.Equal(t, 70, asmts[1].MaxScore())

			assert.Equal(t, 1, repo.updateCalls)
			assert.Same(t, wp, repo.updated)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.generated", audit.events[0].Action)
			assert.Equal(t, tc.actorID, audit.events[0].Fields["actor_user_id"])
			assert.Equal(t, int64(100), audit.events[0].Fields["work_program_id"])
			assert.Equal(t, 2, audit.events[0].Fields["assessments"], "success audit must count generated ФОС items")
		})
	}
}

func TestGenerateDraftUseCase_NonEmptyDraftRejected(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	existing, err := entities.NewGoal("Уже существующая цель", 0)
	require.NoError(t, err)
	require.NoError(t, wp.AddGoal(existing))

	repo := &fakeGenerateRepo{wp: wp}
	gen := &fakeDraftGenerator{result: sampleResult()}
	disc := &fakeDisciplineProvider{info: sampleDisciplineInfo()}
	audit := &recordingAuditSink{}
	uc := NewGenerateDraftUseCase(repo, gen, disc, allowingLimiter(), audit)

	_, err = uc.Execute(context.Background(), authorID, "teacher", 100)
	assert.True(t, errors.Is(err, domain.ErrWorkProgramNotEmpty),
		"generating into a non-empty draft must be rejected, got %v", err)
	assert.Zero(t, gen.calls, "must not call generator for a non-empty draft")
	assert.Zero(t, repo.updateCalls, "must not persist when the draft is non-empty")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.generate_denied", audit.events[0].Action)
	assert.Equal(t, "not_empty", audit.events[0].Fields["reason"])
}

func TestGenerateDraftUseCase_NonEmptyDraftRejectedWhenOnlyAssessmentsPresent(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	existing, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type: domain.AssessmentTypeCurrent, Description: "Уже существующий ФОС", MaxScore: 50,
	})
	require.NoError(t, err)
	require.NoError(t, wp.AddAssessment(existing))

	repo := &fakeGenerateRepo{wp: wp}
	gen := &fakeDraftGenerator{result: sampleResult()}
	disc := &fakeDisciplineProvider{info: sampleDisciplineInfo()}
	audit := &recordingAuditSink{}
	uc := NewGenerateDraftUseCase(repo, gen, disc, allowingLimiter(), audit)

	_, err = uc.Execute(context.Background(), authorID, "teacher", 100)
	assert.True(t, errors.Is(err, domain.ErrWorkProgramNotEmpty),
		"a draft already carrying ФОС must not be regenerated over, got %v", err)
	assert.Zero(t, gen.calls, "must not call generator when assessments already exist")
	assert.Zero(t, repo.updateCalls, "must not persist over an existing ФОС")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.generate_denied", audit.events[0].Action)
	assert.Equal(t, "not_empty", audit.events[0].Fields["reason"])
}

func TestGenerateDraftUseCase_DisciplineProviderErrorPropagates(t *testing.T) {
	const authorID = int64(7)
	sentinel := errors.New("curriculum unavailable")
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	repo := &fakeGenerateRepo{wp: wp}
	disc := &fakeDisciplineProvider{err: sentinel}
	gen := &fakeDraftGenerator{result: sampleResult()}
	uc := NewGenerateDraftUseCase(repo, gen, disc, allowingLimiter(), &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), authorID, "teacher", 100)
	assert.ErrorIs(t, err, sentinel)
	assert.Zero(t, gen.calls, "must not call generator when discipline lookup fails")
	assert.Zero(t, repo.updateCalls)
}

func TestGenerateDraftUseCase_GeneratorErrorPropagates(t *testing.T) {
	const authorID = int64(7)
	sentinel := errors.New("llm upstream 500")
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	repo := &fakeGenerateRepo{wp: wp}
	gen := &fakeDraftGenerator{err: sentinel}
	uc := NewGenerateDraftUseCase(repo, gen, &fakeDisciplineProvider{}, allowingLimiter(), &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), authorID, "teacher", 100)
	assert.ErrorIs(t, err, sentinel)
	assert.Zero(t, repo.updateCalls, "must not persist on generator failure")
}

func TestGenerateDraftUseCase_InvalidGeneratedContentRejected(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	repo := &fakeGenerateRepo{wp: wp}
	gen := &fakeDraftGenerator{result: DraftResult{
		Topics: []TopicDraft{{Kind: "not_a_kind", Title: "X", Hours: 2}},
	}}
	uc := NewGenerateDraftUseCase(repo, gen, &fakeDisciplineProvider{}, allowingLimiter(), &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), authorID, "teacher", 100)
	assert.True(t, errors.Is(err, domain.ErrInvalidWorkProgram),
		"invalid generated content must surface as ErrInvalidWorkProgram, got %v", err)
	assert.Zero(t, repo.updateCalls, "must not persist invalid generated content")
}

func TestGenerateDraftUseCase_UpdateErrorPropagates(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	repo := &fakeGenerateRepo{wp: wp, updateErr: repositories.ErrWorkProgramVersionConflict}
	gen := &fakeDraftGenerator{result: sampleResult()}
	uc := NewGenerateDraftUseCase(repo, gen, &fakeDisciplineProvider{info: sampleDisciplineInfo()}, allowingLimiter(), &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), authorID, "teacher", 100)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramVersionConflict)
}

func TestGenerateDraftUseCase_NilSinkTolerated(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusDraft)
	repo := &fakeGenerateRepo{wp: wp}
	gen := &fakeDraftGenerator{result: sampleResult()}
	uc := NewGenerateDraftUseCase(repo, gen, &fakeDisciplineProvider{info: sampleDisciplineInfo()}, allowingLimiter(), nil)

	got, err := uc.Execute(context.Background(), authorID, "teacher", 100)
	require.NoError(t, err)
	assert.NotNil(t, got)
}
