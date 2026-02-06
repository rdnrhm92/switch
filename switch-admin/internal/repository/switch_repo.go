package repository

import (
	"context"

	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SwitchRepository 定义了对 SwitchModel 模型的数据库操作
type SwitchRepository struct {
	BaseRepository
}

func NewSwitchRepository() *SwitchRepository {
	return &SwitchRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *SwitchRepository) WithTx(tx *gorm.DB) *SwitchRepository {
	return WithTx(r, tx)
}

func (r *SwitchRepository) Create(sw *model.SwitchModel) error {
	return r.GetDB().Create(sw).Error
}

func (r *SwitchRepository) Save(sw *model.SwitchModel) error {
	return r.GetDB().Save(sw).Error
}

// GetSwitchByID 根据 ID 获取开关
func (r *SwitchRepository) GetSwitchByID(id uint) (*model.SwitchModel, error) {
	var sw model.SwitchModel
	err := r.GetDB().First(&sw, id).Error
	return &sw, err
}

// GetByNamespaceTagAndName 根据命名空间Tag和名称获取开关
func (r *SwitchRepository) GetByNamespaceTagAndName(namespaceTag string, name string) (*model.SwitchModel, error) {
	var sw model.SwitchModel
	err := r.GetDB().Where("namespace_tag = ? AND name = ?", namespaceTag, name).First(&sw).Error
	return &sw, err
}

// Update 更新一个开关
func (r *SwitchRepository) Update(sw *model.SwitchModel) error {
	return r.GetDB().Save(sw).Error
}

// UpdateSwitchApproval 更新一个开关的审核状态
func (r *SwitchRepository) UpdateSwitchApproval(id uint, approvalStatus string) error {
	return r.GetDB().Model(&model.SwitchModel{}).Where("id = ?", id).Update("approver_status", approvalStatus).Error
}

// Delete 删除一个开关
func (r *SwitchRepository) Delete(id uint) error {
	return r.GetDB().Delete(&model.SwitchModel{}, id).Error
}

// FindList 开关列表
func (r *SwitchRepository) FindList(ctx *gin.Context, req *dto.SwitchListReq) ([]*model.SwitchModel, error) {
	var results []*model.SwitchModel
	db := FindSwitchWithSuperAdmin(ctx, r.GetDB()).
		Table("switches").
		Select("switches.*")

	if req.Name != "" {
		db = db.Where("switches.name LIKE ?", "%"+req.Name+"%")
	}
	if req.Description != "" {
		db = db.Where("switches.description LIKE ?", "%"+req.Description+"%")
	}

	if req.PageLimit != nil {
		db = db.Offset(int(req.Offset)).Limit(int(req.Limit))
	}

	err := db.Order("switches.update_time desc, switches.create_time desc").Scan(&results).Error
	return results, err
}

// CountAll 开关总数
func (r *SwitchRepository) CountAll(ctx context.Context, req *dto.SwitchListReq) (int64, error) {
	var count int64
	db := FindSwitchWithSuperAdmin(ctx, r.GetDB()).Model(&model.SwitchModel{})

	if req.Name != "" {
		db = db.Where("switches.name LIKE ?", "%"+req.Name+"%")
	}
	if req.Description != "" {
		db = db.Where("switches.description LIKE ?", "%"+req.Description+"%")
	}

	err := db.Count(&count).Error
	return count, err
}

// GetByIDAndVersion 根据ID和版本号获取开关
func (r *SwitchRepository) GetByIDAndVersion(id uint, version int64) (*model.SwitchModel, error) {
	var switchModel model.SwitchModel
	err := r.GetDB().Where("id = ? AND version = ?", id, version).First(&switchModel).Error
	return &switchModel, err
}

// UpdateCurrentEnvTagApprovalStatus 更新开关的当前环境标签跟审核状态
func (r *SwitchRepository) UpdateCurrentEnvTagApprovalStatus(id uint, currentEnvTag string, approvalStatus string) error {
	return r.GetDB().Model(&model.SwitchModel{}).Where("id = ?", id).Update("current_env_tag", currentEnvTag).Update("approver_status", approvalStatus).Error
}
