package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// 提供一个业务可以自己创建解析器的工厂
type ConfigCreator func() ConfigI

// ConfigFactory 配置工厂
type ConfigFactory struct {
	mu sync.RWMutex
	//key => 文件的后缀
	creators map[string]ConfigCreator
}

var (
	defaultFactory *ConfigFactory
	once           sync.Once
)

func GetConfigFactory() *ConfigFactory {
	once.Do(func() {
		defaultFactory = &ConfigFactory{
			creators: make(map[string]ConfigCreator),
		}
	})
	return defaultFactory
}

// Register 注册配置创建器
// ext 是文件后缀比如abc.yaml那么ext就是yaml比如abc.yml那么ext就是yml
func (f *ConfigFactory) Register(ext string, creator ConfigCreator) error {
	if creator == nil {
		return fmt.Errorf("creator function cannot be nil")
	}

	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.creators[ext]; exists {
		return fmt.Errorf("creator for extension %s already registered", ext)
	}

	f.creators[ext] = creator
	return nil
}

// Register 注册配置创建器
// ext 是文件后缀比如abc.yaml那么ext就是yaml比如abc.yml那么ext就是yml
func (f *ConfigFactory) RegisterByFilePath(filePath string, creator ConfigCreator) error {
	if creator == nil {
		return fmt.Errorf("creator function cannot be nil")
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.creators[ext]; exists {
		return fmt.Errorf("creator for extension %s already registered", ext)
	}

	f.creators[ext] = creator
	return nil
}

// CreateByFilePath 根据文件路径创建对应的配置实例
func (f *ConfigFactory) CreateByFilePath(filePath string) (ConfigI, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return nil, fmt.Errorf("file %s has no extension", filePath)
	}

	f.mu.RLock()
	creator, exists := f.creators[ext]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no config creator registered for extension %s", ext)
	}

	return creator(), nil
}

// CreateByExt 根据文件扩展名创建对应的配置实例
func (f *ConfigFactory) CreateByExt(fileExt string) (ConfigI, error) {
	if fileExt == "" {
		return nil, fmt.Errorf("fileExt is empty")
	}
	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}
	ext := strings.ToLower(fileExt)

	f.mu.RLock()
	creator, exists := f.creators[ext]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no config creator registered for extension %s", ext)
	}

	return creator(), nil
}

// Unregister 取消注册配置创建器
func (f *ConfigFactory) Unregister(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	f.mu.Lock()
	delete(f.creators, ext)
	f.mu.Unlock()
}

// GetSupportedExtensions 获取所有支持的文件后缀
func (f *ConfigFactory) GetSupportedExtensions() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	exts := make([]string, 0, len(f.creators))
	for ext := range f.creators {
		exts = append(exts, ext)
	}
	return exts
}
