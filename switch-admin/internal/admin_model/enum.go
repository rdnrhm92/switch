package admin_model

const (
	// PENDING 开关在各个环境下的发布状态-未发布
	PENDING = "PENDING"
	// PUBLISHED 开关在各个环境下的发布状态-已发布
	PUBLISHED = "PUBLISHED"
)

const (
	// ApprovalStatusPending 审批状态 审批中
	ApprovalStatusPending = "PENDING"
	// ApprovalStatusApproved 审批状态 已通过
	ApprovalStatusApproved = "APPROVED"
	// ApprovalStatusRejected 审批状态 已拒绝
	ApprovalStatusRejected = "REJECTED"
)

const (
	NamespaceType = "namespace"
	SwitchType    = "switch"
)
