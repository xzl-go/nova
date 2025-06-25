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

// Config 配置结构体
// 支持 Server、JWT、Log 等常用配置
// 可通过 viper/etcd 动态加载，也可本地静态加载

type ConfigStruct struct {
	Server *ServerConfig `json:"server,omitempty" mapstructure:"server"`
	JWT    *JWTConfig    `json:"jwt,omitempty" mapstructure:"jwt"`
	Log    *LogConfig    `json:"log,omitempty" mapstructure:"log"`
}

type ServerConfig struct {
	Port         *int    `json:"port,omitempty" mapstructure:"port"`
	ReadTimeout  *int    `json:"read_timeout,omitempty" mapstructure:"read_timeout"`
	WriteTimeout *int    `json:"write_timeout,omitempty" mapstructure:"write_timeout"`
	Mode         *string `json:"mode,omitempty" mapstructure:"mode"`
}

type JWTConfig struct {
	Secret     *string `json:"secret,omitempty" mapstructure:"secret"`
	ExpireTime *int    `json:"expire_time,omitempty" mapstructure:"expire_time"`
}

type LogConfig struct {
	Level      *string `json:"level,omitempty" mapstructure:"level"`
	Filename   *string `json:"filename,omitempty" mapstructure:"filename"`
	MaxSize    *int    `json:"max_size,omitempty" mapstructure:"max_size"`
	MaxBackups *int    `json:"max_backups,omitempty" mapstructure:"max_backups"`
	MaxAge     *int    `json:"max_age,omitempty" mapstructure:"max_age"`
	Compress   *bool   `json:"compress,omitempty" mapstructure:"compress"`
}

var defaultConfig = &ConfigStruct{
	Server: &ServerConfig{
		Port:         ptr(8080),
		ReadTimeout:  ptr(60),
		WriteTimeout: ptr(60),
		Mode:         ptr("debug"),
	},
	JWT: &JWTConfig{
		Secret:     ptr("your-secret-key"),
		ExpireTime: ptr(24),
	},
	Log: &LogConfig{
		Level:      ptr("info"),
		Filename:   ptr("logs/app.log"),
		MaxSize:    ptr(100),
		MaxBackups: ptr(10),
		MaxAge:     ptr(30),
		Compress:   ptr(true),
	},
}

func ptr[T any](v T) *T {
	return &v
}

// MergeConfig 合并配置，优先 src
func MergeConfig(dst, src *ConfigStruct) *ConfigStruct {
	if dst == nil {
		dst = &ConfigStruct{}
	}
	if src == nil {
		return dst
	}
	if src.Server != nil {
		if dst.Server == nil {
			dst.Server = &ServerConfig{}
		}
		if src.Server.Port != nil {
			dst.Server.Port = src.Server.Port
		}
		if src.Server.ReadTimeout != nil {
			dst.Server.ReadTimeout = src.Server.ReadTimeout
		}
		if src.Server.WriteTimeout != nil {
			dst.Server.WriteTimeout = src.Server.WriteTimeout
		}
		if src.Server.Mode != nil {
			dst.Server.Mode = src.Server.Mode
		}
	}
	if src.JWT != nil {
		if dst.JWT == nil {
			dst.JWT = &JWTConfig{}
		}
		if src.JWT.Secret != nil {
			dst.JWT.Secret = src.JWT.Secret
		}
		if src.JWT.ExpireTime != nil {
			dst.JWT.ExpireTime = src.JWT.ExpireTime
		}
	}
	if src.Log != nil {
		if dst.Log == nil {
			dst.Log = &LogConfig{}
		}
		if src.Log.Level != nil {
			dst.Log.Level = src.Log.Level
		}
		if src.Log.Filename != nil {
			dst.Log.Filename = src.Log.Filename
		}
		if src.Log.MaxSize != nil {
			dst.Log.MaxSize = src.Log.MaxSize
		}
		if src.Log.MaxBackups != nil {
			dst.Log.MaxBackups = src.Log.MaxBackups
		}
		if src.Log.MaxAge != nil {
			dst.Log.MaxAge = src.Log.MaxAge
		}
		if src.Log.Compress != nil {
			dst.Log.Compress = src.Log.Compress
		}
	}
	return dst
}

// UnmarshalToConfigStruct 将当前配置反序列化到 ConfigStruct 并合并默认值
func (c *Config) UnmarshalToConfigStruct() *ConfigStruct {
	var cfg ConfigStruct
	_ = c.Unmarshal(&cfg)
	return MergeConfig(defaultConfig, &cfg)
}
