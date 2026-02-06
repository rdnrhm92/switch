package admin_model

import (
	"time"

	"gitee.com/fatzeng/switch-sdk-core/model"
)

// Approval 审批表
type Approval struct {
	model.CommonModel

	ApprovableType string `gorm:"size:100;not null;comment:关联业务类型" json:"approvableType"`
	Details        string `gorm:"type:json;comment:审批详情JSON" json:"details"`

	Status        string     `gorm:"size:20;not null;default:'PENDING';comment:状态 (PENDING, REJECTED, APPROVED)" json:"status"`
	RequesterUser uint       `gorm:"not null;comment:发起人ID" json:"requesterUser"`
	ApproverUsers string     `gorm:"type:text;comment:审批人ID列表的JSON字符串" json:"approverUsers"`
	ApproverUser  uint       `gorm:"not null;comment:实际审批人ID" json:"approverUser"`
	ApprovalNotes string     `gorm:"type:text;comment:审批意见" json:"approvalNotes"`
	ApprovalTime  *time.Time `gorm:"comment:审批时间" json:"approvalTime"`
	NamespaceTag  string     `gorm:"size:100;not null;index;comment:命名空间标签" json:"namespaceTag"`
	Namespace     *Namespace `gorm:"foreignKey:Tag;references:NamespaceTag" json:"namespace,omitempty"`
}

func (Approval) TableName() string {
	return "approval"
}

// SwitchApprovalForm 表示一个需要审批的开关变更单子
type SwitchApprovalForm struct {
	SwitchID uint   `json:"switchId"`
	EnvTag   string `json:"envTag"`
	Changes  string `json:"changes"`

	Switch      *model.SwitchModel `json:"switch"`
	Environment *Environment       `json:"environment"`
	Version     *model.Version     `json:"version"`
}

// NamespaceApprovalForm 用户入驻请求
type NamespaceApprovalForm struct {
	NamespaceTag string `json:"namespaceTag"`
}
