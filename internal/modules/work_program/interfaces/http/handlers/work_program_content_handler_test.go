package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// ===== Fake content port =====

// fakeContent records the last call across all nine collection-edit
// methods and returns the configured result/err. A single double keeps
// the test wiring small while still letting each test assert the exact
// arguments threaded from path + body.
type fakeContent struct {
	result *entities.WorkProgram
	err    error

	method   string
	gotActor int64
	gotRole  string
	gotWP    int64
	gotChild int64

	gotText       string
	gotOrderIndex int
	gotCode       string
	gotType       string
	gotDesc       string
	gotTopic      wpUsecases.TopicContentInput
	gotAssessment wpUsecases.AssessmentContentInput
	gotReference  wpUsecases.ReferenceContentInput
}

func (f *fakeContent) AddGoal(_ context.Context, actorID int64, role string, wpID int64, text string, orderIndex int) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotText, f.gotOrderIndex = "AddGoal", actorID, role, wpID, text, orderIndex
	return f.result, f.err
}

func (f *fakeContent) UpdateGoal(_ context.Context, actorID int64, role string, wpID, goalID int64, text string, orderIndex int) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild, f.gotText, f.gotOrderIndex = "UpdateGoal", actorID, role, wpID, goalID, text, orderIndex
	return f.result, f.err
}

func (f *fakeContent) RemoveGoal(_ context.Context, actorID int64, role string, wpID, goalID int64) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild = "RemoveGoal", actorID, role, wpID, goalID
	return f.result, f.err
}

func (f *fakeContent) AddCompetence(_ context.Context, actorID int64, role string, wpID int64, code, ctype, description string) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotCode, f.gotType, f.gotDesc = "AddCompetence", actorID, role, wpID, code, ctype, description
	return f.result, f.err
}

func (f *fakeContent) UpdateCompetence(_ context.Context, actorID int64, role string, wpID, compID int64, code, ctype, description string) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild, f.gotCode, f.gotType, f.gotDesc = "UpdateCompetence", actorID, role, wpID, compID, code, ctype, description
	return f.result, f.err
}

func (f *fakeContent) RemoveCompetence(_ context.Context, actorID int64, role string, wpID, compID int64) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild = "RemoveCompetence", actorID, role, wpID, compID
	return f.result, f.err
}

func (f *fakeContent) AddTopic(_ context.Context, actorID int64, role string, wpID int64, in wpUsecases.TopicContentInput) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotTopic = "AddTopic", actorID, role, wpID, in
	return f.result, f.err
}

func (f *fakeContent) UpdateTopic(_ context.Context, actorID int64, role string, wpID, topicID int64, in wpUsecases.TopicContentInput) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild, f.gotTopic = "UpdateTopic", actorID, role, wpID, topicID, in
	return f.result, f.err
}

func (f *fakeContent) RemoveTopic(_ context.Context, actorID int64, role string, wpID, topicID int64) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild = "RemoveTopic", actorID, role, wpID, topicID
	return f.result, f.err
}

func (f *fakeContent) AddAssessment(_ context.Context, actorID int64, role string, wpID int64, in wpUsecases.AssessmentContentInput) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotAssessment = "AddAssessment", actorID, role, wpID, in
	return f.result, f.err
}

func (f *fakeContent) UpdateAssessment(_ context.Context, actorID int64, role string, wpID, assessmentID int64, in wpUsecases.AssessmentContentInput) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild, f.gotAssessment = "UpdateAssessment", actorID, role, wpID, assessmentID, in
	return f.result, f.err
}

func (f *fakeContent) RemoveAssessment(_ context.Context, actorID int64, role string, wpID, assessmentID int64) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild = "RemoveAssessment", actorID, role, wpID, assessmentID
	return f.result, f.err
}

func (f *fakeContent) AddReference(_ context.Context, actorID int64, role string, wpID int64, in wpUsecases.ReferenceContentInput) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotReference = "AddReference", actorID, role, wpID, in
	return f.result, f.err
}

