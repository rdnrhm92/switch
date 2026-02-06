package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/notifier"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gorm.io/gorm"
)

// SwitchDetailsResponse 定义了开关详情的响应体
type SwitchDetailsResponse struct {
	SwitchModel *model.SwitchModel            `json:"factor"`
	Configs     []*admin_model.SwitchConfig   `json:"configs"`
	Approvals   []*admin_model.SwitchApproval `json:"approvals"`
}

// SwitchService 提供了与开关相关的业务逻辑
type SwitchService struct {
	switchRepo                   *repository.SwitchRepository
	switchConfigRepo             *repository.SwitchConfigRepository
	envRepo                      *repository.EnvironmentRepository
	switchApprovalRepo           *repository.SwitchApprovalRepository
	approvalFormRepo             *repository.ApprovalFormRepository
	switchSnapshotRepositoryRepo *repository.SwitchSnapshotConfigRepository
	switchFactorRepositoryRepo   *repository.SwitchFactorRepository
	namespaceMembersService      *NamespaceMembersService
}

// NewSwitchService 创建一个新的 SwitchService
func NewSwitchService() *SwitchService {
	return &SwitchService{
		switchRepo:                   repository.NewSwitchRepository(),
		switchConfigRepo:             repository.NewSwitchConfigRepository(),
		envRepo:                      repository.NewEnvironmentRepository(),
		switchApprovalRepo:           repository.NewSwitchApprovalRepository(),
		approvalFormRepo:             repository.NewApprovalFormRepository(),
		switchSnapshotRepositoryRepo: repository.NewSwitchSnapshotConfigRepository(),
		switchFactorRepositoryRepo:   repository.NewSwitchFactorRepository(),
		namespaceMembersService:      NewNamespaceMembersService(),
	}
}

// SwitchList 开关列表
func (s *SwitchService) SwitchList(ctx *gin.Context, req *dto.SwitchListReq) map[string]interface{} {
	batchGroup := errgroup.Group{}
	var allRes []*dto.SwitchListResp
	batchGroup.Go(func() error {
		namespaceTag := info.GetSelectNamespace(ctx)
		envs, err := s.envRepo.GetByNamespaceTag(namespaceTag)
		if err != nil {
			return err
		}
		if len(envs) == 0 {
			return fmt.Errorf("namespace envs %v not found", namespaceTag)
		}

		// 按发布顺序排序环境
		sort.Slice(envs, func(i, j int) bool {
			return envs[i].PublishOrder < envs[j].PublishOrder
		})

		all, err := s.switchRepo.FindList(ctx, req)
		if err != nil {
			return err
		}

		switchIDs := make([]uint, 0, len(all))
		for _, switchCombination := range all {
			switchIDs = append(switchIDs, switchCombination.ID)
		}

		configsMap, err := s.switchConfigRepo.GetBatchLatestConfigsBySwitchIDs(switchIDs)
		if err != nil {
			return err
		}

		for _, switchCombination := range all {
			//计算下一个环境信息
			nextEnvInfo := s.calculateNextEnvInfo(switchCombination, envs)
			resp := &dto.SwitchListResp{
				SwitchModel: switchCombination,
				NextEnvInfo: nextEnvInfo,
			}
			if configs, ok := configsMap[switchCombination.ID]; ok {
				resp.SwitchConfigs = configs
			} else {
				resp.SwitchConfigs = make([]*admin_model.SwitchConfig, 0)
			}
			allRes = append(allRes, resp)
		}

		return nil
	})

	var countRes int64
	batchGroup.Go(func() error {
		count, err := s.switchRepo.CountAll(ctx, req)
		if err != nil {
			return err
		}
		countRes = count
		return nil
	})

	err := batchGroup.Wait()
	if err != nil {
		logger.Logger.Errorf("[SwitchService] batch err: %v", err)
		return map[string]interface{}{
			"success": false,
			"total":   0,
			"data":    []interface{}{},
		}
	}

	return map[string]interface{}{
		"success": true,
		"total":   countRes,
		"data":    allRes,
	}
}

