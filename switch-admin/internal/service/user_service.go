package service

import (
	"context"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"golang.org/x/sync/errgroup"
)

// UserService 定义了用户相关的服务
type UserService struct {
	userRepo *repository.UserRepository
	rbacRepo *repository.RBACRepository
}

// NewUserService 创建一个用户服务实例
func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
		rbacRepo: repository.NewRBACRepository(),
	}
}

func (s *UserService) GetUserInfo(userID uint) (*admin_model.User, error) {
	user, err := s.userRepo.FindUserWithAllInfo(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetAllUserLike 模糊查询所有用户
func (s *UserService) GetAllUserLike(ctx context.Context, req *dto.AllUserLikeReq) ([]*admin_model.User, error) {
	allUsers, err := s.userRepo.FindByUsernameLike(ctx, req.UserName)
	if err != nil {
		logger.Logger.Errorf("fail to get likes by username like %v", err)
		return make([]*admin_model.User, 0), nil
	}
	return allUsers, nil
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(ctx context.Context, req *dto.UserListReq) map[string]interface{} {
	var eg errgroup.Group
	users := make([]*dto.UserListResp, 0)
	var count int64

	eg.Go(func() error {
		var err error
		originUsers, err := s.userRepo.GetAll(ctx, req)
		if err == nil && len(originUsers) > 0 {
			for _, user := range originUsers {
				// 检查用户是否有命名空间成员信息
				if len(user.NamespaceMembers) == 0 {
					continue
				}

				roles := make([]*admin_model.Role, 0)
				member := user.NamespaceMembers[0]

				// 检查是否有用户角色
				if len(member.UserRoles) == 0 {
					continue
				}

				// 收集角色信息，过滤掉空角色
				for _, userRole := range member.UserRoles {
					if userRole.Role.ID != 0 {
						roles = append(roles, &userRole.Role)
					}
				}

				// 如果指定了角色过滤但没有匹配的角色，跳过该用户
				if req.Role != "" && len(roles) == 0 {
					continue
				}

				users = append(users, &dto.UserListResp{
					UserInfo: admin_model.User{
						CommonModel: user.CommonModel,
						Username:    user.Username,
					},
					UserRoles: roles,
				})
			}
		}
		return err
	})

	eg.Go(func() error {
		var err error
		count, err = s.userRepo.CountAll(ctx, req)
		return err
	})

	if err := eg.Wait(); err != nil {
		logger.Logger.Errorf("fail to get user list: %v", err)
		return map[string]interface{}{
			"success": false,
			"data":    []interface{}{},
			"total":   0,
		}
	}

	return map[string]interface{}{
		"success": true,
		"data":    users,
		"total":   count,
	}
}

// GetUsersPermissions 获取用户列表
func (s *UserService) GetUsersPermissions(ctx context.Context, req *dto.UserPermissionsListReq) []*dto.UserPermissionsListResp {
	users := make([]*dto.UserPermissionsListResp, 0)

	var err error
	originUsers, err := s.userRepo.GetAllWithPermissions(ctx, req)
	if err == nil && len(originUsers) > 0 {
		for _, user := range originUsers {
			permissionMap := make(map[uint]*admin_model.Permission)

			for _, namespaceMember := range user.NamespaceMembers {
				for _, userRole := range namespaceMember.UserRoles {
					for _, permission := range userRole.Role.Permissions {
						permissionMap[permission.ID] = permission
					}
				}
			}

			userPermissions := make([]*admin_model.Permission, 0, len(permissionMap))
			for _, permission := range permissionMap {
				userPermissions = append(userPermissions, permission)
			}

			userResp := &dto.UserPermissionsListResp{
				UserInfo:        *user,
				UserPermissions: userPermissions,
			}

			users = append(users, userResp)
		}
	}

	return users
}
