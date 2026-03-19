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
	if cfg.Version != "0.1.0" {
		t.Errorf("expected Version '0.1.0', got '%s'", cfg.Version)
	}
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Errorf("expected Server.ReadTimeout 10s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 10*time.Second {
		t.Errorf("expected Server.WriteTimeout 10s, got %v", cfg.Server.WriteTimeout)
	}
	if cfg.Server.IdleTimeout != 120*time.Second {
		t.Errorf("expected Server.IdleTimeout 120s, got %v", cfg.Server.IdleTimeout)
	}
	if cfg.Server.BaseURL != "http://localhost:8080" {
		t.Errorf("expected Server.BaseURL 'http://localhost:8080', got '%s'", cfg.Server.BaseURL)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port 5432, got %d", cfg.Database.Port)
	}
	if cfg.Database.Database != "secretary_methodist" {
		t.Errorf("expected Database.Database 'secretary_methodist', got '%s'", cfg.Database.Database)
	}
	if cfg.Database.Username != "postgres" {
		t.Errorf("expected Database.Username 'postgres', got '%s'", cfg.Database.Username)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("expected Database.MaxOpenConns 25, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 5 {
		t.Errorf("expected Database.MaxIdleConns 5, got %d", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("expected Database.ConnMaxLifetime 5m, got %v", cfg.Database.ConnMaxLifetime)
	}
	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected Redis.Host 'localhost', got '%s'", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("expected Redis.Port 6379, got %d", cfg.Redis.Port)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("expected Redis.DB 0, got %d", cfg.Redis.DB)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected Log.Level 'info', got '%s'", cfg.Log.Level)
	}
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("unexpected CORS.AllowedOrigins: %v", cfg.CORS.AllowedOrigins)
	}
	if cfg.JWT.AccessTTL != 15*time.Minute {
		t.Errorf("expected JWT.AccessTTL 15m, got %v", cfg.JWT.AccessTTL)
	}
	if cfg.JWT.RefreshTTL != 7*24*time.Hour {
		t.Errorf("expected JWT.RefreshTTL 7d, got %v", cfg.JWT.RefreshTTL)
	}
	if cfg.S3.Endpoint != "localhost:9000" {
		t.Errorf("expected S3.Endpoint 'localhost:9000', got '%s'", cfg.S3.Endpoint)
	}
	if cfg.S3.BucketName != "documents" {
		t.Errorf("expected S3.BucketName 'documents', got '%s'", cfg.S3.BucketName)
	}
	if cfg.S3.MaxFileSize != 50*1024*1024 {
		t.Errorf("expected S3.MaxFileSize 50MB, got %d", cfg.S3.MaxFileSize)
	}
	if cfg.S3.UseSSL != false {
		t.Errorf("expected S3.UseSSL false")
	}
	if cfg.Integration.Enabled != false {
		t.Errorf("expected Integration.Enabled false")
	}
	if cfg.Integration.Timeout != 30*time.Second {
		t.Errorf("expected Integration.Timeout 30s, got %v", cfg.Integration.Timeout)
	}
	if cfg.Integration.MaxRetries != 3 {
		t.Errorf("expected Integration.MaxRetries 3, got %d", cfg.Integration.MaxRetries)
	}
	if cfg.Integration.BatchSize != 100 {
		t.Errorf("expected Integration.BatchSize 100, got %d", cfg.Integration.BatchSize)
	}
	if cfg.Tracing.Enabled != false {
		t.Errorf("expected Tracing.Enabled false")
	}
	if cfg.Tracing.SamplingRate != 0.1 {
		t.Errorf("expected Tracing.SamplingRate 0.1, got %f", cfg.Tracing.SamplingRate)
	}
	if cfg.AI.Enabled != false {
		t.Errorf("expected AI.Enabled false")
	}
	if cfg.AI.MaxTokens != 2048 {
		t.Errorf("expected AI.MaxTokens 2048, got %d", cfg.AI.MaxTokens)
	}
	if cfg.AI.Temperature != 0.7 {
		t.Errorf("expected AI.Temperature 0.7, got %f", cfg.AI.Temperature)
	}
	if cfg.AI.ChunkSize != 512 {
		t.Errorf("expected AI.ChunkSize 512, got %d", cfg.AI.ChunkSize)
	}
	if cfg.AI.SearchTopK != 10 {
		t.Errorf("expected AI.SearchTopK 10, got %d", cfg.AI.SearchTopK)
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

func TestLoad_ProductionValidation_DefaultSecrets(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "production")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for production with default JWT secrets")
	}
}

