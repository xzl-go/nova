package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config 配置中心
type Config struct {
	viper    *viper.Viper
	etcd     *clientv3.Client
	watchers []ConfigWatcher
	mu       sync.RWMutex
}

// ConfigWatcher 配置监听器
type ConfigWatcher interface {
	OnConfigChange(key string, value interface{})
}

// NewConfig 创建配置中心
func NewConfig() *Config {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	return &Config{
		viper:    v,
		watchers: make([]ConfigWatcher, 0),
	}
}

// LoadFile 从文件加载配置
func (c *Config) LoadFile(path string) error {
	c.viper.SetConfigFile(path)
	return c.viper.ReadInConfig()
}

// LoadEtcd 从etcd加载配置
func (c *Config) LoadEtcd(endpoints []string, prefix string) error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %v", err)
	}

	c.etcd = client
	c.watchEtcd(prefix)
	return nil
}

// Get 获取配置
func (c *Config) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.Get(key)
}

// GetString 获取字符串配置
func (c *Config) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetString(key)
}

// GetInt 获取整数配置
func (c *Config) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (c *Config) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetBool(key)
}

// GetStringSlice 获取字符串切片配置
func (c *Config) GetStringSlice(key string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetStringSlice(key)
}

// Set 设置配置
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.viper.Set(key, value)
}

// AddWatcher 添加配置监听器
func (c *Config) AddWatcher(watcher ConfigWatcher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.watchers = append(c.watchers, watcher)
}

// watchEtcd 监听etcd配置变化
func (c *Config) watchEtcd(prefix string) {
	go func() {
		watchCh := c.etcd.Watch(context.Background(), prefix, clientv3.WithPrefix())
		for {
			select {
			case resp := <-watchCh:
				for _, ev := range resp.Events {
					key := string(ev.Kv.Key)
					value := string(ev.Kv.Value)
					c.Set(key, value)
					c.notifyWatchers(key, value)
				}
			}
		}
	}()
}

// notifyWatchers 通知所有监听器
func (c *Config) notifyWatchers(key string, value interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, watcher := range c.watchers {
		watcher.OnConfigChange(key, value)
	}
}

// Unmarshal 将配置解析到结构体
func (c *Config) Unmarshal(rawVal interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.Unmarshal(rawVal)
}

// WatchConfig 监听配置文件变化
func (c *Config) WatchConfig() {
	c.viper.WatchConfig()
	c.viper.OnConfigChange(func(e fsnotify.Event) {
		c.notifyWatchers(e.Name, nil)
	})
}
