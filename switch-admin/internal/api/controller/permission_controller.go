package controller

import (
	"net/http"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/service"
	"gitee.com/fatzeng/switch-admin/internal/types"
	"gitee.com/fatzeng/switch-components/bc"
	"github.com/gin-gonic/gin"
)

// PermissionController 权限相关
type PermissionController struct {
	permissionService *service.PermissionService
}

func NewPermissionController() *PermissionController {
	return &PermissionController{
		permissionService: service.NewPermissionService(),
	}
}

// AssignRoles 分配用户角色
func (c *PermissionController) AssignRoles(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.AssignRolesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to retrieve user info")).ToResp())
		return
	}
	if req.UserId == 0 || len(req.RoleIds) <= 0 {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("请检查 UserId RoleIds 是否为空")).ToResp())
		return
	}
	req.NamespaceTag = info.GetSelectNamespace(ctx)
	if err := c.permissionService.AssignRoles(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("fail to AssignRoles")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(new(interface{}))).ToResp())
}

// RolesPermission 角色权限查询
func (c *PermissionController) RolesPermission(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.RoleListReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}
	if req.PageLimit != nil {
		req.ComputeLimit()
	}
	req.NamespaceTag = info.GetSelectNamespace(ctx)

	result := c.permissionService.RolesPermission(&req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(result)).ToResp())
}

// Permissions 权限查询
func (c *PermissionController) Permissions(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	selectNamespaceTag := info.GetSelectNamespace(ctx)

	permissions, err := c.permissionService.Permissions(selectNamespaceTag)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("fail to watch permissions")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(permissions)).ToResp())
}

// UpdatePermission 修改权限
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.UpsertPermissionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}

	if err := c.permissionService.UpsertPermission(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to upsert permission")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}

// InsertPermission 新增权限
func (c *PermissionController) InsertPermission(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.UpsertPermissionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}

	if err := c.permissionService.UpsertPermission(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to upsert permission")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}

// InsertRole 新增角色
func (c *PermissionController) InsertRole(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.UpsertRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}

	req.NamespaceTag = info.GetSelectNamespace(ctx)

	if err := c.permissionService.UpsertRole(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to upsert role")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}

// UpdateRole 修改角色
func (c *PermissionController) UpdateRole(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.UpsertRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}

	req.NamespaceTag = info.GetSelectNamespace(ctx)

	if err := c.permissionService.UpsertRole(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to upsert role")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}