func TestLoad_ProductionValidation_AccessSecretDefault(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "production")
	_ = os.Setenv("JWT_ACCESS_SECRET", "change-this-secret-in-production")
	_ = os.Setenv("JWT_REFRESH_SECRET", "custom-refresh-secret")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("JWT_ACCESS_SECRET")
		_ = os.Unsetenv("JWT_REFRESH_SECRET")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when access secret is default in production")
	}
}

func TestLoad_ProductionValidation_RefreshSecretDefault(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "production")
	_ = os.Setenv("JWT_ACCESS_SECRET", "custom-access-secret")
	_ = os.Setenv("JWT_REFRESH_SECRET", "change-this-refresh-secret-in-production")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("JWT_ACCESS_SECRET")
		_ = os.Unsetenv("JWT_REFRESH_SECRET")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when refresh secret is default in production")
	}
}

func TestLoad_ProductionValidation_CustomSecrets(t *testing.T) {
	_ = os.Setenv("ENVIRONMENT", "production")
	_ = os.Setenv("JWT_ACCESS_SECRET", "my-custom-access-secret")
	_ = os.Setenv("JWT_REFRESH_SECRET", "my-custom-refresh-secret")
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("JWT_ACCESS_SECRET")
		_ = os.Unsetenv("JWT_REFRESH_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error for production with custom secrets, got %v", err)
	}
	if cfg.JWT.AccessSecret != "my-custom-access-secret" {
		t.Errorf("expected custom access secret")
	}
}

