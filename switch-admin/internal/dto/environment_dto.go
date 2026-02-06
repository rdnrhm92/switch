package dto

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
)

// EnvCreateUpdateReq 创建/编辑环境的请求
type EnvCreateUpdateReq struct {
	Id              uint                     `json:"id"`
	Name            string                   `json:"name" binding:"required"`
	Tag             string                   `json:"tag" binding:"required"`
	Description     string                   `json:"description"`
	Drivers         []*CreateUpdateDriverReq `json:"drivers"`
	CreateBy        string                   `json:"createBy"`
	UpdateBy        string                   `json:"updateBy"`
	PublishOrder    int                      `json:"publish_order"`
	Publish         bool                     `json:"publish"`
	SelectNamespace string                   `json:"select_namespace"` //超级管理员
}

type EnvPublishReq struct {
	Id       uint   `json:"id"`
	UpdateBy string `json:"updateBy"`
}

// EnvWithNamespace 查询环境以及空间信息
type EnvWithNamespace struct {
	admin_model.Environment
	NamespaceInfo admin_model.Namespace `gorm:"embedded;embeddedPrefix:namespace_"`
}
