package model

import "gitee.com/fatzeng/switch-sdk-core/driver"

// UsageType 驱动大类别
type UsageType string

const (
	Producer UsageType = "producer" // 生产
	Consumer UsageType = "consumer" // 消费
)

// Driver 一个驱动实现
type Driver struct {
	CommonModel
	Name         string            `gorm:"size:255;not null;comment:驱动名称" json:"name"`
	Usage        UsageType         `gorm:"size:50;not null;comment:驱动用途" json:"usage"`
	DriverType   driver.DriverType `gorm:"size:255;not null;comment:驱动类型 (kafka, webhook)" json:"driverType"`
	DriverConfig JsonRaw           `gorm:"type:text;comment:驱动配置 (JSON格式, YAML格式(前端转json))" json:"driverConfig"`

	EnvironmentID uint `gorm:"comment:所属环境ID" json:"environment_id"`
}

func (Driver) TableName() string {
	return "drivers"
}
