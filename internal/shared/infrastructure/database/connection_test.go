package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
)

func TestNewConnection_PingFails(t *testing.T) {
	// Use a valid DSN format but with an unreachable host so sql.Open succeeds but Ping fails
	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            19999, // unlikely to have postgres here
		Database:        "testdb",
		Username:        "testuser",
		Password:        "testpass",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 1 * time.Minute,
	}

	db, err := NewConnection(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to ping database")
}
