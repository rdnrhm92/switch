package controller

import (
	"net/http"
	"strconv"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/service"
	"gitee.com/fatzeng/switch-admin/internal/types"
	"gitee.com/fatzeng/switch-components/bc"
	"github.com/gin-gonic/gin"
)

// SwitchController 处理开关相关的请求
type SwitchController struct {
	switchService *service.SwitchService
}

func NewSwitchController() *SwitchController {
	return &SwitchController{
		switchService: service.NewSwitchService(),
	}
}

// SwitchList 开关列表
func (c *SwitchController) SwitchList(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.SwitchListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	if req.PageLimit != nil {
		req.ComputeLimit()
	}
	result := c.switchService.SwitchList(ctx, &req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(result)).ToResp())
}

// CreateSwitch 创建开关
func (c *SwitchController) CreateSwitch(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	createSwitchReq := new(dto.CreateUpdateSwitchReq)
	if err := ctx.ShouldBindJSON(createSwitchReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}
	createSwitchReq.NamespaceTag = info.GetSelectNamespace(ctx)
	switchInstance, err := c.switchService.CreateSwitch(ctx, createSwitchReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy(err.Error())).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(switchInstance)).ToResp())
}

// GetSwitchDetails 获取开关详情
func (c *SwitchController) GetSwitchDetails(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	switchID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	details, err := c.switchService.GetSwitchDetails(uint(switchID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, wrapper.SetErrors(types.Busy(err.Error())).ToResp())
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(details)).ToResp())
}

// UpdateSwitch 提交开关修改
func (c *SwitchController) UpdateSwitch(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var updateSwitchReq dto.CreateUpdateSwitchReq
	if err := ctx.ShouldBindJSON(&updateSwitchReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}

	updateSwitchReq.NamespaceTag = info.GetSelectNamespace(ctx)

	switchInstance, err := c.switchService.CreateSwitch(ctx, &updateSwitchReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy(err.Error())).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(switchInstance)).ToResp())
}

// SubmitSwitchPush 开关推送
func (c *SwitchController) SubmitSwitchPush(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var submitSwitchPushReq dto.SubmitSwitchPushReq
	if err := ctx.ShouldBindJSON(&submitSwitchPushReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}

	result, err := c.switchService.PushSwitchChange(ctx, &submitSwitchPushReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy(err.Error())).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(result)).ToResp())
}

// CreateSwitchFactor 因子信息新增
func (c *SwitchController) CreateSwitchFactor(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	switchFactorReq := new(dto.CreateSwitchFactorReq)
	if err := ctx.ShouldBindJSON(switchFactorReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}

	switchFactorInstance, err := c.switchService.CreateUpdateSwitchFactor(ctx, switchFactorReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to create factor meta")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(switchFactorInstance)).ToResp())
}

// UpdateSwitchFactor 因子信息修改
func (c *SwitchController) UpdateSwitchFactor(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	switchFactorReq := new(dto.CreateSwitchFactorReq)
	if err := ctx.ShouldBindJSON(switchFactorReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument(err.Error())).ToResp())
		return
	}

	switchFactorInstance, err := c.switchService.CreateUpdateSwitchFactor(ctx, switchFactorReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wrapper.SetErrors(types.Busy("Failed to update factor meta")).ToResp())
		return
	}

	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(switchFactorInstance)).ToResp())
}

// MetaSwitchList 因子信息列表
func (c *SwitchController) MetaSwitchList(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.SwitchFactorListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	if req.PageLimit != nil {
		req.ComputeLimit()
	}
	result := c.switchService.SwitchFactorList(ctx, &req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(result)).ToResp())
}

// MetaSwitchLike 因子信息列表(全部)
func (c *SwitchController) MetaSwitchLike(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var req dto.SwitchFactorLikeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	result := c.switchService.SwitchFactorLike(ctx, &req)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(result)).ToResp())
}
