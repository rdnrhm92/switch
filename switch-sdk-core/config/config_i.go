package config

// ConfigI 定义配置接口，提供配置的基本操作
type ConfigI interface {
	// Store 载入
	Store(path string) error
	// Load 加载配置
	Load() error
	// Get 获取配置值
	Get(key string) interface{}
	// Set 设置配置值
	Set(key string, value interface{}) error
	// GetString 获取字符串配置
	GetString(key string) string
	// GetInt 获取整数配置
	GetInt(key string) int
	// GetBool 获取布尔值配置
	GetBool(key string) bool
	// GetFloat 获取浮点数配置
	GetFloat(key string) float64
	// GetStringMap 获取字符串map配置
	GetStringMap(key string) map[string]interface{}
	// GetStringSlice 获取字符串切片配置
	GetStringSlice(key string) []string
	// Unmarshal 将配置解析到结构体
	Unmarshal(v interface{}) error
}
