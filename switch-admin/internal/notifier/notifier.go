package notifier

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/fatzeng/switch-admin/internal/admin_driver"
	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-admin/internal/utils"
	"gitee.com/fatzeng/switch-admin/internal/ws"
	"gitee.com/fatzeng/switch-components/drivers"
	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-core/transmit"
	"golang.org/x/sync/errgroup"
)

type service struct {
	namespaceRepo *repository.NamespaceRepository
}

var defaultService *service

func Init() {
	defaultService = &service{
		namespaceRepo: repository.NewNamespaceRepository(),
	}
	logger.Logger.Info("The dynamic notification service has been initialized")
}

// PreloadDrivers 加载所有的驱动实例
func PreloadDrivers(ctx context.Context) (err error) {
	if defaultService == nil {
		err = fmt.Errorf("notification service has not been initialized, unable to preload admin_driver")
		logger.Logger.Error(err)
		return
	}

	logger.Logger.Info("Start preloading notification drivers for all active namespace-environment bindings...")
	relations, err := defaultService.namespaceRepo.GetAll(ctx, nil)
	if err != nil {
		err = fmt.Errorf("preloading admin_driver failed: unable to obtain active environment list: %w", err)
		logger.Logger.Error(err)
		return
	}

	var loadedCount int64
	var failedCount int64
	var wg sync.WaitGroup

	for _, rel := range relations {
		for _, env := range rel.Environments {
			if env.Publish == nil || !*env.Publish {
				continue
			}
			// 以环境为单位 加速驱动的加载
			wg.Add(1)
			go func(namespaceTag, envTag string, drivers []*model.Driver) {
				defer wg.Done()

				if count, err := LoadDrivers(ctx, namespaceTag, envTag, drivers); err != nil {
					logger.Logger.Warnf("Preloading drivers for namespace %s environment %s failed: %v", namespaceTag, envTag, err)
					atomic.AddInt64(&failedCount, 1)
				} else {
					atomic.AddInt64(&loadedCount, int64(count))
					logger.Logger.Debugf("Successfully preloaded %d drivers for namespace %s environment %s", count, namespaceTag, envTag)
				}
			}(rel.Tag, env.Tag, env.Drivers)
		}
	}

	wg.Wait()

	if failedCount > 0 {
		logger.Logger.Warnf("Notification admin_driver preloading completed with %d failures, successfully loaded %d admin_driver instances", failedCount, loadedCount)
	} else {
		logger.Logger.Infof("Notification admin_driver preloading completed successfully, loaded %d admin_driver instances", loadedCount)
	}
	return nil
}

// Notify 发送通知 必须都成功
func Notify(nsTag, envTag string, switchData *model.SwitchModel) error {
	if defaultService == nil {
		return fmt.Errorf("notifier service has not been initialized")
	}
	relation, err := defaultService.namespaceRepo.GetByTagAndNamespaceTag(nsTag, envTag)
	if err != nil {
		err = fmt.Errorf("failed to get environment details: %w", err)
		logger.Logger.Error(err)
		return err
	}

	if len(relation.Drivers) <= 0 {
		logger.Logger.Warnf("Environment '%s' has no drivers configured, skipping notification.", relation.Name)
		return nil
	}

	err = NotifyDrivers(relation.Namespace.Tag, relation.Tag, relation.Drivers, func(driverInterface driver.Driver, driverModel *model.Driver) error {
		notifier, ok := driverInterface.(transmit.Notifier)
		if !ok {
			logger.Logger.Warnf("Driver '%s' (Type: %s) does not implement the Notifier interface. nsTag: %v, envTag: %v.", driverModel.Name, driverModel.DriverType, nsTag, envTag)
			return fmt.Errorf("admin_driver '%s' does not implement Notifier interface", driverModel.Name)
		}

		if err = notifier.Notify(context.Background(), switchData); err != nil {
			logger.Logger.Errorf("Failed to send notification with admin_driver '%s' (ID:%d): %v", driverModel.Name, driverModel.ID, err)
			return err
		}

		return nil
	})
	if err != nil {
		logger.Logger.Errorf("Failed to process drivers for notification: %v", err)
		return fmt.Errorf("failed to process drivers: %w", err)
	}
	return nil
}

