package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvFileConfigLoader 环境配置加载器
type EnvFileConfigLoader struct {
	configManager *ConfigManager
	filePath      string
	env           Environment
}

// SetFilePath 设置文件地址
func (e *EnvFileConfigLoader) SetFilePath(path string) {
	if path == "" {
		panic("empty config path")
	}
	e.filePath = path
}

// GetFilePath 获取文件地址
func (e *EnvFileConfigLoader) GetFilePath() string {
	return e.filePath
}

// GetManager 获取配置管理器
func (e *EnvFileConfigLoader) GetManager() *ConfigManager {
	return e.configManager
}

// SetManager 设置配置管理器
func (e *EnvFileConfigLoader) SetManager(manager *ConfigManager) {
	e.configManager = manager
}

// pathToConfigName 将文件路径转换为配置名
func (e *EnvFileConfigLoader) pathToConfigName(path string) string {
	basename := strings.TrimSuffix(path, filepath.Ext(path))
	configName := strings.ReplaceAll(basename, string(filepath.Separator), ".")
	return configName
}

// LoadConfig LoadConfigs
func (e *EnvFileConfigLoader) LoadConfig() error {
	if fileInfo, err := os.Stat(e.filePath); os.IsNotExist(err) {
		return fmt.Errorf("config directory does not exist: %s", e.filePath)
	} else {
		if fileInfo.IsDir() {
			panic("config directory is not a file please use env_dir_config")
		}
	}

	// 移除扩展名，并将路径分隔符替换为点号，作为配置名
	//比如这样 dev/redis.yaml -> dev.redis
	configName := e.pathToConfigName(e.filePath)

	// 使用工厂创建配置实例
	configInstance, err := GetConfigFactory().CreateByFilePath(e.filePath)
	if err != nil {
		return fmt.Errorf("failed to create config instance for %s: %v", configName, err)
	}

	if err := configInstance.Store(e.filePath); err != nil {
		return fmt.Errorf("failed to store config: %v", err)
	}
	if err := e.configManager.Register(ConfigName(configName), configInstance); err != nil {
		return fmt.Errorf("failed to register config %s: %v", configName, err)
	}

	if err := configInstance.Load(); err != nil {
		return fmt.Errorf("failed to load config %s: %v", configName, err)
	}

	return nil
}

// GetConfigDir 获取配置目录
func (e *EnvFileConfigLoader) GetConfigDir() string {
	return filepath.Dir(e.filePath)
}
