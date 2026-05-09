package usecases

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ===== Bulk-tx fakes =====

// fakeBulkItemsRepo implements repositories.DisciplineItemRepository.
// All 5 methods are stubbed; per-test wires only the ones it exercises.
type fakeBulkItemsRepo struct {
	saveCalls   []*entities.DisciplineItem
	saveErr     error
	idAssigner  func() int64 // returns id для each Save (or zero для default)
	getByIDFn   func(ctx context.Context, id int64) (*entities.DisciplineItem, error)
	updateCalls []*entities.DisciplineItem
	updateErr   error
	deleteCalls []int64
	deleteErr   error
}

func (f *fakeBulkItemsRepo) Save(_ context.Context, d *entities.DisciplineItem) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saveCalls = append(f.saveCalls, d)
	if f.idAssigner != nil {
		d.ID = f.idAssigner()
	}
	return nil
}

func (f *fakeBulkItemsRepo) GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, errors.New("fake: GetByID not wired")
}

func (f *fakeBulkItemsRepo) ListBySectionID(_ context.Context, _ int64) ([]*entities.DisciplineItem, error) {
	return nil, errors.New("fake: ListBySectionID not used by bulk-edit")
}

func (f *fakeBulkItemsRepo) Update(_ context.Context, d *entities.DisciplineItem) error {
	f.updateCalls = append(f.updateCalls, d)
	return f.updateErr
}

func (f *fakeBulkItemsRepo) Delete(_ context.Context, id int64) error {
	f.deleteCalls = append(f.deleteCalls, id)
	return f.deleteErr
}

// fakeBulkSectionsRepo implements repositories.SectionRepository.
type fakeBulkSectionsRepo struct {
	getByIDFn func(ctx context.Context, id int64) (*entities.Section, error)
}

func (f *fakeBulkSectionsRepo) GetByID(ctx context.Context, id int64) (*entities.Section, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, errors.New("fake: Section.GetByID not wired")
}

func (f *fakeBulkSectionsRepo) Save(_ context.Context, _ *entities.Section) error {
	return errors.New("fake: Save not used by bulk-edit")
}
func (f *fakeBulkSectionsRepo) ListByCurriculumID(_ context.Context, _ int64) ([]*entities.Section, error) {
	return nil, errors.New("fake: ListByCurriculumID not used by bulk-edit")
}
func (f *fakeBulkSectionsRepo) Update(_ context.Context, _ *entities.Section) error {
	return errors.New("fake: Update not used by bulk-edit")
}
func (f *fakeBulkSectionsRepo) Delete(_ context.Context, _ int64) error {
	return errors.New("fake: Delete not used by bulk-edit")
}

// fakeBulkCurriculaRepo implements repositories.CurriculumRepository.
type fakeBulkCurriculaRepo struct {
	getByIDFn func(ctx context.Context, id int64) (*entities.Curriculum, error)
}

func (f *fakeBulkCurriculaRepo) GetByID(ctx context.Context, id int64) (*entities.Curriculum, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, errors.New("fake: Curriculum.GetByID not wired")
}

func (f *fakeBulkCurriculaRepo) List(_ context.Context, _ repositories.CurriculumListFilter) (repositories.CurriculumListResult, error) {
	return repositories.CurriculumListResult{}, errors.New("fake: List not used by bulk-edit")
}
func (f *fakeBulkCurriculaRepo) Save(_ context.Context, _ *entities.Curriculum) error {
	return errors.New("fake: Save not used by bulk-edit")
}
func (f *fakeBulkCurriculaRepo) Update(_ context.Context, _ *entities.Curriculum) error {
	return errors.New("fake: Update not used by bulk-edit")
}

// fakeBulkTx implements repositories.BulkDisciplineItemsTx. Tracks
// commit/rollback calls so tests can assert tx lifecycle.
type fakeBulkTx struct {
	items         *fakeBulkItemsRepo
	sections      *fakeBulkSectionsRepo
	curricula     *fakeBulkCurriculaRepo
	commitCalls   int
	rollbackCalls int
	commitErr     error
	rollbackErr   error
}

func (t *fakeBulkTx) Items() repositories.DisciplineItemRepository {
	return t.items
}
func (t *fakeBulkTx) Sections() repositories.SectionRepository {
	return t.sections
}
func (t *fakeBulkTx) Curricula() repositories.CurriculumRepository {
	return t.curricula
}
func (t *fakeBulkTx) Commit() error {
	t.commitCalls++
	return t.commitErr
}
func (t *fakeBulkTx) Rollback() error {
	t.rollbackCalls++
	return t.rollbackErr
}

