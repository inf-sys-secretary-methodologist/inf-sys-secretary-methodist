package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// MockFunFactRepo is a mock implementation of FunFactRepository
type MockFunFactRepo struct {
	mock.Mock
}

func (m *MockFunFactRepo) Create(_ context.Context, fact *entities.FunFact) error {
	args := m.Called(fact)
	return args.Error(0)
}

func (m *MockFunFactRepo) BulkCreate(_ context.Context, facts []entities.FunFact) error {
	args := m.Called(facts)
	return args.Error(0)
}

func (m *MockFunFactRepo) GetRandom(_ context.Context) (*entities.FunFact, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FunFact), args.Error(1)
}

func (m *MockFunFactRepo) GetLeastUsed(_ context.Context) (*entities.FunFact, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FunFact), args.Error(1)
}

func (m *MockFunFactRepo) IncrementUsedCount(_ context.Context, id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFunFactRepo) Count(_ context.Context) (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestGetRandomFact_Success(t *testing.T) {
	mockRepo := new(MockFunFactRepo)
	ps := services.NewPersonalityService()
	uc := NewFunFactUseCase(mockRepo, ps)

	fact := &entities.FunFact{
		ID:       1,
		Content:  "В 1956 году Дартмутский семинар положил начало ИИ как науке.",
		Category: "history",
		Source:   "Wikipedia",
	}

	mockRepo.On("GetRandom").Return(fact, nil)
	mockRepo.On("IncrementUsedCount", int64(1)).Return(nil)

	result, err := uc.GetRandomFact(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, fact.Content, result.Content)
	assert.Equal(t, "history", result.Category)
	assert.Equal(t, "Wikipedia", result.Source)
	mockRepo.AssertExpectations(t)
}

func TestGetRandomFact_EmptyTable(t *testing.T) {
	mockRepo := new(MockFunFactRepo)
	ps := services.NewPersonalityService()
	uc := NewFunFactUseCase(mockRepo, ps)

	mockRepo.On("GetRandom").Return(nil, nil)

	result, err := uc.GetRandomFact(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Content, "Методыч пока не нашёл")
	assert.Equal(t, "system", result.Category)
	mockRepo.AssertExpectations(t)
}

func TestGetRandomFact_RepoError(t *testing.T) {
	mockRepo := new(MockFunFactRepo)
	ps := services.NewPersonalityService()
	uc := NewFunFactUseCase(mockRepo, ps)

	mockRepo.On("GetRandom").Return(nil, errors.New("database error"))

	result, err := uc.GetRandomFact(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get random fact")
	mockRepo.AssertExpectations(t)
}

func TestGetRandomFact_IncrementError(t *testing.T) {
	mockRepo := new(MockFunFactRepo)
	ps := services.NewPersonalityService()
	uc := NewFunFactUseCase(mockRepo, ps)

	fact := &entities.FunFact{
		ID:       2,
		Content:  "Факт об образовании",
		Category: "education",
		Source:   "Источник",
	}

	mockRepo.On("GetRandom").Return(fact, nil)
	mockRepo.On("IncrementUsedCount", int64(2)).Return(errors.New("increment failed"))

	result, err := uc.GetRandomFact(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.ID)
	assert.Equal(t, fact.Content, result.Content)
	mockRepo.AssertExpectations(t)
}
