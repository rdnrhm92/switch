package driver

import (
	"context"
	"io"
)

// DriverType 驱动类型
type DriverType string

// DriverFailureCallback 驱动故障回调函数
type DriverFailureCallback func(failReason string, err error)

// Driver 驱动接口
// 支持按照驱动类型关闭所有驱动跟驱动名关闭驱动
// sync.once加载时机
type Driver interface {
	io.Closer
	// Validate 驱动有效性检查接口
	Validate(driver Driver) error
	// Start 启动驱动
	Start(ctx context.Context) error
	// RecreateFromConfig 回滚策略 跟Creator保持一致
	RecreateFromConfig() (Driver, error)
	// SetFailureCallback 设置故障回调函数
	SetFailureCallback(callback DriverFailureCallback)
	// GetDriverType 获取驱动类型
	GetDriverType() DriverType
	// GetDriverName 获取驱动名称
	GetDriverName() string
	// SetDriverMeta 设置驱动元信息
	SetDriverMeta(name string)
}

// Creator 驱动创建器函数类型
// 业务方传递配置创建驱动
type Creator func() (Driver, error)
