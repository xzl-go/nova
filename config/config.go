package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config 配置结构体
type Config struct {
	Server *ServerConfig `json:"server,omitempty"`
	JWT    *JWTConfig    `json:"jwt,omitempty"`
	Log    *LogConfig    `json:"log,omitempty"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         *int    `json:"port,omitempty"`
	ReadTimeout  *int    `json:"read_timeout,omitempty"`
	WriteTimeout *int    `json:"write_timeout,omitempty"`
	Mode         *string `json:"mode,omitempty"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     *string `json:"secret,omitempty"`
	ExpireTime *int    `json:"expire_time,omitempty"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      *string `json:"level,omitempty"`
	Filename   *string `json:"filename,omitempty"`
	MaxSize    *int    `json:"max_size,omitempty"`
	MaxBackups *int    `json:"max_backups,omitempty"`
	MaxAge     *int    `json:"max_age,omitempty"`
	Compress   *bool   `json:"compress,omitempty"`
}

var (
	config *Config
	once   sync.Once
)

// 默认配置
var defaultConfig = &Config{
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

// ptr 返回指针
func ptr[T any](v T) *T {
	return &v
}

// mergeConfig 合并配置
func mergeConfig(dst, src *Config) *Config {
	if dst == nil {
		dst = &Config{}
	}
	if src == nil {
		return dst
	}

	// 合并 Server 配置
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

	// 合并 JWT 配置
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

	// 合并 Log 配置
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

// Load 加载配置文件
func Load(path string) (*Config, error) {
	once.Do(func() {
		config = &Config{}
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			panic(err)
		}
		config = mergeConfig(defaultConfig, config)
	})
	return config, nil
}

// LoadFromBytes 从字节数组加载配置
func LoadFromBytes(data []byte) (*Config, error) {
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	config = mergeConfig(defaultConfig, cfg)
	return config, nil
}

// LoadFromMap 从map加载配置
func LoadFromMap(m map[string]interface{}) (*Config, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return LoadFromBytes(data)
}

// Get 获取配置实例
func Get() *Config {
	if config == nil {
		config = defaultConfig
	}
	return config
}

// Set 设置配置实例
func Set(cfg *Config) {
	config = mergeConfig(defaultConfig, cfg)
}

// GetServerPort 获取服务器端口
func GetServerPort() int {
	if config == nil || config.Server == nil || config.Server.Port == nil {
		return *defaultConfig.Server.Port
	}
	return *config.Server.Port
}

// GetServerReadTimeout 获取服务器读取超时
func GetServerReadTimeout() int {
	if config == nil || config.Server == nil || config.Server.ReadTimeout == nil {
		return *defaultConfig.Server.ReadTimeout
	}
	return *config.Server.ReadTimeout
}

// GetServerWriteTimeout 获取服务器写入超时
func GetServerWriteTimeout() int {
	if config == nil || config.Server == nil || config.Server.WriteTimeout == nil {
		return *defaultConfig.Server.WriteTimeout
	}
	return *config.Server.WriteTimeout
}

// GetServerMode 获取服务器模式
func GetServerMode() string {
	if config == nil || config.Server == nil || config.Server.Mode == nil {
		return *defaultConfig.Server.Mode
	}
	return *config.Server.Mode
}

// GetJWTSecret 获取JWT密钥
func GetJWTSecret() string {
	if config == nil || config.JWT == nil || config.JWT.Secret == nil {
		return *defaultConfig.JWT.Secret
	}
	return *config.JWT.Secret
}

// GetJWTExpireTime 获取JWT过期时间
func GetJWTExpireTime() int {
	if config == nil || config.JWT == nil || config.JWT.ExpireTime == nil {
		return *defaultConfig.JWT.ExpireTime
	}
	return *config.JWT.ExpireTime
}

// GetLogLevel 获取日志级别
func GetLogLevel() string {
	if config == nil || config.Log == nil || config.Log.Level == nil {
		return *defaultConfig.Log.Level
	}
	return *config.Log.Level
}

// GetLogFilename 获取日志文件名
func GetLogFilename() string {
	if config == nil || config.Log == nil || config.Log.Filename == nil {
		return *defaultConfig.Log.Filename
	}
	return *config.Log.Filename
}

// GetLogMaxSize 获取日志最大大小
func GetLogMaxSize() int {
	if config == nil || config.Log == nil || config.Log.MaxSize == nil {
		return *defaultConfig.Log.MaxSize
	}
	return *config.Log.MaxSize
}

// GetLogMaxBackups 获取日志最大备份数
func GetLogMaxBackups() int {
	if config == nil || config.Log == nil || config.Log.MaxBackups == nil {
		return *defaultConfig.Log.MaxBackups
	}
	return *config.Log.MaxBackups
}

// GetLogMaxAge 获取日志最大保存时间
func GetLogMaxAge() int {
	if config == nil || config.Log == nil || config.Log.MaxAge == nil {
		return *defaultConfig.Log.MaxAge
	}
	return *config.Log.MaxAge
}

// GetLogCompress 获取日志是否压缩
func GetLogCompress() bool {
	if config == nil || config.Log == nil || config.Log.Compress == nil {
		return *defaultConfig.Log.Compress
	}
	return *config.Log.Compress
}
