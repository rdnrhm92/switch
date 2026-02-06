package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gorm.io/gorm"
)

// SwitchApprovalRepository 审批相关
type SwitchApprovalRepository struct {
	BaseRepository
}

func NewSwitchApprovalRepository() *SwitchApprovalRepository {
	return &SwitchApprovalRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *SwitchApprovalRepository) WithTx(tx *gorm.DB) *SwitchApprovalRepository {
	return WithTx(r, tx)
}

// FindBySwitchAndEnv 根据开关ID+环境Tag查找审批人
func (r *SwitchApprovalRepository) FindBySwitchAndEnv(switchID uint, envTag string) (*admin_model.SwitchApproval, error) {
	var approval admin_model.SwitchApproval
	err := r.GetDB().Where("switch_id = ? AND env_tag = ?", switchID, envTag).First(&approval).Error
	return &approval, err
}

// FindBySwitch 根据开关ID查找审批人
func (r *SwitchApprovalRepository) FindBySwitch(switchID uint) ([]*admin_model.SwitchApproval, error) {
	var approval []*admin_model.SwitchApproval
	err := r.GetDB().Where("switch_id = ?", switchID).Find(&approval).Error
	return approval, err
}

// DeleteBySwitchId 删除审批人根据环境ID
func (r *SwitchApprovalRepository) DeleteBySwitchId(switchId uint) error {
	return r.GetDB().Where("switch_id = ? ", switchId).Delete(&admin_model.SwitchApproval{}).Error
}

// Create 创建
func (r *SwitchApprovalRepository) Create(req []*admin_model.SwitchApproval) error {
	return r.GetDB().Create(&req).Error
}
