package persistence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// mockUserRepo is a mock implementation of UserRepository
type mockUserRepo struct {
	createFn          func(ctx context.Context, user *entities.User) error
	saveFn            func(ctx context.Context, user *entities.User) error
	getByIDFn         func(ctx context.Context, id int64) (*entities.User, error)
	getByEmailFn      func(ctx context.Context, email string) (*entities.User, error)
	getByEmailForAuth func(ctx context.Context, email string) (*entities.User, error)
	deleteFn          func(ctx context.Context, id int64) error
	listFn            func(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) Save(ctx context.Context, user *entities.User) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &entities.User{ID: id}, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return &entities.User{Email: email}, nil
}
func (m *mockUserRepo) GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error) {
	if m.getByEmailForAuth != nil {
		return m.getByEmailForAuth(ctx, email)
	}
	return &entities.User{Email: email}, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return []*entities.User{}, nil
}

func newTestCachedRepo(t *testing.T, repo repositories.UserRepository) *CachedUserRepository {
	t.Helper()
	mr := miniredis.RunT(t)
	redisCache, err := cache.NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	logger := logging.NewLogger("test")
	perfLog := logging.NewPerformanceLogger(logger)
	userCache := cache.NewUserCache(redisCache, 5*time.Minute)
	return &CachedUserRepository{
		repo:      repo,
		userCache: userCache,
		perfLog:   perfLog,
	}
}

func TestNewCachedUserRepository(t *testing.T) {
	mr := miniredis.RunT(t)
	redisCache, err := cache.NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	logger := logging.NewLogger("test")
	perfLog := logging.NewPerformanceLogger(logger)
	userCache := cache.NewUserCache(redisCache, 5*time.Minute)
	repo := NewCachedUserRepository(&mockUserRepo{}, userCache, perfLog)
	assert.NotNil(t, repo)
}

func TestCachedRepo_Create_Success(t *testing.T) {
	mockRepo := &mockUserRepo{}
	cached := newTestCachedRepo(t, mockRepo)
	user := entities.NewUser("test@test.com", "hash", "Test", domain.RoleTeacher)
	require.NoError(t, cached.Create(context.Background(), user))
}

func TestCachedRepo_Create_Error(t *testing.T) {
	mockRepo := &mockUserRepo{
		createFn: func(_ context.Context, _ *entities.User) error { return fmt.Errorf("err") },
	}
	cached := newTestCachedRepo(t, mockRepo)
	user := entities.NewUser("test@test.com", "hash", "Test", domain.RoleTeacher)
	assert.Error(t, cached.Create(context.Background(), user))
}

func TestCachedRepo_Save_Success(t *testing.T) {
	mockRepo := &mockUserRepo{}
	cached := newTestCachedRepo(t, mockRepo)
	user := &entities.User{ID: 1, Email: "test@test.com"}
	require.NoError(t, cached.Save(context.Background(), user))
}

func TestCachedRepo_Save_Error(t *testing.T) {
	mockRepo := &mockUserRepo{
		saveFn: func(_ context.Context, _ *entities.User) error { return fmt.Errorf("err") },
	}
	cached := newTestCachedRepo(t, mockRepo)
	user := &entities.User{ID: 1}
	assert.Error(t, cached.Save(context.Background(), user))
}

func TestCachedRepo_GetByID_CacheMiss(t *testing.T) {
	mockRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, id int64) (*entities.User, error) {
			return &entities.User{ID: id, Name: "Test"}, nil
		},
	}
	cached := newTestCachedRepo(t, mockRepo)
	user, err := cached.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Test", user.Name)
}

func TestCachedRepo_GetByID_Error(t *testing.T) {
	mockRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ int64) (*entities.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	cached := newTestCachedRepo(t, mockRepo)
	_, err := cached.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCachedRepo_GetByEmail_CacheMiss(t *testing.T) {
	mockRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (*entities.User, error) {
			return &entities.User{ID: 1, Email: email}, nil
		},
	}
	cached := newTestCachedRepo(t, mockRepo)
	user, err := cached.GetByEmail(context.Background(), "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", user.Email)
}

func TestCachedRepo_GetByEmail_Error(t *testing.T) {
	mockRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	cached := newTestCachedRepo(t, mockRepo)
	_, err := cached.GetByEmail(context.Background(), "nonexistent@test.com")
	assert.Error(t, err)
}

func TestCachedRepo_GetByEmailForAuth(t *testing.T) {
	mockRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (*entities.User, error) {
			return &entities.User{ID: 1, Email: email, Password: "hash"}, nil
		},
	}
	cached := newTestCachedRepo(t, mockRepo)
	user, err := cached.GetByEmailForAuth(context.Background(), "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", user.Email)
}

func TestCachedRepo_Delete_Success(t *testing.T) {
	mockRepo := &mockUserRepo{}
	cached := newTestCachedRepo(t, mockRepo)
	require.NoError(t, cached.Delete(context.Background(), 1))
}

func TestCachedRepo_Delete_Error(t *testing.T) {
	mockRepo := &mockUserRepo{
		deleteFn: func(_ context.Context, _ int64) error { return fmt.Errorf("err") },
	}
	cached := newTestCachedRepo(t, mockRepo)
	assert.Error(t, cached.Delete(context.Background(), 1))
}

func TestCachedRepo_List(t *testing.T) {
	mockRepo := &mockUserRepo{
		listFn: func(_ context.Context, _, _ int) ([]*entities.User, error) {
			return []*entities.User{{ID: 1}, {ID: 2}}, nil
		},
	}
	cached := newTestCachedRepo(t, mockRepo)
	users, err := cached.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}
