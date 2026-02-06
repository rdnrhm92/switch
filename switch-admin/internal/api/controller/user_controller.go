package controller

import (
	"net/http"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/service"
	"gitee.com/fatzeng/switch-admin/internal/types"
	"gitee.com/fatzeng/switch-components/bc"
	"github.com/gin-gonic/gin"
)

// UserController 定义了用户相关的控制器
type UserController struct {
	userService *service.UserService
}

// NewUserController 创建一个新的用户控制器实例
func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
	}
}

func (c *UserController) GetMe(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	// token解析
	userDetail := info.GetUserInfo(ctx)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(userDetail)).ToResp())
}

func (c *UserController) GetAllUserLike(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.AllUserLikeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to retrieve user info")).ToResp())
		return
	}
	if req.UserName == "" {
		ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(make([]admin_model.User, 0))).ToResp())
		return
	}
	allUsers, _ := c.userService.GetAllUserLike(ctx, &req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(allUsers)).ToResp())
}

// ListUsers 列表查询用户
func (c *UserController) ListUsers(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.UserListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to retrieve user info")).ToResp())
		return
	}

	if req.PageLimit != nil {
		req.PageLimit.ComputeLimit()
	}

	data := c.userService.GetUsers(ctx, &req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(data)).ToResp())
}

// ListUsersPermissions 列表查询用户
func (c *UserController) ListUsersPermissions(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.UserPermissionsListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to retrieve user info")).ToResp())
		return
	}

	data := c.userService.GetUsersPermissions(ctx, &req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(data)).ToResp())
}
