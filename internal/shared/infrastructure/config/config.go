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
	Telegram    TelegramConfig
	Integration IntegrationConfig
	WebPush     WebPushConfig
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

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken      string
	BotUsername   string
	WebhookURL    string
	WebhookSecret string
}

// IntegrationConfig holds 1C integration configuration
type IntegrationConfig struct {
	Enabled          bool
	BaseURL          string        // 1C OData base URL (e.g., http://1c-server/base/odata/standard.odata)
	Username         string        // 1C Basic Auth username
	Password         string        // 1C Basic Auth password
	Timeout          time.Duration // HTTP request timeout
	MaxRetries       int           // Max retry attempts for failed requests
	RetryDelay       time.Duration // Delay between retries
	EmployeeCatalog  string        // 1C employee catalog name (e.g., "Catalog_Сотрудники")
	StudentCatalog   string        // 1C student catalog name (e.g., "Catalog_Студенты")
	SyncCronEmployee string        // Cron expression for employee sync (e.g., "0 */6 * * *")
	SyncCronStudent  string        // Cron expression for student sync (e.g., "0 */6 * * *")
	BatchSize        int           // Batch size for sync operations
}


// WebPushConfig contains VAPID configuration for Web Push notifications
type WebPushConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string // Usually mailto: or https:// URL
}

// S3Config holds S3/MinIO storage configuration
type S3Config struct {
	Endpoint        string
	PublicEndpoint  string // External endpoint for presigned URLs (e.g., localhost:9000)
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	UseSSL          bool  // SSL for internal connection to MinIO
	PublicUseSSL    bool  // SSL for public presigned URLs (via reverse proxy like Caddy)
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
			PublicUseSSL:    getEnvAsBool("S3_PUBLIC_USE_SSL", false),
			MaxFileSize:     getEnvAsInt64("S3_MAX_FILE_SIZE", 50*1024*1024), // 50MB default
		},
		Telegram: TelegramConfig{
			BotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
			BotUsername:   getEnv("TELEGRAM_BOT_USERNAME", ""),
			WebhookURL:    getEnv("TELEGRAM_WEBHOOK_URL", ""),
			WebhookSecret: getEnv("TELEGRAM_WEBHOOK_SECRET", ""),
		},
		Integration: IntegrationConfig{
			Enabled:          getEnvAsBool("INTEGRATION_1C_ENABLED", false),
			BaseURL:          getEnv("INTEGRATION_1C_BASE_URL", ""),
			Username:         getEnv("INTEGRATION_1C_USERNAME", ""),
			Password:         getEnv("INTEGRATION_1C_PASSWORD", ""),
			Timeout:          getEnvAsDuration("INTEGRATION_1C_TIMEOUT", 30*time.Second),
			MaxRetries:       getEnvAsInt("INTEGRATION_1C_MAX_RETRIES", 3),
			RetryDelay:       getEnvAsDuration("INTEGRATION_1C_RETRY_DELAY", 5*time.Second),
			EmployeeCatalog:  getEnv("INTEGRATION_1C_EMPLOYEE_CATALOG", "Catalog_Сотрудники"),
			StudentCatalog:   getEnv("INTEGRATION_1C_STUDENT_CATALOG", "Catalog_Студенты"),
			SyncCronEmployee: getEnv("INTEGRATION_1C_SYNC_CRON_EMPLOYEE", "0 */6 * * *"),
			SyncCronStudent:  getEnv("INTEGRATION_1C_SYNC_CRON_STUDENT", "0 */6 * * *"),
			BatchSize:        getEnvAsInt("INTEGRATION_1C_BATCH_SIZE", 100),
		},
		WebPush: WebPushConfig{
			VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
			VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
			VAPIDSubject:    getEnv("VAPID_SUBJECT", ""),
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