// LoadDrivers 加载Producer类型的驱动（用于预加载）
func LoadDrivers(ctx context.Context, namespaceTag, envTag string, drivers []*model.Driver) (int, error) {
	// admin服务驱动启动加载 必须慎重 不能忽略异常
	count := 0
	err := syncProducerDrivers(ctx, namespaceTag, envTag, drivers, DriverOptions{ContinueOnError: false})
	return count, err
}

// RefreshDrivers 刷新Producer类型的驱动（用于环境发布）
func RefreshDrivers(ctx context.Context, namespaceTag, envTag string, drivers []*model.Driver) error {
	// admin服务驱动刷新 必须慎重 不能忽略异常
	if err := syncProducerDrivers(ctx, namespaceTag, envTag, drivers, DriverOptions{ContinueOnError: false}); err != nil {
		return err
	}

	// 客户端驱动增量配置推送
	consumerDrivers := utils.FilterDriversByUsage(drivers, model.Consumer)
	configPayloads := utils.BuildDriverConfigPayloads(consumerDrivers)

	// 增量配置驱动消息推送
	if len(configPayloads) > 0 {
		logger.Logger.Infof("Broadcasting %d driver config changes for namespace: %s, env: %s",
			len(configPayloads), namespaceTag, envTag)
		// 服务端过滤 只发送给正确的客户端
		return ws.BroadcastDriverConfigIncrementChange(configPayloads, pc.WsEndpointChangeConfig, pc.AndFilter(ws.EndpointMatch(pc.WsEndpointChangeConfig), ws.EnvMatch(envTag), ws.NSMatch(namespaceTag)))
	}

	logger.Logger.Infof("No consumer drivers to broadcast for namespace: %s, env: %s", namespaceTag, envTag)
	return nil
}

// NotifyDrivers 处理Producer类型的驱动（用于通知等需要额外处理的场景）
func NotifyDrivers(namespaceTag, envTag string, drivers []*model.Driver, processF func(driverInterface driver.Driver, driverModel *model.Driver) error) error {
	return processDriversInternal(namespaceTag, envTag, drivers, processF, DriverOptions{ContinueOnError: true})
}

// processDriversInternal 内部通用的驱动处理函数
func processDriversInternal(namespaceTag, envTag string, drivers []*model.Driver, processF func(driverInterface driver.Driver, driverModel *model.Driver) error, opts ...DriverOptions) error {
	// 检查是否启用容错模式
	continueOnError := false
	for _, opt := range opts {
		if opt.ContinueOnError {
			continueOnError = true
			break
		}
	}

	// 多驱动加速开关的发布
	accelerate := errgroup.Group{}
	var lastError error
	for _, driverModel := range drivers {
		// 筛选驱动，只处理 Producer 类型的驱动
		if driverModel.Usage != model.Producer {
			continue
		}

		driverInterface, err := getDriver(namespaceTag, envTag, driverModel)
		if err != nil {
			if continueOnError {
				logger.Logger.Warnf("Failed to get admin_driver instance [%s]: %v, continue processing other drivers", driverModel.Name, err)
				lastError = err
				continue
			}
			return fmt.Errorf("failed to get admin_driver instance [%s]: %w", driverModel.Name, err)
		}

		if processF != nil {
			accelerate.Go(func() error {
				if err = processF(driverInterface, driverModel); err != nil {
					if continueOnError {
						logger.Logger.Warnf("驱动处理失败 [%s]: %v，继续处理其他驱动", driverModel.Name, err)
						return nil
					}
					return fmt.Errorf("驱动处理失败 [%s]: %w", driverModel.Name, err)
				}
				return nil
			})
		}
	}

	if err := accelerate.Wait(); err != nil {
		return err
	}

	return lastError
}

