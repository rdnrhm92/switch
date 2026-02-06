package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// Role 角色表
type Role struct {
	model.CommonModel
	Name         string        `gorm:"size:50;not null;unique;comment:角色名称" json:"name"`
	Description  string        `gorm:"size:255;comment:角色描述" json:"description"`
	Permissions  []*Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	NamespaceTag string        `gorm:"size:100;not null;index;comment:命名空间标签" json:"namespaceTag"`
	Namespace    *Namespace    `gorm:"foreignKey:Tag;references:NamespaceTag" json:"namespace,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}
