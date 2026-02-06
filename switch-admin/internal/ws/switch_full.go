package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gorm.io/gorm"
)

// switchFull 获取指定命名空间和环境的全量开关数据
func switchFull(ctx context.Context, namespaceTag, envTag string) ([]*model.SwitchModel, error) {
	logger.Logger.Infof("Fetching switch full data for namespace: %s, env: %s", namespaceTag, envTag)

	cache := GetSwitchCache()
	if cache != nil {
		if cachedData, found := cache.Get(namespaceTag, envTag); found {
			logger.Logger.Infof("Cache hit: returning %d switches for namespace: %s, env: %s",
				len(cachedData), namespaceTag, envTag)
			return cachedData, nil
		}
	}

	logger.Logger.Infof("Cache miss: querying database for namespace: %s, env: %s", namespaceTag, envTag)

	namespaceRepository := repository.NewNamespaceRepository()
	namespace, err := namespaceRepository.GetByNamespaceTagAndEnvTag(namespaceTag, envTag)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Logger.Errorf("Failed to get namespace by tag %s: %v", namespaceTag, err)
		return nil, err
	}
	if namespace == nil {
		return nil, fmt.Errorf("unable to find the corresponding namespace based on %v, %v", namespaceTag, envTag)
	}
	if len(namespace.Environments) != 1 {
		return nil, fmt.Errorf("unable to find the corresponding environments based on %v, %v", namespaceTag, envTag)
	}

	allSwitches, err := fetchAllSwitchesFromDB(namespace)
	if err != nil {
		return nil, err
	}

	if cache != nil {
		cache.Set(namespaceTag, envTag, allSwitches)
	}

	logger.Logger.Infof("Database query completed: found %d switches for namespace: %s, env: %s",
		len(allSwitches), namespaceTag, envTag)
	return allSwitches, nil
}

// fetchAllSwitchesFromDB 从数据库获取所有开关数据
func fetchAllSwitchesFromDB(namespace *admin_model.Namespace) ([]*model.SwitchModel, error) {
	var allSwitches []*model.SwitchModel
	var lastID uint = 0
	const pageSize = 1000
	batchCount := 0
	namespaceTag := namespace.Tag
	envTag := namespace.Environments[0].Tag

	switchSnapshotRepository := repository.NewSwitchSnapshotConfigRepository()

	for {
		batchCount++
		logger.Logger.Debugf("Fetching batch %d with lastID: %d", batchCount, lastID)

		switchSnapshotBatch, err := switchSnapshotRepository.FindByNamespaceAndEnvWithCursor(namespaceTag, envTag, lastID, pageSize)
		if err != nil {
			logger.Logger.Errorf("Failed to find switches batch %d for namespace %s, env %s: %v",
				batchCount, namespaceTag, envTag, err)
			return nil, err
		}

		if len(switchSnapshotBatch) == 0 {
			logger.Logger.Debugf("No more switchSnapshot found, stopping at batch %d", batchCount)
			break
		}

		// 将快照转换为开关模型
		switchBatch, err := convertSnapshotsToSwitches(switchSnapshotBatch)
		if err != nil {
			logger.Logger.Errorf("Failed to convert snapshots to switches for batch %d: %v", batchCount, err)
			return nil, err
		}

		allSwitches = append(allSwitches, switchBatch...)

		lastID = switchSnapshotBatch[len(switchSnapshotBatch)-1].SwitchID

		logger.Logger.Debugf("Batch %d: fetched %d switches, total so far: %d, lastID: %d",
			batchCount, len(switchBatch), len(allSwitches), lastID)

		if len(switchSnapshotBatch) < pageSize {
			logger.Logger.Debugf("Last batch detected (size: %d < %d), stopping", len(switchSnapshotBatch), pageSize)
			break
		}
	}

	return allSwitches, nil
}

// convertSnapshotsToSwitches 批量转换 SwitchSnapshot 为 SwitchModel
func convertSnapshotsToSwitches(snapshots []*admin_model.SwitchSnapshot) ([]*model.SwitchModel, error) {
	switches := make([]*model.SwitchModel, 0, len(snapshots))

	for _, snapshot := range snapshots {
		var switchModel model.SwitchModel
		if err := json.Unmarshal(snapshot.CompleteJSON, &switchModel); err != nil {
			logger.Logger.Errorf("Failed to unmarshal SwitchModel for snapshot (ID: %d, SwitchID: %d): %v",
				snapshot.ID, snapshot.SwitchID, err)
			return nil, fmt.Errorf("failed to unmarshal SwitchModel from CompleteJSON: %w", err)
		}

		switches = append(switches, &switchModel)
	}

	return switches, nil
}
