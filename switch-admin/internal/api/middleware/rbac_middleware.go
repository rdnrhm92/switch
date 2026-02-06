package middleware

import (
	"net/http"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/types"
	"gitee.com/fatzeng/switch-components/bc"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware 一个rbac中间件，用作前端面板的精细化控制
// 支持传入多个权限，满足任意一个即可通过
func RBACMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		respCtx := bc.NewRespContext(c)
		wrapper := bc.GetResp(respCtx)
		//来自jwt
		userInfo := info.GetUserInfo(c)
		if userInfo == nil {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoPermissions("User role not found in token")).ToResp())
			c.Abort()
			return
		}

		// 检查是否包含 canSuperAdmin 特殊权限
		hasCanSuperAdmin := false
		for _, perm := range requiredPermissions {
			if perm == "canSuperAdmin" {
				hasCanSuperAdmin = true
				break
			}
		}

		if hasCanSuperAdmin {
			if userInfo.IsSuperAdmin != nil && *userInfo.IsSuperAdmin {
				c.Next()
				return
			}
			if len(requiredPermissions) == 1 {
				// 只要求 canSuperAdmin，直接拒绝
				c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoPermissions("Super admin permission required")).ToResp())
				c.Abort()
				return
			}
		}

		// 超级管理员拥有所有其他权限，直接通过
		if userInfo.IsSuperAdmin != nil && *userInfo.IsSuperAdmin {
			c.Next()
			return
		}

		if len(userInfo.NamespaceMembers) <= 0 {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoPermissions("You do not have permission to perform this action")).ToResp())
			c.Abort()
			return
		}
		nsTag := userInfo.SelectNamespace
		if nsTag == "" {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoPermissions("You need to choose a namespace")).ToResp())
			c.Abort()
			return
		}

		//用户在某个空间下可能有多个角色对应多个权限
		roles := make([]*admin_model.Role, 0)
		for _, nm := range userInfo.NamespaceMembers {
			if nm.NamespaceTag == nsTag {
				if len(nm.UserRoles) > 0 {
					for _, role := range nm.UserRoles {
						roles = append(roles, &role.Role)
					}
				}
			}
		}
		if len(roles) <= 0 {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoPermissions("You do not have permission to perform this action")).ToResp())
			c.Abort()
			return
		}

		// 检查用户是否拥有任意一个所需权限
		for _, role := range roles {
			for _, permission := range role.Permissions {
				for _, requiredPerm := range requiredPermissions {
					if requiredPerm == "canSuperAdmin" {
						continue
					}
					if permission.Name == requiredPerm {
						c.Next()
						return
					}
				}
			}
		}

		c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoPermissions("You do not have permission to perform this action")).ToResp())
		c.Abort()
		return
	}
}
