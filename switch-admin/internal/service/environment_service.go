package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/notifier"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-core/tool"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// EnvironmentService 提供了与环境相关的业务逻辑
type EnvironmentService struct {
	envRepo    *repository.EnvironmentRepository
	driverRepo *repository.DriverRepository
}

func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{
		envRepo:    repository.NewEnvironmentRepository(),
		driverRepo: repository.NewDriverRepository(),
	}
}

// GetEnvironmentsByNamespace 获取指定命名空间下的所有环境
func (s *EnvironmentService) GetEnvironmentsByNamespace(namespaceTag string) ([]*admin_model.Environment, error) {
	return s.envRepo.GetByNamespaceTag(namespaceTag)
}

// CreateUpdateEnvironment 创建/修改
func (s *EnvironmentService) CreateUpdateEnvironment(ctx *gin.Context, req *dto.EnvCreateUpdateReq) error {
	now := time.Now()
	userInfo := info.GetUserInfo(ctx)
	if userInfo == nil {
		return fmt.Errorf("not found userInfo")
	}
	if req.Id == 0 {
		existingEnv, err := s.envRepo.GetByTag(req.Tag)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check for existing environment: %w", err)
		}
		if existingEnv != nil {
			return fmt.Errorf("environment with tag '%s' already exists", req.Tag)
		}

		selectNamespace := req.SelectNamespace
		if selectNamespace == "" {
			selectNamespace = userInfo.SelectNamespace
		}

		env := &admin_model.Environment{
			Name:         req.Name,
			Tag:          req.Tag,
			Publish:      tool.Bool(req.Publish),
			Description:  req.Description,
			PublishOrder: req.PublishOrder,
			NamespaceTag: selectNamespace,
			CommonModel: model.CommonModel{
				CreatedBy:  req.CreateBy,
				CreateTime: &now,
			},
		}

		return s.envRepo.GetDB().Transaction(func(tx *gorm.DB) error {
			envRepoWithTx := s.envRepo.WithTx(tx)
			driverRepoWithTx := s.driverRepo.WithTx(tx)

			//创建基本信息
			if err = envRepoWithTx.Create(env); err != nil {
				return fmt.Errorf("failed to create environment: %w", err)
			}

			//创建驱动信息，直接设置 EnvironmentID
			drivers := make([]*model.Driver, 0, len(req.Drivers))
			for _, driver := range req.Drivers {
				drivers = append(drivers, &model.Driver{
					CommonModel: model.CommonModel{
						CreatedBy:  req.CreateBy,
						CreateTime: &now,
					},
					Name:          driver.Name,
					Usage:         driver.Usage,
					DriverType:    driver.DriverType,
					DriverConfig:  model.JsonRaw(driver.DriverConfig),
					EnvironmentID: env.ID,
				})
			}
			err = driverRepoWithTx.CreateInBatch(drivers)
			if err != nil {
				return fmt.Errorf("failed to create environment drivers: %w", err)
			}

			return nil
		})
	}

	env := &admin_model.Environment{
		Name:        req.Name,
		Description: req.Description,
		CommonModel: model.CommonModel{
			ID:         req.Id,
			UpdateBy:   req.UpdateBy,
			UpdateTime: &now,
		},
	}

	return s.envRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		envRepoWithTx := s.envRepo.WithTx(tx)
		driverRepoWithTx := s.driverRepo.WithTx(tx)

		//获取已知环境信息
		var existingEnv admin_model.Environment
		if err := envRepoWithTx.GetDB().Preload("Drivers").First(&existingEnv, req.Id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("environment with id %d not found", req.Id)
			}
			return fmt.Errorf("failed to get existing environment: %w", err)
		}

		//更新环境
		env.Publish = tool.Bool(false)
		if err := envRepoWithTx.Update(env); err != nil {
			return fmt.Errorf("failed to update environment basic info: %w", err)
		}

		//获取已经存在的驱动用于选择性更新
		dbDriversMap := make(map[uint]*model.Driver)
		for _, d := range existingEnv.Drivers {
			dbDriversMap[d.ID] = d
		}

		reqDriversMap := make(map[uint]bool)
		finalDrivers := make([]*model.Driver, 0, len(req.Drivers))
		var driversToUpdate []*model.Driver

		for _, reqDriver := range req.Drivers {
			reqDriversMap[reqDriver.Id] = true

			if reqDriver.Id > 0 && dbDriversMap[reqDriver.Id] != nil {
				//目前存在的
				dbDriver := dbDriversMap[reqDriver.Id]
				equal, err := dto.IsEqual(json.RawMessage(dbDriver.DriverConfig), reqDriver.DriverConfig)
				if err != nil {
					logger.Logger.Error(fmt.Errorf("failed to compare admin_driver configs: %w", err))
				}
				if dbDriver.Name != reqDriver.Name ||
					dbDriver.Usage != reqDriver.Usage ||
					dbDriver.DriverType != reqDriver.DriverType || !equal {
					dbDriver.Name = reqDriver.Name
					dbDriver.Usage = reqDriver.Usage
					dbDriver.DriverType = reqDriver.DriverType
					dbDriver.DriverConfig = model.JsonRaw(reqDriver.DriverConfig)
					dbDriver.UpdateBy = req.UpdateBy
					dbDriver.UpdateTime = &now
					driversToUpdate = append(driversToUpdate, dbDriver)
				}
				finalDrivers = append(finalDrivers, dbDriver)
			} else {
				//创建
				newDriver := &model.Driver{
					CommonModel: model.CommonModel{
						CreatedBy:  req.CreateBy,
						CreateTime: &now,
					},
					Name:          reqDriver.Name,
					Usage:         reqDriver.Usage,
					DriverType:    reqDriver.DriverType,
					DriverConfig:  model.JsonRaw(reqDriver.DriverConfig),
					EnvironmentID: req.Id,
				}
				if err := driverRepoWithTx.GetDB().Create(newDriver).Error; err != nil {
					return fmt.Errorf("failed to create new admin_driver: %w", err)
				}
				finalDrivers = append(finalDrivers, newDriver)
			}
		}

		//修改部分驱动
		if len(driversToUpdate) > 0 {
			if err := driverRepoWithTx.GetDB().Save(&driversToUpdate).Error; err != nil {
				return fmt.Errorf("failed to batch update drivers: %w", err)
			}
		}

		// 处理需要删除的驱动
		var driversToDelete []*model.Driver
		for dbDriverID, dbDriver := range dbDriversMap {
			if !reqDriversMap[dbDriverID] {
				driversToDelete = append(driversToDelete, dbDriver)
			}
		}

		// 删除不再需要的驱动
		if len(driversToDelete) > 0 {
			if err := driverRepoWithTx.GetDB().Delete(driversToDelete).Error; err != nil {
				return fmt.Errorf("failed to delete old drivers: %w", err)
			}
		}

		return nil
	})
}

