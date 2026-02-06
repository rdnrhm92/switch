package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// Namespace 命名空间(命名空间下对应着不同的环境)
type Namespace struct {
	model.CommonModel
	Name        string `gorm:"size:100;index;not null;comment:命名空间名称" json:"name"`
	Tag         string `gorm:"size:100;uniqueIndex;not null;comment:命名空间标签" json:"tag"`
	Description string `gorm:"size:255;comment:描述信息" json:"description"`

	//归属者
	OwnerUserId uint `gorm:"not null;comment:归属者" json:"ownerUserId"`

	Environments []*Environment `gorm:"foreignKey:NamespaceTag;references:Tag" json:"environments"`
}

func (Namespace) TableName() string {
	return "namespace"
}
