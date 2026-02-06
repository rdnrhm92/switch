package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gorm.io/gorm"
)

type RoleRepository struct {
	BaseRepository
}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *RoleRepository) WithTx(tx *gorm.DB) *RoleRepository {
	return WithTx(r, tx)
}

// Create 创建角色
func (r *RoleRepository) Create(role *admin_model.Role) error {
	return r.DB.Create(role).Error
}

// Update 更新角色
func (r *RoleRepository) Update(role *admin_model.Role) error {
	return r.DB.Save(role).Error
}

// DeleteByID 删除角色
func (r *RoleRepository) DeleteByID(id uint) error {
	return r.DB.Delete(&admin_model.Role{}, id).Error
}

// FindByID 根据ID查找角色
func (r *RoleRepository) FindByID(id uint) (*admin_model.Role, error) {
	var role admin_model.Role
	err := r.DB.First(&role, id).Error
	return &role, err
}

// GetAll 获取所有角色权限
func (r *RoleRepository) GetAll(req *dto.RoleListReq) ([]*admin_model.Role, error) {
	var roles []*admin_model.Role
	db := r.DB.Model(&admin_model.Role{}).Where("roles.namespace_tag = ? OR roles.namespace_tag = ''", req.NamespaceTag)

	if req.RoleName != "" {
		db = db.Where("roles.name LIKE ?", "%"+req.RoleName+"%")
	}

	if req.PermissionName != "" {
		db = db.Joins("JOIN role_permissions on role_permissions.role_id = roles.id").
			Joins("JOIN permissions on permissions.id = role_permissions.permission_id").
			Where("permissions.name LIKE ?", "%"+req.PermissionName+"%").
			Group("roles.id")
	}

	if req.PageLimit != nil {
		db = db.Offset(int(req.PageLimit.Offset)).Limit(int(req.PageLimit.Limit))
	}

	err := db.Preload("Permissions").Order("id desc").Find(&roles).Error
	return roles, err
}

// CountAll 总数查询
func (r *RoleRepository) CountAll(req *dto.RoleListReq) (int64, error) {
	var count int64
	db := r.DB.Model(&admin_model.Role{}).Where("roles.namespace_tag = ? OR roles.namespace_tag = ''", req.NamespaceTag)

	if req.RoleName != "" {
		db = db.Where("roles.name LIKE ?", "%"+req.RoleName+"%")
	}

	if req.PermissionName != "" {
		db = db.Joins("JOIN role_permissions on role_permissions.role_id = roles.id").
			Joins("JOIN permissions on permissions.id = role_permissions.permission_id").
			Where("permissions.name LIKE ?", "%"+req.PermissionName+"%").
			Group("roles.id")
	}

	err := db.Count(&count).Error
	return count, err
}