func (s *EnvironmentService) GetEnvs(ctx *gin.Context, req *dto.EnvListReq) map[string]interface{} {
	batchGroup := errgroup.Group{}
	var allRes []*admin_model.Environment
	var countRes int64
	batchGroup.Go(func() error {
		all, err := s.envRepo.GetAll(ctx, req)
		if err != nil {
			return err
		}
		allRes = all
		return nil
	})
	batchGroup.Go(func() error {
		count, err := s.envRepo.CountAll(ctx, req)
		if err != nil {
			return err
		}
		countRes = count
		return nil
	})
	if err := batchGroup.Wait(); err != nil {
		logger.Logger.Errorf("[EnvironmentService] batch err: %v", err)
		return map[string]interface{}{
			"success": false,
			"total":   0,
			"data":    []interface{}{},
		}
	}

	return map[string]interface{}{
		"success": true,
		"total":   countRes,
		"data":    allRes,
	}
}

func (s *EnvironmentService) EnvPublish(ctx *gin.Context, req *dto.EnvPublishReq) error {
	now := time.Now()
	return repository.GetDB().Transaction(func(tx *gorm.DB) error {
		envRepoTx := s.envRepo.WithTx(tx)
		err := envRepoTx.Update(&admin_model.Environment{
			CommonModel: model.CommonModel{
				ID:         req.Id,
				UpdateBy:   req.UpdateBy,
				UpdateTime: &now,
			},
			Publish: tool.Bool(true),
		})
		if err != nil {
			logger.Logger.Errorf("[EnvironmentService.EnvPublish] Failed to update environment publish status, environment ID: %d, error: %v", req.Id, err)
			return fmt.Errorf("failed to update environment publish status: %w", err)
		}
		logger.Logger.Infof("[EnvironmentService.EnvPublish] Environment publish status updated successfully, environment ID: %d", req.Id)

		var env admin_model.Environment
		err = tx.Preload("Drivers").Preload("Namespace").Where("id = ?", req.Id).First(&env).Error
		if err != nil {
			logger.Logger.Errorf("[EnvironmentService.EnvPublish] Failed to query environment admin_driver information, environment ID: %d, error: %v", req.Id, err)
			return fmt.Errorf("failed to query environment admin_driver information: %w", err)
		}

		if env.Namespace == nil {
			logger.Logger.Errorf("[EnvironmentService.EnvPublish] No associated namespace")
			return fmt.Errorf("failed to query environment admin_driver information: no associated namespace")
		}

		// 不能用gin的上下文
		if err = notifier.RefreshDrivers(config.GlobalContext, env.Namespace.Tag, env.Tag, env.Drivers); err != nil {
			logger.Logger.Errorf("[EnvironmentService.EnvPublish] Driver initialization failed: %v", err)
			return fmt.Errorf("admin_driver initialization failed: %w", err)
		}

		logger.Logger.Infof("[EnvironmentService.EnvPublish] Environment publish completed, environment ID: %d, operator: %s", req.Id, req.UpdateBy)
		return nil
	})
}
