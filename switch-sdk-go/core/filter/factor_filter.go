// @Author EthanScriptOn
// @Desc
package filter

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/fatzeng/switch-sdk-core/actuator"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-core/statistics"
	"gitee.com/fatzeng/switch-sdk-go/core"
	"gitee.com/fatzeng/switch-sdk-go/core/factor_statistics"
	"gitee.com/fatzeng/switch-sdk-go/core/middleware"
)

type DefaultSwitchFactor struct {
	middlewares []middleware.Middleware
}

func GenerateDefaultSwitchFactor(middlewares ...middleware.Middleware) *DefaultSwitchFactor {
	return &DefaultSwitchFactor{
		middlewares: middlewares,
	}
}

func (d *DefaultSwitchFactor) Filter(ctx *core.SwitchContext, rule *model.SwitchModel) bool {
	//聚合计算
	return d.aggregation(ctx, rule, rule.Rules)
}

// isLogicalNodes 逻辑节点校验
func (d *DefaultSwitchFactor) isLogicalNodes(factorRule *model.RuleNode) bool {
	return factorRule.NodeType != ""
}

// and 逻辑AND
func (d *DefaultSwitchFactor) and(factorRule *model.RuleNode) bool {
	return strings.EqualFold(factorRule.NodeType, "AND")
}

// or 逻辑OR
func (d *DefaultSwitchFactor) or(factorRule *model.RuleNode) bool {
	return strings.EqualFold(factorRule.NodeType, "OR")
}

// aggregation 聚合计算，计算and跟or
func (d *DefaultSwitchFactor) aggregation(ctx *core.SwitchContext, switchRule *model.SwitchModel, factorRule *model.RuleNode) bool {
	//逻辑节点
	if d.isLogicalNodes(factorRule) {
		if d.and(factorRule) {
			if len(factorRule.Children) == 0 {
				return true
			}
			for _, child := range factorRule.Children {
				if !d.aggregation(ctx, switchRule, child) {
					//AND 条件必须都为true
					return false
				}
			}
			return true
		} else if d.or(factorRule) {
			if len(factorRule.Children) == 0 {
				return true
			}
			for _, child := range factorRule.Children {
				if d.aggregation(ctx, switchRule, child) {
					//OR 条件有一个true则true
					return true
				}
			}
			return false
		}
	}
	//非逻辑节点直接执行
	return d.calculate(ctx, switchRule, factorRule)
}

func (d *DefaultSwitchFactor) calculate(ctx *core.SwitchContext, switchRule *model.SwitchModel, factorRule *model.RuleNode) bool {
	factorStat := ctx.UseFactorExecutionRecord(factorRule.Factor)
	factorTimerStat := statistics.TimerBeginWithName(factorRule.Factor)

	defer func() {
		factorStat.SetDuration(factorTimerStat.Complete().Format())
	}()

	if factorRule.Factor == "" {
		factorStat.Error().WithDetails(fmt.Sprintf("开关[%v]-因子[%v]-找不到因子执行器", switchRule.Name, factorRule.Factor))
		return false
	}

	//构建finalHandler执行因子
	var finalHandler middleware.Handler = func(ctx context.Context, switchRule *model.SwitchModel, factorRule *model.RuleNode) bool {
		//factorRule.Config json
		factorResult, err := actuator.Dispatcher(ctx, factorRule.Factor, factorRule.Config)
		if err != nil {
			factorStat.Error().WithDetails(fmt.Sprintf("开关[%v]-因子[%v]-Dispatcher执行异常: %v", switchRule.Name, factorRule.Factor, err.Error()))
		}
		factorStat.SetResult(factorResult)
		return factorResult
	}
	//如果有中间件则执行中间件最终执行finalHandler
	for i := len(d.middlewares) - 1; i >= 0; i-- {
		finalHandler = d.middlewares[i](finalHandler)
	}

	//在一次请求中使用同一个ctx，如果遇到并发调用IsOpen的情况，因子统计项会被污染
	//使用临时ctx解决这个问题，多个临时ctx共同拥有同一个core.SwitchContext
	tempCtx := factor_statistics.NewFactorExecuteStatsContext(ctx, factorStat)

	return finalHandler(tempCtx, switchRule, factorRule)
}
