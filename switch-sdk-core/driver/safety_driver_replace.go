package driver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// ConfigComparator 配置比较函数类型
type ConfigComparator func(oldDriver, newDriver Driver) bool

// DriverReplacementConfig 驱动替换配置
type DriverReplacementConfig struct {
	DriverType        DriverType          // 驱动类型
	DriverName        string              // 驱动名称
	ValidationTimeout time.Duration       // 验证超时时间
	StabilityPeriod   time.Duration       // 稳定期等待时间
	ConfigComparator  ConfigComparator    // 配置比较用于判断是否需要替换驱动
	SkipIfSameConfig  bool                // 是否在配置相同时跳过替换
	SupportedTypes    map[DriverType]bool // 支持的驱动列表
}

// DriverFactory 驱动工厂函数类型
type DriverFactory = Creator

// GracefulDriverReplacer 驱动替换器，确保新驱动没有问题再关闭旧驱动
type GracefulDriverReplacer struct {
	// 是否忽略错误(用于批量替换驱动的时候)
	IgnoreErrors bool
}

func NewGracefulDriverReplacer() *GracefulDriverReplacer {
	return &GracefulDriverReplacer{
		IgnoreErrors: false,
	}
}

// NewGracefulDriverReplacerWithConfig 创建带配置的驱动替换器
func NewGracefulDriverReplacerWithConfig(ignoreErrors bool) *GracefulDriverReplacer {
	return &GracefulDriverReplacer{
		IgnoreErrors: ignoreErrors,
	}
}

// ReplaceDriver 执行优雅替换
func (r *GracefulDriverReplacer) ReplaceDriver(ctx context.Context, config *DriverReplacementConfig, factory DriverFactory) (error, Driver) {
	// 停机替换配置验证
	if err := r.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err), nil
	}

	// 创建新驱动(但不启动)
	logger.Logger.Infof("Creating new driver for %s but not start.", config.DriverName)
	newDriver, err := factory()
	if err != nil {
		return fmt.Errorf("factory validation failed: %w", err), nil
	}

	// 检查是否需要替换（配置比较）
	if config.SkipIfSameConfig && config.ConfigComparator != nil {
		if oldDriver, err := GetManager().GetDriver(config.DriverType, config.DriverName); err == nil {
			logger.Logger.Infof("Checking if replacement is needed for %s", config.DriverName)

			// 比较配置 配置相同关闭临时驱动 直接返回
			if config.ConfigComparator(oldDriver, newDriver) {
				logger.Logger.Infof("Driver %s configuration is identical, skipping replacement", config.DriverName)
				// 关闭临时创建的新驱动(关闭一次更保险)
				if closeErr := newDriver.Close(); closeErr != nil {
					logger.Logger.Warnf("Failed to close temporary driver: %v", closeErr)
				}
				return nil, oldDriver
			}

			logger.Logger.Infof("Driver %s configuration differs, proceeding with replacement", config.DriverName)
		}
	}

	// 验证新驱动(驱动配置有效性验证)
	if err = r.validateDriver(newDriver, config.ValidationTimeout); err != nil {
		// 清理失败的驱动(关闭一次更保险,但是不需要清理)
		newDriver.Close()
		return fmt.Errorf("driver validation failed: %w", err), nil
	}

	// 生成临时驱动名称并注册
	tempDriverName := r.generateTempName(config.DriverName)
	if err = r.registerTempDriver(config.DriverType, tempDriverName, newDriver); err != nil {
		// 注册失败，可能已经注册。需要清理
		r.cleanup(config.DriverType, tempDriverName)
		return fmt.Errorf("failed to register temp driver: %w", err), nil
	}

	// 检查并关闭旧驱动（释放端口等资源）
	var oldDriver Driver
	if oldDriver, err = GetManager().GetDriver(config.DriverType, config.DriverName); err == nil {
		logger.Logger.Infof("Closing old driver %s before starting new one", config.DriverName)
		if closeErr := GetManager().CloseByName(config.DriverType, config.DriverName); closeErr != nil {
			logger.Logger.Warnf("Old driver %s close failed but removed from registry: %v (continuing replacement)", config.DriverName, closeErr)
		} else {
			logger.Logger.Infof("Successfully closed old driver: %s", config.DriverName)
		}
	} else {
		logger.Logger.Infof("Old driver %s not found, proceeding with new driver", config.DriverName)
	}

	// 启动新驱动（此时旧驱动已关闭，端口已释放）
	logger.Logger.Infof("Starting new driver %s", tempDriverName)
	if err = newDriver.Start(ctx); err != nil {
		// 启动失败，清理临时驱动并回滚
		logger.Logger.Errorf("Failed to start new driver %s: %v", tempDriverName, err)
		r.cleanup(config.DriverType, tempDriverName)

		// 回滚旧驱动
		if oldDriver != nil {
			rollbackOldDriver, regErr := GetManager().Register(config.DriverType, config.DriverName, oldDriver.RecreateFromConfig, RegisterOption{UseSyncOnce: false})
			if regErr != nil {
				logger.Logger.Errorf("New driver start failed: %v but! Restore the old driver to fail: %v", err, regErr)
				r.cleanup(config.DriverType, config.DriverName)
				return fmt.Errorf("new driver start failed: %v but! Restore the old driver to fail: %v", err, regErr), nil
			}
			logger.Logger.Warnf("New driver start failed but! Restore the old driver to success: %v", err)
			rollbackOldDriver.Start(ctx)
			return fmt.Errorf("new driver start failed but! Restore the old driver to success: %v", err), nil
		}
		return fmt.Errorf("failed to start new driver and no old driver to rollback: %w", err), nil
	}

	// 稳定期等待
	if config.StabilityPeriod > 0 {
		if err = r.waitForStability(newDriver, config.StabilityPeriod); err != nil {
			// 回滚：关闭新驱动，旧驱动使用策略重新启动
			stabilityErr := err
			r.cleanup(config.DriverType, tempDriverName)

			// 新驱动稳定期失败 回滚老驱动 使用回滚的方式而不是再次start oldDriver 因为无法保证所有驱动的start的幂等
			if oldDriver != nil {
				// 新驱动等待期等待失败 恢复回滚的老驱动
				// 将回滚的驱动注册回管理器
				rollbackOldDriver, regErr := GetManager().Register(config.DriverType, config.DriverName, oldDriver.RecreateFromConfig, RegisterOption{UseSyncOnce: false})
				if regErr != nil {
					logger.Logger.Errorf("New driver stable period waiting for failure: %v but! Restore the old driver to fail: %v", stabilityErr, regErr)
					r.cleanup(config.DriverType, config.DriverName)
					return fmt.Errorf("new driver stable period waiting for failure: %v but! Restore the old driver to fail: %v", stabilityErr, regErr), nil
				}
				// 回滚成功，重新注册并启动旧驱动
				logger.Logger.Warnf("New driver stable period waiting for failure but! Restore the old driver to success: %v", stabilityErr)
				rollbackOldDriver.Start(ctx)
				return fmt.Errorf("new driver stable period waiting for failure but! Restore the old driver to success: %v", stabilityErr), nil
			} else {
				logger.Logger.Errorf("New driver stable period waiting for failure but! Unable to find old driver: %v", stabilityErr)
				return fmt.Errorf("new driver stable period waiting for failure but! Unable to find old driver: %v", stabilityErr), nil
			}
		}
	}

	// 用临时驱动替换目标驱动（只是重命名）
	if err = r.replaceDriverWithTemp(config.DriverType, tempDriverName, config.DriverName); err != nil {
		// 替换失败，清理临时驱动
		r.cleanup(config.DriverType, tempDriverName)
		return fmt.Errorf("failed to replace driver: %w", err), nil
	}

	logger.Logger.Infof("Successfully replaced driver: %s", config.DriverName)
	return nil, newDriver
}

