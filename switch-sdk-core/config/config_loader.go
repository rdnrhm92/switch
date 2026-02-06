package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const DefaultFilePattern = "config-%s.yaml"

// ConfigLoader 配置加载器
type ConfigLoader struct {
	configPath string
	env        Environment
	bindings   []*binding
	isDir      bool
	//内置的文件解析器
	fileConfigLoader *EnvFileConfigLoader
	//内置的目录解析器
	dirConfigLoader *EnvDirConfigLoader
	//此处暂时直接定义，只加载yaml的数据，后续有其他格式，将新增ConfigI路由器
	filePattern string
}

// binding 配置绑定信息
type binding struct {
	target interface{}
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		bindings: make([]*binding, 0),
	}
}

func (l *ConfigLoader) FilePattern(filePattern string) *ConfigLoader {
	if filePattern == "" {
		panic("empty config pattern")
	}
	l.filePattern = filePattern
	return l
}

func (l *ConfigLoader) Env(resolver Resolver) *ConfigLoader {
	if resolver == nil {
		panic("resolver cannot be nil")
	}
	if err := envResolver(resolver); err != nil {
		panic(err)
	}
	environment := GetOsEnvironment()
	if environment == "" {
		panic("environment variable not set if you do not want to distinguish the environment, please do not use Env()")
	}
	l.env = environment
	return l
}

// 做一次全局的初始化
func (l *ConfigLoader) initial() {
	filePattern := l.filePattern
	if filePattern == "" {
		filePattern = DefaultFilePattern
	}
	if l.isDir {
		l.dirConfigLoader.filePattern = filePattern
	}
}

// From 指定从哪加载。可以指定具体的配置文件 也可以指定对应的目录
// 如果指定的是目录，那么加载的是该目录下的对应环境的文件夹下的所有的配置文件
func (l *ConfigLoader) From(paths ...string) *ConfigLoader {
	configPath, err := l.resolverConfigPath(paths...)
	if err != nil || configPath == "" {
		panic(fmt.Sprintf("empty configPath need a dir or file path err: %v", err))
	}
	// 检查路径是否存在
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Errorf("path does not exist: %s", configPath))
		}
		panic(fmt.Errorf("failed to access path: %s, error: %v", configPath, err))
	}

	//如果是目录找目录下对应的环境文件夹
	if fileInfo.IsDir() {
		if l.env == "" {
			panic(fmt.Errorf("environment set to empty, unable to match specific environment directory"))
		}
		environment := l.env
		// 读取目录下的所有文件夹
		entries, err := os.ReadDir(configPath)
		if err != nil {
			panic(fmt.Errorf("failed to read directory: %s, error: %v", configPath, err))
		}

		// 查找环境对应的文件夹
		var envDirFound bool
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == string(environment) {
				envDirFound = true
				//构建环境目录
				configPathEnv := filepath.Join(configPath, entry.Name())
				// 初始化目录配置加载器
				l.dirConfigLoader = &EnvDirConfigLoader{
					configManager: GlobalConfigManager,
					configDir:     configPathEnv,
					env:           environment,
				}
				break
			}
		}

		if !envDirFound {
			panic(fmt.Errorf("environment directory not found: %s in %s", environment, configPath))
		}
	} else {
		// 初始化文件配置加载器
		l.fileConfigLoader = &EnvFileConfigLoader{
			configManager: GlobalConfigManager,
			filePath:      configPath,
			env:           l.env,
		}
	}
	l.configPath = configPath
	l.isDir = fileInfo.IsDir()
	return l
}

// Bind 添加配置绑定
func (l *ConfigLoader) Bind(target interface{}) *ConfigLoader {
	l.bindings = append(l.bindings, &binding{
		target: target,
	})
	return l
}

// Load 加载配置
func (l *ConfigLoader) Load() error {
	l.initial()
	if l.isDir {
		//预加载所有配置
		if err := l.loadFromDir(); err != nil {
			return err
		}
		//映射所有配置
		return l.bind()
	}
	//同上
	if err := l.loadFromFile(); err != nil {
		return err
	}
	return l.bind()
}

// resolverConfigPath 探测逻辑
func (l *ConfigLoader) resolverConfigPath(pathElem ...string) (string, error) {
	if len(pathElem) == 0 {
		if l.filePattern == "" {
			return "", fmt.Errorf("when no directory is specified, probe mode will be used to search for configuration files. Please specify the configuration file format: FilePattern()")
		}
		findPath, err := findConfigPath(l.filePattern)
		if err != nil {
			return "", err
		}
		pathElem = append([]string{findPath}, pathElem...)
	}

	return resolvePath(pathElem...)
}

// loadFromDir 从目录加载配置
func (l *ConfigLoader) loadFromDir() error {
	// 构建环境配置目录路径
	envConfigDir := filepath.Join(l.configPath, string(l.env))

	// 检查目录是否存在
	if _, err := os.Stat(envConfigDir); os.IsNotExist(err) {
		return fmt.Errorf("config directory does not exist: %s", envConfigDir)
	}

	return l.dirConfigLoader.LoadConfigsInDir(envConfigDir)
}

// loadFromFile 从单个文件加载配置
func (l *ConfigLoader) loadFromFile() error {
	return l.fileConfigLoader.LoadConfig()
}

// NameBindRule 默认的绑定规则：配置文件名和结构体名完全匹配
func NameBindRule(configName string, structName string) bool {
	return strings.EqualFold(configName, structName)
}

// FieldBindRule 默认的绑定规则：配置文件名和结构体字段名完全匹配
func FieldBindRule(configName string, structName string) bool {
	return strings.EqualFold(configName, structName)
}

// Bind 执行配置绑定
func (l *ConfigLoader) bind() error {
	if len(l.bindings) == 0 {
		return fmt.Errorf("no bind rules or targets defined")
	}

	// 获取所有已加载的配置
	configNames := GlobalConfigManager.List()

	// 对每个绑定规则和目标进行处理
	for _, buildTarget := range l.bindings {
		target := buildTarget.target

		targetType := reflect.ValueOf(target)
		if targetType.Kind() != reflect.Ptr {
			return fmt.Errorf("v must be a pointer to struct")
		}

		targetInstanceType := targetType.Elem()
		if targetInstanceType.Kind() != reflect.Struct {
			return fmt.Errorf("v must be a pointer to struct")
		}

		// 查找匹配的配置
		for _, configName := range configNames {
			// 获取配置并绑定到结构体
			config, err := GlobalConfigManager.Get(configName)
			if err != nil {
				return fmt.Errorf("failed to get config %s: %v", configName, err)
			}
			if err = config.Unmarshal(target); err != nil {
				return fmt.Errorf("failed to unmarshal config %s: %v", configName, err)
			}
		}
	}

	return nil
}
