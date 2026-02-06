package config

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/fatzeng/switch-sdk-core/config"
	"gopkg.in/yaml.v2"
)

// YamlConfig 承载yaml的结构
type YamlConfig struct {
	configPath string
	config     map[string]interface{}
}

// NewYamlConfig 创建yaml解析
func NewYamlConfig() config.ConfigI {
	return &YamlConfig{
		config: make(map[string]interface{}),
	}
}

func (y *YamlConfig) Store(path string) error {
	y.configPath = path
	return nil
}

// Load Load
func (y *YamlConfig) Load() error {
	// 读取配置文件
	data, err := os.ReadFile(y.configPath)
	if err != nil {
		return fmt.Errorf("read config file error: %v", err)
	}

	// 解析YAML
	if err = yaml.Unmarshal(data, &y.config); err != nil {
		return fmt.Errorf("unmarshal yaml error: %v", err)
	}

	return nil
}

// Get Get
func (y *YamlConfig) Get(key string) interface{} {
	return y.find(key)
}

// Set Set
func (y *YamlConfig) Set(key string, value interface{}) error {
	keys := strings.Split(key, ".")
	currentConfig := y.config

	for i := 0; i < len(keys)-1; i++ {
		if _, ok := currentConfig[keys[i]]; !ok {
			currentConfig[keys[i]] = make(map[string]interface{})
		}
		if current, ok := currentConfig[keys[i]].(map[string]interface{}); !ok {
			return fmt.Errorf("invalid path: %s", key)
		} else {
			currentConfig = current
		}
	}

	currentConfig[keys[len(keys)-1]] = value
	return nil
}

// GetString GetString
func (y *YamlConfig) GetString(key string) string {
	if val := y.Get(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt GetInt
func (y *YamlConfig) GetInt(key string) int {
	if val := y.Get(key); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetBool GetBool
func (y *YamlConfig) GetBool(key string) bool {
	if val := y.Get(key); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetFloat GetFloat
func (y *YamlConfig) GetFloat(key string) float64 {
	if val := y.Get(key); val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

// GetStringMap GetStringMap
func (y *YamlConfig) GetStringMap(key string) map[string]interface{} {
	if val := y.Get(key); val != nil {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// GetStringSlice GetStringSlice
func (y *YamlConfig) GetStringSlice(key string) []string {
	if val := y.Get(key); val != nil {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, v := range slice {
				result[i] = fmt.Sprint(v)
			}
			return result
		}
	}
	return nil
}

// Unmarshal Unmarshal
func (y *YamlConfig) Unmarshal(v interface{}) error {
	data, err := yaml.Marshal(y.config)
	if err != nil {
		return fmt.Errorf("marshal config error: %v", err)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		//Unmarshal提供的错误信息封装一下
		if typeErr, ok := err.(*yaml.TypeError); ok {
			var errDetails strings.Builder
			errDetails.WriteString("yaml unmarshal type error:\n")
			lines := strings.Split(string(data), "\n")

			for _, detail := range typeErr.Errors {
				var lineNum int
				if _, err := fmt.Sscanf(detail, "line %d:", &lineNum); err == nil && lineNum > 0 && lineNum <= len(lines) {
					problematicLine := strings.TrimSpace(lines[lineNum-1])
					fieldParts := strings.SplitN(problematicLine, ":", 2)
					fieldName := strings.TrimSpace(fieldParts[0])
					fieldValue := ""
					if len(fieldParts) > 1 {
						fieldValue = strings.TrimSpace(fieldParts[1])
					}

					errDetails.WriteString(fmt.Sprintf("  - line-num %d:\n", lineNum))
					errDetails.WriteString(fmt.Sprintf("    field-name: %s\n", fieldName))
					errDetails.WriteString(fmt.Sprintf("    curr-val: %s\n", fieldValue))
					errDetails.WriteString(fmt.Sprintf("    err-msg: %s\n", detail))
					errDetails.WriteString(fmt.Sprintf("    origin-line-content: %s\n", problematicLine))
				} else {
					errDetails.WriteString(fmt.Sprintf("  - %s\n", detail))
				}
			}
			return fmt.Errorf("%s", errDetails.String())
		}

		return fmt.Errorf("unmarshal to struct error: %v\nconfig data:\n%s", err, string(data))
	}

	return nil
}

// find 查找配置值（find 按照.去切割的）
func (y *YamlConfig) find(key string) interface{} {
	current := y.config
	keys := strings.Split(key, ".")

	for i, k := range keys {
		if i == len(keys)-1 {
			return current[k]
		}
		if val, ok := current[k].(map[string]interface{}); ok {
			current = val
		} else {
			return nil
		}
	}

	return nil
}

// 注册YAML配置
func init() {
	factory := config.GetConfigFactory()
	_ = factory.Register(".yaml", func() config.ConfigI {
		return NewYamlConfig()
	})
	_ = factory.Register(".yml", func() config.ConfigI {
		return NewYamlConfig()
	})
}
