package service

import (
	"context"
	"time"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// PermissionService 权限相关
type PermissionService struct {
	namespaceUserRoleRepo *repository.NamespaceUserRoleRepository
	namespaceMembersRepo  *repository.NamespaceMembersRepository
	roleRepo              *repository.RoleRepository
	permissionRepo        *repository.PermissionRepository
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		namespaceUserRoleRepo: repository.NewNamespaceUserRoleRepository(),
		namespaceMembersRepo:  repository.NewNamespaceMembersRepository(),
		roleRepo:              repository.NewRoleRepository(),
		permissionRepo:        repository.NewPermissionRepository(),
	}
}

// AssignRoles 为用户分配角色
func (s *PermissionService) AssignRoles(req *dto.AssignRolesReq) error {
	record, err := s.namespaceMembersRepo.FindByUserIdNamespaceTag(req.UserId, req.NamespaceTag)
	if err != nil {
		return err
	}
	return repository.GetDB().Transaction(func(tx *gorm.DB) error {
		namespaceUserRoleRepoTx := s.namespaceUserRoleRepo.WithTx(tx)

		//先删除再新增
		//删除用户现有的所有角色关联
		id := record.ID

		if err = namespaceUserRoleRepoTx.DeleteByNamespaceMembersId(id); err != nil {
			return err
		}

		userRoles := make([]*admin_model.NamespaceUserRole, len(req.RoleIds))
		for i, roleId := range req.RoleIds {
			userRoles[i] = &admin_model.NamespaceUserRole{
				RoleId:             roleId,
				NamespaceMembersId: id,
			}
		}
		if err = namespaceUserRoleRepoTx.CreateInBatch(userRoles); err != nil {
			return err
		}

		return nil
	})
}

// RolesPermission 获取指定命名空间下的所有角色权限
func (s *PermissionService) RolesPermission(req *dto.RoleListReq) map[string]interface{} {
	var eg errgroup.Group
	roles := make([]*admin_model.Role, 0)
	var count int64

	eg.Go(func() error {
		var err error
		roles, err = s.roleRepo.GetAll(req)
		return err
	})

	eg.Go(func() error {
		var err error
		count, err = s.roleRepo.CountAll(req)
		return err
	})

	if err := eg.Wait(); err != nil {
		logger.Logger.Errorf("fail to get role list: %v", err)
		return map[string]interface{}{
			"success": false,
			"data":    []interface{}{},
			"total":   0,
		}
	}

	return map[string]interface{}{
		"success": true,
		"data":    roles,
		"total":   count,
	}
}

// UpsertPermission 新增/修改权限
func (s *PermissionService) UpsertPermission(ctx context.Context, req *dto.UpsertPermissionReq) error {
	permission := &admin_model.Permission{
		Name:        req.Name,
		Description: req.Description,
		CommonModel: model.CommonModel{
			CreatedBy: info.GetUserName(ctx),
		},
		NamespaceTag: info.GetSelectNamespace(ctx),
	}
	if req.ID != 0 {
		permission.ID = req.ID
		return s.permissionRepo.Update(permission)
	}
	return s.permissionRepo.Create(permission)
}

// UpsertRole 新增/修改角色
func (s *PermissionService) UpsertRole(ctx context.Context, req *dto.UpsertRoleReq) error {
	return repository.GetDB().Transaction(func(tx *gorm.DB) error {
		roleRepoTx := s.roleRepo.WithTx(tx)
		now := time.Now()
		var err error
		var role *admin_model.Role

		if req.ID != 0 {
			role, err = roleRepoTx.FindByID(req.ID)
			if err != nil {
				return err
			}
			// 更新字段
			role.Name = req.Name
			role.Description = req.Description
			role.NamespaceTag = req.NamespaceTag
			role.UpdateTime = &now
			role.UpdateBy = info.GetUserName(ctx)
			err = roleRepoTx.Update(role)
		} else {
			// 新增：创建新记录
			role = &admin_model.Role{
				Name:         req.Name,
				Description:  req.Description,
				NamespaceTag: req.NamespaceTag,
				CommonModel: model.CommonModel{
					CreateTime: &now,
					CreatedBy:  info.GetUserName(ctx),
				},
			}
			err = roleRepoTx.Create(role)
		}
		if err != nil {
			return err
		}

		// 先清空旧的角色权限关联
		if err = tx.Where("role_id = ?", role.ID).Delete(&admin_model.RolePermission{}).Error; err != nil {
			return err
		}

		if len(req.PermissionIds) > 0 {
			userName := info.GetUserName(ctx)

			for _, permissionID := range req.PermissionIds {
				rolePermission := &admin_model.RolePermission{
					RoleID:       role.ID,
					PermissionID: permissionID,
					CommonModel: model.CommonModel{
						CreateTime: &now,
						CreatedBy:  userName,
					},
				}
				if err = tx.Create(rolePermission).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Permissions 获取指定命名空间下的所有权限
func (s *PermissionService) Permissions(namespaceTag string) ([]*admin_model.Permission, error) {
	var permissions []*admin_model.Permission
	err := s.permissionRepo.GetDB().
		Where("permissions.namespace_tag = ?", namespaceTag).
		Or("permissions.namespace_tag = ''").
		Distinct().
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}
	return permissions, nil
}