// fakeBulkUoW implements bulkEditUnitOfWork. Tracks Begin call args.
type fakeBulkUoW struct {
	tx         *fakeBulkTx
	beginErr   error
	beginCalls int
	gotOpts    *sql.TxOptions
}

func (u *fakeBulkUoW) Begin(_ context.Context, opts *sql.TxOptions) (repositories.BulkDisciplineItemsTx, error) {
	u.beginCalls++
	u.gotOpts = opts
	if u.beginErr != nil {
		return nil, u.beginErr
	}
	return u.tx, nil
}

// ===== Builders =====

// builtBulkTx returns a wired-up fakeBulkTx with all 3 repo fakes ready
// для per-test customization.
func builtBulkTx() *fakeBulkTx {
	return &fakeBulkTx{
		items:     &fakeBulkItemsRepo{},
		sections:  &fakeBulkSectionsRepo{},
		curricula: &fakeBulkCurriculaRepo{},
	}
}

// validBulkCreateItem returns a complete BulkCreateItem passing
// invariants — used as starting point для denial tests via tweak.
func validBulkCreateItem() BulkCreateItem {
	return BulkCreateItem{
		Title:         "Математический анализ",
		HoursLectures: 36, HoursPractice: 36, HoursLab: 0, HoursSelf: 72,
		ControlForm: entities.ControlFormExam, Credits: 4, Semester: 1, OrderIndex: 0,
	}
}

// ===== Constructor nil-panic =====

func TestNewBulkEditDisciplineItemsUseCase_PanicsOnNilUoW(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("constructor accepted nil uow")
		}
	}()
	NewBulkEditDisciplineItemsUseCase(nil, &recordingAuditSink{}, time.Now)
}

// ===== Empty input =====

func TestBulkEdit_EmptyInput_RejectedBeforeTx(t *testing.T) {
	uow := &fakeBulkUoW{tx: builtBulkTx()}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		// все три коллекции пусты
	})
	assert.Nil(t, res)
	assert.True(t, errors.Is(err, ErrEmptyBulkInput))
	assert.Equal(t, 0, uow.beginCalls,
		"empty input must reject before opening tx")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.bulk_edit_denied", audit.events[0].Action)
	assert.Equal(t, "empty_input", audit.events[0].Fields["reason"])
}

// ===== Section/curriculum gates =====

func TestBulkEdit_SectionNotFound_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return nil, repositories.ErrSectionNotFound
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 999,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
	})
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound))
	assert.Equal(t, 0, tx.commitCalls)
	assert.GreaterOrEqual(t, tx.rollbackCalls, 1, "must rollback when section gone")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.bulk_edit_denied", audit.events[0].Action)
	assert.Equal(t, "section_not_found", audit.events[0].Fields["reason"])
}

func TestBulkEdit_FrozenStatus_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return frozenCurriculumForItem(t, entities.StatusPendingApproval, 42), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
	})
	assert.True(t, errors.Is(err, entities.ErrCannotEditDisciplineItem))
	assert.Equal(t, 0, tx.commitCalls)
	assert.GreaterOrEqual(t, tx.rollbackCalls, 1)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestBulkEdit_NonAuthorMethodist_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 99, false, BulkEditDisciplineItemsInput{ // 99 ≠ 42
		SectionID: 11,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
	})
	assert.True(t, errors.Is(err, entities.ErrDisciplineItemScopeForbidden))
	assert.Equal(t, 0, tx.commitCalls)
	assert.GreaterOrEqual(t, tx.rollbackCalls, 1)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

// ===== Happy creates =====

func TestBulkEdit_HappyCreates_Single(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	idCounter := int64(201)
	tx.items.idAssigner = func() int64 {
		idCounter++
		return idCounter
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, func() time.Time { return frozenNow })

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Created, 1)
	assert.Equal(t, int64(202), res.Created[0].ID)
	assert.Empty(t, res.Conflicts)
	assert.Len(t, tx.items.saveCalls, 1)
	assert.Equal(t, 1, tx.commitCalls, "single tx commit on success")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.bulk_edited", audit.events[0].Action)
	assert.Equal(t, int64(11), audit.events[0].Fields["section_id"])
	assert.Equal(t, 1, audit.events[0].Fields["created_count"])
	assert.Equal(t, 0, audit.events[0].Fields["updated_count"])
	assert.Equal(t, 0, audit.events[0].Fields["deleted_count"])
}

