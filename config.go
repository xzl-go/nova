package nova

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config 配置管理器
type Config struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewConfig 创建配置管理器
func NewConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}

// LoadJSON 从 JSON 文件加载配置
func (c *Config) LoadJSON(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	c.mu.Lock()
	c.data = config
	c.mu.Unlock()

	return nil
}

// SaveJSON 保存配置到 JSON 文件
func (c *Config) SaveJSON(path string) error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c.data, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}

	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// Get 获取配置值
func (c *Config) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := strings.Split(key, ".")
	value := c.data

	for _, k := range keys {
		if v, ok := value[k]; ok {
			if m, ok := v.(map[string]interface{}); ok {
				value = m
			} else {
				return v
			}
		} else {
			return nil
		}
	}

	return value
}

// GetString 获取字符串配置值
func (c *Config) GetString(key string) string {
	value := c.Get(key)
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

// GetInt 获取整数配置值
func (c *Config) GetInt(key string) int {
	value := c.Get(key)
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

// GetFloat 获取浮点数配置值
func (c *Config) GetFloat(key string) float64 {
	value := c.Get(key)
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0
	}
}

// GetBool 获取布尔配置值
func (c *Config) GetBool(key string) bool {
	value := c.Get(key)
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	default:
		return false
	}
}

// GetStringSlice 获取字符串切片配置值
func (c *Config) GetStringSlice(key string) []string {
	value := c.Get(key)
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprint(item)
		}
		return result
	default:
		return nil
	}
}

// GetStringMap 获取字符串映射配置值
func (c *Config) GetStringMap(key string) map[string]interface{} {
	value := c.Get(key)
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case map[string]interface{}:
		return v
	default:
		return nil
	}
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := strings.Split(key, ".")
	lastKey := keys[len(keys)-1]
	keys = keys[:len(keys)-1]

	config := c.data
	for _, k := range keys {
		if v, ok := config[k]; ok {
			if m, ok := v.(map[string]interface{}); ok {
				config = m
			} else {
				config[k] = make(map[string]interface{})
				config = config[k].(map[string]interface{})
			}
		} else {
			config[k] = make(map[string]interface{})
			config = config[k].(map[string]interface{})
		}
	}

	config[lastKey] = value
}

// Delete 删除配置值
func (c *Config) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := strings.Split(key, ".")
	lastKey := keys[len(keys)-1]
	keys = keys[:len(keys)-1]

	config := c.data
	for _, k := range keys {
		if v, ok := config[k]; ok {
			if m, ok := v.(map[string]interface{}); ok {
				config = m
			} else {
				return
			}
		} else {
			return
		}
	}

	delete(config, lastKey)
}

// Clear 清空配置
func (c *Config) Clear() {
	c.mu.Lock()
	c.data = make(map[string]interface{})
	c.mu.Unlock()
}

// All 获取所有配置
func (c *Config) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Has 检查配置是否存在
func (c *Config) Has(key string) bool {
	return c.Get(key) != nil
}

// Merge 合并配置
func (c *Config) Merge(config *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range config.All() {
		c.data[k] = v
	}
}

// Watch 监视配置变化
func (c *Config) Watch(callback func(*Config)) {
	// TODO: 实现配置监视功能
}
