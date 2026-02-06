package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// NamespaceService 提供了与命名空间相关的业务逻辑
type NamespaceService struct {
	nameSpaceRepo         *repository.NamespaceRepository
	userRepo              *repository.UserRepository
	approvalFormRepo      *repository.ApprovalFormRepository
	namespaceMembersRepo  *repository.NamespaceMembersRepository
	namespaceUserRoleRepo *repository.NamespaceUserRoleRepository
}

func NewNamespaceService() *NamespaceService {
	return &NamespaceService{
		nameSpaceRepo:         repository.NewNamespaceRepository(),
		userRepo:              repository.NewUserRepository(),
		approvalFormRepo:      repository.NewApprovalFormRepository(),
		namespaceMembersRepo:  repository.NewNamespaceMembersRepository(),
		namespaceUserRoleRepo: repository.NewNamespaceUserRoleRepository(),
	}
}

// GetNamespaces 获取所有命名空间
func (s *NamespaceService) GetNamespaces(ctx *gin.Context, namespaceListReq *dto.NamespaceListReq) map[string]interface{} {
	batchGroup := errgroup.Group{}
	var allRes []*admin_model.Namespace
	batchGroup.Go(func() error {
		all, err := s.nameSpaceRepo.GetAll(ctx, namespaceListReq)
		if err != nil {
			return err
		}
		allRes = all
		return nil
	})
	var countRes int64
	batchGroup.Go(func() error {
		count, err := s.nameSpaceRepo.CountAll(ctx, namespaceListReq)
		if err != nil {
			return err
		}
		countRes = count
		return nil
	})
	err := batchGroup.Wait()
	if err != nil {
		logger.Logger.Errorf("[NamespaceService] batch err: %v", err)
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

// GetNamespacesLike 获取所有命名空间
func (s *NamespaceService) GetNamespacesLike(ctx *gin.Context, namespaceListLikeReq *dto.NamespaceListLikeReq) []*dto.NamespaceListLikeResp {
	resp := make([]*dto.NamespaceListLikeResp, 0)
	userInfo := info.GetUserInfo(ctx)
	if userInfo == nil {
		return resp
	}
	all, err := s.nameSpaceRepo.GetAllLike(ctx, namespaceListLikeReq)
	if err != nil {
		return resp
	}

	//根据申请人找对应的审核单子(这里只判申请单，拒绝走后门进入空间)
	requests, err := s.approvalFormRepo.GetListByRequestUser(userInfo.ID, admin_model.NamespaceType)
	if err != nil {
		return resp
	}
	// 构建空间Tag跟单子的映射关系
	namespaceStatusMapping := make(map[string]string, len(requests))
	duplicateRemoval := make(map[string]struct{})
	for _, request := range requests {
		ns := new(admin_model.NamespaceApprovalForm)
		if err = json.Unmarshal(unsafe.Slice(unsafe.StringData(request.Details), len(request.Details)), ns); err != nil {
			return resp
		}
		if _, ok := duplicateRemoval[ns.NamespaceTag]; !ok {
			duplicateRemoval[ns.NamespaceTag] = struct{}{}
			namespaceStatusMapping[ns.NamespaceTag] = request.Status
		}
	}

	for _, namespace := range all {
		if _, ok := namespaceStatusMapping[namespace.Tag]; ok {
			resp = append(resp, &dto.NamespaceListLikeResp{
				Namespace: namespace,
				Status:    namespaceStatusMapping[namespace.Tag],
			})
		} else {
			status := ""
			if namespace.OwnerUserId == userInfo.ID {
				status = "APPROVED"
			}
			resp = append(resp, &dto.NamespaceListLikeResp{
				Namespace: namespace,
				Status:    status,
			})
		}
	}
	return resp
}

// AddNamespaceMember 给空间下新增用户并分配角色
func (s *NamespaceService) AddNamespaceMember(tx *gorm.DB, req *admin_model.NamespaceMembers) error {
	namespaceMembersRepoTx := s.namespaceMembersRepo.WithTx(tx)
	if err := namespaceMembersRepoTx.CreateInBatch([]*admin_model.NamespaceMembers{req}); err != nil {
		return err
	}
	return nil
}

// CreateUpdateNamespace 创建一个命名空间/修改
func (s *NamespaceService) CreateUpdateNamespace(ctx *gin.Context, req *dto.CreateUpdateNamespaceReq) (*admin_model.Namespace, error) {
	now := time.Now()
	if req.Id == 0 {
		existingNs, err := s.nameSpaceRepo.GetByTag(req.Tag)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check for existing namespace: %w", err)
		}
		if existingNs != nil {
			return nil, fmt.Errorf("namespace with tag '%s' already exists", req.Tag)
		}
		userId := info.GetUserId(ctx)
		ns := &admin_model.Namespace{
			Name:        req.Name,
			Tag:         req.Tag,
			Description: req.Description,
			OwnerUserId: userId,
			CommonModel: model.CommonModel{
				CreatedBy:  req.CreatedBy,
				CreateTime: &now,
			},
		}
		if err = repository.GetDB().Transaction(func(tx *gorm.DB) error {
			nameSpaceRepo := s.nameSpaceRepo.WithTx(tx)
			namespaceMembersRepo := s.namespaceMembersRepo.WithTx(tx)
			namespaceUserRoleMembersRepo := s.namespaceUserRoleRepo.WithTx(tx)
			// 新增命名空间
			if err = nameSpaceRepo.Create(ns); err != nil {
				return fmt.Errorf("failed to create namespace with environments: %w", err)
			}
			nsMembers := admin_model.NamespaceMembers{
				UserId:       userId,
				NamespaceTag: ns.Tag,
			}
			// 新增当前用户为admin用户
			if err = namespaceMembersRepo.CreateInBatch([]*admin_model.NamespaceMembers{&nsMembers}); err != nil {
				return err
			}

			if err = namespaceUserRoleMembersRepo.Create(&admin_model.NamespaceUserRole{
				RoleId:             repository.AdminRole.ID,
				NamespaceMembersId: nsMembers.ID,
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return ns, nil
	} else {
		ns := &admin_model.Namespace{
			Name:        req.Name,
			Description: req.Description,
			CommonModel: model.CommonModel{
				ID:         req.Id,
				UpdateBy:   req.UpdatedBy,
				UpdateTime: &now,
			},
		}
		if err := s.nameSpaceRepo.Update(ns); err != nil {
			return nil, err
		}
		return ns, nil
	}

}
