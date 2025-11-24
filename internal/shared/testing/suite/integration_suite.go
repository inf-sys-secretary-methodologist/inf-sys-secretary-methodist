package suite

import (
	"database/sql"

	"github.com/stretchr/testify/suite"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/helpers"
)

// IntegrationSuite provides a base suite for integration tests
type IntegrationSuite struct {
	suite.Suite
	DB *sql.DB
}

// SetupSuite runs once before all tests in the suite
func (s *IntegrationSuite) SetupSuite() {
	s.DB = helpers.SetupTestDB(s.T())
}

// TearDownSuite runs once after all tests in the suite
func (s *IntegrationSuite) TearDownSuite() {
	if s.DB != nil {
		s.DB.Close()
	}
}

// SetupTest runs before each test
func (s *IntegrationSuite) SetupTest() {
	// Override in specific test suites if needed
}

// TearDownTest runs after each test
func (s *IntegrationSuite) TearDownTest() {
	// Override in specific test suites if needed
}

// CleanupTables is a helper to cleanup specific tables after tests
func (s *IntegrationSuite) CleanupTables(tables ...string) {
	helpers.CleanupTestDB(s.T(), s.DB, tables...)
}

// TruncateTables is a helper to truncate specific tables after tests
func (s *IntegrationSuite) TruncateTables(tables ...string) {
	helpers.TruncateTestDB(s.T(), s.DB, tables...)
}
