package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected Environment 'development', got '%s'", cfg.Environment)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host 'localhost', got '%s'", cfg.Database.Host)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "production")
	_ = os.Setenv("SERVER_PORT", "9000")
	_ = os.Setenv("DB_HOST", "db.example.com")
	_ = os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-for-production")
	_ = os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-for-production")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("SERVER_PORT")
		_ = os.Unsetenv("DB_HOST")
		_ = os.Unsetenv("JWT_ACCESS_SECRET")
		_ = os.Unsetenv("JWT_REFRESH_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected Environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("expected Server.Port 9000, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected Database.Host 'db.example.com', got '%s'", cfg.Database.Host)
	}
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	dbCfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if got := dbCfg.GetDSN(); got != expected {
		t.Errorf("GetDSN() = %v, want %v", got, expected)
	}
}

func TestGetEnv(t *testing.T) {
	_ = os.Setenv("TEST_KEY", "test_value")
	defer func() {
		_ = os.Unsetenv("TEST_KEY")
	}()

	if got := getEnv("TEST_KEY", "default"); got != "test_value" {
		t.Errorf("getEnv() = %v, want %v", got, "test_value")
	}
	if got := getEnv("NON_EXISTENT", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want %v", got, "default")
	}
}

func TestGetEnvAsInt(t *testing.T) {
	_ = os.Setenv("TEST_INT", "42")
	defer func() {
		_ = os.Unsetenv("TEST_INT")
	}()

	if got := getEnvAsInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 42)
	}
	if got := getEnvAsInt("NON_EXISTENT", 99); got != 99 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 99)
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	_ = os.Setenv("TEST_DURATION", "5s")
	defer func() {
		_ = os.Unsetenv("TEST_DURATION")
	}()

	if got := getEnvAsDuration("TEST_DURATION", 0); got != 5*time.Second {
		t.Errorf("getEnvAsDuration() = %v, want %v", got, 5*time.Second)
	}
	if got := getEnvAsDuration("NON_EXISTENT", 10*time.Second); got != 10*time.Second {
		t.Errorf("getEnvAsDuration() = %v, want %v", got, 10*time.Second)
	}
}
