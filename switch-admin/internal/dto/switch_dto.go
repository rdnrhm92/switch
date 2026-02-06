package dto

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// CreateUpdateSwitchReq 创建修改开关的api请求
type CreateUpdateSwitchReq struct {
	Name         string `json:"name" binding:"required"` // 开关名字
	NamespaceTag string // 命名空间Tag
	SwitchId     uint   `json:"switchId"` // 开关ID

	Description              string                `json:"description"`              // 开关描述
	Rules                    *model.RuleNode       `json:"rules" binding:"required"` // 开关规则
	CreateSwitchApproversReq []*SwitchApproversReq `json:"createSwitchApproversReq"` // 开关审批人
	UseCache                 bool                  `json:"useCache"`                 // 开启缓存
}

// CreateSwitchFactorReq 创建因子信息
type CreateSwitchFactorReq struct {
	Id           uint   `json:"id"`                        // 因子ID
	Factor       string `json:"factor" binding:"required"` // 因子名称
	Description  string `json:"description"`               // 因子描述
	Name         string `json:"name"`                      // 因子显示名称
	JsonSchema   string `json:"jsonSchema"`                // 因子schema
	NamespaceTag string `json:"namespaceTag"`
}

// SwitchFactorListReq 列表操作
type SwitchFactorListReq struct {
	*PageLimit
	*PageTime
	Factor      string `json:"factor"`      // 因子名称
	Description string `json:"description"` // 因子描述
	Name        string `json:"name"`        // 因子显示名称
	CreatedBy   string `json:"createdBy"`
}

// SwitchFactorLikeReq 列表操作
type SwitchFactorLikeReq struct {
	Factor string `json:"factor"` // 因子名字
}

// SwitchListReq 列表操作
type SwitchListReq struct {
	*PageLimit
	Description string `json:"description"` // 开关描述
	Name        string `json:"name"`        // 开关名字
}

// SwitchQueryByTagReq 根据命名空间和环境标签查询开关
type SwitchQueryByTagReq struct {
	*PageLimit
	NamespaceID   uint   `json:"namespaceId" binding:"required"`   // 命名空间ID
	CurrentEnvTag string `json:"currentEnvTag" binding:"required"` // 当前环境标签
}

type SwitchApproversReq struct {
	EnvTag        string `json:"envTag"`
	ApproverUsers []uint `json:"approverUsers"`
}

type SubmitSwitchPushReq struct {
	SwitchID     uint   `json:"switchId" binding:"required"`
	TargetEnvTag string `json:"targetEnvTag" binding:"required"` //目标环境Tag
}

type SwitchListResp struct {
	*model.SwitchModel
	NextEnvInfo   *NextEnvInfo                `json:"nextEnvInfo"`   // 下一个环境的详细信息
	SwitchConfigs []*admin_model.SwitchConfig `json:"switchConfigs"` //所有的环境配置
}

type NextEnvInfo struct {
	EnvTag         string `json:"envTag"`         // 下一个环境标签
	EnvName        string `json:"envName"`        // 下一个环境名称
	ApprovalStatus string `json:"approvalStatus"` // PENDING(审批中), APPROVED(已通过), REJECTED(已拒绝), ""(无审批记录)
	ButtonText     string `json:"buttonText"`     // 按钮显示文本
	ButtonDisabled bool   `json:"buttonDisabled"` // 按钮是否禁用
}
