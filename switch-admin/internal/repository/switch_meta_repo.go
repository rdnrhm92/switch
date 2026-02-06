package repository

import (
	"time"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SwitchFactorRepository 因子信息
type SwitchFactorRepository struct {
	BaseRepository
}

func NewSwitchFactorRepository() *SwitchFactorRepository {
	return &SwitchFactorRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *SwitchFactorRepository) WithTx(tx *gorm.DB) *SwitchFactorRepository {
	return WithTx(r, tx)
}

// FindList 列表检索
func (r *SwitchFactorRepository) FindList(ctx *gin.Context, req *dto.SwitchFactorListReq) ([]*admin_model.SwitchFactor, error) {
	var sms []*admin_model.SwitchFactor
	db := FindSwitchFactorWithSuperAdmin(ctx, r.GetDB()).Model(&admin_model.SwitchFactor{})

	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Factor != "" {
		db = db.Where("factor LIKE ?", "%"+req.Factor+"%")
	}
	if req.Description != "" {
		db = db.Where("description LIKE ?", "%"+req.Description+"%")
	}
	if req.CreatedBy != "" {
		db = db.Where("created_by = ?", req.CreatedBy)
	}
	if req.PageTime != nil {
		if req.StartTime != nil {
			db = db.Where("create_time >= ?", req.StartTime)
		}
		if req.EndTime != nil {
			db = db.Where("create_time <= ?", req.EndTime)
		}
	}
	if req.PageLimit != nil {
		db = db.Offset(int(req.Offset)).Limit(int(req.Limit))
	}

	err := db.Order("update_time desc, create_time desc").Find(&sms).Error
	return sms, err
}

// FindListLike 列表检索(全部)
func (r *SwitchFactorRepository) FindListLike(ctx *gin.Context, req *dto.SwitchFactorLikeReq) ([]*admin_model.SwitchFactor, error) {
	var sms []*admin_model.SwitchFactor
	db := FindSwitchFactorWithSuperAdmin(ctx, r.GetDB()).Model(&admin_model.SwitchFactor{})

	if req.Factor != "" {
		db = db.Where("factor LIKE ?", "%"+req.Factor+"%")
	}

	err := db.Order("update_time desc, create_time desc").Find(&sms).Error
	return sms, err
}

// InsertOne 新增
func (r *SwitchFactorRepository) InsertOne(sm *admin_model.SwitchFactor) (*admin_model.SwitchFactor, error) {
	err := r.GetDB().Create(sm).Error
	return sm, err
}

// BatchInsert 批量新增
func (r *SwitchFactorRepository) BatchInsert(sms []*admin_model.SwitchFactor) ([]*admin_model.SwitchFactor, error) {
	err := r.GetDB().Create(sms).Error
	return sms, err
}

// Update 更新
func (r *SwitchFactorRepository) Update(req *dto.CreateSwitchFactorReq, updateBy string) (*admin_model.SwitchFactor, error) {
	sm := &admin_model.SwitchFactor{}

	updates := map[string]interface{}{
		"description": req.Description,
		"update_by":   updateBy,
		"json_schema": req.JsonSchema,
		"name":        req.Name,
		"update_time": time.Now(),
	}

	db := r.GetDB().Model(sm).Where("id = ?", req.Id).Updates(updates)
	if db.Error != nil {
		return nil, db.Error
	}
	if db.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.GetDB().First(sm, req.Id).Error
	return sm, err
}

// CountAll 总数
func (r *SwitchFactorRepository) CountAll(req *dto.SwitchFactorListReq) (int64, error) {
	var count int64
	db := r.GetDB().Model(&admin_model.SwitchFactor{})

	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Factor != "" {
		db = db.Where("factor LIKE ?", "%"+req.Factor+"%")
	}
	if req.Description != "" {
		db = db.Where("description LIKE ?", "%"+req.Description+"%")
	}
	if req.CreatedBy != "" {
		db = db.Where("created_by = ?", req.CreatedBy)
	}

	err := db.Count(&count).Error
	return count, err
}