func TestBulkEdit_HappyCreates_Multiple(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	idCounter := int64(200)
	tx.items.idAssigner = func() int64 {
		idCounter++
		return idCounter
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	c1 := validBulkCreateItem()
	c2 := validBulkCreateItem()
	c2.Title = "Дискретная математика"
	c3 := validBulkCreateItem()
	c3.Title = "Линейная алгебра"

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Creates:   []BulkCreateItem{c1, c2, c3},
	})
	require.NoError(t, err)
	require.Len(t, res.Created, 3)
	assert.Len(t, tx.items.saveCalls, 3)
	assert.Equal(t, 1, tx.commitCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, 3, audit.events[0].Fields["created_count"])
}

func TestBulkEdit_AdminOverride_HappyCreates(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil // owned by 42
	}
	tx.items.idAssigner = func() int64 { return 202 }
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	res, err := uc.Execute(context.Background(), 99, true, BulkEditDisciplineItemsInput{ // 99 ≠ 42, but isAdmin=true
		SectionID: 11,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
	})
	require.NoError(t, err)
	require.Len(t, res.Created, 1)
	assert.Equal(t, 1, tx.commitCalls)
}

// ===== Invalid create =====

func TestBulkEdit_InvalidCreate_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	bad := validBulkCreateItem()
	bad.Title = "" // invariant fail

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Creates:   []BulkCreateItem{bad},
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidDisciplineItem))
	assert.Empty(t, tx.items.saveCalls,
		"invalid create rejects before any Save is issued")
	assert.Equal(t, 0, tx.commitCalls)
	assert.GreaterOrEqual(t, tx.rollbackCalls, 1)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

// ===== Updates =====

// builtItemForBulk constructs an existing item с specified version
// для update-path tests. section_id matches the canonical builtBulk
// section (11).
func builtItemForBulk(t *testing.T, id int64, version int) *entities.DisciplineItem {
	t.Helper()
	now := time.Now()
	return entities.ReconstituteDisciplineItem(id, 11, "Старое название",
		18, 18, 0, 36, entities.ControlFormZachet, 2, 1, 0, version, now, now)
}

// validBulkUpdateItem returns a complete BulkUpdateItem passing
// invariants для update-path happy tests.
func validBulkUpdateItem(id int64) BulkUpdateItem {
	return BulkUpdateItem{
		ID: id, Title: "Новое название",
		HoursLectures: 36, HoursPractice: 36, HoursLab: 0, HoursSelf: 72,
		ControlForm: entities.ControlFormExam, Credits: 4, Semester: 2, OrderIndex: 1,
	}
}

func TestBulkEdit_HappyUpdate_Single(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		return builtItemForBulk(t, id, 5), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates:   []BulkUpdateItem{validBulkUpdateItem(202)},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Updated, 1)
	assert.Equal(t, "Новое название", res.Updated[0].Title())
	assert.Empty(t, res.Conflicts)
	assert.Empty(t, tx.items.saveCalls, "no creates issued")
	assert.Len(t, tx.items.updateCalls, 1)
	assert.Equal(t, 1, tx.commitCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.bulk_edited", audit.events[0].Action)
	assert.Equal(t, 1, audit.events[0].Fields["updated_count"])
}

func TestBulkEdit_HappyUpdate_Multiple(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		return builtItemForBulk(t, id, 1), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates: []BulkUpdateItem{
			validBulkUpdateItem(201),
			validBulkUpdateItem(202),
			validBulkUpdateItem(203),
		},
	})
	require.NoError(t, err)
	require.Len(t, res.Updated, 3)
	assert.Len(t, tx.items.updateCalls, 3)
	assert.Equal(t, 1, tx.commitCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, 3, audit.events[0].Fields["updated_count"])
}

func TestBulkEdit_UpdateItemNotFound_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.getByIDFn = func(_ context.Context, _ int64) (*entities.DisciplineItem, error) {
		return nil, repositories.ErrDisciplineItemNotFound
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates:   []BulkUpdateItem{validBulkUpdateItem(999)},
	})
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
	assert.Equal(t, 0, tx.commitCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestBulkEdit_UpdateCrossSection_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	// Item belongs to section 99, but request targets section 11 — cross-section.
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		now := time.Now()
		return entities.ReconstituteDisciplineItem(id, 99, "X", 18, 18, 0, 36,
			entities.ControlFormZachet, 2, 1, 0, 1, now, now), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates:   []BulkUpdateItem{validBulkUpdateItem(202)},
	})
	assert.True(t, errors.Is(err, ErrCrossSectionBulkEdit))
	assert.Equal(t, 0, tx.commitCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "cross_section", audit.events[0].Fields["reason"])
}

