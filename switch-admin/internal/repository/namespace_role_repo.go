package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gorm.io/gorm"
)

type NamespaceUserRoleRepository struct {
	BaseRepository
}

func NewNamespaceUserRoleRepository() *NamespaceUserRoleRepository {
	return &NamespaceUserRoleRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *NamespaceUserRoleRepository) WithTx(tx *gorm.DB) *NamespaceUserRoleRepository {
	return WithTx(r, tx)
}

func (r *NamespaceUserRoleRepository) Create(nsRole *admin_model.NamespaceUserRole) error {
	return r.DB.Model(nsRole).Save(nsRole).Error
}

func (r *NamespaceUserRoleRepository) FindByUserId(nsRole *admin_model.NamespaceUserRole) error {
	return r.DB.Model(nsRole).Save(nsRole).Error
}

func (r *NamespaceUserRoleRepository) DeleteByNamespaceMembersId(namespaceMembersId uint) error {
	return r.DB.Where("namespace_members_id = ?", namespaceMembersId).Delete(&admin_model.NamespaceUserRole{}).Error
}

// CreateInBatch 批量创建用户和角色的关联
func (r *NamespaceUserRoleRepository) CreateInBatch(roles []*admin_model.NamespaceUserRole) error {
	if len(roles) == 0 {
		return nil
	}
	return r.GetDB().Create(&roles).Error
}
