// @Author EthanScriptOn
// @Desc
package core

import (
	"gitee.com/fatzeng/switch-sdk-core/model"
)

type Filter interface {
	Filter(ctx *SwitchContext, sm *model.SwitchModel) bool
}
