package adapters

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// v0.153.3 #196 coverage push: fun_fact_seeder.go was at 0% covered.
// Mock-based tests pin SeedIfEmpty + getDefaultFacts behavior.

type fakeFunFactRepo struct {
	count           int64
	countErr        error
	bulkCreateErr   error
	bulkCreateCount int
	bulkCreateFacts []entities.FunFact
}

func (r *fakeFunFactRepo) Create(_ context.Context, _ *entities.FunFact) error { return nil }
func (r *fakeFunFactRepo) BulkCreate(_ context.Context, facts []entities.FunFact) error {
	r.bulkCreateCount++
	r.bulkCreateFacts = facts
	return r.bulkCreateErr
}
func (r *fakeFunFactRepo) GetRandom(_ context.Context) (*entities.FunFact, error)    { return nil, nil }
func (r *fakeFunFactRepo) GetLeastUsed(_ context.Context) (*entities.FunFact, error) { return nil, nil }
func (r *fakeFunFactRepo) IncrementUsedCount(_ context.Context, _ int64) error       { return nil }
func (r *fakeFunFactRepo) Count(_ context.Context) (int64, error) {
	return r.count, r.countErr
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewFunFactSeeder_StoresDeps(t *testing.T) {
	repo := &fakeFunFactRepo{}
	logger := quietLogger()
	seeder := NewFunFactSeeder(repo, logger)
	require.NotNil(t, seeder)
	assert.Equal(t, repo, seeder.repo)
	assert.Equal(t, logger, seeder.logger)
}

func TestFunFactSeeder_SeedIfEmpty_HappyPath(t *testing.T) {
	repo := &fakeFunFactRepo{count: 0}
	seeder := NewFunFactSeeder(repo, quietLogger())

	err := seeder.SeedIfEmpty(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, repo.bulkCreateCount, "BulkCreate must be called when table empty")
	assert.NotEmpty(t, repo.bulkCreateFacts, "BulkCreate must receive default facts")
	for _, f := range repo.bulkCreateFacts {
		assert.Equal(t, "ru", f.Language, "all default facts must be Russian")
		assert.True(t, f.IsApproved, "all default facts must be pre-approved")
		assert.NotEmpty(t, f.Content)
		assert.NotEmpty(t, f.Category)
	}
}

func TestFunFactSeeder_SeedIfEmpty_AlreadySeeded(t *testing.T) {
	repo := &fakeFunFactRepo{count: 42}
	seeder := NewFunFactSeeder(repo, quietLogger())

	err := seeder.SeedIfEmpty(context.Background())
	require.NoError(t, err, "non-zero count must be a no-op success")
	assert.Equal(t, 0, repo.bulkCreateCount, "BulkCreate must NOT be called when table non-empty")
}

func TestFunFactSeeder_SeedIfEmpty_CountError(t *testing.T) {
	repo := &fakeFunFactRepo{countErr: errors.New("db connection lost")}
	seeder := NewFunFactSeeder(repo, quietLogger())

	err := seeder.SeedIfEmpty(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count facts")
	assert.Equal(t, 0, repo.bulkCreateCount, "BulkCreate must NOT be called when Count fails")
}

func TestFunFactSeeder_SeedIfEmpty_BulkCreateError(t *testing.T) {
	repo := &fakeFunFactRepo{
		count:         0,
		bulkCreateErr: errors.New("postgres write failed"),
	}
	seeder := NewFunFactSeeder(repo, quietLogger())

	err := seeder.SeedIfEmpty(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to seed fun facts")
	assert.Equal(t, 1, repo.bulkCreateCount, "BulkCreate is attempted даже if it fails")
}

// TestGetDefaultFacts_ShapeAndCategories pins the static seed dataset shape:
// non-empty content + category, language=ru, IsApproved=true. Defends
// against silent regression в the embedded раздаточный list.
func TestGetDefaultFacts_ShapeAndCategories(t *testing.T) {
	facts := getDefaultFacts()
	require.NotEmpty(t, facts, "default facts must be non-empty (used by SeedIfEmpty)")
	assert.GreaterOrEqual(t, len(facts), 30, "expected at least 30 default facts")

	categorySet := map[string]bool{}
	for _, f := range facts {
		assert.NotEmpty(t, f.Content, "every default fact must have non-empty content")
		assert.NotEmpty(t, f.Category, "every default fact must have non-empty category")
		assert.Equal(t, "ru", f.Language, "default seed is Russian-only")
		assert.True(t, f.IsApproved, "default facts must be pre-approved")
		categorySet[f.Category] = true
	}

	// Sanity: multiple categories represented (history/world/science/etc).
	assert.GreaterOrEqual(t, len(categorySet), 5, "expected at least 5 distinct categories в seed")
}
