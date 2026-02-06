package admin_model

import (
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// NamespaceMembers 命名空间成员表
type NamespaceMembers struct {
	model.CommonModel
	UserId       uint   `gorm:"not null;comment:用户ID" json:"userId"`
	NamespaceTag string `gorm:"size:100;not null;index;comment:命名空间标签" json:"namespaceTag"`

	//表关联数据
	UserRoles []*NamespaceUserRole `gorm:"foreignKey:NamespaceMembersId" json:"userRoles"`
	Namespace Namespace            `gorm:"foreignKey:Tag;references:NamespaceTag" json:"namespace,omitempty"`
}

func (NamespaceMembers) TableName() string {
	return "namespace_members"
}
