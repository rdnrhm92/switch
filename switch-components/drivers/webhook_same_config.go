package drivers

import (
	"strings"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// WebhookProducerConfigComparator Webhook Producer 配置比较器
func WebhookProducerConfigComparator(oldDriver, newDriver driver.Driver) bool {
	oldWebhook, oldOk := oldDriver.(*WebhookProducer)
	newWebhook, newOk := newDriver.(*WebhookProducer)

	if !oldOk || !newOk {
		logger.Logger.Warnf("Driver type mismatch in webhook producer config comparison")
		return false
	}

	return isProducerConfigEqual(oldWebhook.config, newWebhook.config)
}

// WebhookConsumerConfigComparator Webhook Consumer 配置比较器
func WebhookConsumerConfigComparator(oldDriver, newDriver driver.Driver) bool {
	oldWebhook, oldOk := oldDriver.(*WebhookConsumer)
	newWebhook, newOk := newDriver.(*WebhookConsumer)

	if !oldOk || !newOk {
		logger.Logger.Warnf("Driver type mismatch in webhook consumer config comparison")
		return false
	}

	return isConsumerConfigEqual(oldWebhook.config, newWebhook.config)
}

// isProducerConfigEqual 比较两个 Producer 配置是否相等
func isProducerConfigEqual(oldConfig, newConfig *WebhookProducerConfig) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	// 比较黑名单IPs
	if !isStringSliceEqual(oldConfig.BlacklistIPs, newConfig.BlacklistIPs) {
		return false
	}

	// 比较端口配置
	if oldConfig.Port != newConfig.Port {
		return false
	}

	// 比较基础配置
	if oldConfig.IgnoreExceptions != newConfig.IgnoreExceptions ||
		oldConfig.TimeOut != newConfig.TimeOut {
		return false
	}

	// 比较重试配置
	if !isRetryConfigEqual(oldConfig.Retry, newConfig.Retry) {
		return false
	}

	// 比较安全配置
	if !isSecurityConfigEqual(oldConfig.Security, newConfig.Security) {
		return false
	}

	return true
}

// isConsumerConfigEqual 比较两个 Consumer 配置是否相等
func isConsumerConfigEqual(oldConfig, newConfig *WebhookConsumerConfig) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	// 比较黑名单IPs
	if !isStringSliceEqual(oldConfig.BlacklistIPs, newConfig.BlacklistIPs) {
		return false
	}

	// 比较端口配置
	if oldConfig.Port != newConfig.Port {
		return false
	}

	// 比较重试配置
	if !isRetryConfigEqual(oldConfig.Retry, newConfig.Retry) {
		return false
	}

	// 比较安全配置
	if !isSecurityConfigEqual(oldConfig.Security, newConfig.Security) {
		return false
	}

	return true
}

// isStringSliceEqual 比较字符串切片是否相等（忽略url顺序）
func isStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// 比较
	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, url := range a {
		normalizedURL := normalizeURL(url)
		countA[normalizedURL]++
	}

	for _, url := range b {
		normalizedURL := normalizeURL(url)
		countB[normalizedURL]++
	}

	// 比较两个url map是否相等
	if len(countA) != len(countB) {
		return false
	}

	for url, count := range countA {
		if countB[url] != count {
			return false
		}
	}

	return true
}

// normalizeURL 标准化URL以便比较
func normalizeURL(url string) string {
	// 转换为小写
	normalized := strings.ToLower(strings.TrimSpace(url))

	// 去除尾部斜杠
	normalized = strings.TrimSuffix(normalized, "/")

	// 如果没有协议，添加默认协议
	if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
		normalized = "http://" + normalized
	}

	return normalized
}

// isSecurityConfigEqual 比较安全配置是否相等(配置内容的匹配)
func isSecurityConfigEqual(a, b *WebhookSecurityConfig) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Secret == b.Secret
}
