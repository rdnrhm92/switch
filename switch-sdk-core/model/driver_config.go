package model

// DriverConfig 是一个通用的驱动配置结构体
type DriverConfig struct {
	// Type 驱动类型，例如 "kafka", "webhook"。
	Type string `json:"type"`
	// Properties json yaml
	Properties map[string]interface{} `json:"properties"`
}
