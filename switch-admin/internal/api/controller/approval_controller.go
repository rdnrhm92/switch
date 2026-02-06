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

// ApprovalController 审批相关的api
type ApprovalController struct {
	approvalService *service.ApprovalService
}

func NewApprovalController() *ApprovalController {
	return &ApprovalController{
		approvalService: service.NewApprovalService(),
	}
}

// GetMyRequests 获取我的审批
func (c *ApprovalController) GetMyRequests(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	uid := info.GetUserId(ctx)
	namespaceTag := info.GetSelectNamespace(ctx)
	var myRequestReqBody dto.MyRequestReqBody
	if err := ctx.ShouldBindJSON(&myRequestReqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}
	myRequestReqBody.UId = uid
	myRequestReqBody.NamespaceTag = namespaceTag
	requests := c.approvalService.GetRequestsByRequester(&myRequestReqBody)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(requests)).ToResp())
}

// GetAllRequests 获取所有审批
func (c *ApprovalController) GetAllRequests(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	uid := info.GetUserId(ctx)
	namespaceTag := info.GetSelectNamespace(ctx)
	var myRequestReqBody dto.MyRequestReqBody
	if err := ctx.ShouldBindJSON(&myRequestReqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}
	myRequestReqBody.UId = uid
	myRequestReqBody.NamespaceTag = namespaceTag
	requests := c.approvalService.GetAllRequests(&myRequestReqBody)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(requests)).ToResp())
}

// JudgeRequest 审批一个请求
func (c *ApprovalController) JudgeRequest(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	uName := info.GetUserName(ctx)
	uId := info.GetUserId(ctx)
	var approvalReqBody dto.SwitchApprovalReqBody
	if err := ctx.ShouldBindJSON(&approvalReqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}
	approvalReqBody.UName = uName
	approvalReqBody.UId = uId
	err := c.approvalService.ApproveRequest(&approvalReqBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy(err.Error())).ToResp())
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success("")).ToResp())
}

// ApplyToJoin 申请加入空间
func (c *ApprovalController) ApplyToJoin(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	user := info.GetUserInfo(ctx)
	if user == nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid user")).ToResp())
		return
	}
	var namespaceJoinApprovalReqBody dto.NamespaceJoinApprovalReqBody
	if err := ctx.ShouldBindJSON(&namespaceJoinApprovalReqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy("Invalid request body")).ToResp())
		return
	}
	namespaceJoinApprovalReqBody.UserName = user.Username
	namespaceJoinApprovalReqBody.UserId = user.ID
	err := c.approvalService.ApplyToJoin(&namespaceJoinApprovalReqBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy(err.Error())).ToResp())
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success("")).ToResp())
}
