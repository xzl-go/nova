package middleware

import (
	"net/http"

	"github.com/xzl/nova/core"
)

// Security 安全中间件
func Security() core.HandlerFunc {
	return func(c *core.Context) {
		// CORS
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理 OPTIONS 请求
		if c.Request.Method == "OPTIONS" {
			c.Abort()
			return
		}

		// XSS 防护
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'")

		// 安全头部
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Download-Options", "noopen")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 移除敏感头部
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// CSRF CSRF 防护中间件
func CSRF(secret string) core.HandlerFunc {
	return func(c *core.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.Request.Header.Get("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "CSRF token missing",
			})
			c.Abort()
			return
		}

		// 验证 token
		if !validateCSRFToken(token, secret) {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "Invalid CSRF token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateCSRFToken 验证 CSRF token
func validateCSRFToken(token, secret string) bool {
	// TODO: 实现 CSRF token 验证逻辑
	return true
}
