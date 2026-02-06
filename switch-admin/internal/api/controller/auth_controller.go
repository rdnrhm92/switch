package controller

import (
	"net/http"

	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/service"
	"gitee.com/fatzeng/switch-admin/internal/types"
	"gitee.com/fatzeng/switch-components/bc"
	"github.com/gin-gonic/gin"
)

// AuthController 处理与认证相关的请求
type AuthController struct {
	authService *service.AuthService
}

// NewAuthController 创建一个新的 AuthController
func NewAuthController(cfg *config.Config) *AuthController {
	return &AuthController{
		authService: service.NewAuthService(cfg),
	}
}

// Register 处理用户注册
func (c *AuthController) Register(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	registerReq := new(dto.RegisterReq)
	if err := ctx.ShouldBindJSON(registerReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()))
		return
	}

	user, err := c.authService.Register(ctx, registerReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Register(err.Error())).ToResp())
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(user)).ToResp())
}

// Login 处理用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	loginReq := new(dto.LoginReq)
	if err := ctx.ShouldBindJSON(loginReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}

	token, err := c.authService.Login(ctx, loginReq.Username, loginReq.Password, loginReq.AutoLogin)
	if err != nil {
		ctx.JSON(http.StatusOK, wrapper.SetErrors(types.Login(err.Error())).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(map[string]interface{}{"token": token})).ToResp())
}

// RefreshToken 刷新令牌
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	refreshToken := new(dto.RefreshToken)
	if err := ctx.ShouldBindJSON(refreshToken); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}

	token, err := c.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		ctx.JSON(http.StatusOK, wrapper.SetErrors(types.Login(err.Error())).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(map[string]interface{}{"token": token})).ToResp())
}
