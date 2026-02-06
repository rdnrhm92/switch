package config

import (
	"fmt"
	"sync"
)

type ConfigName string

var GlobalConfigManager = NewConfigManager()

// ConfigManager 配置管理器
type ConfigManager struct {
	configs map[ConfigName]ConfigI
	mu      sync.RWMutex
}

// NewConfigManager 创建新的配置管理器实例
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs: make(map[ConfigName]ConfigI),
	}
}

// Register 注册配置实例
func (m *ConfigManager) Register(name ConfigName, config ConfigI) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; exists {
		return fmt.Errorf("config with name '%s' already exists", name)
	}

	m.configs[name] = config
	return nil
}

// Get 获取指定名称的配置实例
func (m *ConfigManager) Get(name ConfigName) (ConfigI, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("config with name '%s' not found", name)
	}

	return config, nil
}

// MustGet 获取指定名称的配置实例，如果不存在则panic
func (m *ConfigManager) MustGet(name ConfigName) ConfigI {
	config, err := m.Get(name)
	if err != nil {
		panic(err)
	}
	return config
}

// LoadAll 加载所有已注册的配置
func (m *ConfigManager) LoadAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for name, config := range m.configs {
		if err := config.Load(); err != nil {
			errs = append(errs, fmt.Errorf("failed to load config '%s': %v", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to load all configs: %v", errs)
	}
	return nil
}

// Remove 移除指定名称的配置
func (m *ConfigManager) Remove(name ConfigName) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; !exists {
		return fmt.Errorf("config with name '%s' not found", name)
	}

	delete(m.configs, name)
	return nil
}

// List 列出所有已注册的配置名称
func (m *ConfigManager) List() []ConfigName {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var names []ConfigName
	for name := range m.configs {
		names = append(names, name)
	}
	return names
}

// Clear 清空所有配置
func (m *ConfigManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs = make(map[ConfigName]ConfigI)
}

// Exists 检查指定名称的配置是否存在
func (m *ConfigManager) Exists(name ConfigName) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.configs[name]
	return exists
}