// 生成临时驱动名称
func (r *GracefulDriverReplacer) generateTempName(originalName string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_temp_%d", originalName, timestamp)
}

// 验证驱动
func (r *GracefulDriverReplacer) validateDriver(driver Driver, timeout time.Duration) error {
	// 如果验证超时表明验证失败
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- driver.Validate(driver)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("validation timeout after %v", timeout)
	}
}

// 等待稳定期
func (r *GracefulDriverReplacer) waitForStability(driver Driver, period time.Duration) error {
	logger.Logger.Infof("Waiting for v stability period: %v", period)
	time.Sleep(period)
	return nil
}

// 清理失败的驱动
func (r *GracefulDriverReplacer) cleanup(driverType DriverType, driverName string) {
	if err := GetManager().CloseByName(driverType, driverName); err != nil {
		// 可能注册根本没成功
		logger.Logger.Debugf("Cleanup driver %s (may not exist): %v", driverName, err)
	} else {
		logger.Logger.Infof("Successfully cleaned up failed driver: %s", driverName)
	}
}

// 关闭旧驱动
func (r *GracefulDriverReplacer) closeOldDriver(driverType DriverType, driverName string) error {
	return GetManager().CloseByName(driverType, driverName)
}

// registerTempDriver 注册临时驱动
func (r *GracefulDriverReplacer) registerTempDriver(driverType DriverType, tempName string, driverInstance Driver) error {
	creator := func() (Driver, error) {
		return driverInstance, nil
	}

	opts := RegisterOption{UseSyncOnce: false}
	_, err := GetManager().Register(driverType, tempName, creator, opts)
	if err != nil {
		return fmt.Errorf("failed to register temp driver %s: %w", tempName, err)
	}

	logger.Logger.Infof("Temporary driver registered: %s", tempName)
	return nil
}

