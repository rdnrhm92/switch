package drivers

import (
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/driver"
)

// WebhookProducerValidator Webhook驱动验证器
// 1.配置有效性的检查
type WebhookProducerValidator struct {
	webhookDriver *WebhookProducer
}

func (w *WebhookProducerValidator) Validate(driver driver.Driver) error {
	webhookDriver, ok := driver.(*WebhookProducer)
	if !ok {
		return fmt.Errorf("admin_driver is not a WebhookProducer")
	}
	w.webhookDriver = webhookDriver
	//向webhook发送消息，校验配置有效性
	if err := w.webhookDriver.config.isValid(); err != nil {
		return fmt.Errorf("webhook producer config validation failed: %w", err)
	}

	return nil
}