func (f *fakeContent) UpdateReference(_ context.Context, actorID int64, role string, wpID, referenceID int64, in wpUsecases.ReferenceContentInput) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild, f.gotReference = "UpdateReference", actorID, role, wpID, referenceID, in
	return f.result, f.err
}

func (f *fakeContent) RemoveReference(_ context.Context, actorID int64, role string, wpID, referenceID int64) (*entities.WorkProgram, error) {
	f.method, f.gotActor, f.gotRole, f.gotWP, f.gotChild = "RemoveReference", actorID, role, wpID, referenceID
	return f.result, f.err
}

func newContentRouter(fake *fakeContent, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if fake == nil {
		fake = &fakeContent{}
	}
	h := NewWorkProgramContentHandler(fake)
	api := r.Group("/api/v1")
	for _, m := range mw {
		api.Use(m)
	}
	RegisterWorkProgramContentRoutes(api, h)
	return r
}

// ===== Goals =====

func TestContentHandler_AddGoal_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/goals",
		AddGoalRequest{Text: "Освоить нормализацию БД", OrderIndex: 1})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "AddGoal", f.method)
	assert.Equal(t, int64(42), f.gotActor, "actor derives from JWT")
	assert.Equal(t, "teacher", f.gotRole)
	assert.Equal(t, int64(99), f.gotWP, "wp id from path")
	assert.Equal(t, "Освоить нормализацию БД", f.gotText)
	assert.Equal(t, 1, f.gotOrderIndex)
}

func TestContentHandler_AddGoal_Unauthorized(t *testing.T) {
	r := newContentRouter(nil) // no withAuth
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/goals",
		AddGoalRequest{Text: "x"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContentHandler_AddGoal_BadID(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/abc/goals",
		AddGoalRequest{Text: "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_AddGoal_MissingText(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/goals",
		AddGoalRequest{Text: ""})
	assert.Equal(t, http.StatusBadRequest, w.Code, "text is required by binding")
}

func TestContentHandler_UpdateGoal_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/goals/5",
		AddGoalRequest{Text: "Обновлённая цель", OrderIndex: 2})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "UpdateGoal", f.method)
	assert.Equal(t, int64(99), f.gotWP)
	assert.Equal(t, int64(5), f.gotChild, "goal id from path")
	assert.Equal(t, "Обновлённая цель", f.gotText)
}

func TestContentHandler_UpdateGoal_BadChildID(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/goals/0",
		AddGoalRequest{Text: "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_RemoveGoal_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/goals/5", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "RemoveGoal", f.method)
	assert.Equal(t, int64(99), f.gotWP)
	assert.Equal(t, int64(5), f.gotChild)
}

