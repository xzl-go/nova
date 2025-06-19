package middleware

import (
	"context"
	"framework/router"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// 全局 Redis 客户端
var redisClient *redis.Client

// InitRedis 初始化 Redis 客户端
func InitRedis(addr, password string, db int) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// Health 健康检查中间件
func Health() router.HandlerFunc {
	return func(c *router.Context) {
		ctx := context.Background()
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
					"status": "redis unavailable",
					"error":  err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	}
}
