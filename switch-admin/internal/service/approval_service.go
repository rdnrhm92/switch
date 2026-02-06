package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-core/tool"
	"gorm.io/gorm"
)

// ApprovalService 提供了与审批相关的业务逻辑
type ApprovalService struct {
	approvalRepo            *repository.ApprovalFormRepository
	userRepo                *repository.UserRepository
	namespaceUserRoleRepo   *repository.NamespaceUserRoleRepository
	switchConfigRepo        *repository.SwitchConfigRepository
	envRepo                 *repository.EnvironmentRepository
	switchRepo              *repository.SwitchRepository
	switchService           *SwitchService
	namespaceService        *NamespaceService
	namespaceMembersService *NamespaceMembersService
}

func NewApprovalService() *ApprovalService {
	return &ApprovalService{
		approvalRepo:            repository.NewApprovalFormRepository(),
		userRepo:                repository.NewUserRepository(),
		namespaceUserRoleRepo:   repository.NewNamespaceUserRoleRepository(),
		switchConfigRepo:        repository.NewSwitchConfigRepository(),
		switchRepo:              repository.NewSwitchRepository(),
		envRepo:                 repository.NewEnvironmentRepository(),
		switchService:           NewSwitchService(),
		namespaceService:        NewNamespaceService(),
		namespaceMembersService: NewNamespaceMembersService(),
	}
}

// ApproveRequest 批准一个审批请求
func (s *ApprovalService) ApproveRequest(req *dto.SwitchApprovalReqBody) error {
	now := time.Now()
	return repository.GetDB().Transaction(func(tx *gorm.DB) error {
		txApprovalRepo := s.approvalRepo.WithTx(tx)
		form, err := txApprovalRepo.GetByID(req.Id)
		if err != nil {
			return fmt.Errorf("failed to get approval request: %w", err)
		}

		if form.Status != admin_model.ApprovalStatusPending {
			return fmt.Errorf("only pending requests can be approved")
		}
		if req.Status == 0 {
			form.Status = admin_model.ApprovalStatusRejected
		} else {
			form.Status = admin_model.ApprovalStatusApproved
		}
		form.ApprovalNotes = req.Notes
		form.ApproverUser = req.UId
		form.ApprovalTime = &now
		form.ApprovalTime = &now
		form.UpdateBy = tool.ToString(req.UName)

		if err = txApprovalRepo.Update(form); err != nil {
			return fmt.Errorf("failed to update approval form: %w", err)
		}

		switch form.ApprovableType {
		case admin_model.NamespaceType:
			//命名空间的审批
			var nsApproval = admin_model.NamespaceApprovalForm{}
			if err = json.Unmarshal([]byte(form.Details), &nsApproval); err != nil {
				return err
			}
			members := admin_model.NamespaceMembers{
				UserId:       form.RequesterUser,
				NamespaceTag: nsApproval.NamespaceTag,
			}
			if err = s.namespaceService.AddNamespaceMember(tx, &members); err != nil {
				return err
			}
			nsRoleTx := s.namespaceUserRoleRepo.WithTx(tx)
			return nsRoleTx.Create(&admin_model.NamespaceUserRole{
				NamespaceMembersId: members.ID,
				RoleId:             repository.OrdinaryRole.ID,
			})
		case admin_model.SwitchType:
			//开关的审批
			var switchApprovalForm = admin_model.SwitchApprovalForm{}
			if err = json.Unmarshal([]byte(form.Details), &switchApprovalForm); err != nil {
				return err
			}

			//获取规则配置
			var ruleNodes = model.RuleNode{}
			if err = json.Unmarshal([]byte(switchApprovalForm.Changes), &ruleNodes); err != nil {
				return err
			}

			switchModel := switchApprovalForm.Switch

			switchRepoTx := s.switchRepo.WithTx(tx)
			_, err = switchRepoTx.GetByIDAndVersion(switchModel.ID, switchModel.Version.Version)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 看看能不能根据这个版本号查到，如果查询不到说明开关在被审批的这个时间内又修改了，记录一个日志直接返回就行了
					logger.Logger.Warnf("Switch (ID: %d, Version: %d) not found during approval, it may have been modified during the approval process", switchModel.ID, switchModel.Version)
					return nil
				}
				return err
			}

			// 设置开关主体的审核状态
			if err = switchRepoTx.UpdateSwitchApproval(switchModel.ID, form.Status); err != nil {
				return err
			}

			return nil
		default:
			return fmt.Errorf("unknown approval type")
		}
	})
}

// ApplyToJoin 加入空间
func (s *ApprovalService) ApplyToJoin(req *dto.NamespaceJoinApprovalReqBody) error {
	now := time.Now()
	return repository.GetDB().Transaction(func(tx *gorm.DB) error {
		txApprovalRepo := s.approvalRepo.WithTx(tx)

		approverUsers, err := s.namespaceMembersService.FindApprovePermissionsByNamespaceTag(req.NamespaceTag)
		if err != nil || len(approverUsers) == 0 {
			return fmt.Errorf("无法创建申请：空间 %d 没有找到审批人", req.NamespaceTag)
		}

		nsDetail, err := json.Marshal(&admin_model.NamespaceApprovalForm{NamespaceTag: req.NamespaceTag})
		if err != nil {
			return fmt.Errorf("序列化审批详情失败: %w", err)
		}
		approverUsersJSON, err := json.Marshal(approverUsers)
		if err != nil {
			return fmt.Errorf("序列化审批人列表失败: %w", err)
		}

		approval := &admin_model.Approval{
			CommonModel: model.CommonModel{
				CreatedBy:  req.UserName,
				CreateTime: &now,
			},
			ApprovableType: admin_model.NamespaceType,
			Details:        string(nsDetail),
			Status:         admin_model.ApprovalStatusPending,
			RequesterUser:  req.UserId,
			ApproverUsers:  string(approverUsersJSON),
			NamespaceTag:   req.NamespaceTag,
		}

		return txApprovalRepo.Create(approval)
	})
}

