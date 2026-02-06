package resp

import (
	_ "fmt"
	"reflect"
)

type FieldMasksFun func(interface{}) interface{}

// SecurityConfig 安全配置
type SecurityConfig struct {
	MaskFields    []string                 `json:"maskFields"`   // 脱敏的字段
	FilterFields  []string                 `json:"filterFields"` // 排除的字段
	PassDirection bool                     `json:"include"`      //配置正向逆向
	FieldMasks    map[string]FieldMasksFun `json:"FieldMasks"`   // 字段脱敏规则
}

// FilterFields 过滤字段
func FilterFields(data interface{}, config *SecurityConfig) interface{} {
	if config == nil || data == nil {
		return data
	}

	// 处理map类型
	if dataMap, ok := data.(map[string]interface{}); ok {
		result := make(map[string]interface{})

		if config.PassDirection {
			// 只包含指定字段
			for _, field := range config.FilterFields {
				if value, exists := dataMap[field]; exists {
					result[field] = value
				}
			}
		} else {
			// 排除指定字段
			for key, value := range dataMap {
				shouldExclude := false
				for _, field := range config.FilterFields {
					if key == field {
						shouldExclude = true
						break
					}
				}
				if !shouldExclude {
					result[key] = value
				}
			}
		}

		return result
	}

	// 处理结构体类型
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() == reflect.Struct {
		result := make(map[string]interface{})
		typ := value.Type()

		for i := 0; i < value.NumField(); i++ {
			fieldName := typ.Field(i).Name
			fieldValue := value.Field(i).Interface()

			if config.PassDirection {
				// 只包含指定字段
				for _, field := range config.FilterFields {
					if fieldName == field {
						result[fieldName] = fieldValue
						break
					}
				}
			} else {
				// 排除指定字段
				shouldExclude := false
				for _, field := range config.FilterFields {
					if fieldName == field {
						shouldExclude = true
						break
					}
				}
				if !shouldExclude {
					result[fieldName] = fieldValue
				}
			}
		}

		return result
	}

	return data
}

// MaskFields 脱敏字段
func MaskFields(data interface{}, config *SecurityConfig) interface{} {
	if data == nil || config == nil {
		return data
	}
	if dataMap, ok := data.(map[string]interface{}); !ok {
		return data
	} else {
		//配置正向
		if config.PassDirection {
			for _, field := range config.MaskFields {
				if process, ok := config.FieldMasks[field]; ok {
					if processData, ok := dataMap[field]; ok {
						dataMap[field] = process(processData)
					}
				}
			}
		} else {
			//配置逆向
			needProcess := make(map[string]FieldMasksFun)
			//准备处理的
			prepareProcess := make(map[string]struct{})
			for _, field := range config.MaskFields {
				prepareProcess[field] = struct{}{}
			}
			if len(prepareProcess) <= 0 {
				return data
			}

			for field := range dataMap {
				_, prepareOk := prepareProcess[field]
				process, processOk := config.FieldMasks[field]
				if !prepareOk && processOk {
					needProcess[field] = process
				}
			}
			if len(needProcess) <= 0 {
				return data
			}

			for field, process := range needProcess {
				dataMap[field] = process(process)
			}
		}
	}
	return data
}
