package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvDirConfigLoader 目录相关的加载器
type EnvDirConfigLoader struct {
	configManager *ConfigManager
	configDir     string
	env           Environment
	filePattern   string
}

// SetFilePattern 设置配置文件模式
func (e *EnvDirConfigLoader) SetFilePattern(pattern string) {
	if pattern == "" {
		panic("empty config pattern")
	}
	e.filePattern = pattern
}

// GetManager 获取配置管理器
func (e *EnvDirConfigLoader) GetManager() *ConfigManager {
	return e.configManager
}

// SetManager 设置配置管理器
func (e *EnvDirConfigLoader) SetManager(manager *ConfigManager) {
	e.configManager = manager
}

// LoadConfigsInDir 递归加载目录下的所有配置
func (e *EnvDirConfigLoader) LoadConfigsInDir(dirPath string) error {
	// 读取目录内容
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dirPath, err)
	}

	foundConfig := false
	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// 如果是目录，递归处理该目录
			if err := e.LoadConfigsInDir(path); err != nil {
				return err
			}
			continue
		}

		// 检查是否匹配文件模式
		match, err := filepath.Match(e.filePattern, entry.Name())
		if err != nil {
			return fmt.Errorf("invalid file pattern: %v", err)
		}

		if !match {
			continue
		}

		foundConfig = true
		// 从环境目录开始的相对路径作为配置名
		relPath, err := filepath.Rel(e.configDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// 移除扩展名，并将路径分隔符替换为点号，作为配置名
		configName := e.pathToConfigName(relPath)

		// 使用工厂创建配置实例
		configInstance, err := GetConfigFactory().CreateByFilePath(path)
		if err != nil {
			return fmt.Errorf("failed to create config instance for %s: %v", path, err)
		}

		if err = configInstance.Store(path); err != nil {
			return fmt.Errorf("failed to store config: %v", err)
		}
		if err = e.configManager.Register(ConfigName(configName), configInstance); err != nil {
			return fmt.Errorf("failed to register config %s: %v", configName, err)
		}
		if err = configInstance.Load(); err != nil {
			return fmt.Errorf("failed to load config %s: %v", configName, err)
		}
	}

	if !foundConfig && dirPath == e.configDir {
		return fmt.Errorf("no config files matching pattern '%s' were found in directory '%s' or its subdirectories", e.filePattern, e.configDir)
	}

	return nil
}

// pathToConfigName 将文件路径转换为配置名
func (e *EnvDirConfigLoader) pathToConfigName(path string) string {
	basename := strings.TrimSuffix(path, filepath.Ext(path))
	configName := strings.ReplaceAll(basename, string(filepath.Separator), ".")
	return configName
}

// GetEnv 获取当前环境
func (e *EnvDirConfigLoader) GetEnv() Environment {
	return e.env
}

// GetConfigDir 获取配置目录
func (e *EnvDirConfigLoader) GetConfigDir() string {
	return e.configDir
}

// SetConfigDir 设置配置目录
func (e *EnvDirConfigLoader) SetConfigDir(configDir string) {
	e.configDir = configDir
}