// filterAndValidateProducerDrivers 筛选和验证Producer驱动
func filterAndValidateProducerDrivers(allDrivers []*model.Driver, continueOnError bool) ([]*model.Driver, []map[string]interface{}, error) {
	// 使用通用筛选函数
	producerDrivers := utils.FilterDriversByUsage(allDrivers, model.Producer)

	producerModels := make([]*model.Driver, 0)
	driverMaps := make([]map[string]interface{}, 0)

	for _, driverModel := range producerDrivers {
		driverMap, err := driverModel.DriverConfig.ToMap()
		if err != nil {
			logger.Logger.Errorf("Failed to convert driver config to map for driver %s (ID: %d, Type: %s): %v",
				driverModel.Name, driverModel.ID, driverModel.DriverType, err)
			if continueOnError {
				logger.Logger.Warnf("Skipping driver %s due to config conversion error in continue-on-error mode", driverModel.Name)
				continue
			}
			return nil, nil, fmt.Errorf("driver config conversion failed for %s: %w", driverModel.Name, err)
		}

		producerModels = append(producerModels, driverModel)
		driverMaps = append(driverMaps, driverMap)
	}

	return producerModels, driverMaps, nil
}

// buildReplacementConfigs 构建驱动替换配置
func buildReplacementConfigs(producerModels []*model.Driver, driverMaps []map[string]interface{}, namespaceTag, envTag string) ([]*driver.DriverReplacementConfig, []driver.DriverFactory, error) {
	replacementConfigs := make([]*driver.DriverReplacementConfig, 0)
	driverFactories := make([]driver.DriverFactory, 0)

	for i, driverModel := range producerModels {
		driverMap := driverMaps[i]
		driverName := buildDriverInstanceName(namespaceTag, envTag, driverModel)

		switch driverModel.DriverType {
		case drivers.KafkaProducerDriverType:
			// 获取Kafka Producer配置
			validationTimeout, err := config.GlobalConfig.Replacement.GetKafkaProducerValidationTimeout()
			if err != nil {
				logger.Logger.Warnf("Failed to get kafka producer validation timeout: %v, using default", err)
				validationTimeout = 30 * time.Second
			}
			stabilityPeriod, err := config.GlobalConfig.Replacement.GetKafkaProducerStabilityPeriod()
			if err != nil {
				logger.Logger.Warnf("Failed to get kafka producer stability period: %v, using default", err)
				stabilityPeriod = 5 * time.Second
			}

			replacementConfigs = append(replacementConfigs, &driver.DriverReplacementConfig{
				DriverType:        driverModel.DriverType,
				DriverName:        driverName,
				ValidationTimeout: validationTimeout,
				StabilityPeriod:   stabilityPeriod,
				ConfigComparator: func(oldDriver, newDriver driver.Driver) bool {
					return drivers.KafkaProducerConfigComparatorWithOptions(oldDriver, newDriver, drivers.KafkaConfigCompareOptions{
						StrictBrokerOrder: config.GlobalConfig.Replacement.GetKafkaProducerVerifyBrokers(),
					})
				},
				SkipIfSameConfig: true,
				SupportedTypes:   drivers.SupportedTypes,
			})

			driverFactories = append(driverFactories, func() (driver.Driver, error) {
				return admin_driver.KafkaProducerDriver(driverMap)
			})
		case drivers.WebhookProducerDriverType:
			// 获取Webhook Producer配置
			validationTimeout, err := config.GlobalConfig.Replacement.GetWebhookProducerValidationTimeout()
			if err != nil {
				logger.Logger.Warnf("Failed to get webhook producer validation timeout: %v, using default", err)
				validationTimeout = 30 * time.Second
			}
			stabilityPeriod, err := config.GlobalConfig.Replacement.GetWebhookProducerStabilityPeriod()
			if err != nil {
				logger.Logger.Warnf("Failed to get webhook producer stability period: %v, using default", err)
				stabilityPeriod = 3 * time.Second
			}

			replacementConfigs = append(replacementConfigs, &driver.DriverReplacementConfig{
				DriverType:        driverModel.DriverType,
				DriverName:        driverName,
				ValidationTimeout: validationTimeout,
				StabilityPeriod:   stabilityPeriod,
				ConfigComparator:  drivers.WebhookProducerConfigComparator,
				SkipIfSameConfig:  true,
				SupportedTypes:    drivers.SupportedTypes,
			})
			driverFactories = append(driverFactories, func() (driver.Driver, error) {
				return admin_driver.WebhookProducerDriver(driverMap)
			})
		case drivers.PollingProducerDriverType:
			// 获取Polling Producer配置
			validationTimeout, err := config.GlobalConfig.Replacement.GetPollingProducerValidationTimeout()
			if err != nil {
				logger.Logger.Warnf("Failed to get polling producer validation timeout: %v, using default", err)
				validationTimeout = 30 * time.Second
			}
			stabilityPeriod, err := config.GlobalConfig.Replacement.GetPollingProducerStabilityPeriod()
			if err != nil {
				logger.Logger.Warnf("Failed to get polling producer stability period: %v, using default", err)
				stabilityPeriod = 3 * time.Second
			}

			replacementConfigs = append(replacementConfigs, &driver.DriverReplacementConfig{
				DriverType:        driverModel.DriverType,
				DriverName:        driverName,
				ValidationTimeout: validationTimeout,
				StabilityPeriod:   stabilityPeriod,
				ConfigComparator:  drivers.PollingProducerConfigComparator,
				SkipIfSameConfig:  true,
				SupportedTypes:    drivers.SupportedTypes,
			})
			driverFactories = append(driverFactories, func() (driver.Driver, error) {
				return admin_driver.PollingProducerDriver(driverMap)
			})
		default:
			logger.Logger.Warnf("Unsupported driver type: %v", driverModel.DriverType)
		}
	}

	return replacementConfigs, driverFactories, nil
}

