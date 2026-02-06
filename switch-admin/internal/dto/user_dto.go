package dto

import "gitee.com/fatzeng/switch-admin/internal/admin_model"

// UserListReq 定义了查询用户列表的api request
type UserListReq struct {
	*PageLimit
	*PageTime
	Username string `json:"username" form:"username"` // 用户名
	Role     string `json:"role" form:"role"`         // 角色名
}

// UserPermissionsListReq 定义了查询用户(用户+权限)
type UserPermissionsListReq struct {
	Username string `json:"username" form:"username"` // 用户名
}

// UserListResp 定义了查询用户列表的api response
type UserListResp struct {
	UserInfo  admin_model.User    `json:"userInfo"`
	UserRoles []*admin_model.Role `json:"userRoles"`
}

// UserPermissionsListResp 定义了查询用户列表的api response
type UserPermissionsListResp struct {
	UserInfo        admin_model.User          `json:"userInfo"`
	UserPermissions []*admin_model.Permission `json:"userPermissions"`
}

// AssignRolesReq 定义了为用户分配角色的api request
type AssignRolesReq struct {
	NamespaceTag string
	UserId       uint   `json:"userId"`  // 用户ID
	RoleIds      []uint `json:"roleIds"` // 角色ID列表
}

// AllUserLikeReq 模糊查询用户
type AllUserLikeReq struct {
	UserName string `json:"username"`
}

type UserInfo struct {
	*admin_model.User
	SelectNamespace string `json:"select_namespace"`
}

// UpsertPermissionReq 定义了新增/修改权限的api request
type UpsertPermissionReq struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpsertRoleReq 定义了新增/修改角色的api request
type UpsertRoleReq struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	NamespaceTag  string
	PermissionIds []uint `json:"permissionIds"`
}
