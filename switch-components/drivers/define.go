package drivers

import "gitee.com/fatzeng/switch-sdk-core/driver"

var SupportedTypes = map[driver.DriverType]bool{
	KafkaConsumerDriverType:   true,
	KafkaProducerDriverType:   true,
	WebhookConsumerDriverType: true,
	WebhookProducerDriverType: true,
	PollingConsumerDriverType: true,
	PollingProducerDriverType: true,
}

// RetryConfig 定义重试配置
type RetryConfig struct {
	Count   int    `json:"count" mapstructure:"count" yaml:"count"`
	Backoff string `json:"backoff" mapstructure:"backoff" yaml:"backoff"`
}

// isRetryConfigEqual 比较 RetryConfig 是否相等
func isRetryConfigEqual(oldRetry, newRetry *RetryConfig) bool {
	if oldRetry == nil && newRetry == nil {
		return true
	}
	if oldRetry == nil || newRetry == nil {
		return false
	}

	return oldRetry.Count == newRetry.Count &&
		oldRetry.Backoff == newRetry.Backoff
}
