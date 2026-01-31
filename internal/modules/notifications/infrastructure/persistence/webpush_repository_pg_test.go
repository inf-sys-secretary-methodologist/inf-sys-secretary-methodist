package persistence_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/infrastructure/persistence"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/helpers"
	testSuite "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/suite"
)

// WebPushRepositoryTestSuite tests the PostgreSQL WebPush repository
type WebPushRepositoryTestSuite struct {
	testSuite.IntegrationSuite
	repo repositories.WebPushRepository
}

// SetupSuite runs once before all tests
func (s *WebPushRepositoryTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = persistence.NewWebPushRepositoryPG(s.DB)
}

// TearDownTest runs after each test
func (s *WebPushRepositoryTestSuite) TearDownTest() {
	s.TruncateTables("webpush_subscriptions")
}

// createTestSubscription creates a test subscription
func (s *WebPushRepositoryTestSuite) createTestSubscription(userID int64, endpoint string) *entities.WebPushSubscription {
	sub := entities.NewWebPushSubscription(
		userID,
		endpoint,
		"test-p256dh-key-base64",
		"test-auth-key-base64",
	)
	sub.UserAgent = "Mozilla/5.0 Test Browser"
	sub.DeviceName = "Test Device"
	return sub
}

// TestCreate tests subscription creation
func (s *WebPushRepositoryTestSuite) TestCreate() {
	ctx := helpers.TestContext(s.T())
	sub := s.createTestSubscription(1, "https://push.example.com/sub1")

	err := s.repo.Create(ctx, sub)
	s.NoError(err)
	s.NotZero(sub.ID, "Subscription ID should be set after creation")
	s.NotZero(sub.CreatedAt, "CreatedAt should be set")
	s.True(sub.IsActive, "Subscription should be active by default")
}

// TestCreateDuplicateEndpoint tests that creating subscription with duplicate endpoint updates it
func (s *WebPushRepositoryTestSuite) TestCreateDuplicateEndpoint() {
	ctx := helpers.TestContext(s.T())
	endpoint := "https://push.example.com/duplicate"

	sub1 := s.createTestSubscription(1, endpoint)
	err := s.repo.Create(ctx, sub1)
	s.NoError(err)
	originalID := sub1.ID

	// Create with same endpoint but different user should update
	sub2 := s.createTestSubscription(2, endpoint)
	sub2.DeviceName = "Updated Device"
	err = s.repo.Create(ctx, sub2)
	s.NoError(err)

	// Should have same ID (upsert)
	s.Equal(originalID, sub2.ID)
}

// TestGetByID tests getting subscription by ID
func (s *WebPushRepositoryTestSuite) TestGetByID() {
	ctx := helpers.TestContext(s.T())
	sub := s.createTestSubscription(1, "https://push.example.com/getbyid")

	err := s.repo.Create(ctx, sub)
	s.NoError(err)

	fetched, err := s.repo.GetByID(ctx, sub.ID)
	s.NoError(err)
	s.NotNil(fetched)
	s.Equal(sub.ID, fetched.ID)
	s.Equal(sub.UserID, fetched.UserID)
	s.Equal(sub.Endpoint, fetched.Endpoint)
	s.Equal(sub.P256dhKey, fetched.P256dhKey)
	s.Equal(sub.AuthKey, fetched.AuthKey)
}

// TestGetByIDNotFound tests getting non-existent subscription
func (s *WebPushRepositoryTestSuite) TestGetByIDNotFound() {
	ctx := helpers.TestContext(s.T())

	fetched, err := s.repo.GetByID(ctx, 99999)
	s.NoError(err)
	s.Nil(fetched)
}

// TestGetByEndpoint tests getting subscription by endpoint
func (s *WebPushRepositoryTestSuite) TestGetByEndpoint() {
	ctx := helpers.TestContext(s.T())
	endpoint := "https://push.example.com/byendpoint"
	sub := s.createTestSubscription(1, endpoint)

	err := s.repo.Create(ctx, sub)
	s.NoError(err)

	fetched, err := s.repo.GetByEndpoint(ctx, endpoint)
	s.NoError(err)
	s.NotNil(fetched)
	s.Equal(sub.ID, fetched.ID)
	s.Equal(endpoint, fetched.Endpoint)
}

// TestGetByEndpointNotFound tests getting non-existent endpoint
func (s *WebPushRepositoryTestSuite) TestGetByEndpointNotFound() {
	ctx := helpers.TestContext(s.T())

	fetched, err := s.repo.GetByEndpoint(ctx, "https://nonexistent.com/sub")
	s.NoError(err)
	s.Nil(fetched)
}

// TestGetByUserID tests getting all subscriptions for a user
func (s *WebPushRepositoryTestSuite) TestGetByUserID() {
	ctx := helpers.TestContext(s.T())
	userID := int64(1)

	// Create multiple subscriptions for user
	sub1 := s.createTestSubscription(userID, "https://push.example.com/user1-sub1")
	sub2 := s.createTestSubscription(userID, "https://push.example.com/user1-sub2")
	sub3 := s.createTestSubscription(2, "https://push.example.com/user2-sub1") // Different user

	s.NoError(s.repo.Create(ctx, sub1))
	s.NoError(s.repo.Create(ctx, sub2))
	s.NoError(s.repo.Create(ctx, sub3))

	fetched, err := s.repo.GetByUserID(ctx, userID)
	s.NoError(err)
	s.Len(fetched, 2)
}

