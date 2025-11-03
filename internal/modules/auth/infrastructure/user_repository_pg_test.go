package persistence_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	persistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	testSuite "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/suite"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/fixtures"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/helpers"
)

// UserRepositoryTestSuite tests the PostgreSQL user repository
type UserRepositoryTestSuite struct {
	testSuite.IntegrationSuite
	repo *persistence.UserRepositoryPG
}

// SetupSuite runs once before all tests
func (s *UserRepositoryTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = persistence.NewUserRepositoryPG(s.DB)
}

// TearDownTest runs after each test
func (s *UserRepositoryTestSuite) TearDownTest() {
	s.TruncateTables("users")
}

// TestCreate tests user creation
func (s *UserRepositoryTestSuite) TestCreate() {
	ctx := helpers.TestContext(s.T())
	user := fixtures.NewUserBuilder().
		WithEmail("newuser@example.com").
		Build()

	err := s.repo.Create(ctx, user)
	s.NoError(err)
	s.NotZero(user.ID, "User ID should be set after creation")
	s.NotZero(user.CreatedAt, "CreatedAt should be set")
	s.NotZero(user.UpdatedAt, "UpdatedAt should be set")
}

// TestCreateDuplicateEmail tests that creating user with duplicate email fails
func (s *UserRepositoryTestSuite) TestCreateDuplicateEmail() {
	ctx := helpers.TestContext(s.T())
	email := "duplicate@example.com"

	user1 := fixtures.NewUserBuilder().WithEmail(email).Build()
	err := s.repo.Create(ctx, user1)
	s.NoError(err)

	user2 := fixtures.NewUserBuilder().WithEmail(email).Build()
	err = s.repo.Create(ctx, user2)
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrAlreadyExists)
}

// TestGetByID tests getting user by ID
func (s *UserRepositoryTestSuite) TestGetByID() {
	ctx := helpers.TestContext(s.T())
	user := fixtures.StudentUser()

	err := s.repo.Create(ctx, user)
	s.NoError(err)

	fetched, err := s.repo.GetByID(ctx, user.ID)
	s.NoError(err)
	s.Equal(user.ID, fetched.ID)
	s.Equal(user.Email, fetched.Email)
	s.Equal(user.Role, fetched.Role)
}

// TestGetByIDNotFound tests getting non-existent user
func (s *UserRepositoryTestSuite) TestGetByIDNotFound() {
	ctx := helpers.TestContext(s.T())

	_, err := s.repo.GetByID(ctx, 99999)
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrNotFound)
}

// TestGetByEmail tests getting user by email
func (s *UserRepositoryTestSuite) TestGetByEmail() {
	ctx := helpers.TestContext(s.T())
	user := fixtures.TeacherUser()

	err := s.repo.Create(ctx, user)
	s.NoError(err)

	fetched, err := s.repo.GetByEmail(ctx, user.Email)
	s.NoError(err)
	s.Equal(user.ID, fetched.ID)
	s.Equal(user.Email, fetched.Email)
}

// TestGetByEmailNotFound tests getting user with non-existent email
func (s *UserRepositoryTestSuite) TestGetByEmailNotFound() {
	ctx := helpers.TestContext(s.T())

	_, err := s.repo.GetByEmail(ctx, "nonexistent@example.com")
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrNotFound)
}

// TestSave tests updating user
func (s *UserRepositoryTestSuite) TestSave() {
	ctx := helpers.TestContext(s.T())
	user := fixtures.StudentUser()

	err := s.repo.Create(ctx, user)
	s.NoError(err)

	// Update user
	user.Email = "updated@example.com"
	err = s.repo.Save(ctx, user)
	s.NoError(err)

	// Verify update
	fetched, err := s.repo.GetByID(ctx, user.ID)
	s.NoError(err)
	s.Equal("updated@example.com", fetched.Email)
}

// TestDelete tests deleting user
func (s *UserRepositoryTestSuite) TestDelete() {
	ctx := helpers.TestContext(s.T())
	user := fixtures.StudentUser()

	err := s.repo.Create(ctx, user)
	s.NoError(err)

	err = s.repo.Delete(ctx, user.ID)
	s.NoError(err)

	// Verify deletion
	_, err = s.repo.GetByID(ctx, user.ID)
	s.Error(err)
	s.ErrorIs(err, domainErrors.ErrNotFound)
}

// TestList tests listing users
func (s *UserRepositoryTestSuite) TestList() {
	ctx := helpers.TestContext(s.T())

	// Create multiple users
	users := []*fixtures.UserBuilder{
		fixtures.NewUserBuilder().WithEmail("user1@example.com"),
		fixtures.NewUserBuilder().WithEmail("user2@example.com"),
		fixtures.NewUserBuilder().WithEmail("user3@example.com"),
	}

	for _, builder := range users {
		user := builder.Build()
		err := s.repo.Create(ctx, user)
		s.NoError(err)
	}

	// Test listing
	fetched, err := s.repo.List(ctx, 10, 0)
	s.NoError(err)
	s.Len(fetched, 3)
}

// TestListWithPagination tests pagination
func (s *UserRepositoryTestSuite) TestListWithPagination() {
	ctx := helpers.TestContext(s.T())

	// Create 5 users
	for i := 1; i <= 5; i++ {
		user := fixtures.NewUserBuilder().
			WithEmail(fmt.Sprintf("user%d@example.com", i)).
			Build()
		err := s.repo.Create(ctx, user)
		s.NoError(err)
	}

	// First page
	page1, err := s.repo.List(ctx, 2, 0)
	s.NoError(err)
	s.Len(page1, 2)

	// Second page
	page2, err := s.repo.List(ctx, 2, 2)
	s.NoError(err)
	s.Len(page2, 2)

	// Verify no overlap
	s.NotEqual(page1[0].ID, page2[0].ID)
}

// TestSuite runs the test suite
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
