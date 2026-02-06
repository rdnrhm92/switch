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

// EnvironmentController 环境相关的api
type EnvironmentController struct {
	envService *service.EnvironmentService
}

func NewEnvironmentController() *EnvironmentController {
	return &EnvironmentController{
		envService: service.NewEnvironmentService(),
	}
}

// GetEnvironmentsByNamespace 获取某个命名空间下的所有环境
func (c *EnvironmentController) GetEnvironmentsByNamespace(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	namespaceTag := ctx.Param("namespaceTag")
	environments, err := c.envService.GetEnvironmentsByNamespace(namespaceTag)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve environments"})
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(environments)).ToResp())
}

// GetAllEnvs 获取环境信息
func (c *EnvironmentController) GetAllEnvs(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var envListReq dto.EnvListReq
	if err := ctx.ShouldBindJSON(&envListReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	envListReq.ComputeLimit()
	envs := c.envService.GetEnvs(ctx, &envListReq)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(envs)).ToResp())
}

// CreateEnv 创建环境
func (c *EnvironmentController) CreateEnv(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.EnvCreateUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	req.CreateBy = info.GetUserName(ctx)

	if req.Name == "" || req.Tag == "" || req.Description == "" || len(req.Drivers) <= 0 || req.PublishOrder == 0 {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument("请检查参数是否为空")).ToResp())
		return
	}

	if err := c.envService.CreateUpdateEnvironment(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy(err)).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}

// UpdateEnv 编辑环境信息
func (c *EnvironmentController) UpdateEnv(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.EnvCreateUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	req.UpdateBy = info.GetUserName(ctx)

	if req.Name == "" || req.Description == "" || len(req.Drivers) <= 0 {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument("请检查参数是否为空")).ToResp())
		return
	}

	if err := c.envService.CreateUpdateEnvironment(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy(err)).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}

// PublishEnv 发布环境
func (c *EnvironmentController) PublishEnv(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.EnvPublishReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	req.UpdateBy = info.GetUserName(ctx)

	if req.Id == 0 {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument("请检查参数是否为空")).ToResp())
		return
	}

	if err := c.envService.EnvPublish(ctx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy(err)).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(nil)).ToResp())
}