// TestGetActiveByUserID tests getting only active subscriptions
func (s *WebPushRepositoryTestSuite) TestGetActiveByUserID() {
	ctx := helpers.TestContext(s.T())
	userID := int64(1)

	// Create active and inactive subscriptions
	sub1 := s.createTestSubscription(userID, "https://push.example.com/active1")
	sub2 := s.createTestSubscription(userID, "https://push.example.com/active2")

	s.NoError(s.repo.Create(ctx, sub1))
	s.NoError(s.repo.Create(ctx, sub2))

	// Deactivate one
	s.NoError(s.repo.Deactivate(ctx, sub2.ID))

	fetched, err := s.repo.GetActiveByUserID(ctx, userID)
	s.NoError(err)
	s.Len(fetched, 1)
	s.Equal(sub1.ID, fetched[0].ID)
}

// TestUpdate tests subscription update
func (s *WebPushRepositoryTestSuite) TestUpdate() {
	ctx := helpers.TestContext(s.T())
	sub := s.createTestSubscription(1, "https://push.example.com/update")

	s.NoError(s.repo.Create(ctx, sub))

	sub.DeviceName = "Updated Device Name"
	sub.IsActive = false
	err := s.repo.Update(ctx, sub)
	s.NoError(err)

	fetched, err := s.repo.GetByID(ctx, sub.ID)
	s.NoError(err)
	s.Equal("Updated Device Name", fetched.DeviceName)
	s.False(fetched.IsActive)
}

// TestUpdateLastUsed tests updating last used timestamp
func (s *WebPushRepositoryTestSuite) TestUpdateLastUsed() {
	ctx := helpers.TestContext(s.T())
	sub := s.createTestSubscription(1, "https://push.example.com/lastused")

	s.NoError(s.repo.Create(ctx, sub))

	// Initially LastUsedAt should be nil
	fetched, _ := s.repo.GetByID(ctx, sub.ID)
	s.Nil(fetched.LastUsedAt)

	// Update last used
	err := s.repo.UpdateLastUsed(ctx, sub.ID)
	s.NoError(err)

	fetched, _ = s.repo.GetByID(ctx, sub.ID)
	s.NotNil(fetched.LastUsedAt)
	s.WithinDuration(time.Now(), *fetched.LastUsedAt, 5*time.Second)
}

// TestDeactivate tests subscription deactivation
func (s *WebPushRepositoryTestSuite) TestDeactivate() {
	ctx := helpers.TestContext(s.T())
	sub := s.createTestSubscription(1, "https://push.example.com/deactivate")

	s.NoError(s.repo.Create(ctx, sub))
	s.True(sub.IsActive)

	err := s.repo.Deactivate(ctx, sub.ID)
	s.NoError(err)

	fetched, _ := s.repo.GetByID(ctx, sub.ID)
	s.False(fetched.IsActive)
}

// TestDelete tests subscription deletion
func (s *WebPushRepositoryTestSuite) TestDelete() {
	ctx := helpers.TestContext(s.T())
	sub := s.createTestSubscription(1, "https://push.example.com/delete")

	s.NoError(s.repo.Create(ctx, sub))

	err := s.repo.Delete(ctx, sub.ID)
	s.NoError(err)

	fetched, _ := s.repo.GetByID(ctx, sub.ID)
	s.Nil(fetched)
}

// TestDeleteByEndpoint tests deletion by endpoint
func (s *WebPushRepositoryTestSuite) TestDeleteByEndpoint() {
	ctx := helpers.TestContext(s.T())
	endpoint := "https://push.example.com/deletebyendpoint"
	sub := s.createTestSubscription(1, endpoint)

	s.NoError(s.repo.Create(ctx, sub))

	err := s.repo.DeleteByEndpoint(ctx, endpoint)
	s.NoError(err)

	fetched, _ := s.repo.GetByEndpoint(ctx, endpoint)
	s.Nil(fetched)
}

// TestDeleteByUserID tests deletion of all user subscriptions
func (s *WebPushRepositoryTestSuite) TestDeleteByUserID() {
	ctx := helpers.TestContext(s.T())
	userID := int64(1)

	// Create multiple subscriptions
	sub1 := s.createTestSubscription(userID, "https://push.example.com/deluser1")
	sub2 := s.createTestSubscription(userID, "https://push.example.com/deluser2")
	sub3 := s.createTestSubscription(2, "https://push.example.com/deluser3") // Different user

	s.NoError(s.repo.Create(ctx, sub1))
	s.NoError(s.repo.Create(ctx, sub2))
	s.NoError(s.repo.Create(ctx, sub3))

	err := s.repo.DeleteByUserID(ctx, userID)
	s.NoError(err)

	// User 1's subscriptions should be deleted
	fetched, _ := s.repo.GetByUserID(ctx, userID)
	s.Len(fetched, 0)

	// User 2's subscription should remain
	fetched, _ = s.repo.GetByUserID(ctx, 2)
	s.Len(fetched, 1)
}

// TestCountByUserID tests counting subscriptions
func (s *WebPushRepositoryTestSuite) TestCountByUserID() {
	ctx := helpers.TestContext(s.T())
	userID := int64(1)

	// Initially zero
	count, err := s.repo.CountByUserID(ctx, userID)
	s.NoError(err)
	s.Equal(int64(0), count)

	// Add subscriptions
	s.NoError(s.repo.Create(ctx, s.createTestSubscription(userID, "https://push.example.com/count1")))
	s.NoError(s.repo.Create(ctx, s.createTestSubscription(userID, "https://push.example.com/count2")))
	s.NoError(s.repo.Create(ctx, s.createTestSubscription(userID, "https://push.example.com/count3")))

	count, err = s.repo.CountByUserID(ctx, userID)
	s.NoError(err)
	s.Equal(int64(3), count)
}

// TestSuite runs the test suite
func TestWebPushRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(WebPushRepositoryTestSuite))
}
