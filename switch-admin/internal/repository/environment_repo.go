package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EnvironmentRepository 定义了对 Environment 模型的数据库操作
type EnvironmentRepository struct {
	BaseRepository
}

func NewEnvironmentRepository() *EnvironmentRepository {
	return &EnvironmentRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *EnvironmentRepository) WithTx(tx *gorm.DB) *EnvironmentRepository {
	return WithTx(r, tx)
}

// Create 创建一个新环境
func (r *EnvironmentRepository) Create(env *admin_model.Environment) error {
	return r.DB.Create(env).Error
}

// CreateInBatch 批量创建环境
func (r *EnvironmentRepository) CreateInBatch(envs []*admin_model.Environment) error {
	if len(envs) == 0 {
		return nil
	}
	return r.GetDB().Create(&envs).Error
}

// GetAll 获取所有的环境配置
func (r *EnvironmentRepository) GetAll(ctx *gin.Context, req *dto.EnvListReq) ([]*admin_model.Environment, error) {
	var envs []*admin_model.Environment

	db := FindWithSuperAdminWithJoin(ctx, r.GetDB())

	query := db.Model(&admin_model.Environment{}).
		Preload("Drivers").
		Preload("Namespace")

	if req != nil {
		if req.Name != "" {
			query = query.Where("environments.name LIKE ?", "%"+req.Name+"%")
		}
		if req.Tag != "" {
			query = query.Where("environments.tag LIKE ?", "%"+req.Tag+"%")
		}
		if req.Description != "" {
			query = query.Where("environments.description LIKE ?", "%"+req.Description+"%")
		}
		if req.CreatedBy != "" {
			query = query.Where("environments.created_by LIKE ?", "%"+req.CreatedBy+"%")
		}
		if req.NamespaceTag != "" {
			query = query.Where("environments.namespace_tag = ?", req.NamespaceTag)
		}

		if req.PageTime != nil {
			if req.StartTime != nil {
				query = query.Where("environments.create_time >= ?", req.StartTime)
			}
			if req.EndTime != nil {
				query = query.Where("environments.create_time <= ?", req.EndTime)
			}
		}

		if req.PageLimit != nil {
			query = query.Offset(int(req.PageLimit.Offset)).Limit(int(req.PageLimit.Limit))
		}
	}

	err := query.Debug().Order("environments.update_time desc , environments.create_time desc ").Find(&envs).Error
	return envs, err
}

// CountAll 获取所有环境配置数量
func (r *EnvironmentRepository) CountAll(ctx *gin.Context, req *dto.EnvListReq) (int64, error) {
	var count int64

	db := FindWithSuperAdminWithJoin(ctx, r.GetDB())

	query := db.Model(&admin_model.Environment{})

	if req != nil {
		if req.Name != "" {
			query = query.Where("name LIKE ?", "%"+req.Name+"%")
		}
		if req.Tag != "" {
			query = query.Where("tag LIKE ?", "%"+req.Tag+"%")
		}
		if req.Description != "" {
			query = query.Where("description LIKE ?", "%"+req.Description+"%")
		}
		if req.CreatedBy != "" {
			query = query.Where("environments.created_by LIKE ?", "%"+req.CreatedBy+"%")
		}
		if req.NamespaceTag != "" {
			query = query.Where("namespace_tag = ?", req.NamespaceTag)
		}
		if req.PageTime != nil {
			if req.StartTime != nil {
				query = query.Where("create_time >= ?", req.StartTime)
			}
			if req.EndTime != nil {
				query = query.Where("create_time <= ?", req.EndTime)
			}
		}
	}

	err := query.Count(&count).Error
	return count, err
}

// GetByNamespaceTag 根据命名空间Tag获取环境
func (r *EnvironmentRepository) GetByNamespaceTag(namespaceTag string) ([]*admin_model.Environment, error) {
	var envs []*admin_model.Environment
	err := r.GetDB().
		Where("namespace_tag = ?", namespaceTag).
		Where("publish = 1").
		Order("publish_order asc").
		Find(&envs).Error
	return envs, err
}

// GetByID 根据ID获取环境
func (r *EnvironmentRepository) GetByID(id uint) (*admin_model.Environment, error) {
	var env admin_model.Environment
	err := r.GetDB().First(&env, id).Error
	return &env, err
}

// Update 更新环境信息
func (r *EnvironmentRepository) Update(env *admin_model.Environment) error {
	return r.GetDB().Model(env).Updates(env).Error
}

// GetByTag 根据tag去找环境
func (r *EnvironmentRepository) GetByTag(tag string) (*admin_model.Environment, error) {
	var env admin_model.Environment
	err := r.GetDB().Where("tag = ?", tag).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}
