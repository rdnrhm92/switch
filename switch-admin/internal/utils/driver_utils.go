package utils

import (
	"encoding/json"

	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// FilterDriversByUsage 通用的驱动筛选函数
func FilterDriversByUsage(allDrivers []*model.Driver, usage model.UsageType) []*model.Driver {
	var filteredDrivers []*model.Driver
	skippedCount := 0

	for _, driver := range allDrivers {
		if driver == nil {
			skippedCount++
			continue
		}
		if driver.Usage == usage {
			filteredDrivers = append(filteredDrivers, driver)
		}
	}

	logger.Logger.Debugf("Filtered drivers by usage %s: found %d, skipped %d null drivers, total input %d",
		usage, len(filteredDrivers), skippedCount, len(allDrivers))
	return filteredDrivers
}

// BuildDriverConfigPayloads 构建驱动配置载荷（用于WebSocket推送）
func BuildDriverConfigPayloads(drivers []*model.Driver) []*pc.DriverConfigPayload {
	var configPayloads []*pc.DriverConfigPayload
	skippedCount := 0
	failedCount := 0

	logger.Logger.Debugf("Building driver config payloads for %d drivers", len(drivers))

	for _, driver := range drivers {
		if driver == nil {
			skippedCount++
			continue
		}

		driverBytes, err := json.Marshal(driver)
		if err != nil {
			failedCount++
			logger.Logger.Errorf("Failed to marshal driver %s (ID: %d, Type: %s): %v",
				driver.Name, driver.ID, driver.DriverType, err)
			continue
		}

		configPayload := &pc.DriverConfigPayload{
			Type:   driver.DriverType,
			Config: json.RawMessage(driverBytes),
		}
		configPayloads = append(configPayloads, configPayload)

		logger.Logger.Debugf("Built config payload for driver %s (ID: %d, Type: %s)",
			driver.Name, driver.ID, driver.DriverType)
	}

	logger.Logger.Infof("Built %d driver config payloads (skipped %d null, failed %d serialization)",
		len(configPayloads), skippedCount, failedCount)
	return configPayloads
}
