package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter — структура для Redis-based rate limiting
type RateLimiter struct {
	redisClient *redis.Client
	requests    int
	window      time.Duration
}

// NewRateLimiter создаёт новый rate limiter с Redis
func NewRateLimiter(redisClient *redis.Client, requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		requests:    requests,
		window:      window,
	}
}

// RateLimitMiddleware возвращает HTTP middleware для rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getRealIP(c.Request)

		key := fmt.Sprintf("rate_limit:%s", ip)

		count, retryAfter, err := rl.incrementAndCheck(key)
		if err != nil {
			c.Next() // разрешаем запрос, если Redis недоступен
			return
		}

		if count > int64(rl.requests) {
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC3339))

			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		remaining := rl.requests - int(count)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", time.Now().Add(rl.window).Format(time.RFC3339))

		c.Next()
	}
}

// incrementAndCheck — увеличивает счётчик и проверяет лимит
func (rl *RateLimiter) incrementAndCheck(key string) (int64, int64, error) {
	ctx := context.Background()

	// Lua-скрипт для atomic increment и TTL
	luaScript := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local current = redis.call("GET", key)
		
		if current == false then
			redis.call("SET", key, 1)
			redis.call("EXPIRE", key, window)
			return {1, window}
		end
		
		local count = tonumber(current) + 1
		redis.call("SET", key, count)
		local ttl = redis.call("TTL", key)
		
		return {count, ttl}
	`

	result, err := rl.redisClient.Eval(ctx, luaScript, []string{key}, rl.window.Seconds()).Result()
	if err != nil {
		return 0, 0, err
	}

	resSlice := result.([]interface{})
	count := resSlice[0].(int64)
	ttl := resSlice[1].(int64)

	return count, ttl, nil
}

// getRealIP — получает реальный IP клиента (учитывает X-Forwarded-For, X-Real-IP)
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
