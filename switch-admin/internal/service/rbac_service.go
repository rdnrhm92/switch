package service

import (
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gorm.io/gorm"
)

type RBACService struct {
	db *gorm.DB
}

func NewRBACService() *RBACService {
	return &RBACService{db: repository.GetDB()}
}

// CheckPermission 检查roleID下是否可以进行requiredPermission
func (s *RBACService) CheckPermission(roleID uint, requiredPermission string) (bool, error) {
	var count int64

	err := s.db.Table("role_permissions as rp").
		Joins("join permissions as p on p.id = rp.permission_id").
		Where("rp.role_id = ? AND p.name = ?", roleID, requiredPermission).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