// CreateSwitch 创建一个新开关，并为该开关所属命名空间下的环境创建配置
func (s *SwitchService) CreateSwitch(ctx *gin.Context, req *dto.CreateUpdateSwitchReq) (*model.SwitchModel, error) {
	// 创建跟修改前如果设置了审批人，要看审批人是否有审批权限
	if len(req.CreateSwitchApproversReq) != 0 {
		approverUsers, err := s.namespaceMembersService.FindApprovePermissionsByNamespaceTag(req.NamespaceTag)
		if err != nil || len(approverUsers) == 0 {
			return nil, fmt.Errorf("无法创建/修改开关：空间 %d 没有找到审批人", req.NamespaceTag)
		}
		userApproverMap := make(map[uint]struct{}, len(approverUsers))
		for _, userId := range approverUsers {
			userApproverMap[userId] = struct{}{}
		}
		for _, approversReq := range req.CreateSwitchApproversReq {
			for _, userId := range approversReq.ApproverUsers {
				if _, ok := userApproverMap[userId]; !ok {
					// 不存在 疑似在提交这个过程中，审批人信息发生变更 属于极端情况
					return nil, fmt.Errorf("无法创建/修改开关：空间 %d 下审批人疑似变更,请刷新重试", req.NamespaceTag)
				}
			}
		}
	}

	uName := info.GetUserName(ctx)
	now := time.Now()

	if req.SwitchId != 0 {
		sw, err := s.switchRepo.GetSwitchByID(req.SwitchId)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain factor: %w", err)
		}

		//变更当前环境
		envs, err := s.envRepo.GetByNamespaceTag(sw.NamespaceTag)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the environment list under namespace: %w", err)
		}
		if len(envs) == 0 {
			return nil, fmt.Errorf("there are no environments in this namespace, so switches cannot be update")
		}

		if req.Description != "" {
			sw.Description = req.Description
		}
		if req.Rules != nil {
			sw.Rules = req.Rules
		}
		//开关更新后从头开始
		sw.CurrentEnvTag = ""
		sw.UpdateBy = uName
		sw.UpdateTime = &now
		sw.UseCache = req.UseCache
		sw.ApproverStatus = ""
		sw.Version = model.Version{
			Version: sw.Version.Version + 1,
		}

		//审批人的删除
		//审批人的变更
		//审批单(旧单不变增新单)

		// 只有开关创建者才能修改审批人 需要check
		var sas []*admin_model.SwitchApproval
		if sw.CreatedBy != uName {
			for _, sar := range req.CreateSwitchApproversReq {
				approverUsers, err := json.Marshal(&sar.ApproverUsers)
				if err != nil {
					return nil, err
				}
				sas = append(sas, &admin_model.SwitchApproval{
					SwitchId:      req.SwitchId,
					EnvTag:        sar.EnvTag,
					ApproverUsers: string(approverUsers),
					CommonModel: model.CommonModel{
						UpdateBy:   uName,
						UpdateTime: &now,
					},
				})
			}
		}

		if err = repository.GetDB().Transaction(func(tx *gorm.DB) error {
			switchRepo := s.switchRepo.WithTx(tx)
			switchApprovalRepo := s.switchApprovalRepo.WithTx(tx)
			if err = switchRepo.Save(sw); err != nil {
				return fmt.Errorf("failed to save factor: %w", err)
			}
			if err = switchApprovalRepo.DeleteBySwitchId(req.SwitchId); err != nil {
				return err
			}

			if len(sas) > 0 {
				if err = switchApprovalRepo.Create(sas); err != nil {
					return fmt.Errorf("failed to save factor: %w", err)
				}
			}

			// 开关修改后 已经存在的审批单不用管了，当审批的时候会判断version 如果不一致 审批单自动失效 属于惰性处理

			return nil
		}); err != nil {
			return nil, err
		}
		return sw, nil
	} else {
		//开关的创建必须有关联环境
		envs, err := s.envRepo.GetByNamespaceTag(req.NamespaceTag)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the environment list under namespace: %w", err)
		}
		if len(envs) == 0 {
			return nil, fmt.Errorf("there are no environments in this namespace, so switches cannot be created")
		}

		//创建开关模型
		sw := &model.SwitchModel{
			NamespaceTag:  req.NamespaceTag,
			Name:          req.Name,
			CurrentEnvTag: "",
			Rules:         req.Rules,
			Description:   req.Description,
			UseCache:      req.UseCache,
			CommonModel: model.CommonModel{
				CreatedBy:  uName,
				UpdateBy:   uName,
				CreateTime: &now,
			},
			Version: model.Version{
				Version: 1,
			},
		}

		// 创建开关在各个环境下的配置
		// 如果将来环境发生了变更SwitchConfig将会在开关推送的时候自动维护
		// 修改不做维护提升执行效率
		var configs []*admin_model.SwitchConfig
		for _, env := range envs {
			configs = append(configs, &admin_model.SwitchConfig{
				Version: model.Version{
					Version: 1,
				},
				EnvTag:      env.Tag,
				ConfigValue: req.Rules,
				Status:      admin_model.PENDING,
			})
		}

		//创建开关在各个环境下的审批人
		var sas []*admin_model.SwitchApproval
		for _, approversReq := range req.CreateSwitchApproversReq {
			approverUsers, err := json.Marshal(&approversReq.ApproverUsers)
			if err != nil {
				return nil, err
			}
			sas = append(sas, &admin_model.SwitchApproval{
				EnvTag:        approversReq.EnvTag,
				ApproverUsers: string(approverUsers),
				CommonModel: model.CommonModel{
					CreatedBy:  uName,
					CreateTime: &now,
				},
			})
		}

		if err = repository.GetDB().Transaction(func(tx *gorm.DB) error {
			if err = tx.Create(sw).Error; err != nil {
				return err
			}

			if configs != nil && len(configs) > 0 {
				for _, config := range configs {
					config.SwitchID = sw.ID
				}
				if err = tx.Create(&configs).Error; err != nil {
					return err
				}
			}

			if sas != nil && len(sas) > 0 {
				for _, sa := range sas {
					sa.SwitchId = sw.ID
				}
				if err = tx.Create(&sas).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return sw, nil
	}
}

// calculateNextEnvInfo 计算下一个环境的信息
func (s *SwitchService) calculateNextEnvInfo(switchModel *model.SwitchModel, envs []*admin_model.Environment) *dto.NextEnvInfo {
	if len(envs) == 0 {
		return nil
	}

	// 计算下一个环境的索引
	nextIndex := s.calculateNextEnvIndex(switchModel.CurrentEnvTag, envs)
	if nextIndex == -1 {
		// 没有下一个环境
		return nil
	}

	// 获取下一个环境
	nextEnv := envs[nextIndex]

	// 构建 NextEnvInfo
	nextEnvInfo := &dto.NextEnvInfo{
		EnvTag:  nextEnv.Tag,
		EnvName: nextEnv.Name,
	}

	// 有配置记录，根据状态设置按钮
	nextEnvInfo.ApprovalStatus = switchModel.ApproverStatus

	switch nextEnvInfo.ApprovalStatus {
	case "PENDING":
		nextEnvInfo.ButtonText = nextEnv.Tag + "审批中"
		nextEnvInfo.ButtonDisabled = true
	case "APPROVED":
		nextEnvInfo.ButtonText = nextEnv.Tag + "已通过"
		nextEnvInfo.ButtonDisabled = false
	case "REJECTED":
		nextEnvInfo.ButtonText = nextEnv.Tag + "已拒绝"
		nextEnvInfo.ButtonDisabled = false
	default:
		nextEnvInfo.ButtonText = nextEnv.Tag
		nextEnvInfo.ButtonDisabled = false
	}

	return nextEnvInfo
}

// validateJsonSchema 校验 JSON Schema 格式
func validateJsonSchema(jsonSchemaStr string) error {
	if strings.TrimSpace(jsonSchemaStr) == "" {
		return errors.New("JSON Schema 不能为空")
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(jsonSchemaStr), &schema); err != nil {
		return fmt.Errorf("JSON Schema 格式错误: %w", err)
	}

	if _, exists := schema["type"]; !exists {
		return errors.New("JSON Schema 必须包含 type 字段")
	}

	typeValue, ok := schema["type"].(string)
	if !ok {
		return errors.New("JSON Schema 的 type 字段必须是字符串")
	}

	validTypes := []string{"string", "number", "integer", "boolean", "object", "array", "null"}
	isValidType := false
	for _, validType := range validTypes {
		if typeValue == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("JSON Schema 的 type 字段值无效: %s", typeValue)
	}

	return nil
}

// CreateUpdateSwitchFactor 因子信息的新建和更新
func (s *SwitchService) CreateUpdateSwitchFactor(ctx *gin.Context, req *dto.CreateSwitchFactorReq) (*admin_model.SwitchFactor, error) {
	// 校验 JSON Schema 格式
	if err := validateJsonSchema(req.JsonSchema); err != nil {
		return nil, fmt.Errorf("JSON Schema 校验失败: %w", err)
	}
	userInfo := info.GetUserInfo(ctx)
	if userInfo == nil {
		return nil, fmt.Errorf("user info not found")
	}
	selectNamespace := req.NamespaceTag
	if selectNamespace == "" {
		selectNamespace = userInfo.SelectNamespace
	}
	now := time.Now()
	if req.Id == 0 {
		//新增
		sm := admin_model.SwitchFactor{
			Factor:       req.Factor,
			Description:  req.Description,
			NamespaceTag: selectNamespace,
			JsonSchema:   req.JsonSchema,
			Name:         req.Name,
			CommonModel: model.CommonModel{
				CreatedBy:  userInfo.Username,
				CreateTime: &now,
			},
		}
		metaOne, err := s.switchFactorRepositoryRepo.InsertOne(&sm)
		if err != nil {
			return nil, fmt.Errorf("failed to create SwitchFactor: %w", err)
		}
		return metaOne, nil
	}

	//更新
	return s.switchFactorRepositoryRepo.Update(req, userInfo.Username)
}

// SwitchFactorList 因子信息列表
func (s *SwitchService) SwitchFactorList(ctx *gin.Context, req *dto.SwitchFactorListReq) map[string]interface{} {
	batchGroup := errgroup.Group{}
	var allRes []*admin_model.SwitchFactor
	batchGroup.Go(func() error {
		all, err := s.switchFactorRepositoryRepo.FindList(ctx, req)
		if err != nil {
			return err
		}
		allRes = all
		return nil
	})

	var countRes int64
	batchGroup.Go(func() error {
		count, err := s.switchFactorRepositoryRepo.CountAll(req)
		if err != nil {
			return err
		}
		countRes = count
		return nil
	})

	err := batchGroup.Wait()
	if err != nil {
		logger.Logger.Errorf("[SwitchService] batch err: %v", err)
		return map[string]interface{}{
			"success": false,
			"total":   0,
			"data":    []interface{}{},
		}
	}

	return map[string]interface{}{
		"success": true,
		"total":   countRes,
		"data":    allRes,
	}
}

// SwitchFactorLike 因子信息列表(全部)
func (s *SwitchService) SwitchFactorLike(ctx *gin.Context, req *dto.SwitchFactorLikeReq) []*admin_model.SwitchFactor {
	var allRes = make([]*admin_model.SwitchFactor, 0)
	allRes, err := s.switchFactorRepositoryRepo.FindListLike(ctx, req)
	if err != nil {
		return allRes
	}

	return allRes
}

// GetSwitchDetails 获取开关的详细信息，包括其在所有环境下的配置
func (s *SwitchService) GetSwitchDetails(switchID uint) (*SwitchDetailsResponse, error) {
	sw, err := s.switchRepo.GetSwitchByID(switchID)
	if err != nil {
		return nil, err
	}

	configs, err := s.switchConfigRepo.GetBySwitchID(switchID)
	if err != nil {
		return nil, err
	}

	approvals, err := s.switchApprovalRepo.FindBySwitch(sw.ID)
	if err != nil {
		return nil, err
	}

	return &SwitchDetailsResponse{
		SwitchModel: sw,
		Configs:     configs,
		Approvals:   approvals,
	}, nil
}

func (s *SwitchService) PushSwitchChange(ctx *gin.Context, req *dto.SubmitSwitchPushReq) (string, error) {
	now := time.Now()

	switchInstance, err := s.switchRepo.GetSwitchByID(req.SwitchID)
	if err != nil {
		return "", fmt.Errorf("failed to find factor with ID %d: %w", req.SwitchID, err)
	}
	if switchInstance == nil {
		return "", fmt.Errorf("failed to find factor with ID %d", req.SwitchID)
	}

	envs, err := s.envRepo.GetByNamespaceTag(switchInstance.NamespaceTag)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve the environment list under namespace: %w", err)
	}
	if len(envs) == 0 {
		return "", fmt.Errorf("there are no environments in this namespace, so switches cannot be update")
	}

	//要推送的环境详细信息
	var env *admin_model.Environment
	for _, environment := range envs {
		if environment.Tag == req.TargetEnvTag {
			env = environment
			break
		}
	}

	// 操作开关过程中，环境发生改变，需要刷新重试(计算要推送的环境是否还存在)
	if env == nil {
		return "", fmt.Errorf("failed to find environment with Tag %v: %w", req.TargetEnvTag, err)
	}

	// 根据开关当前最新态，计算下一个环境标签
	nextEnvTag := s.getNextEnvTag(switchInstance.CurrentEnvTag, envs)
	if nextEnvTag == switchInstance.CurrentEnvTag {
		// 属于无效推送 不响应错误 也不执行后续逻辑
		return "推送成功", nil
	}

	// 比如两个页面，a页面长时间不刷新，b页面已经推送了某开关，此时再操作a就会有问题(计算要推送的环境是否跟开关的下一个环境一致)
	if nextEnvTag != env.Tag {
		return "请刷新重试", fmt.Errorf("need refresh")
	}

	//查询开关配置-判断审批状态
	switchConfig, err := s.switchConfigRepo.GetBySwitchIdEnvTagVersion(req.SwitchID, env.Tag)
	if err != nil {
		// 要推送到某某某环境但是环境对应的switch config不存在，可能是在开关创建后新增了环境导致的，手动维护switch config
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config := &admin_model.SwitchConfig{
				Version: model.Version{
					Version: 1,
				},
				SwitchID:    req.SwitchID,
				EnvTag:      env.Tag,
				ConfigValue: switchInstance.Rules,
				Status:      admin_model.PENDING,
			}
			// 不需要事务
			err = s.switchConfigRepo.Create(config)
			if err != nil {
				return "", err
			}
			switchConfig = config
		} else {
			return "", err
		}
	}

	// 根据审批状态处理
	switch switchInstance.ApproverStatus {
	case admin_model.ApprovalStatusPending:
		return "", fmt.Errorf("审批中-不可重复提交")
	case admin_model.ApprovalStatusApproved:
		// 审批通过走推送逻辑 - 直接推送最新内容
		return "推送成功", s.executeInitialPush(ctx, switchInstance, switchConfig, env, now)
	case admin_model.ApprovalStatusRejected:
		// 审批拒绝-重新发起审批单
		return s.handleApprovalFlow(ctx, switchInstance, switchConfig, env, now, true)
	}

	// 默认情况：检查要推送的环境是否需要审批
	return s.handleApprovalFlow(ctx, switchInstance, switchConfig, env, now, false)
}

