// @Author EthanScriptOn
// @Desc
package filter

import (
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-go/core"
)

// DefaultSwitchFilter 开关执行过滤器
type DefaultSwitchFilter struct {
	factorSwitch core.Filter
}

func GenerateDefaultSwitchFilter(factorSwitch core.Filter) *DefaultSwitchFilter {
	return &DefaultSwitchFilter{
		factorSwitch: factorSwitch,
	}
}

func (d *DefaultSwitchFilter) Filter(ctx *core.SwitchContext, rule *model.SwitchModel) bool {
	if rule == nil {
		ctx.Error().WithDetails(fmt.Sprintf("执行开关失败,规则rule为空"))
		return false
	}
	return d.factorSwitch.Filter(ctx, rule)
}
