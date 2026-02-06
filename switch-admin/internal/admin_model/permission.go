package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// Permission 权限表
type Permission struct {
	model.CommonModel
	Name         string     `gorm:"size:100;not null;unique;comment:权限名称，例如 'switches:create'" json:"name"`
	Description  string     `gorm:"size:255;comment:权限描述" json:"description"`
	NamespaceTag string     `gorm:"size:100;not null;index;comment:命名空间标签" json:"namespaceTag"`
	Namespace    *Namespace `gorm:"foreignKey:Tag;references:NamespaceTag" json:"namespace,omitempty"`
}

func (Permission) TableName() string {
	return "permissions"
}
