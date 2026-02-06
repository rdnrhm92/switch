package drivers

import (
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/driver"
)

// PollingProducerValidator 长轮询生产者驱动验证器
// 1.配置有效性检查
type PollingProducerValidator struct {
	pollingDriver *PollingProducer
}

// Validate 验证长轮询生产者驱动
func (p *PollingProducerValidator) Validate(driver driver.Driver) error {
	pollingDriver, ok := driver.(*PollingProducer)
	if !ok {
		return fmt.Errorf("admin_driver is not a PollingProducer")
	}
	p.pollingDriver = pollingDriver

	// 验证配置有效性
	if err := p.pollingDriver.config.isValid(); err != nil {
		return fmt.Errorf("polling producer config validation failed: %w", err)
	}

	return nil
}