func TestLoad_AllEnvVars(t *testing.T) {
	envVars := map[string]string{
		"ENVIRONMENT":                        "staging",
		"VERSION":                            "1.2.3",
		"SERVER_PORT":                        "3000",
		"SERVER_READ_TIMEOUT":                "30s",
		"SERVER_WRITE_TIMEOUT":               "30s",
		"SERVER_IDLE_TIMEOUT":                "60s",
		"SERVER_BASE_URL":                    "https://example.com",
		"DB_HOST":                            "dbhost",
		"DB_PORT":                            "5433",
		"DB_NAME":                            "testdb",
		"DB_USER":                            "testuser",
		"DB_PASSWORD":                        "testpass",
		"DB_MAX_OPEN_CONNS":                  "50",
		"DB_MAX_IDLE_CONNS":                  "10",
		"DB_CONN_MAX_LIFETIME":               "10m",
		"REDIS_HOST":                         "redis-host",
		"REDIS_PORT":                         "6380",
		"REDIS_PASSWORD":                     "redispass",
		"REDIS_DB":                           "1",
		"LOG_LEVEL":                          "debug",
		"CORS_ALLOWED_ORIGINS":               "http://a.com,http://b.com",
		"CORS_ALLOWED_METHODS":               "GET,POST",
		"CORS_ALLOWED_HEADERS":               "Authorization",
		"JWT_ACCESS_SECRET":                  "access-sec",
		"JWT_REFRESH_SECRET":                 "refresh-sec",
		"JWT_ACCESS_TTL":                     "30m",
		"JWT_REFRESH_TTL":                    "48h",
		"COMPOSIO_API_KEY":                   "composio-key",
		"COMPOSIO_ENTITY_ID":                 "composio-entity",
		"COMPOSIO_MCP_CONFIG_ID":             "composio-mcp",
		"S3_ENDPOINT":                        "s3.example.com",
		"S3_PUBLIC_ENDPOINT":                 "s3-public.example.com",
		"S3_ACCESS_KEY_ID":                   "s3key",
		"S3_SECRET_ACCESS_KEY":               "s3secret",
		"S3_BUCKET_NAME":                     "mybucket",
		"S3_REGION":                          "eu-west-1",
		"S3_USE_SSL":                         "true",
		"S3_PUBLIC_USE_SSL":                  "true",
		"S3_MAX_FILE_SIZE":                   "104857600",
		"TELEGRAM_BOT_TOKEN":                 "tg-token",
		"TELEGRAM_BOT_USERNAME":              "tg-user",
		"TELEGRAM_WEBHOOK_URL":               "https://tg.example.com/hook",
		"TELEGRAM_WEBHOOK_SECRET":            "tg-secret",
		"INTEGRATION_1C_ENABLED":             "true",
		"INTEGRATION_1C_BASE_URL":            "http://1c.example.com",
		"INTEGRATION_1C_USERNAME":            "1cuser",
		"INTEGRATION_1C_PASSWORD":            "1cpass",
		"INTEGRATION_1C_TIMEOUT":             "60s",
		"INTEGRATION_1C_MAX_RETRIES":         "5",
		"INTEGRATION_1C_RETRY_DELAY":         "10s",
		"INTEGRATION_1C_EMPLOYEE_CATALOG":    "Catalog_Emp",
		"INTEGRATION_1C_STUDENT_CATALOG":     "Catalog_Stu",
		"INTEGRATION_1C_SYNC_CRON_EMPLOYEE":  "0 */3 * * *",
		"INTEGRATION_1C_SYNC_CRON_STUDENT":   "0 */4 * * *",
		"INTEGRATION_1C_BATCH_SIZE":          "200",
		"VAPID_PUBLIC_KEY":                   "vapid-pub",
		"VAPID_PRIVATE_KEY":                  "vapid-priv",
		"VAPID_SUBJECT":                      "mailto:test@example.com",
		"TRACING_ENABLED":                    "true",
		"TRACING_OTLP_ENDPOINT":              "otel:4317",
		"TRACING_SAMPLING_RATE":              "0.5",
		"TRACING_SERVICE_NAME":               "test-service",
		"AI_ENABLED":                         "true",
		"AI_PROVIDER":                        "anthropic",
		"AI_TIMEOUT":                         "120s",
		"OPENAI_API_KEY":                     "openai-key",
		"OPENAI_BASE_URL":                    "https://openai.example.com",
		"ANTHROPIC_API_KEY":                  "anthropic-key",
		"ANTHROPIC_BASE_URL":                 "https://anthropic.example.com",
		"AI_CHAT_API_KEY":                    "chat-key",
		"AI_CHAT_BASE_URL":                   "https://chat.example.com",
		"AI_CHAT_MODEL":                      "gpt-4",
		"AI_MAX_TOKENS":                      "4096",
		"AI_TEMPERATURE":                     "0.5",
		"AI_EMBEDDING_PROVIDER":              "gemini",
		"AI_EMBEDDING_API_KEY":               "embed-key",
		"AI_EMBEDDING_MODEL":                 "embed-model",
		"AI_EMBEDDING_DIMENSIONALITY":        "768",
		"AI_CHUNK_SIZE":                      "1024",
		"AI_CHUNK_OVERLAP":                   "200",
		"AI_SEARCH_TOP_K":                    "5",
		"AI_SEARCH_THRESHOLD":                "0.8",
		"AI_FALLBACK_PROVIDER":               "groq",
		"AI_FALLBACK_API_KEY":                "fb-key",
		"AI_FALLBACK_BASE_URL":               "https://fb.example.com",
		"AI_FALLBACK_CHAT_MODEL":             "fb-model",
		"AI_FALLBACK_EMBEDDING_PROVIDER":     "openai",
		"AI_FALLBACK_EMBEDDING_API_KEY":      "fbe-key",
		"AI_FALLBACK_EMBEDDING_BASE_URL":     "https://fbe.example.com",
		"AI_FALLBACK_EMBEDDING_MODEL":        "fbe-model",
		"AI_FALLBACK_EMBEDDING_DIMENSIONALITY": "512",
	}

	for k, v := range envVars {
		_ = os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			_ = os.Unsetenv(k)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Environment != "staging" {
		t.Errorf("expected staging, got %s", cfg.Environment)
	}
	if cfg.Version != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", cfg.Version)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("expected 3000, got %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("expected 5433, got %d", cfg.Database.Port)
	}
	if cfg.Database.MaxOpenConns != 50 {
		t.Errorf("expected 50, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("expected 6380, got %d", cfg.Redis.Port)
	}
	if cfg.Redis.DB != 1 {
		t.Errorf("expected 1, got %d", cfg.Redis.DB)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected debug, got %s", cfg.Log.Level)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Errorf("expected 2 CORS origins, got %d", len(cfg.CORS.AllowedOrigins))
	}
	if cfg.JWT.AccessTTL != 30*time.Minute {
		t.Errorf("expected 30m, got %v", cfg.JWT.AccessTTL)
	}
	if cfg.JWT.RefreshTTL != 48*time.Hour {
		t.Errorf("expected 48h, got %v", cfg.JWT.RefreshTTL)
	}
	if cfg.S3.Endpoint != "s3.example.com" {
		t.Errorf("expected s3.example.com, got %s", cfg.S3.Endpoint)
	}
	if cfg.S3.PublicEndpoint != "s3-public.example.com" {
		t.Errorf("expected s3-public.example.com, got %s", cfg.S3.PublicEndpoint)
	}
	if cfg.S3.UseSSL != true {
		t.Error("expected S3.UseSSL true")
	}
	if cfg.S3.PublicUseSSL != true {
		t.Error("expected S3.PublicUseSSL true")
	}
	if cfg.S3.MaxFileSize != 104857600 {
		t.Errorf("expected 104857600, got %d", cfg.S3.MaxFileSize)
	}
	if cfg.Telegram.BotToken != "tg-token" {
		t.Errorf("expected tg-token, got %s", cfg.Telegram.BotToken)
	}
	if cfg.Integration.Enabled != true {
		t.Error("expected Integration.Enabled true")
	}
	if cfg.Integration.BaseURL != "http://1c.example.com" {
		t.Errorf("expected http://1c.example.com, got %s", cfg.Integration.BaseURL)
	}
	if cfg.Integration.Timeout != 60*time.Second {
		t.Errorf("expected 60s, got %v", cfg.Integration.Timeout)
	}
	if cfg.Integration.MaxRetries != 5 {
		t.Errorf("expected 5, got %d", cfg.Integration.MaxRetries)
	}
	if cfg.Integration.BatchSize != 200 {
		t.Errorf("expected 200, got %d", cfg.Integration.BatchSize)
	}
	if cfg.WebPush.VAPIDPublicKey != "vapid-pub" {
		t.Errorf("expected vapid-pub, got %s", cfg.WebPush.VAPIDPublicKey)
	}
	if cfg.Tracing.Enabled != true {
		t.Error("expected Tracing.Enabled true")
	}
	if cfg.Tracing.SamplingRate != 0.5 {
		t.Errorf("expected 0.5, got %f", cfg.Tracing.SamplingRate)
	}
	if cfg.AI.Enabled != true {
		t.Error("expected AI.Enabled true")
	}
	if cfg.AI.Provider != "anthropic" {
		t.Errorf("expected anthropic, got %s", cfg.AI.Provider)
	}
	if cfg.AI.Timeout != 120*time.Second {
		t.Errorf("expected 120s, got %v", cfg.AI.Timeout)
	}
	if cfg.AI.MaxTokens != 4096 {
		t.Errorf("expected 4096, got %d", cfg.AI.MaxTokens)
	}
	if cfg.AI.Temperature != 0.5 {
		t.Errorf("expected 0.5, got %f", cfg.AI.Temperature)
	}
	if cfg.AI.EmbeddingDimensionality != 768 {
		t.Errorf("expected 768, got %d", cfg.AI.EmbeddingDimensionality)
	}
	if cfg.AI.ChunkSize != 1024 {
		t.Errorf("expected 1024, got %d", cfg.AI.ChunkSize)
	}
	if cfg.AI.SearchTopK != 5 {
		t.Errorf("expected 5, got %d", cfg.AI.SearchTopK)
	}
	if cfg.AI.SearchThreshold != 0.8 {
		t.Errorf("expected 0.8, got %f", cfg.AI.SearchThreshold)
	}
	if cfg.AI.FallbackProvider != "groq" {
		t.Errorf("expected groq, got %s", cfg.AI.FallbackProvider)
	}
	if cfg.AI.FallbackEmbeddingDimensionality != 512 {
		t.Errorf("expected 512, got %d", cfg.AI.FallbackEmbeddingDimensionality)
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
	_ = os.Setenv("TEST_INT_INVALID", "not_a_number")
	defer func() {
		_ = os.Unsetenv("TEST_INT")
		_ = os.Unsetenv("TEST_INT_INVALID")
	}()

	if got := getEnvAsInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 42)
	}
	if got := getEnvAsInt("NON_EXISTENT", 99); got != 99 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 99)
	}
	if got := getEnvAsInt("TEST_INT_INVALID", 77); got != 77 {
		t.Errorf("getEnvAsInt() with invalid value = %v, want %v", got, 77)
	}
}