// ChildNotFound is the new behavior 12b adds to the error map: a missing
// collection element surfaces 404, not 500.
func TestContentHandler_RemoveGoal_ChildNotFound(t *testing.T) {
	f := &fakeContent{err: domain.ErrChildNotFound}
	r := newContentRouter(f, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/goals/777", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// A non-admin author-scope denial collapses to 404 (IDOR mitigation),
// mirroring the resource-scoped transition endpoints.
func TestContentHandler_RemoveGoal_ForbiddenNonAdmin_Collapses404(t *testing.T) {
	f := &fakeContent{err: domain.ErrWorkProgramScopeForbidden}
	r := newContentRouter(f, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/goals/5", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestContentHandler_RemoveGoal_ForbiddenAdmin_Gets403(t *testing.T) {
	f := &fakeContent{err: domain.ErrWorkProgramScopeForbidden}
	r := newContentRouter(f, withAuth(1, "system_admin"))
	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/goals/5", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestContentHandler_AddGoal_WorkProgramNotFound(t *testing.T) {
	f := &fakeContent{err: repositories.ErrWorkProgramNotFound}
	r := newContentRouter(f, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/goals",
		AddGoalRequest{Text: "x"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// A domain invariant violation (status gate / bad value) maps to 422.
func TestContentHandler_AddGoal_InvalidWorkProgram(t *testing.T) {
	f := &fakeContent{err: domain.ErrInvalidWorkProgram}
	r := newContentRouter(f, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/goals",
		AddGoalRequest{Text: "x"})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== Competences =====

func TestContentHandler_AddCompetence_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/competences",
		CompetenceContentRequest{Code: "ОПК-1", Type: "professional", Description: "Способен применять СУБД"})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "AddCompetence", f.method)
	assert.Equal(t, int64(99), f.gotWP)
	assert.Equal(t, "ОПК-1", f.gotCode)
	assert.Equal(t, "professional", f.gotType)
	assert.Equal(t, "Способен применять СУБД", f.gotDesc)
}

func TestContentHandler_AddCompetence_MissingCode(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/competences",
		CompetenceContentRequest{Type: "professional", Description: "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_UpdateCompetence_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/competences/8",
		CompetenceContentRequest{Code: "ОПК-2", Type: "general", Description: "y"})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "UpdateCompetence", f.method)
	assert.Equal(t, int64(8), f.gotChild)
	assert.Equal(t, "ОПК-2", f.gotCode)
}

func TestContentHandler_RemoveCompetence_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/competences/8", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "RemoveCompetence", f.method)
	assert.Equal(t, int64(8), f.gotChild)
}

func TestContentHandler_UpdateCompetence_ChildNotFound(t *testing.T) {
	f := &fakeContent{err: domain.ErrChildNotFound}
	r := newContentRouter(f, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/competences/777",
		CompetenceContentRequest{Code: "ОПК-2", Type: "general", Description: "y"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===== Topics =====

func TestContentHandler_AddTopic_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	week := 3
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/topics",
		TopicContentRequest{
			Kind: "lecture", Title: "Реляционная модель", Hours: 4,
			WeekNumber: &week, LearningOutcomes: "Знать модель", OrderIndex: 1,
		})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "AddTopic", f.method)
	assert.Equal(t, int64(99), f.gotWP)
	assert.Equal(t, "lecture", f.gotTopic.Kind)
	assert.Equal(t, "Реляционная модель", f.gotTopic.Title)
	assert.Equal(t, 4, f.gotTopic.Hours)
	require.NotNil(t, f.gotTopic.WeekNumber)
	assert.Equal(t, 3, *f.gotTopic.WeekNumber)
}

func TestContentHandler_AddTopic_MissingTitle(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/topics",
		TopicContentRequest{Kind: "lecture", Hours: 4})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_UpdateTopic_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/topics/12",
		TopicContentRequest{Kind: "practice", Title: "Нормальные формы", Hours: 2, OrderIndex: 2})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "UpdateTopic", f.method)
	assert.Equal(t, int64(12), f.gotChild)
	assert.Equal(t, "practice", f.gotTopic.Kind)
}

func TestContentHandler_RemoveTopic_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/topics/12", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "RemoveTopic", f.method)
	assert.Equal(t, int64(12), f.gotChild)
}

// ===== Assessments (ФОС) =====

func TestContentHandler_AddAssessment_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/assessments",
		AssessmentContentRequest{
			Type: "exam", Description: "Итоговый экзамен", MaxScore: 40,
			ExampleQuestions: []string{"Что такое 3НФ?", "Опишите ACID"},
		})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "AddAssessment", f.method)
	assert.Equal(t, int64(99), f.gotWP)
	assert.Equal(t, "exam", f.gotAssessment.Type)
	assert.Equal(t, "Итоговый экзамен", f.gotAssessment.Description)
	assert.Equal(t, 40, f.gotAssessment.MaxScore)
	assert.Equal(t, []string{"Что такое 3НФ?", "Опишите ACID"}, f.gotAssessment.ExampleQuestions)
}

func TestContentHandler_AddAssessment_MissingType(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/assessments",
		AssessmentContentRequest{Description: "x", MaxScore: 10})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_UpdateAssessment_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/assessments/15",
		AssessmentContentRequest{Type: "test", Description: "Контрольная", MaxScore: 20})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "UpdateAssessment", f.method)
	assert.Equal(t, int64(15), f.gotChild)
	assert.Equal(t, "test", f.gotAssessment.Type)
}