// buildUserMapping 构建用户ID到用户名的映射
func (s *ApprovalService) buildUserMapping(approvals []*admin_model.Approval) (map[uint]string, map[uint][]uint) {
	allUserIDs := make([]uint, 0)
	approverLists := make(map[uint][]uint)

	for _, approval := range approvals {
		allUserIDs = append(allUserIDs, approval.RequesterUser)
		allUserIDs = append(allUserIDs, approval.ApproverUser)

		var approverUsers []uint
		if err := json.Unmarshal([]byte(approval.ApproverUsers), &approverUsers); err != nil {
			logger.Logger.Warnf("failed to process ApproverUsers for approval ID %d, err: %v", approval.ID, err)
			approverLists[approval.ID] = []uint{}
			continue
		}
		allUserIDs = append(allUserIDs, approverUsers...)
		approverLists[approval.ID] = approverUsers
	}

	userMapping := make(map[uint]string)
	if len(allUserIDs) > 0 {
		uniqueUserIDs := tool.UniqueSlice(allUserIDs)
		userInfos, err := s.userRepo.FindByUserIds(uniqueUserIDs)
		if err != nil {
			logger.Logger.Warnf("failed to fetch all users: %v", err)
		} else {
			for _, uInfo := range userInfos {
				userMapping[uInfo.ID] = uInfo.Username
			}
		}
	}

	return userMapping, approverLists
}

// buildApprovalView 构建单个审批的视图对象
func (s *ApprovalService) buildApprovalView(approval *admin_model.Approval, userMapping map[uint]string, approverLists map[uint][]uint) *dto.ApprovalDetailView {
	view := &dto.ApprovalDetailView{
		Approval: approval,
	}

	// 处理审批类型
	switch approval.ApprovableType {
	case admin_model.SwitchType:
		var switchForm admin_model.SwitchApprovalForm
		if err := json.Unmarshal([]byte(approval.Details), &switchForm); err != nil {
			logger.Logger.Errorf("failed to unmarshal switch approval details for approval ID %d: %v", approval.ID, err)
			return nil
		}
		view.Approvable = switchForm
		view.ApprovableType = "开关变更"
	case admin_model.NamespaceType:
		var namespaceForm admin_model.NamespaceApprovalForm
		if err := json.Unmarshal([]byte(approval.Details), &namespaceForm); err != nil {
			logger.Logger.Errorf("failed to unmarshal namespace approval details for approval ID %d: %v", approval.ID, err)
			return nil
		}
		view.Approvable = namespaceForm
		view.ApprovableType = "命名空间变更"
	default:
		logger.Logger.Warnf("unknown approvable type '%s' for approval ID %d", approval.ApprovableType, approval.ID)
	}

	// 处理状态
	switch approval.Status {
	case admin_model.ApprovalStatusPending:
		view.StatusStr = "审批中"
	case admin_model.ApprovalStatusApproved:
		view.StatusStr = "已通过"
	case admin_model.ApprovalStatusRejected:
		view.StatusStr = "已拒绝"
	default:
		logger.Logger.Warnf("unknown approval status '%s' for approval ID %d", approval.ApprovableType, approval.ID)
	}

	// 设置用户名
	if userName, ok := userMapping[approval.RequesterUser]; ok {
		view.RequesterUserStr = userName
	} else {
		view.RequesterUserStr = ""
	}

	if userName, ok := userMapping[approval.ApproverUser]; ok {
		view.ApproverUserStr = userName
	} else {
		view.ApproverUserStr = ""
	}

	// 设置审批人列表
	if approverIDs, ok := approverLists[approval.ID]; ok && len(approverIDs) > 0 {
		approverNames := make([]string, 0, len(approverIDs))
		for _, userID := range approverIDs {
			if userName, ok := userMapping[userID]; ok {
				approverNames = append(approverNames, userName)
			}
		}
		view.ApproverUsersStr = fmt.Sprintf("[%s]", strings.Join(approverNames, ","))
	}

	return view
}

// processApprovals 处理审批列表的公共逻辑
func (s *ApprovalService) processApprovals(approvals []*admin_model.Approval, total int64) map[string]interface{} {
	userMapping, approverLists := s.buildUserMapping(approvals)

	views := make([]*dto.ApprovalDetailView, 0, len(approvals))
	for _, approval := range approvals {
		if view := s.buildApprovalView(approval, userMapping, approverLists); view != nil {
			views = append(views, view)
		}
	}

	return map[string]interface{}{
		"success": true,
		"total":   total,
		"data":    views,
	}
}

// GetRequestsByRequester 获取所有单子(区分发起人跟受邀人)
func (s *ApprovalService) GetRequestsByRequester(req *dto.MyRequestReqBody) map[string]interface{} {
	approvals, total, err := s.approvalRepo.Find(req)
	if err != nil {
		logger.Logger.Error("failed to fetch approval forms: %w", err)
		return map[string]interface{}{"success": false, "data": []interface{}{}, "total": 0}
	}

	return s.processApprovals(approvals, total)
}

// GetAllRequests 获取所有单子(不区分发起人跟受邀人)
func (s *ApprovalService) GetAllRequests(req *dto.MyRequestReqBody) map[string]interface{} {
	approvals, total, err := s.approvalRepo.FindAll(req)
	if err != nil {
		logger.Logger.Error("failed to fetch all approval forms: %w", err)
		return map[string]interface{}{"success": false, "data": []interface{}{}, "total": 0}
	}

	return s.processApprovals(approvals, total)
}