func TestBulkEdit_UpdateInvalidInput_RollsBack(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		return builtItemForBulk(t, id, 1), nil
	}
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	bad := validBulkUpdateItem(202)
	bad.Title = "" // invariant fail
	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates:   []BulkUpdateItem{bad},
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidDisciplineItem))
	assert.Equal(t, 0, tx.commitCalls)
	assert.Empty(t, tx.items.updateCalls,
		"invalid update rejects before any tx.Items().Update issued")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

// ===== Version-conflict collect-all =====

func TestBulkEdit_UpdateVersionConflict_Single(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	currentVersion := 7
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		// First load returns expected version 5 (caller's snapshot).
		// Second load (after conflict) returns "current" version 7.
		// Implementation re-fetches via GetByID after conflict to read
		// current version для UI feedback.
		return builtItemForBulk(t, id, currentVersion), nil
	}
	tx.items.updateErr = repositories.ErrDisciplineItemVersionConflict
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates:   []BulkUpdateItem{validBulkUpdateItem(202)},
	})
	assert.True(t, errors.Is(err, ErrBulkVersionConflict))
	require.NotNil(t, res, "result should be returned alongside conflict err for handler 409 mapping")
	require.Len(t, res.Conflicts, 1)
	assert.Equal(t, int64(202), res.Conflicts[0].ID)
	assert.Equal(t, 0, tx.commitCalls, "no commit when conflicts present")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "version_conflict", audit.events[0].Fields["reason"])
}

func TestBulkEdit_UpdateVersionConflict_CollectsAll(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		return builtItemForBulk(t, id, 7), nil
	}
	tx.items.updateErr = repositories.ErrDisciplineItemVersionConflict
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	res, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Updates: []BulkUpdateItem{
			validBulkUpdateItem(201),
			validBulkUpdateItem(202),
			validBulkUpdateItem(203),
		},
	})
	assert.True(t, errors.Is(err, ErrBulkVersionConflict))
	require.NotNil(t, res)
	require.Len(t, res.Conflicts, 3,
		"collect-all per ADR-12 — every conflict reported, не short-circuit on first")
	assert.Equal(t, 0, tx.commitCalls)
}

// ===== Combined =====

func TestBulkEdit_CombinedCreatesUpdates_AtomicRollbackOnConflict(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.idAssigner = func() int64 { return 250 }
	tx.items.getByIDFn = func(_ context.Context, id int64) (*entities.DisciplineItem, error) {
		return builtItemForBulk(t, id, 5), nil
	}
	tx.items.updateErr = repositories.ErrDisciplineItemVersionConflict
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
		Updates:   []BulkUpdateItem{validBulkUpdateItem(202)},
	})
	assert.True(t, errors.Is(err, ErrBulkVersionConflict))
	// Both Save (creates) и Update (updates) invoked within tx; rollback
	// throws away both. Fake records calls regardless — we assert the
	// transactional guarantee via tx.commitCalls == 0.
	assert.Len(t, tx.items.saveCalls, 1, "create was attempted within tx")
	assert.Len(t, tx.items.updateCalls, 1, "update was attempted within tx")
	assert.Equal(t, 0, tx.commitCalls, "no commit — atomic rollback on any conflict")
	assert.GreaterOrEqual(t, tx.rollbackCalls, 1)
}

// ===== Isolation level =====

func TestBulkEdit_BeginsTxWithRepeatableRead(t *testing.T) {
	tx := builtBulkTx()
	tx.sections.getByIDFn = func(_ context.Context, _ int64) (*entities.Section, error) {
		return builtSectionForItemTests(t), nil
	}
	tx.curricula.getByIDFn = func(_ context.Context, _ int64) (*entities.Curriculum, error) {
		return draftCurriculumForItem(t, 42), nil
	}
	tx.items.idAssigner = func() int64 { return 202 }
	uow := &fakeBulkUoW{tx: tx}
	audit := &recordingAuditSink{}
	uc := NewBulkEditDisciplineItemsUseCase(uow, audit, time.Now)

	_, err := uc.Execute(context.Background(), 42, false, BulkEditDisciplineItemsInput{
		SectionID: 11,
		Creates:   []BulkCreateItem{validBulkCreateItem()},
	})
	require.NoError(t, err)
	require.NotNil(t, uow.gotOpts, "Begin must be called с TxOptions, not nil")
	assert.Equal(t, sql.LevelRepeatableRead, uow.gotOpts.Isolation,
		"v0.128.3 ADR-12: bulk-edit uses Repeatable Read for phantom-prevention")
}
