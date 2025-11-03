package persistence_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure/persistence"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	testSuite "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/suite"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/fixtures"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/helpers"
)

// SessionRepositoryTestSuite tests the PostgreSQL session repository
type SessionRepositoryTestSuite struct {
	testSuite.IntegrationSuite
	repo   repositories.SessionRepository
	userID int64
}

// SetupSuite runs once before all tests
func (s *SessionRepositoryTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = persistence.NewSessionRepositoryPG(s.DB)

	// Create a test user
	ctx := helpers.TestContext(s.T())
	userRepo := persistence.NewUserRepositoryPG(s.DB)
	user := fixtures.StudentUser()
	err := userRepo.Create(ctx, user)
	s.Require().NoError(err)
	s.userID = user.ID
}

// TearDownTest runs after each test
func (s *SessionRepositoryTestSuite) TearDownTest() {
	s.TruncateTables("sessions")
}

// TearDownSuite runs once after all tests
func (s *SessionRepositoryTestSuite) TearDownSuite() {
	s.TruncateTables("users")
}

// TestCreate tests session creation
func (s *SessionRepositoryTestSuite) TestCreate() {
	ctx := helpers.TestContext(s.T())
	session := &entities.Session{
		UserID:       s.userID,
		RefreshToken: "test_token_123",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, session)
	s.NoError(err)
	s.NotZero(session.ID, "Session ID should be set after creation")
}

// TestCreateDuplicateToken tests that creating session with duplicate token fails
func (s *SessionRepositoryTestSuite) TestCreateDuplicateToken() {
	ctx := helpers.TestContext(s.T())
	token := "duplicate_token"

	session1 := &entities.Session{
		UserID:       s.userID,
		RefreshToken: token,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := s.repo.Create(ctx, session1)
	s.NoError(err)

	session2 := &entities.Session{
		UserID:       s.userID,
		RefreshToken: token,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.2",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = s.repo.Create(ctx, session2)
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrAlreadyExists)
}

// TestGetByRefreshToken tests getting session by token
func (s *SessionRepositoryTestSuite) TestGetByRefreshToken() {
	ctx := helpers.TestContext(s.T())
	token := "test_token_get"
	session := &entities.Session{
		UserID:       s.userID,
		RefreshToken: token,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, session)
	s.NoError(err)

	fetched, err := s.repo.GetByRefreshToken(ctx, token)
	s.NoError(err)
	s.Equal(session.ID, fetched.ID)
	s.Equal(session.RefreshToken, fetched.RefreshToken)
	s.Equal(session.UserID, fetched.UserID)
}

// TestGetByRefreshTokenNotFound tests getting non-existent session
func (s *SessionRepositoryTestSuite) TestGetByRefreshTokenNotFound() {
	ctx := helpers.TestContext(s.T())

	_, err := s.repo.GetByRefreshToken(ctx, "nonexistent_token")
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrNotFound)
}

// TestDelete tests deleting session
func (s *SessionRepositoryTestSuite) TestDelete() {
	ctx := helpers.TestContext(s.T())
	token := "test_token_delete"
	session := &entities.Session{
		UserID:       s.userID,
		RefreshToken: token,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, session)
	s.NoError(err)

	err = s.repo.Delete(ctx, token)
	s.NoError(err)

	// Verify deletion
	_, err = s.repo.GetByRefreshToken(ctx, token)
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrNotFound)
}

// TestDeleteByUserID tests deleting all user sessions
func (s *SessionRepositoryTestSuite) TestDeleteByUserID() {
	ctx := helpers.TestContext(s.T())

	// Create multiple sessions for user
	for i := 0; i < 3; i++ {
		session := &entities.Session{
			UserID:       s.userID,
			RefreshToken: fmt.Sprintf("token_%d", i),
			UserAgent:    "Mozilla/5.0",
			IPAddress:    "192.168.1.1",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := s.repo.Create(ctx, session)
		s.NoError(err)
	}

	err := s.repo.DeleteByUserID(ctx, s.userID)
	s.NoError(err)

	// Verify all sessions deleted
	sessions, err := s.repo.GetActiveByUserID(ctx, s.userID)
	s.NoError(err)
	s.Empty(sessions)
}

// TestDeleteExpired tests deleting expired sessions
func (s *SessionRepositoryTestSuite) TestDeleteExpired() {
	ctx := helpers.TestContext(s.T())

	// Create expired session
	expiredSession := &entities.Session{
		UserID:       s.userID,
		RefreshToken: "expired_token",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := s.repo.Create(ctx, expiredSession)
	s.NoError(err)

	// Create active session
	activeSession := &entities.Session{
		UserID:       s.userID,
		RefreshToken: "active_token",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour), // Active
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = s.repo.Create(ctx, activeSession)
	s.NoError(err)

	// Delete expired
	err = s.repo.DeleteExpired(ctx)
	s.NoError(err)

	// Verify expired is deleted
	_, err = s.repo.GetByRefreshToken(ctx, "expired_token")
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrNotFound)

	// Verify active still exists
	_, err = s.repo.GetByRefreshToken(ctx, "active_token")
	s.NoError(err)
}

// TestGetActiveByUserID tests getting active sessions for user
func (s *SessionRepositoryTestSuite) TestGetActiveByUserID() {
	ctx := helpers.TestContext(s.T())

	// Create active sessions
	for i := 0; i < 2; i++ {
		session := &entities.Session{
			UserID:       s.userID,
			RefreshToken: fmt.Sprintf("active_%d", i),
			UserAgent:    "Mozilla/5.0",
			IPAddress:    "192.168.1.1",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := s.repo.Create(ctx, session)
		s.NoError(err)
	}

	// Create expired session
	expiredSession := &entities.Session{
		UserID:       s.userID,
		RefreshToken: "expired",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := s.repo.Create(ctx, expiredSession)
	s.NoError(err)

	// Get active sessions
	sessions, err := s.repo.GetActiveByUserID(ctx, s.userID)
	s.NoError(err)
	s.Len(sessions, 2, "Should return only active sessions")
}

// TestSuite runs the test suite
func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}