// syncProducerDrivers 同步Producer驱动（用于驱动加载和刷新）
func syncProducerDrivers(ctx context.Context, namespaceTag, envTag string, allDrivers []*model.Driver, opts ...DriverOptions) error {
	// 检查是否启用容错模式
	continueOnError := false
	for _, opt := range opts {
		if opt.ContinueOnError {
			continueOnError = true
			break
		}
	}

	// 筛选和验证Producer驱动
	producerModels, driverMaps, err := filterAndValidateProducerDrivers(allDrivers, continueOnError)
	if err != nil {
		return err
	}

	if len(producerModels) <= 0 {
		logger.Logger.Infof("No producer drivers found for namespace '%s' environment '%s', skipping batch replacement",
			namespaceTag, envTag)
		return nil
	}

	// 构建驱动替换配置
	replacementConfigs, driverFactories, err := buildReplacementConfigs(producerModels, driverMaps, namespaceTag, envTag)
	if err != nil {
		return err
	}

	if len(replacementConfigs) <= 0 {
		logger.Logger.Warnf("Unable to find the driver configuration set to be replaced")
		// 有驱动但是找不到驱动配置 不合法
		return fmt.Errorf("unable to find the driver configuration set to be replaced")
	}

	// 执行批量驱动替换
	replacer := driver.NewGracefulDriverReplacerWithConfig(continueOnError)
	return replacer.BatchReplaceParallel(ctx, replacementConfigs, driverFactories)
}

// buildDriverInstanceName 构建驱动实例名称(对照sdk-go中名称保持一致)
func buildDriverInstanceName(namespaceTag, envTag string, d *model.Driver) string {
	return fmt.Sprintf("%s_%s_%s_%d", namespaceTag, envTag, d.DriverType, d.ID)
}

// validateDriverParams 验证驱动参数
func validateDriverParams(namespaceTag, envTag string, d *model.Driver) error {
	if namespaceTag == "" || envTag == "" {
		return fmt.Errorf("invalid namespaceTag, envTag configuration")
	}
	if d == nil || d.DriverType == "" || d.DriverConfig == nil || len(d.DriverConfig) <= 0 {
		return fmt.Errorf("invalid admin_driver configuration")
	}
	return nil
}

// DriverOptions 驱动操作选项
type DriverOptions struct {
	//忽略错误 刷新驱动(忽略则不影响后续驱动刷新) 消息推送(忽略则不影响后续消息发送)
	ContinueOnError bool
}

// getDriver 获取驱动
func getDriver(namespaceTag, envTag string, d *model.Driver) (driver.Driver, error) {
	if err := validateDriverParams(namespaceTag, envTag, d); err != nil {
		return nil, err
	}

	driverName := buildDriverInstanceName(namespaceTag, envTag, d)

	return driver.GetManager().GetDriver(d.DriverType, driverName)
}
