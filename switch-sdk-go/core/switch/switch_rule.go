// @Author EthanScriptOn
// @Desc
package _switch

import (
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/model"
)

var (
	ruleContainer = make(map[string]*model.SwitchModel)
	mu            sync.RWMutex
)

func GetRule(switchName string) (*model.SwitchModel, bool) {
	mu.RLock()
	defer mu.RUnlock()
	rule, ok := ruleContainer[switchName]
	return rule, ok
}

func UnregisterRule(switchName string) {
	mu.Lock()
	defer mu.Unlock()
	delete(ruleContainer, switchName)
}

// ClearAllRules 退出清空所有
func ClearAllRules() {
	mu.Lock()
	defer mu.Unlock()
	ruleContainer = make(map[string]*model.SwitchModel)
}

func RegisterRule(switchName string, rule *model.SwitchModel) {
	mu.Lock()
	defer mu.Unlock()

	existingRule, ok := ruleContainer[switchName]
	if !ok || rule.Version.Version > existingRule.Version.Version {
		ruleContainer[switchName] = rule
	}
}
