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

// ===== Path tests =====

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
