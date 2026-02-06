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

// NamespaceController 命名空间api
type NamespaceController struct {
	namespaceService *service.NamespaceService
}

func NewNamespaceController() *NamespaceController {
	return &NamespaceController{
		namespaceService: service.NewNamespaceService(),
	}
}

// GetAllNamespacesLike 模糊查询所有的空间用作下拉选择
func (c *NamespaceController) GetAllNamespacesLike(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var namespaceListLikeReq dto.NamespaceListLikeReq
	if err := ctx.ShouldBindJSON(&namespaceListLikeReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	namespaces := c.namespaceService.GetNamespacesLike(ctx, &namespaceListLikeReq)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(namespaces)).ToResp())
}

// GetAllNamespaces 获取所有的命名空间
func (c *NamespaceController) GetAllNamespaces(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var namespaceListReq dto.NamespaceListReq
	if err := ctx.ShouldBindJSON(&namespaceListReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	namespaceListReq.ComputeLimit()
	namespaces := c.namespaceService.GetNamespaces(ctx, &namespaceListReq)
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(namespaces)).ToResp())
}

// CreateNamespace 创建命名空间
func (c *NamespaceController) CreateNamespace(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var createNamespaceReq dto.CreateUpdateNamespaceReq
	if err := ctx.ShouldBindJSON(&createNamespaceReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	if createNamespaceReq.Name == "" || createNamespaceReq.Description == "" || createNamespaceReq.Tag == "" {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument("请检查 name tag description 的值是否为空")).ToResp())
		return
	}

	createNamespaceReq.CreatedBy = info.GetUserName(ctx)

	namespace, err := c.namespaceService.CreateUpdateNamespace(ctx, &createNamespaceReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy()).ToResp())
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(namespace)).ToResp())
}

// UpdateNamespace 编辑命名空间(只能编辑名称跟描述)
func (c *NamespaceController) UpdateNamespace(ctx *gin.Context) {
	respCtx := bc.NewRespContext(ctx)
	wrapper := bc.GetResp(respCtx)
	var updateNamespaceReq dto.CreateUpdateNamespaceReq
	if err := ctx.ShouldBindJSON(&updateNamespaceReq); err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument()).ToResp())
		return
	}
	if updateNamespaceReq.Name == "" || updateNamespaceReq.Description == "" {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.InvalidArgument("请检查 name description 的值是否为空")).ToResp())
		return
	}
	updateNamespaceReq.UpdatedBy = info.GetUserName(ctx)
	namespace, err := c.namespaceService.CreateUpdateNamespace(ctx, &updateNamespaceReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wrapper.SetErrors(types.Busy()).ToResp())
		return
	}
	ctx.JSON(http.StatusOK, wrapper.SetSuccess(types.Success(namespace)).ToResp())
}
