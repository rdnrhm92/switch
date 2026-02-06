package dto

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
)

type SwitchApprovalReqBody struct {
	Id     uint   `json:"id"` //审批单子ID
	UId    uint   //审批人ID
	UName  string //审批人名字
	Notes  string `json:"notes"`  //审批意见
	Status uint   `json:"status"` //0-拒绝 1-同意
}

type NamespaceJoinApprovalReqBody struct {
	NamespaceTag string `json:"namespaceTag"` //空间Tag
	UserId       uint   //申请人ID
	UserName     string //申请人昵称
	ApproverIDs  []uint //审批人列表
}

type MyRequestReqBody struct {
	*PageLimit
	Invited        uint   `json:"invited"` //0-发起人 1-受邀人
	UId            uint   //当前用户ID
	NamespaceTag   string //当前空间Tag
	Status         uint   `json:"status"`         //1-PENDING 2-APPROVED 3-REJECTED
	ApprovableType uint   `json:"approvableType"` //1-namespace 2-switch
	ApproverUser   uint   `json:"approverUser"`   //审批人ID
	RequesterUser  uint   `json:"requesterUser"`  //发起人ID
}

type ApprovalDetailView struct {
	Approvable       interface{} `json:"approvable"`
	RequesterUserStr string      `json:"requesterUserStr"`
	ApproverUserStr  string      `json:"approverUserStr"`
	ApproverUsersStr string      `json:"approverUsersStr"`
	StatusStr        string      `json:"statusStr"`
	*admin_model.Approval
}
