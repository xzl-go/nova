package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config 缓存配置
type Config struct {
	Type     string // redis
	Host     string
	Port     int
	Password string
	DB       int
}

// Cache 缓存接口
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}

// Redis Redis 缓存
type Redis struct {
	client *redis.Client
}

// NewRedis 创建 Redis 缓存实例
func NewRedis(config *Config) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Host + ":" + string(config.Port),
			Password: config.Password,
			DB:       config.DB,
		}),
	}
}

// Set 设置缓存
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (r *Redis) Get(ctx context.Context, key string, value interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// Delete 删除缓存
func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists 检查缓存是否存在
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Close 关闭缓存连接
func (r *Redis) Close() error {
	return r.client.Close()
}

// NewCache 创建缓存实例
func NewCache(config *Config) (Cache, error) {
	switch config.Type {
	case "redis":
		return NewRedis(config), nil
	default:
		return nil, nil
	}
}
