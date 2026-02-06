package dto

import "gitee.com/fatzeng/switch-admin/internal/admin_model"

// CreateUpdateNamespaceReq 定义了创建/编辑命名空间的 API 请求体
type CreateUpdateNamespaceReq struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"` //方便页面展示的命名空间名字 举例Name:淘搜 Tag:tb_search
	Tag         string `json:"tag"`  //真实用来区分命名空间的
	Description string `json:"description"`
	CreatedBy   string //创建人
	UpdatedBy   string //修改人
}

// NamespaceListReq 列表的查询
type NamespaceListReq struct {
	*PageLimit
	*PageTime
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	CreatedBy   string `json:"createdBy"`
}

func (n *NamespaceListReq) Initial() {
	if n.PageLimit != nil {
		n.PageLimit.ComputeLimit()
	}
}

// NamespaceListLikeReq 下拉查询
type NamespaceListLikeReq struct {
	Search string `json:"search"`
	All    bool   `json:"all"`
}

// NamespaceListLikeResp 下拉查询
type NamespaceListLikeResp struct {
	*admin_model.Namespace
	Status string `json:"status"`
}
