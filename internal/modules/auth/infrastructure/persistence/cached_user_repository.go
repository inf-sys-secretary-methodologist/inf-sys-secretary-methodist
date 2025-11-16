package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// CachedUserRepository wraps UserRepository with caching layer
type CachedUserRepository struct {
	repo      repositories.UserRepository
	userCache *cache.UserCache
	perfLog   *logging.PerformanceLogger
}

// NewCachedUserRepository creates a new cached repository
func NewCachedUserRepository(
	repo repositories.UserRepository,
	userCache *cache.UserCache,
	perfLog *logging.PerformanceLogger,
) repositories.UserRepository {
	return &CachedUserRepository{
		repo:      repo,
		userCache: userCache,
		perfLog:   perfLog,
	}
}

// Create creates a user and invalidates cache
func (r *CachedUserRepository) Create(ctx context.Context, user *entities.User) error {
	if err := r.repo.Create(ctx, user); err != nil {
		return err
	}

	// Cache the newly created user
	_ = r.userCache.SetUser(ctx, user.ID, user)
	_ = r.userCache.SetUserByEmail(ctx, user.Email, user)

	return nil
}

// Save updates a user and invalidates cache
func (r *CachedUserRepository) Save(ctx context.Context, user *entities.User) error {
	if err := r.repo.Save(ctx, user); err != nil {
		return err
	}

	// Invalidate cache
	_ = r.userCache.InvalidateUser(ctx, user.ID)

	return nil
}

// GetByID gets user by ID with caching
func (r *CachedUserRepository) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	// Try cache first
	var user entities.User
	hit, err := r.userCache.GetUser(ctx, id, &user)
	if err == nil && hit {
		r.perfLog.LogCacheOperation(ctx, "get_by_id", fmt.Sprintf("user:%d", id), true)
		return &user, nil
	}

	r.perfLog.LogCacheOperation(ctx, "get_by_id", fmt.Sprintf("user:%d", id), false)

	// Cache miss - get from database
	start := time.Now()
	dbUser, err := r.repo.GetByID(ctx, id)
	duration := time.Since(start)

	r.perfLog.LogDatabaseQuery(ctx, "SELECT user BY ID", duration, 1)

	if err != nil {
		return nil, err
	}

	// Update cache
	_ = r.userCache.SetUser(ctx, dbUser.ID, dbUser)

	return dbUser, nil
}

// GetByEmail gets user by email with caching
func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	// Try cache first
	var user entities.User
	hit, err := r.userCache.GetUserByEmail(ctx, email, &user)
	if err == nil && hit {
		r.perfLog.LogCacheOperation(ctx, "get_by_email", fmt.Sprintf("user:email:%s", email), true)
		return &user, nil
	}

	r.perfLog.LogCacheOperation(ctx, "get_by_email", fmt.Sprintf("user:email:%s", email), false)

	// Cache miss - get from database
	start := time.Now()
	dbUser, err := r.repo.GetByEmail(ctx, email)
	duration := time.Since(start)

	r.perfLog.LogDatabaseQuery(ctx, "SELECT user BY EMAIL", duration, 1)

	if err != nil {
		return nil, err
	}

	// Update cache
	_ = r.userCache.SetUser(ctx, dbUser.ID, dbUser)
	_ = r.userCache.SetUserByEmail(ctx, email, dbUser)

	return dbUser, nil
}

// GetByEmailForAuth gets user by email for authentication purposes
// ALWAYS bypasses cache and fetches directly from database to ensure password field is populated
// Password field has json:"-" tag so it's excluded from cache serialization
func (r *CachedUserRepository) GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error) {
	start := time.Now()
	dbUser, err := r.repo.GetByEmail(ctx, email)
	duration := time.Since(start)

	r.perfLog.LogDatabaseQuery(ctx, "SELECT user BY EMAIL FOR AUTH", duration, 1)

	if err != nil {
		return nil, err
	}

	return dbUser, nil
}

// Delete deletes a user and invalidates cache
func (r *CachedUserRepository) Delete(ctx context.Context, id int64) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	_ = r.userCache.InvalidateUser(ctx, id)

	return nil
}

// List lists users (no caching for lists to avoid stale data)
func (r *CachedUserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	start := time.Now()
	users, err := r.repo.List(ctx, limit, offset)
	duration := time.Since(start)

	r.perfLog.LogDatabaseQuery(ctx, "SELECT users LIST", duration, int64(len(users)))

	return users, err
}
