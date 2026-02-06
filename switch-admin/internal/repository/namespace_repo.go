package repository

import (
	"context"
	"errors"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NamespaceRepository struct {
	BaseRepository
}

func NewNamespaceRepository() *NamespaceRepository {
	return &NamespaceRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *NamespaceRepository) WithTx(tx *gorm.DB) *NamespaceRepository {
	return WithTx(r, tx)
}

// GetAll 获取所有命名空间
func (r *NamespaceRepository) GetAll(ctx context.Context, namespaceListReq *dto.NamespaceListReq) ([]*admin_model.Namespace, error) {
	var namespaces []*admin_model.Namespace

	db := r.GetDB()

	if namespaceListReq != nil {
		if namespaceListReq.Name != "" {
			db = db.Where("name LIKE ?", "%"+namespaceListReq.Name+"%")
		}
		if namespaceListReq.Tag != "" {
			db = db.Where("tag LIKE ?", "%"+namespaceListReq.Tag+"%")
		}
		if namespaceListReq.PageTime != nil {
			if namespaceListReq.StartTime != nil {
				db = db.Where("create_time >= ?", namespaceListReq.StartTime)
			}
			if namespaceListReq.EndTime != nil {
				db = db.Where("create_time <= ?", namespaceListReq.EndTime)
			}
		}
		if namespaceListReq.Description != "" {
			db = db.Where("description LIKE ?", "%"+namespaceListReq.Description+"%")
		}
		if namespaceListReq.CreatedBy != "" {
			db = db.Where("created_by = ?", namespaceListReq.CreatedBy)
		}
		if namespaceListReq.PageLimit != nil {
			db = db.Offset(int(namespaceListReq.PageLimit.Offset)).Limit(int(namespaceListReq.PageLimit.Limit))
		}
	}
	err := db.Preload("Environments.Drivers").Order("id desc").Find(&namespaces).Error
	return namespaces, err
}

// GetAllLike 获取所有命名空间 like查询
func (r *NamespaceRepository) GetAllLike(ctx context.Context, namespaceListLikeReq *dto.NamespaceListLikeReq) ([]*admin_model.Namespace, error) {
	var namespaces []*admin_model.Namespace
	var db *gorm.DB
	db = r.GetDB()

	if !namespaceListLikeReq.All {
		db = FindWithSuperAdmin(ctx, r.GetDB())
	}

	if namespaceListLikeReq.Search != "" {
		db = db.Where("name LIKE ?", "%"+namespaceListLikeReq.Search+"%").
			Or("tag LIKE ?", "%"+namespaceListLikeReq.Search+"%").
			Or("description LIKE ?", "%"+namespaceListLikeReq.Search+"%")
	}
	err := db.Debug().Order("id asc").Find(&namespaces).Error
	return namespaces, err
}

// GetByTagAndNamespaceTag 根据环境Tag和其所属的命名空间Tag获取环境信息
func (r *NamespaceRepository) GetByTagAndNamespaceTag(nsTag, envTag string) (*admin_model.Environment, error) {
	var environment admin_model.Environment

	err := r.GetDB().
		Preload("Namespace").
		Preload("Drivers").
		Where("tag = ? AND namespace_tag = ?", envTag, nsTag).
		First(&environment).Error

	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// CountAll 获取所有命名空间数量
func (r *NamespaceRepository) CountAll(ctx *gin.Context, namespaceListReq *dto.NamespaceListReq) (int64, error) {
	var count int64

	db := r.GetDB()
	if namespaceListReq.Name != "" {
		db = db.Where("name LIKE ?", "%"+namespaceListReq.Name+"%")
	}
	if namespaceListReq.Tag != "" {
		db = db.Where("tag LIKE ?", "%"+namespaceListReq.Tag+"%")
	}
	if namespaceListReq.PageTime != nil {
		if namespaceListReq.StartTime != nil {
			db = db.Where("create_time >= ?", namespaceListReq.StartTime)
		}
		if namespaceListReq.EndTime != nil {
			db = db.Where("create_time <= ?", namespaceListReq.EndTime)
		}
	}
	if namespaceListReq.Description != "" {
		db = db.Where("description LIKE ?", "%"+namespaceListReq.Description+"%")
	}
	if namespaceListReq.CreatedBy != "" {
		db = db.Where("created_by = ?", namespaceListReq.CreatedBy)
	}
	err := r.GetDB().Debug().Model(&admin_model.Namespace{}).Count(&count).Error
	return count, err
}

// GetByTag 根据Tag获取命名空间
func (r *NamespaceRepository) GetByTag(tag string) (*admin_model.Namespace, error) {
	var ns admin_model.Namespace
	err := r.GetDB().Where("tag = ?", tag).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ns, nil
}

// GetById 根据Id获取命名空间
func (r *NamespaceRepository) GetById(id uint) (*admin_model.Namespace, error) {
	var ns admin_model.Namespace
	err := r.GetDB().Where("id = ?", id).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ns, nil
}

// Update 修改命名空间
func (r *NamespaceRepository) Update(ns *admin_model.Namespace) error {
	return r.GetDB().Model(ns).Updates(ns).Error
}

// Create 创建命名空间
func (r *NamespaceRepository) Create(ns *admin_model.Namespace) error {
	return r.DB.Model(ns).Save(ns).Error
}

// GetByNamespaceTagAndEnvTag 根据命名空间Tag和环境Tag获取命名空间及指定环境
func (r *NamespaceRepository) GetByNamespaceTagAndEnvTag(namespaceTag, envTag string) (*admin_model.Namespace, error) {
	var ns admin_model.Namespace

	err := r.GetDB().
		Preload("Environments", "tag = ?", envTag).
		Preload("Environments.Drivers").
		Where("tag = ?", namespaceTag).
		First(&ns).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &ns, nil
}
