package plugin

import (
	"fmt"
	"sync"
)

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string
	// Version 返回插件版本
	Version() string
	// Init 初始化插件
	Init(config map[string]interface{}) error
	// Start 启动插件
	Start() error
	// Stop 停止插件
	Stop() error
}

// Manager 插件管理器
type Manager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

// Register 注册插件
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	m.plugins[name] = plugin
	return nil
}

// Unregister 卸载插件
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := plugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin %s: %v", name, err)
	}

	delete(m.plugins, name)
	return nil
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// InitPlugin 初始化插件
func (m *Manager) InitPlugin(name string, config map[string]interface{}) error {
	plugin, exists := m.GetPlugin(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Init(config)
}

// StartPlugin 启动插件
func (m *Manager) StartPlugin(name string) error {
	plugin, exists := m.GetPlugin(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Start()
}

// StopPlugin 停止插件
func (m *Manager) StopPlugin(name string) error {
	plugin, exists := m.GetPlugin(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Stop()
}

// InitAll 初始化所有插件
func (m *Manager) InitAll(configs map[string]map[string]interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, plugin := range m.plugins {
		config, exists := configs[name]
		if !exists {
			config = make(map[string]interface{})
		}

		if err := plugin.Init(config); err != nil {
			return fmt.Errorf("failed to init plugin %s: %v", name, err)
		}
	}
	return nil
}

// StartAll 启动所有插件
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, plugin := range m.plugins {
		if err := plugin.Start(); err != nil {
			return fmt.Errorf("failed to start plugin %s: %v", name, err)
		}
	}
	return nil
}

// StopAll 停止所有插件
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, plugin := range m.plugins {
		if err := plugin.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin %s: %v", name, err)
		}
	}
	return nil
}
