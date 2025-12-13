// Package config handles application configuration loading and management.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Version     string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	S3          S3Config
	Log         LogConfig
	CORS        CORSConfig
	JWT         JWTConfig
	Composio    ComposioConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BaseURL      string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// ComposioConfig holds Composio integration configuration
type ComposioConfig struct {
	APIKey      string
	EntityID    string
	MCPConfigID string
}

// S3Config holds S3/MinIO storage configuration
type S3Config struct {
	Endpoint        string
	PublicEndpoint  string // External endpoint for presigned URLs (e.g., localhost:9000)
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	UseSSL          bool
	MaxFileSize     int64 // max file size in bytes
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("VERSION", "0.1.0"),
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			BaseURL:      getEnv("SERVER_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Database:        getEnv("DB_NAME", "secretary_methodist"),
			Username:        getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "change-this-secret-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "change-this-refresh-secret-in-production"),
			AccessTTL:     getEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvAsDuration("JWT_REFRESH_TTL", 7*24*time.Hour), // 7 days
		},
		Composio: ComposioConfig{
			APIKey:      getEnv("COMPOSIO_API_KEY", ""),
			EntityID:    getEnv("COMPOSIO_ENTITY_ID", ""),
			MCPConfigID: getEnv("COMPOSIO_MCP_CONFIG_ID", ""),
		},
		S3: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", "localhost:9000"),
			PublicEndpoint:  getEnv("S3_PUBLIC_ENDPOINT", getEnv("S3_ENDPOINT", "localhost:9000")),
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", "minioadmin"),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", "minioadmin"),
			BucketName:      getEnv("S3_BUCKET_NAME", "documents"),
			Region:          getEnv("S3_REGION", "us-east-1"),
			UseSSL:          getEnvAsBool("S3_USE_SSL", false),
			MaxFileSize:     getEnvAsInt64("S3_MAX_FILE_SIZE", 50*1024*1024), // 50MB default
		},
	}

	// Validate JWT secrets in production
	if config.Environment == "production" {
		if config.JWT.AccessSecret == "change-this-secret-in-production" ||
			config.JWT.RefreshSecret == "change-this-refresh-secret-in-production" {
			return nil, fmt.Errorf("JWT secrets must be set in production environment")
		}
	}

	return config, nil
}

// GetDSN returns database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	// Split by comma, support spaces
	var result []string
	for _, v := range splitAndTrim(valueStr, ",") {
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
