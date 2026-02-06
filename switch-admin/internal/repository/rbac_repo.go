package repository

import "gorm.io/gorm"

// RBACRepository rbac操作
type RBACRepository struct {
	BaseRepository
}

func NewRBACRepository() *RBACRepository {
	return &RBACRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *RBACRepository) WithTx(tx *gorm.DB) *RBACRepository {
	return WithTx(r, tx)
}

// GetPermissionsByRoleID 获取权限
func (r *RBACRepository) GetPermissionsByRoleID(roleID uint) ([]string, error) {
	var permissionNames []string

	err := r.GetDB().Table("permissions as p").
		Select("p.name").
		Joins("join role_permissions as rp on rp.permission_id = p.id").
		Where("rp.role_id = ?", roleID).
		Pluck("p.name", &permissionNames).Error

	if err != nil {
		return nil, err
	}

	return permissionNames, nil
}
