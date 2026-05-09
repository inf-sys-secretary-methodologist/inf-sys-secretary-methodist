package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ===== Fakes =====

type fakeDisciplineItemSaveRepo struct {
	saved      *entities.DisciplineItem
	idAssigned int64
	saveErr    error
	saveCalls  int
}

func (f *fakeDisciplineItemSaveRepo) Save(_ context.Context, d *entities.DisciplineItem) error {
	f.saveCalls++
	f.saved = d
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.idAssigned > 0 {
		d.ID = f.idAssigned
	}
	return nil
}

type fakeDisciplineItemGetRepo struct {
	got    *entities.DisciplineItem
	getErr error
}

func (f *fakeDisciplineItemGetRepo) GetByID(_ context.Context, id int64) (*entities.DisciplineItem, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.got != nil && f.got.ID != id {
		return nil, errors.New("fake: GetByID id mismatch")
	}
	return f.got, nil
}

type fakeDisciplineItemListRepo struct {
	got     []*entities.DisciplineItem
	listErr error
	gotID   int64
}

func (f *fakeDisciplineItemListRepo) ListBySectionID(_ context.Context, sectionID int64) ([]*entities.DisciplineItem, error) {
	f.gotID = sectionID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.got, nil
}

type fakeDisciplineItemUpdateRepo struct {
	got         *entities.DisciplineItem
	getErr      error
	updateErr   error
	updateCalls int
}

func (f *fakeDisciplineItemUpdateRepo) GetByID(_ context.Context, id int64) (*entities.DisciplineItem, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.got != nil && f.got.ID != id {
		return nil, errors.New("fake: GetByID id mismatch")
	}
	return f.got, nil
}

func (f *fakeDisciplineItemUpdateRepo) Update(_ context.Context, d *entities.DisciplineItem) error {
	f.updateCalls++
	return f.updateErr
}

type fakeDisciplineItemDeleteRepo struct {
	got         *entities.DisciplineItem
	getErr      error
	deleteErr   error
	deleteCalls int
}

func (f *fakeDisciplineItemDeleteRepo) GetByID(_ context.Context, id int64) (*entities.DisciplineItem, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.got, nil
}

func (f *fakeDisciplineItemDeleteRepo) Delete(_ context.Context, id int64) error {
	f.deleteCalls++
	return f.deleteErr
}

type fakeSectionLookup struct {
	got    *entities.Section
	getErr error
}

func (f *fakeSectionLookup) GetByID(_ context.Context, id int64) (*entities.Section, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.got, nil
}

// builtSection helper для cross-aggregate path. Section.curriculum_id = 7.
func builtSectionForItemTests(t *testing.T) *entities.Section {
	t.Helper()
	now := time.Now()
	return entities.ReconstituteSection(11, 7, "Базовая часть", "", 0, 0, now, now)
}

// draftCurriculumForItem builds curriculum в draft (editable) status.
func draftCurriculumForItem(t *testing.T, createdBy int64) *entities.Curriculum {
	t.Helper()
	c, err := entities.NewCurriculum(entities.NewCurriculumParams{
		Title: "ИВТ-2026", Code: fmt.Sprintf("C-%d", createdBy),
		Specialty: "ИВТ", Year: 2026, Description: "",
		CreatedBy: createdBy, Now: time.Now(),
	})
	require.NoError(t, err)
	c.ID = 7
	return c
}

func frozenCurriculumForItem(t *testing.T, status entities.CurriculumStatus, createdBy int64) *entities.Curriculum {
	t.Helper()
	now := time.Now()
	return entities.ReconstituteCurriculum(7, "ИВТ-2026", "C-1", "ИВТ", 2026, "",
		status, createdBy, nil, nil, now, now)
}

func validCreateInput() CreateDisciplineItemInput {
	return CreateDisciplineItemInput{
		SectionID: 11, Title: "Математический анализ",
		HoursLectures: 36, HoursPractice: 36, HoursLab: 0, HoursSelf: 72,
		ControlForm: entities.ControlFormExam, Credits: 4, Semester: 1, OrderIndex: 0,
	}
}

// ===== Constructor nil-panics =====