func TestGetEnvAsInt64(t *testing.T) {
	_ = os.Setenv("TEST_INT64", "9999999999")
	_ = os.Setenv("TEST_INT64_INVALID", "abc")
	defer func() {
		_ = os.Unsetenv("TEST_INT64")
		_ = os.Unsetenv("TEST_INT64_INVALID")
	}()

	if got := getEnvAsInt64("TEST_INT64", 0); got != 9999999999 {
		t.Errorf("getEnvAsInt64() = %v, want %v", got, int64(9999999999))
	}
	if got := getEnvAsInt64("NON_EXISTENT", 123); got != 123 {
		t.Errorf("getEnvAsInt64() = %v, want %v", got, 123)
	}
	if got := getEnvAsInt64("TEST_INT64_INVALID", 456); got != 456 {
		t.Errorf("getEnvAsInt64() with invalid = %v, want %v", got, 456)
	}
}

func TestGetEnvAsBool(t *testing.T) {
	_ = os.Setenv("TEST_BOOL_TRUE", "true")
	_ = os.Setenv("TEST_BOOL_FALSE", "false")
	_ = os.Setenv("TEST_BOOL_INVALID", "maybe")
	defer func() {
		_ = os.Unsetenv("TEST_BOOL_TRUE")
		_ = os.Unsetenv("TEST_BOOL_FALSE")
		_ = os.Unsetenv("TEST_BOOL_INVALID")
	}()

	if got := getEnvAsBool("TEST_BOOL_TRUE", false); got != true {
		t.Error("expected true")
	}
	if got := getEnvAsBool("TEST_BOOL_FALSE", true); got != false {
		t.Error("expected false")
	}
	if got := getEnvAsBool("NON_EXISTENT", true); got != true {
		t.Error("expected default true")
	}
	if got := getEnvAsBool("TEST_BOOL_INVALID", true); got != true {
		t.Error("expected default true for invalid")
	}
}

