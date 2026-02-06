package factor

import (
	"gitee.com/fatzeng/switch-sdk-core/tool"
)

//这里的因子执行规则是跟因子一一绑定的

type ContextKv struct {
	ContextKey string      `json:"context_key" description:"用于从上下文中取值的键"`
	ContextVal interface{} `json:"context_val" description:"用于比较的值，支持多种类型"`
}

// CustomWholeConfig 自定义map匹配,规则map中的k跟v必须在上下文中找到
type CustomWholeConfig = []ContextKv

// CustomArbitrarilyConfig 自定义map匹配,规则map中的k跟v有任意一个可以上下文中找到
type CustomArbitrarilyConfig = []ContextKv

// UserNickConfig 用户昵称匹配
type UserNickConfig struct {
	ContextKv
	IsBlack bool `json:"isBlack" description:"是否是黑名单"`
}

// TelNumConfig 用户手机号精确匹配
type TelNumConfig = ContextKv

// SingleConfig 单一开关
type SingleConfig struct {
	Enabled bool `json:"enabled" description:"开关是否启用"`
}

// TimeRangeConfig 时间范围匹配
type TimeRangeConfig struct {
	StartTime string `json:"start_time" description:"开始时间 (秒)"`
	EndTime   string `json:"end_time" description:"结束时间 (秒)"`
}

// UserIdConfig 用户ID匹配
type UserIdConfig = ContextKv

// LocationConfig 区域匹配
type LocationConfig = ContextKv

// UserNameConfig 用户名匹配
type UserNameConfig = ContextKv

// IpConfig 用户IP匹配
type IpConfig struct {
	ContextKv
	IsBlack bool `json:"isBlack" description:"是否是黑名单"`
}

// ContainsString 检查字符串切片中是否包含某个字符串
func ContainsString(checkValue string, slice []string) bool {
	for _, item := range slice {
		if item == checkValue {
			return true
		}
	}
	return false
}

// MatchValue 通用值匹配函数，支持多种数据类型
func MatchValue(contextValue interface{}, configValue interface{}) bool {
	if contextValue == nil || configValue == nil {
		return false
	}
	switch cv := configValue.(type) {
	case string:
		return tool.ToString(contextValue) == cv
	case []interface{}:
		contextStr := tool.ToString(contextValue)
		for _, item := range cv {
			if tool.ToString(item) == contextStr {
				return true
			}
		}
		return false
	case []string:
		return ContainsString(tool.ToString(contextValue), cv)
	case int, int64, float64:
		return tool.ToString(contextValue) == tool.ToString(cv)
	case bool:
		contextBool := tool.ToBool(contextValue)
		return contextBool == cv
	default:
		return tool.ToString(contextValue) == tool.ToString(cv)
	}
}