func TestNewCreateDisciplineItemUseCase_PanicsOnNilDeps(t *testing.T) {
	cases := []struct {
		name string
		fn   func()
	}{
		{"nil repo", func() {
			NewCreateDisciplineItemUseCase(nil, &fakeSectionLookup{}, &fakeCurriculumLookup{}, &recordingAuditSink{}, nil)
		}},
		{"nil sectionRepo", func() {
			NewCreateDisciplineItemUseCase(&fakeDisciplineItemSaveRepo{}, nil, &fakeCurriculumLookup{}, &recordingAuditSink{}, nil)
		}},
		{"nil curriculumRepo", func() {
			NewCreateDisciplineItemUseCase(&fakeDisciplineItemSaveRepo{}, &fakeSectionLookup{}, nil, &recordingAuditSink{}, nil)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("constructor accepted nil dep (%s)", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

func TestNewGetDisciplineItemUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic on nil repo")
		}
	}()
	NewGetDisciplineItemUseCase(nil)
}

func TestNewListDisciplineItemsBySectionUseCase_PanicsOnNilDeps(t *testing.T) {
	cases := []struct {
		name string
		fn   func()
	}{
		{"nil repo", func() {
			NewListDisciplineItemsBySectionUseCase(nil, &fakeSectionLookup{})
		}},
		{"nil sectionRepo", func() {
			NewListDisciplineItemsBySectionUseCase(&fakeDisciplineItemListRepo{}, nil)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("constructor accepted nil dep (%s)", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

// ===== Create =====

func TestCreateDisciplineItem_HappyPath(t *testing.T) {
	repo := &fakeDisciplineItemSaveRepo{idAssigned: 202}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	uc := NewCreateDisciplineItemUseCase(repo, section, curriculum, audit, func() time.Time { return frozenNow })
	d, err := uc.Execute(context.Background(), 42, false, validCreateInput())
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, int64(202), d.ID)
	assert.Equal(t, 1, repo.saveCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.created", audit.events[0].Action)
	assert.Equal(t, "curriculum_section_item", audit.events[0].Resource)
	assert.Equal(t, int64(202), audit.events[0].Fields["item_id"])
	assert.Equal(t, int64(11), audit.events[0].Fields["section_id"])
	assert.Equal(t, int64(7), audit.events[0].Fields["curriculum_id"])
}

func TestCreateDisciplineItem_NonAuthorMethodistDenied(t *testing.T) {
	repo := &fakeDisciplineItemSaveRepo{}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewCreateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 99, false, validCreateInput())
	assert.True(t, errors.Is(err, entities.ErrDisciplineItemScopeForbidden))
	assert.Equal(t, 0, repo.saveCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.create_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestCreateDisciplineItem_FrozenStatusDenied(t *testing.T) {
	repo := &fakeDisciplineItemSaveRepo{}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: frozenCurriculumForItem(t, entities.StatusPendingApproval, 42)}
	audit := &recordingAuditSink{}

	uc := NewCreateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, validCreateInput())
	assert.True(t, errors.Is(err, entities.ErrCannotEditDisciplineItem))
	assert.Equal(t, 0, repo.saveCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.create_denied", audit.events[0].Action)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestCreateDisciplineItem_AdminOverride(t *testing.T) {
	repo := &fakeDisciplineItemSaveRepo{idAssigned: 202}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewCreateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	d, err := uc.Execute(context.Background(), 99, true, validCreateInput())
	require.NoError(t, err)
	assert.Equal(t, int64(202), d.ID)
}

func TestCreateDisciplineItem_SectionNotFound(t *testing.T) {
	repo := &fakeDisciplineItemSaveRepo{}
	section := &fakeSectionLookup{getErr: repositories.ErrSectionNotFound}
	curriculum := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewCreateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, validCreateInput())
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound))
	require.Len(t, audit.events, 1)
	assert.Equal(t, "section_not_found", audit.events[0].Fields["reason"])
}

func TestCreateDisciplineItem_InvalidInput(t *testing.T) {
	repo := &fakeDisciplineItemSaveRepo{}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	in := validCreateInput()
	in.Title = "" // invariant fail
	uc := NewCreateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, in)
	assert.True(t, errors.Is(err, entities.ErrInvalidDisciplineItem))
	assert.Equal(t, 0, repo.saveCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.create_denied", audit.events[0].Action)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

// ===== Get =====

func TestGetDisciplineItem_HappyPath(t *testing.T) {
	want := entities.ReconstituteDisciplineItem(202, 11, "T", 36, 36, 0, 72,
		entities.ControlFormExam, 4, 1, 0, 0, time.Now(), time.Now())
	repo := &fakeDisciplineItemGetRepo{got: want}

	uc := NewGetDisciplineItemUseCase(repo)
	got, err := uc.Execute(context.Background(), 202)
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestGetDisciplineItem_NotFound(t *testing.T) {
	repo := &fakeDisciplineItemGetRepo{getErr: repositories.ErrDisciplineItemNotFound}
	uc := NewGetDisciplineItemUseCase(repo)
	got, err := uc.Execute(context.Background(), 999)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
}

// ===== List =====

func TestListDisciplineItemsBySection_HappyPath(t *testing.T) {
	now := time.Now()
	items := []*entities.DisciplineItem{
		entities.ReconstituteDisciplineItem(202, 11, "Дисциплина 1", 18, 18, 0, 36,
			entities.ControlFormZachet, 2, 1, 0, 0, now, now),
		entities.ReconstituteDisciplineItem(203, 11, "Дисциплина 2", 36, 36, 0, 72,
			entities.ControlFormExam, 4, 1, 1, 0, now, now),
	}
	repo := &fakeDisciplineItemListRepo{got: items}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	uc := NewListDisciplineItemsBySectionUseCase(repo, section)
	got, err := uc.Execute(context.Background(), 11)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(11), repo.gotID)
}

func TestListDisciplineItemsBySection_EmptyResult(t *testing.T) {
	repo := &fakeDisciplineItemListRepo{got: nil}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	uc := NewListDisciplineItemsBySectionUseCase(repo, section)
	got, err := uc.Execute(context.Background(), 11)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

// TestListDisciplineItemsBySection_SectionNotFound pins the cross-aggregate
// guard added in v0.128.2 — clients distinguish "no items" from "section
// gone" instead of seeing identical empty 200 responses (closes v0.128.1
// retroactive review Tier 1 #1).
func TestListDisciplineItemsBySection_SectionNotFound(t *testing.T) {
	repo := &fakeDisciplineItemListRepo{}
	section := &fakeSectionLookup{getErr: repositories.ErrSectionNotFound}
	uc := NewListDisciplineItemsBySectionUseCase(repo, section)
	got, err := uc.Execute(context.Background(), 99)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound),
		"List must propagate ErrSectionNotFound — caller maps to 404")
	assert.Equal(t, int64(0), repo.gotID,
		"must NOT call ListBySectionID when section absent (extra DB load)")
}

// TestListDisciplineItemsBySection_PropagatesOpaqueSectionLookupError pins
// transport-error path (DB connection issues, context cancellation) ahead
// of the items query.
func TestListDisciplineItemsBySection_PropagatesOpaqueSectionLookupError(t *testing.T) {
	opaque := errors.New("db down")
	repo := &fakeDisciplineItemListRepo{}
	section := &fakeSectionLookup{getErr: opaque}
	uc := NewListDisciplineItemsBySectionUseCase(repo, section)
	got, err := uc.Execute(context.Background(), 99)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, opaque))
	assert.Equal(t, int64(0), repo.gotID)
}

// ===== Update =====

func TestUpdateDisciplineItem_HappyPath(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "Старый", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 5, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, func() time.Time { return frozenNow })
	got, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "Новый",
		HoursLectures: 36, HoursPractice: 36, HoursLab: 0, HoursSelf: 72,
		ControlForm: entities.ControlFormExam, Credits: 4, Semester: 2, OrderIndex: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Новый", got.Title())
	assert.Equal(t, 1, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.updated", audit.events[0].Action)
}

func TestUpdateDisciplineItem_VersionConflict(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 3, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d, updateErr: repositories.ErrDisciplineItemVersionConflict}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "T",
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemVersionConflict))
	require.Len(t, audit.events, 1)
	assert.Equal(t, "version_conflict", audit.events[0].Fields["reason"])
}

