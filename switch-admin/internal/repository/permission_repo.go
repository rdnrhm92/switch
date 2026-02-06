package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	BaseRepository
}

func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *PermissionRepository) WithTx(tx *gorm.DB) *PermissionRepository {
	return WithTx(r, tx)
}

// Create 创建权限
func (r *PermissionRepository) Create(permission *admin_model.Permission) error {
	return r.DB.Create(permission).Error
}

// Update 更新权限
func (r *PermissionRepository) Update(permission *admin_model.Permission) error {
	return r.DB.Save(permission).Error
}

// DeleteByID 删除权限
func (r *PermissionRepository) DeleteByID(id uint) error {
	return r.DB.Delete(&admin_model.Permission{}, id).Error
}

// FindByID 根据ID查找权限
func (r *PermissionRepository) FindByID(id uint) (*admin_model.Permission, error) {
	var permission admin_model.Permission
	err := r.DB.First(&permission, id).Error
	return &permission, err
}
