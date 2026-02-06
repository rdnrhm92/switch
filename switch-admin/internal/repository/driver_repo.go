package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gorm.io/gorm"
)

// DriverRepository 驱动
type DriverRepository struct {
	BaseRepository
}

func NewDriverRepository() *DriverRepository {
	return &DriverRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *DriverRepository) WithTx(tx *gorm.DB) *DriverRepository {
	return WithTx(r, tx)
}

// GetAll 获取所有驱动
func (r *DriverRepository) GetAll(req *dto.DriverListReq) ([]*model.Driver, error) {
	var drivers []*model.Driver
	db := r.GetDB().Model(&model.Driver{})

	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Usage != "" {
		db = db.Where("usage = ?", req.Usage)
	}
	if req.DriverType != "" {
		db = db.Where("driver_type = ?", req.DriverType)
	}

	if req.PageLimit != nil {
		db = db.Offset(int(req.PageLimit.Offset)).Limit(int(req.PageLimit.Limit))
	}

	err := db.Debug().Find(&drivers).Error
	return drivers, err
}

// CountAll 获取所有驱动数量
func (r *DriverRepository) CountAll(req *dto.DriverListReq) (int64, error) {
	var count int64
	db := r.GetDB().Model(&model.Driver{})

	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Usage != "" {
		db = db.Where("usage = ?", req.Usage)
	}
	if req.DriverType != "" {
		db = db.Where("driver_type = ?", req.DriverType)
	}

	err := db.Debug().Count(&count).Error
	return count, err
}

// Create 创建
func (r *DriverRepository) Create(driver *model.Driver) error {
	return r.GetDB().Create(driver).Error
}

// CreateInBatch 批量创建
func (r *DriverRepository) CreateInBatch(drivers []*model.Driver) error {
	if len(drivers) == 0 {
		return nil
	}
	return r.GetDB().Create(drivers).Error
}

// DeleteByID 删除
func (r *DriverRepository) DeleteByID(driverIds []uint) error {
	return r.GetDB().Where("id in ?", driverIds).Delete(&model.Driver{}).Error
}

// Update 更新
func (r *DriverRepository) Update(driver *model.Driver) error {
	return r.GetDB().Model(&model.Driver{}).Updates(driver).Error
}

// GetByEnvironmentID 根据环境ID查询所有驱动
func (r *DriverRepository) GetByEnvironmentID(environmentID uint) ([]*model.Driver, error) {
	var drivers []*model.Driver
	err := r.GetDB().Where("environment_id = ?", environmentID).Find(&drivers).Error
	return drivers, err
}
