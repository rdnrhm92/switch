package admin_model

import (
	"encoding/json"

	"gitee.com/fatzeng/switch-sdk-core/model"
)

// SwitchSnapshot 存储开关在特定命名空间和环境下的完整快照
type SwitchSnapshot struct {
	model.CommonModel
	model.Version
	NamespaceTag string          `gorm:"size:100;not null;comment:命名空间tag" json:"namespaceTag"`
	EnvTag       string          `gorm:"size:50;comment:环境标签" json:"envTag"`
	SwitchID     uint            `gorm:"not null;comment:开关ID" json:"switchId"`
	CompleteJSON json.RawMessage `gorm:"type:json;not null;comment:完整的开关JSON镜像用于diff" json:"completeJson"`
}

func (SwitchSnapshot) TableName() string {
	return "switch_snapshots"
}
