package driver

import (
	"fmt"
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// RegisterOption 注册opts
type RegisterOption struct {
	UseSyncOnce bool
}

// driverWrapper 驱动包装器
type driverWrapper struct {
	once       sync.Once
	driver     Driver
	creator    Creator
	driverName string
}

// DriverManager 驱动管理器
type DriverManager struct {
	sync.RWMutex
	drivers map[DriverType]map[string]*driverWrapper
}

var (
	defaultManager = NewDriverManager()
)

// GetManager 获取默认驱动管理器
func GetManager() *DriverManager {
	return defaultManager
}

// NewDriverManager 创建新的驱动管理器
func NewDriverManager() *DriverManager {
	return &DriverManager{
		drivers: make(map[DriverType]map[string]*driverWrapper),
	}
}

// Register 注册驱动实例 使用opts自定义是否单例 默认单例 可取消
func (m *DriverManager) Register(driverType DriverType, name string, creator Creator, opts ...RegisterOption) (Driver, error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.drivers[driverType]; !ok {
		m.drivers[driverType] = make(map[string]*driverWrapper)
	}

	// 如果已存在，返回现有的驱动
	if wrapper, exists := m.drivers[driverType][name]; exists {
		if wrapper.driver != nil {
			return wrapper.driver, nil
		}
	}

	var opt = RegisterOption{
		UseSyncOnce: true,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}

	wrapper := &driverWrapper{
		creator: func() (Driver, error) {
			driver, err := creator()
			if err != nil {
				return nil, err
			}
			driver.SetDriverMeta(name)
			driver.SetFailureCallback(func(failReason string, err error) {
				logger.Logger.Errorf("Driver %v name %v triggered failure callback to remove from driver pool failReason: %v err: %v", driverType, name, failReason, err)
				// 异步清理 避免死锁
				go func() {
					_ = GetManager().CloseByName(driverType, name)
				}()
			})
			return driver, err
		},
		driverName: name,
		once:       sync.Once{},
	}

	//创建单利模式下的驱动实力
	if opt.UseSyncOnce {
		m.drivers[driverType][name] = wrapper
		return wrapper.onceDriver()
	}

	driver, err := wrapper.creator()
	if err != nil {
		return nil, fmt.Errorf("create driver error: %v", err)
	}
	wrapper.driver = driver
	m.drivers[driverType][name] = wrapper
	return driver, nil
}

// onceDriver
func (w *driverWrapper) onceDriver() (Driver, error) {
	var err error
	w.once.Do(func() {
		w.driver, err = w.creator()
	})
	if err != nil {
		return nil, fmt.Errorf("create driver error: %v", err)
	}
	return w.driver, nil
}

// getDriver 获取或创建驱动实例
func (w *driverWrapper) getDriver() (Driver, error) {
	if w.driver == nil {
		return nil, fmt.Errorf("not found [%v] driver please use Register", w.driverName)
	}
	return w.driver, nil
}

// GetDriver 获取驱动实例
func (m *DriverManager) GetDriver(driverType DriverType, name string) (Driver, error) {
	m.RLock()
	defer m.RUnlock()

	if driverMap, ok := m.drivers[driverType]; ok {
		if wrapper, exists := driverMap[name]; exists {
			return wrapper.driver, nil
		}
	}
	return nil, fmt.Errorf("driver not found: type=%s, name=%s", driverType, name)
}

// GetDriversByType 获取驱动实例(根据类型)
func (m *DriverManager) GetDriversByType(driverType DriverType) []Driver {
	m.RLock()
	defer m.RUnlock()

	drivers := make([]Driver, 0)
	if driverMap, ok := m.drivers[driverType]; ok {
		for _, wrapper := range driverMap {
			if wrapper.driver != nil {
				drivers = append(drivers, wrapper.driver)
			}
		}
	}
	return drivers
}

// GetOrCreate 获取或创建驱动实例
func (m *DriverManager) GetOrCreate(driverType DriverType, name string, creator Creator) (Driver, error) {
	m.RLock()
	if driverMap, ok := m.drivers[driverType]; ok {
		if wrapper, exists := driverMap[name]; exists {
			m.RUnlock()
			return wrapper.onceDriver()
		}
	}
	m.RUnlock()

	return m.Register(driverType, name, creator)
}

// CloseByType 关闭指定类型的所有驱动
func (m *DriverManager) CloseByType(driverType DriverType) error {
	m.Lock()
	defer m.Unlock()

	if driverMap, ok := m.drivers[driverType]; ok {
		var errs []error
		for name, wrapper := range driverMap {
			if wrapper.driver != nil {
				if err := wrapper.driver.Close(); err != nil {
					errs = append(errs, fmt.Errorf("close %s driver '%s' error: %v", driverType, name, err))
				}
			}
		}
		delete(m.drivers, driverType)
		if len(errs) > 0 {
			return fmt.Errorf("close drivers error: %v", errs)
		}
	}
	return nil
}

// CloseByName 关闭指定类型和名称的驱动
func (m *DriverManager) CloseByName(driverType DriverType, name string) error {
	m.Lock()
	defer m.Unlock()

	if driverMap, ok := m.drivers[driverType]; ok {
		if wrapper, exists := driverMap[name]; exists {
			var closeErr error
			if wrapper.driver != nil {
				closeErr = wrapper.driver.Close()
				if closeErr != nil {
					logger.Logger.Errorf("Failed to close driver %s/%s: %v (forcing removal from registry)", driverType, name, closeErr)
				}
			}
			// 无论关闭成功与否 都从容器中删除 避免产生僵尸驱动
			delete(driverMap, name)

			if closeErr != nil {
				return fmt.Errorf("close %s driver '%s' error (driver already removed): %w", driverType, name, closeErr)
			}
			return nil
		}
	}
	return fmt.Errorf("driver not found: type=%s, name=%s", driverType, name)
}

// RenameDriver 重命名驱动（不关闭驱动实例）
func (m *DriverManager) RenameDriver(driverType DriverType, oldName, newName string) error {
	m.Lock()
	defer m.Unlock()

	// 检查源驱动是否存在
	driverMap, ok := m.drivers[driverType]
	if !ok {
		// driverType类型的驱动都没有的话 没办法名字替换
		return fmt.Errorf("driver type not found: %s", driverType)
	}

	// 给老的改名。老的必须要存在
	wrapper, exists := driverMap[oldName]
	if !exists {
		return fmt.Errorf("source driver not found: type=%s, name=%s", driverType, oldName)
	}

	// 检查目标名称是否已存在
	if _, exists = driverMap[newName]; exists {
		logger.Logger.Warnf("target driver name already exists: type=%s, name=%s", driverType, newName)
		return nil
	}

	// 更新wrapper中的驱动名称
	wrapper.driverName = newName

	// 在map中进行重命名：添加新名称，删除旧名称
	driverMap[newName] = wrapper
	delete(driverMap, oldName)

	return nil
}

// ReplaceDriverWithInstance 用新驱动实例替换现有驱动（会关闭旧驱动）
func (m *DriverManager) ReplaceDriverWithInstance(driverType DriverType, name string, newDriver Driver, creator Creator) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.drivers[driverType]; !ok {
		m.drivers[driverType] = make(map[string]*driverWrapper)
	}

	driverMap := m.drivers[driverType]

	// 如果存在旧驱动，先关闭它
	if oldWrapper, exists := driverMap[name]; exists {
		if oldWrapper.driver != nil {
			if err := oldWrapper.driver.Close(); err != nil {
				return fmt.Errorf("failed to close old driver '%s': %v", name, err)
			}
		}
	}

	newDriver.SetDriverMeta(name)
	newDriver.SetFailureCallback(func(failReason string, err error) {
		logger.Logger.Errorf("Driver %v name %v triggered failure callback to remove from driver pool failReason: %v err: %v", driverType, name, failReason, err)
		// 异步清理 避免死锁
		go func() {
			_ = GetManager().CloseByName(driverType, name)
		}()
	})

	// 创建新的wrapper
	newWrapper := &driverWrapper{
		driver:     newDriver,
		creator:    creator,
		driverName: name,
		once:       sync.Once{},
	}

	// 设置新驱动
	driverMap[name] = newWrapper

	return nil
}

// Close 关闭所有驱动
func (m *DriverManager) Close() error {
	m.Lock()
	defer m.Unlock()

	var errs []error
	for driverType, driverMap := range m.drivers {
		for name, wrapper := range driverMap {
			if wrapper.driver != nil {
				if err := wrapper.driver.Close(); err != nil {
					errs = append(errs, fmt.Errorf("close %s driver '%s' error: %v", driverType, name, err))
				}
			}
		}
	}
	m.drivers = make(map[DriverType]map[string]*driverWrapper)

	if len(errs) > 0 {
		return fmt.Errorf("close drivers error: %v", errs)
	}
	return nil
}
