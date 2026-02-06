package dto

// RoleListReq 角色列表查询参数
type RoleListReq struct {
	*PageLimit
	NamespaceTag   string
	RoleName       string `json:"roleName"`
	PermissionName string `json:"permissionName"`
}
