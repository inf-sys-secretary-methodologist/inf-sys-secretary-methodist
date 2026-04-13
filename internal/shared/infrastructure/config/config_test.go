package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	envVars := []string{
		"ENVIRONMENT", "VERSION", "SERVER_PORT", "SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT", "SERVER_BASE_URL",
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"LOG_LEVEL", "JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET",
		"JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_METHODS", "CORS_ALLOWED_HEADERS",
		"TRACING_ENABLED", "TRACING_OTLP_ENDPOINT", "TRACING_SAMPLING_RATE",
		"TRACING_SERVICE_NAME", "N8N_ENABLED", "N8N_WEBHOOK_URL",
		"AI_ENABLED", "AI_PROVIDER",
	}
	saved := make(map[string]string)
	for _, k := range envVars {
		saved[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	})

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "0.1.0", cfg.Version)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "http://localhost:8080", cfg.Server.BaseURL)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "secretary_methodist", cfg.Database.Database)
	assert.Equal(t, "postgres", cfg.Database.Username)
	assert.Equal(t, "postgres", cfg.Database.Password)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxLifetime)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)

	assert.Equal(t, "info", cfg.Log.Level)

	assert.Equal(t, []string{"http://localhost:3000"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, cfg.CORS.AllowedMethods)
	assert.Equal(t, []string{"Content-Type", "Authorization"}, cfg.CORS.AllowedHeaders)

	assert.Equal(t, "change-this-secret-in-production", cfg.JWT.AccessSecret)
	assert.Equal(t, "change-this-refresh-secret-in-production", cfg.JWT.RefreshSecret)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, cfg.JWT.RefreshTTL)

	assert.False(t, cfg.Tracing.Enabled)
	assert.Equal(t, "otel-collector:4317", cfg.Tracing.OTLPEndpoint)
	assert.InDelta(t, 0.1, cfg.Tracing.SamplingRate, 0.001)
	assert.Equal(t, "inf-sys-secretary-methodist", cfg.Tracing.ServiceName)

	assert.False(t, cfg.N8N.Enabled)
	assert.Equal(t, "http://localhost:5678", cfg.N8N.WebhookURL)

	assert.False(t, cfg.AI.Enabled)
	assert.Equal(t, "openai", cfg.AI.Provider)
}

func TestLoad_EnvOverrides(t *testing.T) {
	envs := map[string]string{
		"ENVIRONMENT":           "staging",
		"VERSION":               "1.2.3",
		"SERVER_PORT":           "9090",
		"SERVER_READ_TIMEOUT":   "30s",
		"DB_HOST":               "db.example.com",
		"DB_PORT":               "5433",
		"REDIS_HOST":            "redis.example.com",
		"REDIS_PORT":            "6380",
		"REDIS_DB":              "2",
		"LOG_LEVEL":             "debug",
		"JWT_ACCESS_SECRET":     "my-access-secret",
		"JWT_REFRESH_SECRET":    "my-refresh-secret",
		"JWT_ACCESS_TTL":        "1h",
		"CORS_ALLOWED_ORIGINS":  "https://example.com, https://app.example.com",
		"TRACING_ENABLED":       "true",
		"TRACING_SAMPLING_RATE": "0.5",
		"AI_ENABLED":            "true",
		"AI_TEMPERATURE":        "0.3",
		"S3_MAX_FILE_SIZE":      "104857600",
		"S3_USE_SSL":            "true",
	}

	saved := make(map[string]string)
	for k, v := range envs {
		saved[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	})

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "staging", cfg.Environment)
	assert.Equal(t, "1.2.3", cfg.Version)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, 2, cfg.Redis.DB)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "my-access-secret", cfg.JWT.AccessSecret)
	assert.Equal(t, "my-refresh-secret", cfg.JWT.RefreshSecret)
	assert.Equal(t, 1*time.Hour, cfg.JWT.AccessTTL)
	assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, cfg.CORS.AllowedOrigins)
	assert.True(t, cfg.Tracing.Enabled)
	assert.InDelta(t, 0.5, cfg.Tracing.SamplingRate, 0.001)
	assert.True(t, cfg.AI.Enabled)
	assert.InDelta(t, 0.3, cfg.AI.Temperature, 0.001)
	assert.Equal(t, int64(104857600), cfg.S3.MaxFileSize)
	assert.True(t, cfg.S3.UseSSL)
}

func TestLoad_ProductionValidation(t *testing.T) {
	saved := map[string]string{
		"ENVIRONMENT":        os.Getenv("ENVIRONMENT"),
		"JWT_ACCESS_SECRET":  os.Getenv("JWT_ACCESS_SECRET"),
		"JWT_REFRESH_SECRET": os.Getenv("JWT_REFRESH_SECRET"),
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	})

	os.Setenv("ENVIRONMENT", "production")
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")

	cfg, err := Load()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secrets must be set in production")

	os.Setenv("JWT_ACCESS_SECRET", "prod-access-secret")
	os.Setenv("JWT_REFRESH_SECRET", "prod-refresh-secret")
	cfg, err = Load()
	require.NoError(t, err)
	assert.Equal(t, "production", cfg.Environment)
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	db := DatabaseConfig{
		Host:     "myhost",
		Port:     5433,
		Username: "myuser",
		Password: "mypass",
		Database: "mydb",
	}

	dsn := db.GetDSN()
	assert.Equal(t, "host=myhost port=5433 user=myuser password=mypass dbname=mydb sslmode=disable", dsn)
}

