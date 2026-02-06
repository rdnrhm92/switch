package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// RolePermission 角色ID跟权限ID对照表
type RolePermission struct {
	model.CommonModel
	RoleID       uint `gorm:"not null;uniqueIndex:idx_role_perm,priority:1;comment:角色ID"`
	PermissionID uint `gorm:"not null;uniqueIndex:idx_role_perm,priority:2;comment:权限ID"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