func TestUpdateDisciplineItem_ItemNotFound(t *testing.T) {
	repo := &fakeDisciplineItemUpdateRepo{getErr: repositories.ErrDisciplineItemNotFound}
	section := &fakeSectionLookup{}
	curriculum := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{ID: 999, Title: "T"})
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

// TestUpdateDisciplineItem_FrozenStatusDenied pins the not_editable
// audit reason on the Update denial path.
func TestUpdateDisciplineItem_FrozenStatusDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: frozenCurriculumForItem(t, entities.StatusPendingApproval, 42)}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "T",
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, entities.ErrCannotEditDisciplineItem))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

// TestUpdateDisciplineItem_NonAuthorMethodistDenied pins the forbidden
// audit reason on the Update denial path (non-author methodist trying
// to edit someone else's curriculum's items).
func TestUpdateDisciplineItem_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)} // owned by 42
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 99, false, UpdateDisciplineItemInput{ // actor 99 ≠ owner 42
		ID: 202, Title: "T",
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, entities.ErrDisciplineItemScopeForbidden))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

// TestUpdateDisciplineItem_InvalidInput pins the invalid audit reason
// on the Update denial path (UpdateBasics invariant fail).
func TestUpdateDisciplineItem_InvalidInput(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "", // invariant fail
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidDisciplineItem))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

// ===== Delete =====

func TestDeleteDisciplineItem_HappyPath(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemDeleteRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 42, false, 202)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.deleted", audit.events[0].Action)
}

func TestDeleteDisciplineItem_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemDeleteRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 99, false, 202)
	assert.True(t, errors.Is(err, entities.ErrDisciplineItemScopeForbidden))
	assert.Equal(t, 0, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.delete_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

// TestDeleteDisciplineItem_FrozenStatusDenied pins the not_editable
// audit reason on the Delete denial path.
func TestDeleteDisciplineItem_FrozenStatusDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemDeleteRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: frozenCurriculumForItem(t, entities.StatusPendingApproval, 42)}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 42, false, 202)
	assert.True(t, errors.Is(err, entities.ErrCannotEditDisciplineItem))
	assert.Equal(t, 0, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.delete_denied", audit.events[0].Action)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestDeleteDisciplineItem_ItemNotFound(t *testing.T) {
	repo := &fakeDisciplineItemDeleteRepo{getErr: repositories.ErrDisciplineItemNotFound}
	section := &fakeSectionLookup{}
	curriculum := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 42, false, 999)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.delete_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}
