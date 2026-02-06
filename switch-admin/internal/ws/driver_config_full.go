package ws

import (
	"context"

	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-admin/internal/utils"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// driverConfigFull 获取指定命名空间和环境的全量驱动配置数据
func driverConfigFull(ctx context.Context, namespaceTag, envTag string) ([]*model.Driver, error) {
	logger.Logger.Infof("Fetching driver config full data for namespace: %s, env: %s", namespaceTag, envTag)

	environmentRepository := repository.NewEnvironmentRepository()
	env, err := environmentRepository.GetByTag(envTag)
	if err != nil {
		logger.Logger.Errorf("Failed to get environment by tag %s: %v", envTag, err)
		return nil, err
	}
	driverRepository := repository.NewDriverRepository()
	allDrivers, err := driverRepository.GetByEnvironmentID(env.ID)
	if err != nil {
		logger.Logger.Errorf("Failed to get drivers for environment ID %d: %v", env.ID, err)
		return nil, err
	}

	filteredDrivers := utils.FilterDriversByUsage(allDrivers, model.Consumer)

	logger.Logger.Infof("Found %d consumer drivers (total: %d) for environment: %s",
		len(filteredDrivers), len(allDrivers), envTag)
	return filteredDrivers, nil
}
