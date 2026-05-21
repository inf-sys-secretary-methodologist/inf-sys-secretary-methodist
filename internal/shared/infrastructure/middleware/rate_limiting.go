package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter — структура для Redis-based rate limiting с поддержкой burst
type RateLimiter struct {
	redisClient *redis.Client
	requests    int           // Количество запросов в минуту
	burst       int           // Дополнительные запросы для кратковременных всплесков
	window      time.Duration // Временное окно (обычно 1 минута)
}

// NewRateLimiter создаёт новый rate limiter с Redis и поддержкой burst
func NewRateLimiter(redisClient *redis.Client, requestsPerMinute int, burst int) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		requests:    requestsPerMinute,
		burst:       burst,
		window:      time.Minute, // Фиксированное окно в 1 минуту
	}
}

// RateLimitMiddleware returns the IP-keyed limiter middleware. Suitable for
// pre-auth surfaces (login, public branding). Trusts X-Forwarded-For for
// reverse-proxy deployments — known limitation: client-set header bypasses
// the limit. Post-auth surfaces should prefer RateLimitByUserMiddleware
// so NAT'd users do not share a bucket.
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getRealIP(c.Request)

		key := fmt.Sprintf("rate_limit:%s", ip)

		count, retryAfter, err := rl.incrementAndCheck(key)
		if err != nil {
			c.Next() // разрешаем запрос, если Redis недоступен
			return
		}

		// Общий лимит = базовый лимит + burst
		totalLimit := rl.requests + rl.burst

		if count > int64(totalLimit) {
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
			c.Header("X-RateLimit-Burst", strconv.Itoa(rl.burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC3339))

			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		remaining := totalLimit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
		c.Header("X-RateLimit-Burst", strconv.Itoa(rl.burst))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", time.Now().Add(rl.window).Format(time.RFC3339))

		c.Next()
	}
}

// RateLimitByUserMiddleware returns a limiter middleware keyed by the
// authenticated user_id ctx value. Must be mounted AFTER JWT middleware so
// the key is the authenticated principal — NAT'd students no longer share
// a bucket (which is the security goal for AI / chat endpoints where each
// request has dollar-cost). On missing ctx (mis-wired chain) falls back к
// IP-keyed to fail closed. Issue #263 ADR-3.
func (rl *RateLimiter) RateLimitByUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		if rawUserID, exists := c.Get("user_id"); exists {
			if uid, ok := rawUserID.(int64); ok && uid > 0 {
				key = fmt.Sprintf("rate_limit:user:%d", uid)
			}
		}
		if key == "" {
			// Fallback: protect against mis-wired chains by still applying
			// IP-keyed limit (fail closed). Production deployments should
			// never hit this branch because JWT middleware populates user_id.
			key = fmt.Sprintf("rate_limit:ip-fallback:%s", getRealIP(c.Request))
		}

		count, retryAfter, err := rl.incrementAndCheck(key)
		if err != nil {
			c.Next()
			return
		}

		totalLimit := rl.requests + rl.burst

		if count > int64(totalLimit) {
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
			c.Header("X-RateLimit-Burst", strconv.Itoa(rl.burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC3339))

			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		remaining := totalLimit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
		c.Header("X-RateLimit-Burst", strconv.Itoa(rl.burst))
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

	resSlice, _ := result.([]interface{})
	count, _ := resSlice[0].(int64)
	ttl, _ := resSlice[1].(int64)

	return count, ttl, nil
}

// getRealIP — получает реальный IP клиента (учитывает X-Forwarded-For, X-Real-IP)
// Deprecated by getRealIPWithTrustedProxies (v0.159.0 ADR-3b). Kept as a thin
// shim so legacy call sites keep compiling; new code must pass the trusted-
// proxy CIDR slice so spoofed X-Forwarded-For headers do not bypass the
// rate limit. Issue #279.
func getRealIP(r *http.Request) string {
	return getRealIPWithTrustedProxies(r, nil)
}

// getRealIPWithTrustedProxies returns the client IP for rate-limiting,
// honoring X-Forwarded-For / X-Real-IP ONLY when r.RemoteAddr falls
// inside the supplied trusted-proxy CIDR allowlist. With no trusted
// CIDRs (the secure default) the proxy headers must be ignored
// entirely and the TCP peer used directly. Issue #279 ADR-3.
//
// RED stub — currently delegates to the legacy unconditional logic
// (X-Forwarded-For trumps RemoteAddr); proper trusted-CIDR enforcement
// lands in the GREEN pair.
func getRealIPWithTrustedProxies(r *http.Request, _ []*net.IPNet) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// ParseTrustedProxyCIDRs parses a comma-separated list of CIDR notations
// into *net.IPNet entries. Empty / malformed entries are silently
// skipped (the call site can log the parse skip if desired). Intended
// to be called once at startup with the TRUSTED_PROXY_CIDRS env value.
// Issue #279 ADR-3.
//
// RED stub — returns nil regardless of input. Real parsing lands in
// the GREEN pair.
func ParseTrustedProxyCIDRs(_ string) []*net.IPNet {
	return nil
}
