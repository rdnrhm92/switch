package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// NamespaceUserRole 记录命名空间跟角色的关联关系
type NamespaceUserRole struct {
	model.CommonModel
	RoleId             uint `gorm:"not null;comment:角色ID" json:"roleId"`
	NamespaceMembersId uint `gorm:"not null;comment:namespaceMembers表主键" json:"namespaceMembersId"`

	//关联表
	Role Role `gorm:"foreignKey:RoleId" json:"role,omitempty"`
}

// TableName 表名
func (NamespaceUserRole) TableName() string {
	return "namespace_user_role"
}