// replaceDriverWithTemp 用临时驱动替换目标驱动
func (r *GracefulDriverReplacer) replaceDriverWithTemp(driverType DriverType, tempDriverName, targetDriverName string) error {
	// 注意：旧驱动已经在 ReplaceDriver 中关闭了，这里只需要重命名
	// 临时驱动重命名 完成替换
	if err := GetManager().RenameDriver(driverType, tempDriverName, targetDriverName); err != nil {
		return fmt.Errorf("failed to rename driver from %s to %s: %w", tempDriverName, targetDriverName, err)
	}

	logger.Logger.Infof("Successfully replaced driver %s with temp driver %s", targetDriverName, tempDriverName)
	return nil
}

// validateDriverType 验证驱动类型是否支持
func (r *GracefulDriverReplacer) validateDriverType(supportedTypes map[DriverType]bool, driverType DriverType) error {
	if supportedTypes == nil || len(supportedTypes) <= 0 {
		return nil
	}
	if !supportedTypes[driverType] {
		return fmt.Errorf("unsupported driver type: %s", driverType)
	}
	return nil
}

// validateConfig 验证替换配置
func (r *GracefulDriverReplacer) validateConfig(config *DriverReplacementConfig) error {
	if config.DriverName == "" {
		return fmt.Errorf("driver name cannot be empty")
	}

	if err := r.validateDriverType(config.SupportedTypes, config.DriverType); err != nil {
		return err
	}

	if config.ValidationTimeout <= 0 {
		logger.Logger.Warn("Validation timeout not set, using default 30s")
		config.ValidationTimeout = 30 * time.Second
	}

	if config.StabilityPeriod < 0 {
		logger.Logger.Warn("StabilityPeriod timeout not set, using default 3s")
		config.StabilityPeriod = 3 * time.Second
	}

	return nil
}

// BatchReplaceParallel 批量替换驱动（并行）
func (r *GracefulDriverReplacer) BatchReplaceParallel(ctx context.Context, configs []*DriverReplacementConfig, factories []DriverFactory) error {
	if len(configs) != len(factories) {
		return fmt.Errorf("configs and factories length mismatch: %d vs %d", len(configs), len(factories))
	}

	totalCount := len(configs)
	logger.Logger.Infof("Starting batch driver replacement for %d drivers with IgnoreErrors %v", totalCount, r.IgnoreErrors)

	var mu sync.Mutex
	var errors []error
	var successCount int
	var stopFlag bool

	var wg sync.WaitGroup

	for i, config := range configs {
		wg.Add(1)
		go func(index int, cfg *DriverReplacementConfig, factory DriverFactory) {
			defer wg.Done()

			// 提前终止后续的驱动
			mu.Lock()
			if stopFlag {
				mu.Unlock()
				return
			}
			mu.Unlock()

			logger.Logger.Infof("Replacing driver %d/%d: %s", index+1, len(configs), cfg.DriverName)

			if err, _ := r.ReplaceDriver(ctx, cfg, factory); err != nil {
				errorMsg := fmt.Errorf("failed to replace driver %s: %w", cfg.DriverName, err)
				logger.Logger.Errorf("%v", errorMsg)

				mu.Lock()
				errors = append(errors, errorMsg)
				if !r.IgnoreErrors {
					stopFlag = true
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			logger.Logger.Infof("Successfully replaced driver %d/%d: %s", index+1, len(configs), cfg.DriverName)
		}(i, config, factories[i])
	}

	wg.Wait()

	if len(errors) > 0 && !r.IgnoreErrors {
		return fmt.Errorf("batch replacement completed with %d errors: %v", len(errors), errors)
	}

	logger.Logger.Infof("Concurrent batch driver replacement completed: %d/%d drivers replaced successfully, %d failed",
		successCount, totalCount, len(errors))
	return nil
}

// BatchReplaceSerial 批量替换(串行)
func (r *GracefulDriverReplacer) BatchReplaceSerial(ctx context.Context, configs []*DriverReplacementConfig, factories []DriverFactory) error {
	totalCount := len(configs)
	var errors []error
	var successCount int

	logger.Logger.Infof("Starting serial batch driver replacement for %d drivers with IgnoreErrors %v", totalCount, r.IgnoreErrors)

	for i, config := range configs {
		logger.Logger.Infof("Replacing driver %d/%d: %s", i+1, len(configs), config.DriverName)

		if err, _ := r.ReplaceDriver(ctx, config, factories[i]); err != nil {
			errorMsg := fmt.Errorf("failed to replace driver %s: %w", config.DriverName, err)
			errors = append(errors, errorMsg)
			logger.Logger.Errorf("%v", errorMsg)

			if !r.IgnoreErrors {
				return errorMsg
			}
			continue
		}

		successCount++
		logger.Logger.Infof("Successfully replaced driver %d/%d: %s", i+1, len(configs), config.DriverName)
	}

	if len(errors) > 0 && !r.IgnoreErrors {
		return fmt.Errorf("batch replacement completed with %d errors: %v", len(errors), errors)
	}

	logger.Logger.Infof("Serial batch driver replacement completed: %d/%d drivers replaced successfully, %d failed",
		successCount, totalCount, len(errors))
	return nil
}
