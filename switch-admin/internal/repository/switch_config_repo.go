package repository

import (
	"strings"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gorm.io/gorm"
)

// SwitchConfigRepository 定义了对 SwitchConfig 模型的数据库操作
type SwitchConfigRepository struct {
	BaseRepository
}

func NewSwitchConfigRepository() *SwitchConfigRepository {
	return &SwitchConfigRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *SwitchConfigRepository) WithTx(tx *gorm.DB) *SwitchConfigRepository {
	return WithTx(r, tx)
}

// GetBySwitchID 获取一个开关在所有环境下的配置
func (r *SwitchConfigRepository) GetBySwitchID(switchID uint) ([]*admin_model.SwitchConfig, error) {
	var configs []*admin_model.SwitchConfig
	err := r.GetDB().Where("switch_id = ?", switchID).Find(&configs).Error
	return configs, err
}

// GetBySwitchIdEnvTagVersion 获取一个开关的配置
func (r *SwitchConfigRepository) GetBySwitchIdEnvTagVersion(switchID uint, envTag string) (*admin_model.SwitchConfig, error) {
	var config *admin_model.SwitchConfig
	err := r.GetDB().Where("switch_id = ? and env_tag = ? ", switchID, envTag).First(&config).Error
	return config, err
}

// Update 修改SwitchConfig
func (r *SwitchConfigRepository) Update(cfg *admin_model.SwitchConfig) error {
	return r.GetDB().Save(cfg).Error
}

func (r *SwitchConfigRepository) UpdateBySwitchIdEnvTag(cfg *admin_model.SwitchConfig) error {
	return r.GetDB().Save(cfg).Error
}

// Create 创建SwitchConfig
func (r *SwitchConfigRepository) Create(cfg *admin_model.SwitchConfig) error {
	return r.GetDB().Create(cfg).Error
}

// GetBatchConfigsByIDsAndVersions 批量获取开关配置信息
func (r *SwitchConfigRepository) GetBatchConfigsByIDsAndVersions(switchIDVersions map[uint]int64) (map[uint][]*admin_model.SwitchConfig, error) {
	if len(switchIDVersions) == 0 {
		return make(map[uint][]*admin_model.SwitchConfig), nil
	}

	var configs []*admin_model.SwitchConfig

	var conditions []string
	var args []interface{}
	for switchID, version := range switchIDVersions {
		conditions = append(conditions, "(switch_id = ? AND version = ?)")
		args = append(args, switchID, version)
	}

	query := r.GetDB().Where("("+strings.Join(conditions, " OR ")+")", args...)
	err := query.Debug().Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// 按开关ID分组
	result := make(map[uint][]*admin_model.SwitchConfig)
	for _, config := range configs {
		result[config.SwitchID] = append(result[config.SwitchID], config)
	}

	return result, nil
}

// GetBatchLatestConfigsBySwitchIDs 批量获取开关的配置信息
func (r *SwitchConfigRepository) GetBatchLatestConfigsBySwitchIDs(switchIDs []uint) (map[uint][]*admin_model.SwitchConfig, error) {
	if len(switchIDs) == 0 {
		return make(map[uint][]*admin_model.SwitchConfig), nil
	}

	var configs []*admin_model.SwitchConfig

	err := r.GetDB().
		Where("switch_id IN (?)", switchIDs).
		Order("id").
		Find(&configs).Error

	if err != nil {
		return nil, err
	}

	// 按开关ID分组
	result := make(map[uint][]*admin_model.SwitchConfig)
	for _, config := range configs {
		result[config.SwitchID] = append(result[config.SwitchID], config)
	}

	return result, nil
}
