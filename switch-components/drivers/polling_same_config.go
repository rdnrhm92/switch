package drivers

import (
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// PollingConsumerConfigComparator 长轮询配置比较器
func PollingConsumerConfigComparator(oldDriver, newDriver driver.Driver) bool {
	oldPolling, oldOk := oldDriver.(*PollingConsumer)
	newPolling, newOk := newDriver.(*PollingConsumer)

	if !oldOk || !newOk {
		logger.Logger.Warnf("Driver type mismatch in polling consumer config comparison")
		return false
	}

	return isPollingConsumerConfigEqual(oldPolling.config, newPolling.config)
}

// PollingProducerConfigComparator 长轮询生产者配置比较器
func PollingProducerConfigComparator(oldDriver, newDriver driver.Driver) bool {
	oldPolling, oldOk := oldDriver.(*PollingProducer)
	newPolling, newOk := newDriver.(*PollingProducer)

	if !oldOk || !newOk {
		logger.Logger.Warnf("Driver type mismatch in polling producer config comparison")
		return false
	}

	return isPollingProducerConfigEqual(oldPolling.config, newPolling.config)
}

// isPollingConsumerConfigEqual 比较两个长轮询消费者配置是否相等
func isPollingConsumerConfigEqual(oldConfig, newConfig *PollingConsumerConfig) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	// 比较URL配置
	if oldConfig.URL != newConfig.URL {
		return false
	}

	// 比较轮询配置
	if !isPollingTimingConfigEqual(oldConfig, newConfig) {
		return false
	}

	// 比较重试配置
	if !isRetryConfigEqual(oldConfig.Retry, newConfig.Retry) {
		return false
	}

	// 比较HTTP配置
	if !isPollingHTTPConfigEqual(oldConfig, newConfig) {
		return false
	}

	// 比较安全配置
	if !isPollingConsumerSecurityConfigEqual(oldConfig.Security, newConfig.Security) {
		return false
	}

	// 比较其他配置
	if oldConfig.IgnoreExceptions != newConfig.IgnoreExceptions {
		return false
	}

	return true
}

// isPollingProducerConfigEqual 比较两个长轮询生产者配置是否相等
func isPollingProducerConfigEqual(oldConfig, newConfig *PollingProducerConfig) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	// 比较端口配置
	if oldConfig.Port != newConfig.Port {
		return false
	}

	// 比较长轮询超时
	if oldConfig.LongPollTimeout != newConfig.LongPollTimeout {
		return false
	}

	// 比较服务器超时配置
	if !isPollingServerTimeoutConfigEqual(oldConfig, newConfig) {
		return false
	}

	// 比较重试配置
	if !isRetryConfigEqual(oldConfig.Retry, newConfig.Retry) {
		return false
	}

	// 比较安全配置
	if !isPollingProducerSecurityConfigEqual(oldConfig.Security, newConfig.Security) {
		return false
	}

	return true
}

// isPollingServerTimeoutConfigEqual 比较服务器超时配置
func isPollingServerTimeoutConfigEqual(oldConfig, newConfig *PollingProducerConfig) bool {
	return oldConfig.ServerReadTimeout == newConfig.ServerReadTimeout &&
		oldConfig.ServerWriteTimeout == newConfig.ServerWriteTimeout &&
		oldConfig.ServerIdleTimeout == newConfig.ServerIdleTimeout
}

// isPollingTimingConfigEqual 比较轮询时间配置
func isPollingTimingConfigEqual(oldConfig, newConfig *PollingConsumerConfig) bool {
	return oldConfig.PollInterval == newConfig.PollInterval &&
		oldConfig.RequestTimeout == newConfig.RequestTimeout
}

// isPollingHTTPConfigEqual 比较HTTP配置
func isPollingHTTPConfigEqual(oldConfig, newConfig *PollingConsumerConfig) bool {
	// 比较Headers
	if !isStringMapEqual(oldConfig.Headers, newConfig.Headers) {
		return false
	}

	// 比较UserAgent
	return oldConfig.UserAgent == newConfig.UserAgent
}

// isPollingProducerSecurityConfigEqual 比较安全配置 producer
func isPollingProducerSecurityConfigEqual(oldSec, newSec *PollingServerSecurityConfig) bool {
	if oldSec == nil && newSec == nil {
		return true
	}
	if oldSec == nil || newSec == nil {
		return false
	}

	// 比较HTTPS配置
	if oldSec.CertFile != newSec.CertFile || oldSec.KeyFile != newSec.KeyFile {
		return false
	}

	// 比较ValidTokens
	return isTokenStringSliceEqual(oldSec.ValidTokens, newSec.ValidTokens)
}

// isPollingConsumerSecurityConfigEqual 比较安全配置 consumer
func isPollingConsumerSecurityConfigEqual(oldSec, newSec *PollingClientSecurityConfig) bool {
	if oldSec == nil && newSec == nil {
		return true
	}
	if oldSec == nil || newSec == nil {
		return false
	}

	return oldSec.InsecureSkipVerify == newSec.InsecureSkipVerify &&
		oldSec.Token == newSec.Token
}

// isStringMapEqual 比较字符串映射是否相等
func isStringMapEqual(a, b map[string]string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists || valueA != valueB {
			return false
		}
	}

	return true
}

// isTokenStringSliceEqual 比较字符串切片是否相等
func isTokenStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, valueA := range a {
		if valueA != b[i] {
			return false
		}
	}

	return true
}
