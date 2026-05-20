package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
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

// ===== Builders =====

// builtSectionForItemTests returns a section с curriculum_id=7, section_id=11 —
// canonical fixture for cross-aggregate path tests.
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

// frozenCurriculumForItem builds curriculum в non-editable status (e.g.,
// pending_approval, approved, archived).
func frozenCurriculumForItem(t *testing.T, status entities.CurriculumStatus, createdBy int64) *entities.Curriculum {
	t.Helper()
	now := time.Now()
	return entities.ReconstituteCurriculum(7, "ИВТ-2026", "C-1", "ИВТ", 2026, "",
		status, createdBy, nil, nil, now, now, 0)
}

// validCreateInput returns a complete CreateDisciplineItemInput passing all
// invariants (used as starting point for "tweak one field" denial tests).
func validCreateInput() CreateDisciplineItemInput {
	return CreateDisciplineItemInput{
		SectionID: 11, Title: "Математический анализ",
		HoursLectures: 36, HoursPractice: 36, HoursLab: 0, HoursSelf: 72,
		ControlForm: entities.ControlFormExam, Credits: 4, Semester: 1, OrderIndex: 0,
	}
}