func TestGetEnvHelpers(t *testing.T) {
	t.Run("getEnv", func(t *testing.T) {
		assert.Equal(t, "fallback", getEnv("TEST_NONEXISTENT_KEY_12345", "fallback"))
		os.Setenv("TEST_GETENV_HELPER", "hello")
		defer os.Unsetenv("TEST_GETENV_HELPER")
		assert.Equal(t, "hello", getEnv("TEST_GETENV_HELPER", "fallback"))
	})

	t.Run("getEnvAsInt", func(t *testing.T) {
		assert.Equal(t, 42, getEnvAsInt("TEST_NONEXISTENT_KEY_12345", 42))
		os.Setenv("TEST_INT_VAL", "100")
		defer os.Unsetenv("TEST_INT_VAL")
		assert.Equal(t, 100, getEnvAsInt("TEST_INT_VAL", 42))

		os.Setenv("TEST_INT_BAD", "notanumber")
		defer os.Unsetenv("TEST_INT_BAD")
		assert.Equal(t, 42, getEnvAsInt("TEST_INT_BAD", 42))
	})

	t.Run("getEnvAsInt64", func(t *testing.T) {
		assert.Equal(t, int64(999), getEnvAsInt64("TEST_NONEXISTENT_KEY_12345", 999))
		os.Setenv("TEST_INT64_VAL", "123456789")
		defer os.Unsetenv("TEST_INT64_VAL")
		assert.Equal(t, int64(123456789), getEnvAsInt64("TEST_INT64_VAL", 999))

		os.Setenv("TEST_INT64_BAD", "bad")
		defer os.Unsetenv("TEST_INT64_BAD")
		assert.Equal(t, int64(999), getEnvAsInt64("TEST_INT64_BAD", 999))
	})

	t.Run("getEnvAsBool", func(t *testing.T) {
		assert.False(t, getEnvAsBool("TEST_NONEXISTENT_KEY_12345", false))
		os.Setenv("TEST_BOOL_VAL", "true")
		defer os.Unsetenv("TEST_BOOL_VAL")
		assert.True(t, getEnvAsBool("TEST_BOOL_VAL", false))

		os.Setenv("TEST_BOOL_BAD", "notabool")
		defer os.Unsetenv("TEST_BOOL_BAD")
		assert.False(t, getEnvAsBool("TEST_BOOL_BAD", false))
	})

	t.Run("getEnvAsFloat", func(t *testing.T) {
		assert.InDelta(t, 1.5, getEnvAsFloat("TEST_NONEXISTENT_KEY_12345", 1.5), 0.001)
		os.Setenv("TEST_FLOAT_VAL", "3.14")
		defer os.Unsetenv("TEST_FLOAT_VAL")
		assert.InDelta(t, 3.14, getEnvAsFloat("TEST_FLOAT_VAL", 1.5), 0.001)

		os.Setenv("TEST_FLOAT_BAD", "notafloat")
		defer os.Unsetenv("TEST_FLOAT_BAD")
		assert.InDelta(t, 1.5, getEnvAsFloat("TEST_FLOAT_BAD", 1.5), 0.001)
	})

	t.Run("getEnvAsDuration", func(t *testing.T) {
		assert.Equal(t, 5*time.Second, getEnvAsDuration("TEST_NONEXISTENT_KEY_12345", 5*time.Second))
		os.Setenv("TEST_DUR_VAL", "30s")
		defer os.Unsetenv("TEST_DUR_VAL")
		assert.Equal(t, 30*time.Second, getEnvAsDuration("TEST_DUR_VAL", 5*time.Second))

		os.Setenv("TEST_DUR_BAD", "badduration")
		defer os.Unsetenv("TEST_DUR_BAD")
		assert.Equal(t, 5*time.Second, getEnvAsDuration("TEST_DUR_BAD", 5*time.Second))
	})

	t.Run("getEnvAsSlice", func(t *testing.T) {
		defaults := []string{"a", "b"}
		assert.Equal(t, defaults, getEnvAsSlice("TEST_NONEXISTENT_KEY_12345", defaults))

		os.Setenv("TEST_SLICE_VAL", "x, y, z")
		defer os.Unsetenv("TEST_SLICE_VAL")
		assert.Equal(t, []string{"x", "y", "z"}, getEnvAsSlice("TEST_SLICE_VAL", defaults))

		os.Setenv("TEST_SLICE_EMPTY", "")
		defer os.Unsetenv("TEST_SLICE_EMPTY")
		assert.Equal(t, defaults, getEnvAsSlice("TEST_SLICE_EMPTY", defaults))
	})
}

func TestSplitAndTrim(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitAndTrim("a, b, c", ","))
	assert.Equal(t, []string{"hello"}, splitAndTrim("hello", ","))
	assert.Empty(t, splitAndTrim("", ","))
	assert.Equal(t, []string{"x", "y"}, splitAndTrim("  x  ,  y  ", ","))
}

func TestSplitString(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitString("a,b,c", ","))
	assert.Equal(t, []string{"hello"}, splitString("hello", ","))
	assert.Empty(t, splitString("", ","))
	assert.Equal(t, []string{"a", "b"}, splitString("a::b", "::"))
}

func TestTrimSpace(t *testing.T) {
	assert.Equal(t, "hello", trimSpace("  hello  "))
	assert.Equal(t, "hello", trimSpace("\t\nhello\r\n"))
	assert.Equal(t, "", trimSpace("   "))
	assert.Equal(t, "", trimSpace(""))
	assert.Equal(t, "a b", trimSpace("  a b  "))
}
