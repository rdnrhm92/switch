package admin_model

import (
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// Environment 定义了项目的运行环境
type Environment struct {
	model.CommonModel
	Name         string `gorm:"size:255;not null;uniqueIndex:idx_name_tag,priority:1;comment:环境名称" json:"name"`
	Tag          string `gorm:"size:255;not null;uniqueIndex:idx_name_tag,priority:2;comment:环境标签" json:"tag"`
	Description  string `gorm:"size:500;comment:描述" json:"description"`
	PublishOrder int    `gorm:"comment:发布顺序" json:"publish_order"`
	Publish      *bool  `gorm:"comment:发布" json:"publish"`

	NamespaceTag string     `gorm:"size:100;not null;index;comment:所属命名空间标签" json:"namespaceTag"`
	Namespace    *Namespace `gorm:"foreignKey:Tag;references:NamespaceTag" json:"namespace"`

	Drivers []*model.Driver `gorm:"foreignKey:EnvironmentID" json:"drivers"`
}

func (Environment) TableName() string {
	return "environments"
}
