package middleware

import (
	"fmt"
	"github.com/xzl-go/nova"
	"github.com/xzl-go/nova/logger"
	"github.com/xzl-go/nova/pkg/errors"
	"net/http"
	"time"
)

// Logger 日志中间件
func Logger() nova.HandlerFunc {
	return func(c *nova.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取客户端IP
		clientIP := c.ClientIP()
		// 获取请求方法
		method := c.Request.Method
		// 获取状态码
		statusCode := c.Writer.Status

		// 如果有查询参数，添加到路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 记录日志
		logger.Infof("[%s] %s %s %d %v",
			clientIP,
			method,
			path,
			statusCode,
			latency,
		)
	}
}

// Recovery 恢复中间件
func Recovery() nova.HandlerFunc {
	return func(c *nova.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				logger.Errorf("panic recovered: %v", err)

				// 返回 500 错误
				c.Error(errors.New(errors.ErrInternal, "Internal Server Error"))
			}
		}()

		c.Next()
	}
}

// CORS 跨域中间件
func CORS() nova.HandlerFunc {
	return func(c *nova.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.Abort()
			return
		}

		c.Next()
	}
}

// Auth 认证中间件
func Auth() nova.HandlerFunc {
	return func(c *nova.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		// TODO: 实现 JWT 验证
		c.Next()
	}
}

// RequestID 为每个请求生成唯一 traceID
func RequestID() nova.HandlerFunc {
	return func(c *nova.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Next()
	}
}

// Chain 中间件链
func Chain(middlewares ...nova.HandlerFunc) nova.HandlerFunc {
	return func(c *nova.Context) {
		// 保存当前中间件索引
		index := c.Index
		// 设置中间件索引
		c.Index = -1
		// 执行中间件链
		for i := 0; i < len(middlewares); i++ {
			middlewares[i](c)
			if c.IsAborted() {
				return
			}
		}
		// 恢复中间件索引
		c.Index = index
	}
}
