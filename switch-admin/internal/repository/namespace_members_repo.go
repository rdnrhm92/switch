package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gorm.io/gorm"
)

type NamespaceMembersRepository struct {
	BaseRepository
}

func NewNamespaceMembersRepository() *NamespaceMembersRepository {
	return &NamespaceMembersRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *NamespaceMembersRepository) WithTx(tx *gorm.DB) *NamespaceMembersRepository {
	return WithTx(r, tx)
}

// FindApprovePermissionsByNamespaceTag 找这个空间Tag下的带审批权限的用户
func (r *NamespaceMembersRepository) FindApprovePermissionsByNamespaceTag(namespaceTag string) ([]*admin_model.NamespaceMembers, error) {
	var members []*admin_model.NamespaceMembers
	err := r.GetDB().
		Joins("JOIN namespace_user_role on namespace_user_role.namespace_members_id = namespace_members.id").
		Joins("JOIN roles on roles.id = namespace_user_role.role_id").
		Joins("JOIN role_permissions on role_permissions.role_id = roles.id").
		Joins("JOIN permissions on permissions.id = role_permissions.permission_id").
		Where("namespace_members.namespace_tag = ? AND permissions.name = ?", namespaceTag, "approvals:approve").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// FindByUserIdNamespaceTag 根据用户ID,namespaceTag查询
func (r *NamespaceMembersRepository) FindByUserIdNamespaceTag(userId uint, namespaceTag string) (*admin_model.NamespaceMembers, error) {
	var member *admin_model.NamespaceMembers
	err := r.GetDB().Where("user_id = ?", userId).Where("namespace_tag = ?", namespaceTag).First(&member).Error
	return member, err
}

// CreateInBatch 批量创建用户和角色的关联
func (r *NamespaceMembersRepository) CreateInBatch(userRoles []*admin_model.NamespaceMembers) error {
	if len(userRoles) == 0 {
		return nil
	}
	return r.GetDB().Create(&userRoles).Error
}
