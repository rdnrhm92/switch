package admin_model

import "gitee.com/fatzeng/switch-sdk-core/model"

// SwitchApproval 审批表(作为因子信息的存在,记录开关都需要在什么环境下交由谁进行审批)
type SwitchApproval struct {
	model.CommonModel
	SwitchId      uint   `gorm:"not null;uniqueIndex:idx_switch_env,priority:1;comment:开关ID" json:"switchId"`
	EnvTag        string `gorm:"size:100;not null;uniqueIndex:idx_switch_env,priority:2;comment:环境标签" json:"envTag"`
	ApproverUsers string `gorm:"type:text;comment:审批人ID列表的JSON字符串" json:"approverUsers"`
}

func (SwitchApproval) TableName() string {
	return "switch_approvals"
}
