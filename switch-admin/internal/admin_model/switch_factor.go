package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// SwitchFactor 因子信息表
type SwitchFactor struct {
	model.CommonModel
	Factor       string     `gorm:"size:100;not null;comment:因子名称'" json:"factor"`
	Description  string     `gorm:"size:255;comment:因子描述" json:"description"`
	Name         string     `gorm:"size:255;comment:因子显示名称" json:"name"`
	JsonSchema   string     `gorm:"type:text;comment:因子配置的JSON Schema" json:"jsonSchema"`
	NamespaceTag string     `gorm:"size:100;not null;index;comment:命名空间标签" json:"namespaceTag"`
	Namespace    *Namespace `gorm:"foreignKey:Tag;references:NamespaceTag" json:"namespace,omitempty"`
}

func (SwitchFactor) TableName() string {
	return "switch_factor"
}
