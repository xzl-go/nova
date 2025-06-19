package middleware

import (
	"net/http"
	"time"

	"github.com/xzl/nova/core"
)

// RateLimiter 限流器
type RateLimiter struct {
	limit   int
	window  time.Duration
	clients map[string][]time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string][]time.Time),
	}
}

// Allow 检查是否允许请求
func (r *RateLimiter) Allow(clientIP string) bool {
	now := time.Now()
	windowStart := now.Add(-r.window)

	// 清理过期的请求记录
	if times, ok := r.clients[clientIP]; ok {
		valid := make([]time.Time, 0)
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		r.clients[clientIP] = valid
	}

	// 检查是否超过限制
	if len(r.clients[clientIP]) >= r.limit {
		return false
	}

	// 记录新的请求
	r.clients[clientIP] = append(r.clients[clientIP], now)
	return true
}

// RateLimit 限流中间件
func RateLimit(limit int, window time.Duration) core.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *core.Context) {
		if !limiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    429,
				"message": "Too Many Requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
