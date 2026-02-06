package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// SwitchConfig 存储一个开关在特定环境下的配置,switch_snapshot的作用主要是diff以及回溯
type SwitchConfig struct {
	model.CommonModel
	model.Version
	SwitchID    uint            `gorm:"not null;uniqueIndex:idx_switch_env;comment:开关ID 关联switch_model" json:"switchId"`
	EnvTag      string          `gorm:"size:50;not null;uniqueIndex:idx_switch_env;comment:当前环境" json:"envTag"`
	ConfigValue *model.RuleNode `gorm:"type:json;not null;comment:当前开关在当前环境下的规则配置" json:"configValue"`
	Status      string          `gorm:"size:50;not null;default:'PENDING';comment:当前开关在当前环境下的发布状态" json:"status"`
}

func (SwitchConfig) TableName() string {
	return "switch_configs"
}