// handleApprovalFlow 处理审批流程的通用逻辑
func (s *SwitchService) handleApprovalFlow(ctx *gin.Context, switchInstance *model.SwitchModel, switchConfig *admin_model.SwitchConfig, env *admin_model.Environment, now time.Time, isRejectedResubmit bool) (string, error) {
	approvalRule, err := s.switchApprovalRepo.FindBySwitchAndEnv(switchInstance.ID, env.Tag)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有审批规则，直接推送
			return "推送成功", s.executeInitialPush(ctx, switchInstance, switchConfig, env, now)
		}
		return "", err
	}

	// 需要审批
	if len(approvalRule.ApproverUsers) > 0 {
		if isRejectedResubmit {
			return "已重新提交审批，等待审核人审批", s.executeRejectedResubmit(ctx, approvalRule, switchInstance, env, now)
		}

		return "已提交审批，等待审核人审批", s.executeRejectedResubmit(ctx, approvalRule, switchInstance, env, now)
	}

	// 没有审批人，直接推送
	return "推送成功", s.executeInitialPush(ctx, switchInstance, switchConfig, env, now)
}

// executeRejectedResubmit 执行审批拒绝后的重新提交逻辑
func (s *SwitchService) executeRejectedResubmit(ctx context.Context, approvalRule *admin_model.SwitchApproval, switchInstance *model.SwitchModel, env *admin_model.Environment, now time.Time) error {
	userInfo := info.GetUserInfo(ctx)
	if userInfo == nil {
		return fmt.Errorf("user info is nil")
	}
	changes, err := json.Marshal(switchInstance.Rules)
	if err != nil {
		return err
	}

	form := admin_model.SwitchApprovalForm{
		SwitchID: approvalRule.SwitchId,
		EnvTag:   approvalRule.EnvTag,
		Changes:  string(changes),
		Switch:   switchInstance,
		Version: &model.Version{
			Version: switchInstance.Version.Version,
		},
		Environment: env,
	}
	details, err := json.Marshal(&form)
	if err != nil {
		return err
	}

	return repository.GetDB().Transaction(func(tx *gorm.DB) error {
		// 创建审批单
		if len(approvalRule.ApproverUsers) > 0 {
			approvalFormRepoTx := s.approvalFormRepo.WithTx(tx)
			forms := &admin_model.Approval{
				CommonModel: model.CommonModel{
					CreatedBy:  userInfo.Username,
					CreateTime: &now,
				},
				ApprovableType: admin_model.SwitchType,
				Details:        string(details),
				Status:         admin_model.ApprovalStatusPending,
				RequesterUser:  userInfo.ID,
				ApproverUsers:  approvalRule.ApproverUsers,
				NamespaceTag:   switchInstance.NamespaceTag,
			}
			if err = approvalFormRepoTx.Create(forms); err != nil {
				return fmt.Errorf("failed to create approval form: %w", err)
			}
		}

		// switch_config什么都不变

		// switch主体修改审批状态
		switchInstance.ApproverStatus = admin_model.ApprovalStatusPending
		switchRepoTx := s.switchRepo.WithTx(tx)
		// 设置开关主体的审核状态
		if err = switchRepoTx.UpdateSwitchApproval(switchInstance.ID, switchInstance.ApproverStatus); err != nil {
			return err
		}
		return nil
	})
}

