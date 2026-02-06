// @Author EthanScriptOn
// @Desc
package factor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/actuator"
	"gitee.com/fatzeng/switch-sdk-core/factor"
	"gitee.com/fatzeng/switch-sdk-core/tool"
)

func init() {
	actuator.Register(factor.IP, IP_FactorActuator)
	actuator.Register(factor.UserName, UserName_FactorActuator)
	actuator.Register(factor.Location, Location_FactorActuator)
	actuator.Register(factor.UserId, UserId_FactorActuator)
	actuator.Register(factor.TimeRange, TimeRange_FactorActuator)
	actuator.Register(factor.Single, Single_FactorActuator)
	actuator.Register(factor.TelNum, TelNum_FactorActuator)
	actuator.Register(factor.UserNick, UserNick_FactorActuator)
	actuator.Register(factor.Custom_Arbitrarily, Custom_Arbitrarily_FactorActuator)
	actuator.Register(factor.Custom_Whole, Custom_Whole_FactorActuator)
}

// IP_FactorActuator 系统原数据 - IP比较
var IP_FactorActuator = func(ctx context.Context, config *factor.IpConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.IpConfig")
	}
	contain, err := executeMatchFactor(ctx, config.ContextKey, config.ContextVal)
	if err != nil {
		return false, err
	}

	if config.IsBlack {
		// 黑名单包含则false
		return !contain, nil
	} else {
		// 白名单包含则true
		return contain, nil
	}
}

// UserName_FactorActuator 系统原数据 - 用户名匹配
var UserName_FactorActuator = func(ctx context.Context, config *factor.UserNameConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.UserNameConfig")
	}
	return executeMatchFactor(ctx, config.ContextKey, config.ContextVal)
}

// Location_FactorActuator 系统原数据 - 区域匹配
var Location_FactorActuator = func(ctx context.Context, config *factor.LocationConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.LocationConfig")
	}
	return executeMatchFactor(ctx, config.ContextKey, config.ContextVal)
}

// UserId_FactorActuator 系统原数据 - 用户ID匹配
var UserId_FactorActuator = func(ctx context.Context, config *factor.UserIdConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.UserIdConfig")
	}
	return executeMatchFactor(ctx, config.ContextKey, config.ContextVal)
}

// TimeRange_FactorActuator 系统原数据 - 时间范围匹配
var TimeRange_FactorActuator = func(ctx context.Context, config *factor.TimeRangeConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.TimeRangeConfig")
	}
	startTime, err := strconv.ParseInt(config.StartTime, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid start_time: %w", err)
	}
	endTime, err := strconv.ParseInt(config.EndTime, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid end_time: %w", err)
	}

	now := time.Now().Unix()
	return now >= startTime && now <= endTime, nil
}

// Single_FactorActuator 系统原数据 - 单一开关
var Single_FactorActuator = func(ctx context.Context, config *factor.SingleConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.SingleConfig")
	}
	return config.Enabled, nil
}

// TelNum_FactorActuator 系统原数据 - 手机号匹配
var TelNum_FactorActuator = func(ctx context.Context, config *factor.TelNumConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.TelNumConfig")
	}
	return executeMatchFactor(ctx, config.ContextKey, config.ContextVal)
}

// UserNick_FactorActuator 系统原数据 - 用户昵称匹配
var UserNick_FactorActuator = func(ctx context.Context, config *factor.UserNickConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("invalid *factor.UserNickConfig")
	}
	contain, err := executeMatchFactor(ctx, config.ContextKey, config.ContextVal)
	if err != nil {
		return false, err
	}

	if config.IsBlack {
		return !contain, nil
	} else {
		return contain, nil
	}
}

// Custom_Arbitrarily_FactorActuator 系统原数据 - 自定义任意匹配
var Custom_Arbitrarily_FactorActuator = func(ctx context.Context, config *factor.CustomArbitrarilyConfig) (bool, error) {
	if config == nil || len(*config) == 0 {
		return false, fmt.Errorf("invalid *factor.CustomArbitrarilyConfig")
	}
	for _, c := range *config {
		match, err := executeMatchFactor(ctx, c.ContextKey, c.ContextVal)
		if err != nil {
			continue
		}
		if match {
			//任意一个匹配成功即可
			return true, nil
		}
	}
	return false, nil
}

// Custom_Whole_FactorActuator 系统原数据 - 自定义全量匹配
var Custom_Whole_FactorActuator = func(ctx context.Context, config *factor.CustomWholeConfig) (bool, error) {
	if config == nil || len(*config) == 0 {
		return false, fmt.Errorf("invalid *factor.CustomWholeConfig")
	}
	for _, c := range *config {
		match, err := executeMatchFactor(ctx, c.ContextKey, c.ContextVal)
		if err != nil {
			return false, err
		}
		if !match {
			//任意一个不匹配则整体失败
			return false, nil
		}
	}
	return true, nil
}

// executeStringMatchFactor 是处理通用字符串匹配的辅助函数（向后兼容）
func executeStringMatchFactor(ctx context.Context, key string, configValues []string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("invalid context key")
	}
	if len(configValues) == 0 {
		return false, fmt.Errorf("invalid context val")
	}

	value := ctx.Value(key)
	if value == nil {
		return false, nil
	}

	return factor.ContainsString(tool.ToString(value), configValues), nil
}

// executeMatchFactor 是处理通用值匹配的辅助函数，支持多种数据类型
func executeMatchFactor(ctx context.Context, key string, configValue interface{}) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("invalid context key")
	}
	if configValue == nil {
		return false, fmt.Errorf("invalid context val")
	}

	contextValue := ctx.Value(key)
	if contextValue == nil {
		return false, nil
	}

	return factor.MatchValue(contextValue, configValue), nil
}