func TestContentHandler_RemoveAssessment_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/assessments/15", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "RemoveAssessment", f.method)
	assert.Equal(t, int64(15), f.gotChild)
}

// Domain enforces the per-item MaxScore range; the handler surfaces it as 422.
func TestContentHandler_AddAssessment_InvalidWorkProgram(t *testing.T) {
	f := &fakeContent{err: domain.ErrInvalidWorkProgram}
	r := newContentRouter(f, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/assessments",
		AssessmentContentRequest{Type: "exam", Description: "x", MaxScore: 9999})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestContentHandler_UpdateAssessment_ChildNotFound(t *testing.T) {
	f := &fakeContent{err: domain.ErrChildNotFound}
	r := newContentRouter(f, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/assessments/777",
		AssessmentContentRequest{Type: "test", Description: "x", MaxScore: 5})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===== References =====

func TestContentHandler_AddReference_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	year := 2021
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/references",
		ReferenceContentRequest{
			Kind: "main", Citation: "Дейт К. Введение в системы баз данных",
			Year: &year, ISBN: "978-5-8459-0788-2", OrderIndex: 1,
		})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "AddReference", f.method)
	assert.Equal(t, int64(99), f.gotWP)
	assert.Equal(t, "main", f.gotReference.Kind)
	assert.Equal(t, "Дейт К. Введение в системы баз данных", f.gotReference.Citation)
	require.NotNil(t, f.gotReference.Year)
	assert.Equal(t, 2021, *f.gotReference.Year)
}

func TestContentHandler_AddReference_MissingCitation(t *testing.T) {
	r := newContentRouter(nil, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs/99/references",
		ReferenceContentRequest{Kind: "main"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_UpdateReference_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPut, "/api/v1/work-programs/99/references/21",
		ReferenceContentRequest{Kind: "additional", Citation: "Гарсиа-Молина Г. Системы баз данных", OrderIndex: 2})

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "UpdateReference", f.method)
	assert.Equal(t, int64(21), f.gotChild)
	assert.Equal(t, "additional", f.gotReference.Kind)
}

func TestContentHandler_RemoveReference_HappyPath(t *testing.T) {
	f := &fakeContent{result: sampleWP(t)}
	r := newContentRouter(f, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/references/21", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "RemoveReference", f.method)
	assert.Equal(t, int64(21), f.gotChild)
}

func TestContentHandler_RemoveReference_ForbiddenNonAdmin_Collapses404(t *testing.T) {
	f := &fakeContent{err: domain.ErrWorkProgramScopeForbidden}
	r := newContentRouter(f, withAuth(7, "methodist"))
	w := doJSON(t, r, http.MethodDelete, "/api/v1/work-programs/99/references/21", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===== Nil-port guard =====

func TestNewWorkProgramContentHandler_NilPortPanics(t *testing.T) {
	assert.Panics(t, func() { NewWorkProgramContentHandler(nil) })
}

// Content routes share the /work-programs/:id subtree with the main РПД
// routes and the revision routes. gin panics at registration time on a
// path/param conflict, so registering all three on one engine guards
// against the conflict that unit-level routers (which mount only one set)
// cannot see.
func TestRegisterWorkProgramContentRoutes_NoConflictWithSiblingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")

	wpH := NewWorkProgramHandler(
		&fakeCreate{}, &fakeGet{}, &fakeList{},
		&fakeSubmit{}, &fakeApprove{}, &fakeReject{}, &fakeDiscard{}, &fakeGenerate{},
	)
	revH := NewRevisionHandler(&fakeCreateRevision{}, &fakeSubmitRevision{}, &fakeApproveRevision{}, &fakeRejectRevision{})
	contentH := NewWorkProgramContentHandler(&fakeContent{})

	assert.NotPanics(t, func() {
		RegisterWorkProgramRoutes(api, wpH)
		RegisterRevisionRoutes(api, revH)
		RegisterWorkProgramContentRoutes(api, contentH)
	})
}
