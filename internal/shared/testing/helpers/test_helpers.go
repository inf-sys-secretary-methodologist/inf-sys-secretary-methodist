// Package helpers provides test helper utilities.
package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver for integration tests

	"github.com/stretchr/testify/require"
)

// TestContext creates a context with timeout for tests
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// SetupTestDB sets up a test database connection
// For now, it returns a connection to the actual test database
// In the future, this can be replaced with testcontainers
func SetupTestDB(t *testing.T) *sql.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=inf_sys_db_test sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	if err != nil {
		t.Skip("Test database not available, skipping test")
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// CleanupTestDB cleans up test data from database
func CleanupTestDB(t *testing.T, db *sql.DB, tables ...string) {
	ctx := TestContext(t)

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table) // #nosec G201 -- table name is from test parameters, not user input //nolint:gosec
		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err, "Failed to cleanup table %s", table)
	}
}

// TruncateTestDB truncates tables in test database
func TruncateTestDB(t *testing.T, db *sql.DB, tables ...string) {
	ctx := TestContext(t)

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err, "Failed to truncate table %s", table)
	}
}

// AssertErrorCode checks if error has expected code
func AssertErrorCode(t *testing.T, err error, expectedCode string) {
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedCode)
}

// MustExec executes SQL and fails test on error
func MustExec(t *testing.T, db *sql.DB, query string, args ...interface{}) {
	ctx := TestContext(t)
	_, err := db.ExecContext(ctx, query, args...)
	require.NoError(t, err, "Failed to execute query")
}

// MustQuery executes SQL query and fails test on error
func MustQuery(t *testing.T, db *sql.DB, query string, args ...interface{}) *sql.Rows {
	ctx := TestContext(t)
	rows, err := db.QueryContext(ctx, query, args...)
	require.NoError(t, err, "Failed to execute query")
	return rows
}
