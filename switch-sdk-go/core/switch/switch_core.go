// @Author EthanScriptOn
// @Desc
package _switch

import (
	"fmt"
	"strconv"
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-core/statistics"
	"gitee.com/fatzeng/switch-sdk-go/core"
	"gitee.com/fatzeng/switch-sdk-go/core/filter"
	"gitee.com/fatzeng/switch-sdk-go/core/middleware"
	"golang.org/x/net/context"
)

// 默认的执行器
var defaultFilter core.Filter
var once sync.Once

// getFilter 获取执行器
func getFilter() core.Filter {
	once.Do(func() {
		defaultFilter = filter.GenerateDefaultSwitchFilter(filter.GenerateDefaultSwitchFactor(middleware.Middlewares()...))
	})
	return defaultFilter
}

// logSwitchStatistics 打印开关执行统计信息
func logSwitchStatistics(switchContext *core.SwitchContext, switchName string) {
	logger.Logger.Infof("[Switch Statistics] RequestID=%s, Switch=%s, Result=%v, Duration=%v, Factors=%v, Errors=%v",
		switchContext.RequestID(), // 交给业务方设置吧
		switchName,
		switchContext.SwitchExecutionResult(),
		switchContext.SwitchExecutionDuration(),
		formatFactorRecords(switchContext.ExecutionRecords()),
		formatErrors(switchContext.Error()))
}

// formatFactorRecords 格式化因子执行记录
func formatFactorRecords(records []*core.FactorExecutionRecord) string {
	if len(records) == 0 {
		return "[]"
	}
	result := "["
	for i, record := range records {
		if i > 0 {
			result += "; "
		}
		if record.Stat != nil {
			result += record.Name + "(" + record.Stat.Duration() + "," + strconv.FormatBool(record.Stat.Result()) + ")"
			result += record.Name
		}
	}
	result += "]"
	return result
}

// formatErrors 格式化错误信息
func formatErrors(err *reply.Error) string {
	if err == nil || len(err.Details) == 0 {
		return "[]"
	}
	return fmt.Sprintf("%v", err.Details)
}

// IsSwitchOpen 传递开关项
func IsSwitchOpen(ctx context.Context, sm *model.SwitchModel) bool {
	if sm == nil {
		logger.Logger.Warn("IsSwitchOpen switch admin_model is nil")
		return false
	}
	if !GlobalClient.IsInitialized() {
		logger.Logger.Warn("SDK not initialized. Please call core.Start() first")
		return false
	}

	// 开始计时
	factorTimerStat := statistics.TimerBeginWithName(sm.Name)

	//上下文设置
	switchContext, ok := core.FromContext(ctx)
	if !ok {
		switchContext = core.NewSwitchContext(ctx)
	}

	// 获取开关执行器
	flt := getFilter()

	executeBool := flt.Filter(switchContext, sm)
	if sm.Name != core.IS_OPEN_SWITCH_STATISTIC {
		//统计项开启
		switchContext.SetExecutionStatsOpen(IsOpen(switchContext, core.IS_OPEN_SWITCH_STATISTIC))
		// 开关执行结果
		switchContext.SetSwitchExecutionResult(executeBool)
		// 开关执行时间
		switchContext.SetSwitchExecutionDuration(factorTimerStat.Complete().Format())
		// 如果开启了统计
		if switchContext.ExecutionStatsOpen() {
			logSwitchStatistics(switchContext, sm.Name)
		}
	}
	return executeBool
}

// IsOpen 传递开关名
func IsOpen(ctx context.Context, switchName string) bool {
	if switchName == "" {
		logger.Logger.Warn("IsOpen switch named is empty!")
		return false
	}

	//上下文设置
	switchContext, ok := core.FromContext(ctx)
	if !ok {
		switchContext = core.NewSwitchContext(ctx)
	}

	//寻找sm规则
	rule, canFound := GetRule(switchName)
	if !canFound {
		logger.Logger.Warnf("unable to find switch rule named :[%v]", switchName)
		return false
	}
	return IsSwitchOpen(switchContext, rule)
}
