package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	core "github.com/xzl-go/nova"

	"github.com/redis/go-redis/v9"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Addr       string        // Redis 地址
	Password   string        // Redis 密码
	DB         int           // Redis 数据库
	Expiration time.Duration // 缓存过期时间
}

// Cache 缓存中间件
func Cache(config CacheConfig) core.HandlerFunc {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return func(c *core.Context) {
		// 只缓存 GET 请求
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
		key := fmt.Sprintf("cache:%s", c.Request.URL.String())

		// 尝试从缓存获取
		val, err := client.Get(context.Background(), key).Result()
		if err == nil {
			// 缓存命中，直接返回
			var data interface{}
			if err := json.Unmarshal([]byte(val), &data); err == nil {
				c.JSON(200, data)
				c.Abort()
				return
			}
		}

		// 缓存未命中，继续处理请求
		c.Next()

		// 如果响应状态码是 200，则缓存响应
		if c.Writer.Status == 200 {
			// 获取响应数据
			data := c.Data
			if data != nil {
				// 序列化数据
				if bytes, err := json.Marshal(data); err == nil {
					// 设置缓存
					client.Set(context.Background(), key, bytes, config.Expiration)
				}
			}
		}
	}
}

// ClearCache 清除缓存
func ClearCache(pattern string) core.HandlerFunc {
	return func(c *core.Context) {
		// 创建 Redis 客户端
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		// 删除匹配的键
		iter := client.Scan(context.Background(), 0, pattern, 0).Iterator()
		for iter.Next(context.Background()) {
			client.Del(context.Background(), iter.Val())
		}

		c.Next()
	}
}
