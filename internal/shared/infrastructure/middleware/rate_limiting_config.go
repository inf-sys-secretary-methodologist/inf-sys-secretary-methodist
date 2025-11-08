package middleware

import (
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// RateLimitConfig — конфигурация для rate limiting
type RateLimitConfig struct {
	// Public endpoints (неаутентифицированные)
	PublicRequestsPerMinute int
	PublicBurst             int

	// Authenticated endpoints (аутентифицированные)
	AuthRequestsPerMinute int
	AuthBurst             int
}

// LoadRateLimitConfig загружает конфигурацию из environment variables
func LoadRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		PublicRequestsPerMinute: getEnvInt("RATE_LIMIT_PUBLIC_RPM", 10),
		PublicBurst:             getEnvInt("RATE_LIMIT_PUBLIC_BURST", 5),
		AuthRequestsPerMinute:   getEnvInt("RATE_LIMIT_AUTH_RPM", 60),
		AuthBurst:               getEnvInt("RATE_LIMIT_AUTH_BURST", 10),
	}
}

// getEnvInt — вспомогательная функция для чтения int из env с default значением
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetPublicRateLimiter создаёт rate limiter для публичных endpoints
func (cfg *RateLimitConfig) GetPublicRateLimiter(redisClient interface{}) *RateLimiter {
	return NewRateLimiter(
		redisClient.(*redis.Client),
		cfg.PublicRequestsPerMinute,
		cfg.PublicBurst,
	)
}

// GetAuthRateLimiter создаёт rate limiter для аутентифицированных endpoints
func (cfg *RateLimitConfig) GetAuthRateLimiter(redisClient interface{}) *RateLimiter {
	return NewRateLimiter(
		redisClient.(*redis.Client),
		cfg.AuthRequestsPerMinute,
		cfg.AuthBurst,
	)
}
