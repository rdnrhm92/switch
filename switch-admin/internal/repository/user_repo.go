package repository

import (
	"context"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gorm.io/gorm"
)

// UserRepository 定义了用户仓库
type UserRepository struct {
	BaseRepository
}

func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return WithTx(r, tx)
}

// NewUserRepository 创建实例
func NewUserRepository() *UserRepository {
	return &UserRepository{BaseRepository: BaseRepository{
		DB: GetDB(),
	}}
}

// buildUserQuery 根据请求参数构建查询
func (r *UserRepository) buildUserQuery(req *dto.UserListReq) *gorm.DB {
	db := r.GetDB().Model(&admin_model.User{})
	if req != nil {
		if req.Username != "" {
			db = db.Where("username LIKE ?", "%"+req.Username+"%")
		}
		if req.PageTime != nil {
			if req.PageTime.StartTime != nil {
				db = db.Where("create_time >= ?", req.PageTime.StartTime)
			}
			if req.PageTime.EndTime != nil {
				db = db.Where("create_time <= ?", req.PageTime.EndTime)
			}
		}
	}
	// 超级管理员不在列表展示范围内
	db = db.Where("is_super_admin = 0")
	return db
}

// buildUserQueryWithRole 根据请求参数构建查询（包含角色过滤）
func (r *UserRepository) buildUserQueryWithRole(ctx context.Context, req *dto.UserListReq) *gorm.DB {
	db := r.buildUserQuery(req)
	db = FindWithSuperAdminV2(ctx, db)

	if req != nil && req.Role != "" {
		db = db.Joins("JOIN namespace_user_role ON namespace_user_role.namespace_members_id = namespace_members.id").
			Joins("JOIN roles ON roles.id = namespace_user_role.role_id AND roles.name = ?", req.Role)
	}

	return db
}

// GetAll 获取所有用户
func (r *UserRepository) GetAll(ctx context.Context, req *dto.UserListReq) ([]*admin_model.User, error) {
	var users []*admin_model.User

	db := r.buildUserQueryWithRole(ctx, req)

	if req.PageLimit != nil {
		db = db.Offset(int(req.PageLimit.Offset)).Limit(int(req.PageLimit.Limit))
	}

	// 预加载用户的角色信息
	if req.Role != "" {
		db = db.Preload("NamespaceMembers.UserRoles.Role", "name = ?", req.Role)
	} else {
		db = db.Preload("NamespaceMembers.UserRoles.Role")
	}

	err := db.Order("id desc").Find(&users).Error
	return users, err
}

// GetAllWithPermissions 获取所有用户(带权限)
func (r *UserRepository) GetAllWithPermissions(ctx context.Context, req *dto.UserPermissionsListReq) ([]*admin_model.User, error) {
	var users []*admin_model.User

	db := r.GetDB().Model(&admin_model.User{})

	if req != nil && req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}

	db = FindWithSuperAdminV2(ctx, db)

	err := db.Preload("NamespaceMembers.UserRoles.Role.Permissions").
		Order("id desc").
		Find(&users).Error

	return users, err
}

// CountAll 获取所有用户数量
func (r *UserRepository) CountAll(ctx context.Context, req *dto.UserListReq) (int64, error) {
	var count int64

	db := r.buildUserQueryWithRole(ctx, req)

	err := db.Count(&count).Error
	return count, err
}

// Create 创建一个新的用户记录
func (r *UserRepository) Create(user *admin_model.User) error {
	return r.GetDB().Create(user).Error
}

// FindByUsername 通过用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*admin_model.User, error) {
	var user admin_model.User
	err := r.GetDB().Where("username = ?", username).First(&user).Error
	return &user, err
}

// FindByUsernameLike 通过用户名查找用户(模糊)
func (r *UserRepository) FindByUsernameLike(ctx context.Context, username string) ([]*admin_model.User, error) {
	var users []*admin_model.User
	db := r.GetDB()
	db = FindWithSuperAdminV2(ctx, db)
	err := db.Where("username LIKE ?", "%"+username+"%").Find(&users).Error
	return users, err
}

// FindByUserId 通过用户id查找用户
func (r *UserRepository) FindByUserId(uid uint) (*admin_model.User, error) {
	var user admin_model.User
	err := r.GetDB().Where("id = ?", uid).First(&user).Error
	return &user, err
}

// FindByUserIds 通过用户ids查找用户
func (r *UserRepository) FindByUserIds(uids []uint) ([]*admin_model.User, error) {
	var users []*admin_model.User
	err := r.GetDB().Where("id in ?", uids).Find(&users).Error
	return users, err
}

// FindByIDWithRole 通过用户ID查找用户并加载关联的角色
func (r *UserRepository) FindByIDWithRole(userId uint) (*admin_model.User, error) {
	var user admin_model.User
	err := r.GetDB().Preload("Role").First(&user, userId).Error
	return &user, err
}

// FindUserWithAllInfo 查询用户及其所有关联信息
func (r *UserRepository) FindUserWithAllInfo(userId uint) (*admin_model.User, error) {
	var user admin_model.User

	err := r.GetDB().
		Preload("NamespaceMembers.UserRoles.Role.Permissions").
		Preload("NamespaceMembers.Namespace").
		First(&user, userId).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
