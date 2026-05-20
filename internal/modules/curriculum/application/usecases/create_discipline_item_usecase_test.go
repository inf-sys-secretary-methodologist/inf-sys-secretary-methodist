package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ===== Constructor nil-panic =====

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

// ===== Path tests =====

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

func TestCreateDisciplineItem_NonAuthorDenied(t *testing.T) {
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