// executeInitialPush 执行初始推送逻辑
func (s *SwitchService) executeInitialPush(ctx context.Context, switchInstance *model.SwitchModel, switchConfig *admin_model.SwitchConfig, env *admin_model.Environment, now time.Time) error {
	userInfo := info.GetUserInfo(ctx)
	if userInfo == nil {
		return fmt.Errorf("can not get user info")
	}

	// 更新开关的当前环境标签
	switchConfig.EnvTag = env.Tag
	switchInstance.CurrentEnvTag = env.Tag

	// 构建config
	switchConfig.CommonModel.UpdateTime = &now
	switchConfig.CommonModel.UpdateBy = userInfo.Username
	switchConfig.ConfigValue = switchInstance.Rules
	switchConfig.Status = admin_model.PUBLISHED
	switchConfig.Version = switchInstance.Version

	changes, err := json.Marshal(switchInstance)
	if err != nil {
		return err
	}

	// 构建镜像
	snapshot := admin_model.SwitchSnapshot{
		CommonModel: model.CommonModel{
			CreatedBy:  userInfo.Username,
			CreateTime: &now,
		},
		Version: model.Version{
			Version: switchInstance.Version.Version,
		},
		SwitchID:     switchInstance.ID,
		NamespaceTag: switchInstance.NamespaceTag,
		EnvTag:       env.Tag,
		CompleteJSON: changes,
	}

	// 无需审批，直接推送
	return s.executeInitialDirectPush(&snapshot, switchInstance, switchConfig)
}