func TestGetEnvAsFloat(t *testing.T) {
	_ = os.Setenv("TEST_FLOAT", "3.14")
	_ = os.Setenv("TEST_FLOAT_INVALID", "not_float")
	defer func() {
		_ = os.Unsetenv("TEST_FLOAT")
		_ = os.Unsetenv("TEST_FLOAT_INVALID")
	}()

	if got := getEnvAsFloat("TEST_FLOAT", 0.0); got != 3.14 {
		t.Errorf("getEnvAsFloat() = %v, want %v", got, 3.14)
	}
	if got := getEnvAsFloat("NON_EXISTENT", 1.5); got != 1.5 {
		t.Errorf("getEnvAsFloat() = %v, want %v", got, 1.5)
	}
	if got := getEnvAsFloat("TEST_FLOAT_INVALID", 2.0); got != 2.0 {
		t.Errorf("getEnvAsFloat() with invalid = %v, want %v", got, 2.0)
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	_ = os.Setenv("TEST_DURATION", "5s")
	_ = os.Setenv("TEST_DURATION_INVALID", "not_duration")
	defer func() {
		_ = os.Unsetenv("TEST_DURATION")
		_ = os.Unsetenv("TEST_DURATION_INVALID")
	}()

	if got := getEnvAsDuration("TEST_DURATION", 0); got != 5*time.Second {
		t.Errorf("getEnvAsDuration() = %v, want %v", got, 5*time.Second)
	}
	if got := getEnvAsDuration("NON_EXISTENT", 10*time.Second); got != 10*time.Second {
		t.Errorf("getEnvAsDuration() = %v, want %v", got, 10*time.Second)
	}
	if got := getEnvAsDuration("TEST_DURATION_INVALID", 20*time.Second); got != 20*time.Second {
		t.Errorf("getEnvAsDuration() with invalid = %v, want %v", got, 20*time.Second)
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	_ = os.Setenv("TEST_SLICE", "a,b,c")
	_ = os.Setenv("TEST_SLICE_SPACES", " a , b , c ")
	_ = os.Setenv("TEST_SLICE_EMPTY", "")
	_ = os.Setenv("TEST_SLICE_COMMAS", ",,,")
	defer func() {
		_ = os.Unsetenv("TEST_SLICE")
		_ = os.Unsetenv("TEST_SLICE_SPACES")
		_ = os.Unsetenv("TEST_SLICE_EMPTY")
		_ = os.Unsetenv("TEST_SLICE_COMMAS")
	}()

	defaultSlice := []string{"default"}

	got := getEnvAsSlice("TEST_SLICE", defaultSlice)
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("getEnvAsSlice() = %v, want [a b c]", got)
	}

	got = getEnvAsSlice("TEST_SLICE_SPACES", defaultSlice)
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("getEnvAsSlice() with spaces = %v, want [a b c]", got)
	}

	got = getEnvAsSlice("TEST_SLICE_EMPTY", defaultSlice)
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("getEnvAsSlice() empty = %v, want [default]", got)
	}

	got = getEnvAsSlice("NON_EXISTENT", defaultSlice)
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("getEnvAsSlice() missing = %v, want [default]", got)
	}

	got = getEnvAsSlice("TEST_SLICE_COMMAS", defaultSlice)
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("getEnvAsSlice() commas only = %v, want [default]", got)
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{" a , b , c ", ",", []string{"a", "b", "c"}},
		{"", ",", []string{}},
		{"  ,  ,  ", ",", []string{}},
		{"hello", ",", []string{"hello"}},
		{"a::b::c", "::", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		got := splitAndTrim(tt.input, tt.sep)
		if len(got) != len(tt.expected) {
			t.Errorf("splitAndTrim(%q, %q) = %v, want %v", tt.input, tt.sep, got, tt.expected)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitAndTrim(%q, %q)[%d] = %q, want %q", tt.input, tt.sep, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{"", ",", []string{}},
		{"hello", ",", []string{"hello"}},
		{"a::b::c", "::", []string{"a", "b", "c"}},
		{"abc", "abc", []string{"", ""}},
		{",", ",", []string{"", ""}},
	}

	for _, tt := range tests {
		got := splitString(tt.input, tt.sep)
		if len(got) != len(tt.expected) {
			t.Errorf("splitString(%q, %q) = %v (len %d), want %v (len %d)", tt.input, tt.sep, got, len(got), tt.expected, len(tt.expected))
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitString(%q, %q)[%d] = %q, want %q", tt.input, tt.sep, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\t\nhello\r\n", "hello"},
		{"hello", "hello"},
		{"   ", ""},
		{"", ""},
		{" hello world ", "hello world"},
	}

	for _, tt := range tests {
		got := trimSpace(tt.input)
		if got != tt.expected {
			t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