// executeInitialDirectPush 执行初始直接推送（无需审批）
func (s *SwitchService) executeInitialDirectPush(snapshot *admin_model.SwitchSnapshot, switchInstance *model.SwitchModel, config *admin_model.SwitchConfig) error {
	// 无需审批，直接执行变更
	if err := s.ExecutePublishRequest(snapshot, config, switchInstance); err != nil {
		return fmt.Errorf("failed to execute publish request directly: %w", err)
	}
	return nil
}

func (s *SwitchService) ExecutePublishRequest(snapshot *admin_model.SwitchSnapshot, config *admin_model.SwitchConfig, switchInstance *model.SwitchModel) error {
	if err := repository.GetDB().Transaction(func(tx *gorm.DB) error {
		switchSnapshotRepositoryRepoTx := s.switchSnapshotRepositoryRepo.WithTx(tx)
		switchConfigRepoTx := s.switchConfigRepo.WithTx(tx)
		switchRepoTx := s.switchRepo.WithTx(tx)

		//创建快照
		if err := switchSnapshotRepositoryRepoTx.Create(snapshot); err != nil {
			return fmt.Errorf("failed to create snapshot in transaction: %w", err)
		}

		//修改开关配置
		if err := switchConfigRepoTx.UpdateBySwitchIdEnvTag(config); err != nil {
			return fmt.Errorf("failed to update factor config in transaction: %w", err)
		}

		//修改开关信息当前环境标签
		if err := switchRepoTx.UpdateCurrentEnvTagApprovalStatus(switchInstance.ID, switchInstance.CurrentEnvTag, ""); err != nil {
			return fmt.Errorf("failed to update factor config in transaction: %w", err)
		}

		//发送通知消息
		if err := notifier.Notify(snapshot.NamespaceTag, snapshot.EnvTag, switchInstance); err != nil {
			return fmt.Errorf("failed to send notification, transaction rolled back: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// calculateNextEnvIndex 计算下一个环境的索引
func (s *SwitchService) calculateNextEnvIndex(currentEnvTag string, envs []*admin_model.Environment) int {
	envCount := len(envs)
	if envCount == 0 {
		return -1
	}

	// 如果当前环境标签为空，返回第一个环境的索引
	if currentEnvTag == "" {
		return 0
	}

	// 查找当前环境在列表中的位置
	currentIndex := -1
	for i, env := range envs {
		if env.Tag == currentEnvTag {
			currentIndex = i
			break
		}
	}

	// 如果找不到当前环境，返回第一个环境的索引
	if currentIndex == -1 {
		return 0
	}

	// 如果是最后一个环境，返回 -1 表示没有下一个
	if currentIndex >= envCount-1 {
		return -1
	}

	// 返回下一个环境的索引
	return currentIndex + 1
}

// getNextEnvTag 获取下一个环境的标签
// 如果没有下一个环境，返回当前环境标签
func (s *SwitchService) getNextEnvTag(currentEnvTag string, envs []*admin_model.Environment) string {
	nextIndex := s.calculateNextEnvIndex(currentEnvTag, envs)
	if nextIndex == -1 {
		return currentEnvTag // 没有下一个环境，保持当前环境
	}
	return envs[nextIndex].Tag
}
