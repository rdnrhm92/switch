package drivers

import (
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// KafkaConfigCompareOptions Kafka 配置比较选项
type KafkaConfigCompareOptions struct {
	// 是否对比broker的顺序
	StrictBrokerOrder bool
}

// DefaultKafkaConfigCompareOptions 默认比较选项
var DefaultKafkaConfigCompareOptions = KafkaConfigCompareOptions{
	StrictBrokerOrder: false,
}

// KafkaConsumerConfigComparator Kafka Consumer 配置比较器
func KafkaConsumerConfigComparator(oldDriver, newDriver driver.Driver) bool {
	return KafkaConsumerConfigComparatorWithOptions(oldDriver, newDriver)
}

// KafkaConsumerConfigComparatorWithOptions Kafka Consumer 配置比较器
func KafkaConsumerConfigComparatorWithOptions(oldDriver, newDriver driver.Driver, options ...KafkaConfigCompareOptions) bool {
	oldKafka, oldOk := oldDriver.(*KafkaConsumer)
	newKafka, newOk := newDriver.(*KafkaConsumer)

	if !oldOk || !newOk {
		logger.Logger.Warnf("Driver type mismatch in kafka consumer config comparison")
		return false
	}

	var option KafkaConfigCompareOptions
	if len(options) > 0 {
		option = options[0]
	} else {
		option = DefaultKafkaConfigCompareOptions
	}

	return isKafkaConsumerConfigEqualWithOptions(oldKafka.config, newKafka.config, option)
}

// KafkaProducerConfigComparator Kafka Producer 配置比较器
func KafkaProducerConfigComparator(oldDriver, newDriver driver.Driver) bool {
	return KafkaProducerConfigComparatorWithOptions(oldDriver, newDriver)
}

// KafkaProducerConfigComparatorWithOptions Kafka Producer 配置比较器
func KafkaProducerConfigComparatorWithOptions(oldDriver, newDriver driver.Driver, options ...KafkaConfigCompareOptions) bool {
	oldKafka, oldOk := oldDriver.(*KafkaProducer)
	newKafka, newOk := newDriver.(*KafkaProducer)

	if !oldOk || !newOk {
		logger.Logger.Warnf("Driver type mismatch in kafka producer config comparison")
		return false
	}

	var option KafkaConfigCompareOptions
	if len(options) > 0 {
		option = options[0]
	} else {
		option = DefaultKafkaConfigCompareOptions
	}

	return isKafkaProducerConfigEqualWithOptions(oldKafka.config, newKafka.config, option)
}

// isKafkaConsumerConfigEqualWithOptions 比较两个 Kafka 消费者配置是否相等（带选项）
func isKafkaConsumerConfigEqualWithOptions(oldConfig, newConfig *KafkaConsumerConfig, options KafkaConfigCompareOptions) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	// 比较基础配置
	if !isKafkaConsumerBaseConfigEqualWithOptions(oldConfig, newConfig, options) {
		return false
	}

	// 比较安全配置
	if !isKafkaSecurityConfigEqual(oldConfig.Security, newConfig.Security) {
		return false
	}

	// 比较重试配置
	if !isRetryConfigEqual(oldConfig.Retry, newConfig.Retry) {
		return false
	}

	// 比较消费者特有配置
	if !isKafkaConsumerSpecificConfigEqual(oldConfig, newConfig) {
		return false
	}

	return true
}

// isKafkaProducerConfigEqualWithOptions 比较两个 Kafka 生产者配置是否相等（带选项）
func isKafkaProducerConfigEqualWithOptions(oldConfig, newConfig *KafkaProducerConfig, options KafkaConfigCompareOptions) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	// 比较基础配置
	if !isKafkaProducerBaseConfigEqualWithOptions(oldConfig, newConfig, options) {
		return false
	}

	// 比较安全配置
	if !isKafkaSecurityConfigEqual(oldConfig.Security, newConfig.Security) {
		return false
	}

	// 比较生产者特有配置
	if !isKafkaProducerSpecificConfigEqual(oldConfig, newConfig) {
		return false
	}

	return true
}

// isKafkaConsumerBaseConfigEqualWithOptions 比较 Kafka 消费者基础配置
func isKafkaConsumerBaseConfigEqualWithOptions(oldConfig, newConfig *KafkaConsumerConfig, options KafkaConfigCompareOptions) bool {
	var brokersEqual bool
	if options.StrictBrokerOrder {
		brokersEqual = isStringSliceEqualKafkaOrdered(oldConfig.Brokers, newConfig.Brokers)
	} else {
		brokersEqual = isStringSliceEqualKafka(oldConfig.Brokers, newConfig.Brokers)
	}

	if !brokersEqual {
		return false
	}

	if oldConfig.ReadTimeout != newConfig.ReadTimeout {
		return false
	}

	if oldConfig.CommitTimeout != newConfig.CommitTimeout {
		return false
	}

	return oldConfig.Topic == newConfig.Topic
}

// isKafkaProducerBaseConfigEqualWithOptions 比较 Kafka 生产者基础配置
func isKafkaProducerBaseConfigEqualWithOptions(oldConfig, newConfig *KafkaProducerConfig, options KafkaConfigCompareOptions) bool {
	var brokersEqual bool
	if options.StrictBrokerOrder {
		brokersEqual = isStringSliceEqualKafkaOrdered(oldConfig.Brokers, newConfig.Brokers)
	} else {
		brokersEqual = isStringSliceEqualKafka(oldConfig.Brokers, newConfig.Brokers)
	}

	if !brokersEqual {
		return false
	}

	return oldConfig.Topic == newConfig.Topic
}

// isKafkaConsumerSpecificConfigEqual 比较消费者特有配置
func isKafkaConsumerSpecificConfigEqual(oldConfig, newConfig *KafkaConsumerConfig) bool {
	return oldConfig.GroupID == newConfig.GroupID &&
		oldConfig.AutoOffsetReset == newConfig.AutoOffsetReset &&
		oldConfig.EnableAutoCommit == newConfig.EnableAutoCommit &&
		oldConfig.AutoCommitInterval == newConfig.AutoCommitInterval
}

// isKafkaProducerSpecificConfigEqual 比较生产者特有配置
func isKafkaProducerSpecificConfigEqual(oldConfig, newConfig *KafkaProducerConfig) bool {
	return oldConfig.RequiredAcks == newConfig.RequiredAcks &&
		oldConfig.Timeout == newConfig.Timeout &&
		oldConfig.BatchTimeout == newConfig.BatchTimeout &&
		oldConfig.BatchBytes == newConfig.BatchBytes &&
		oldConfig.BatchSize == newConfig.BatchSize &&
		oldConfig.Retries == newConfig.Retries &&
		oldConfig.RetryBackoffMin == newConfig.RetryBackoffMin &&
		oldConfig.RetryBackoffMax == newConfig.RetryBackoffMax &&
		oldConfig.Compression == newConfig.Compression
}

// isKafkaSecurityConfigEqual 比较安全配置
func isKafkaSecurityConfigEqual(oldSec, newSec *SecurityConfig) bool {
	if oldSec == nil && newSec == nil {
		return true
	}
	if oldSec == nil || newSec == nil {
		return false
	}

	// 比较 SASL 配置
	if !isSASLConfigEqual(oldSec.SASL, newSec.SASL) {
		return false
	}

	// 比较 TLS 配置
	if !isTLSConfigEqual(oldSec.TLS, newSec.TLS) {
		return false
	}

	return true
}

// isSASLConfigEqual 比较 SASL 配置
func isSASLConfigEqual(oldSASL, newSASL *SASLConfig) bool {
	if oldSASL == nil && newSASL == nil {
		return true
	}
	if oldSASL == nil || newSASL == nil {
		return false
	}

	return oldSASL.Enabled == newSASL.Enabled &&
		oldSASL.Mechanism == newSASL.Mechanism &&
		oldSASL.Username == newSASL.Username &&
		oldSASL.Password == newSASL.Password
}

// isTLSConfigEqual 比较 TLS 配置
func isTLSConfigEqual(oldTLS, newTLS *TLSConfig) bool {
	if oldTLS == nil && newTLS == nil {
		return true
	}
	if oldTLS == nil || newTLS == nil {
		return false
	}

	return oldTLS.Enabled == newTLS.Enabled &&
		oldTLS.CaFile == newTLS.CaFile &&
		oldTLS.CertFile == newTLS.CertFile &&
		oldTLS.KeyFile == newTLS.KeyFile &&
		oldTLS.InsecureSkipVerify == newTLS.InsecureSkipVerify
}

// isStringSliceEqualKafka 不考虑顺序的情况下比较字符串切片是否相等
func isStringSliceEqualKafka(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, broker := range a {
		countA[broker]++
	}

	for _, broker := range b {
		countB[broker]++
	}

	if len(countA) != len(countB) {
		return false
	}

	for broker, count := range countA {
		if countB[broker] != count {
			return false
		}
	}

	return true
}

// isStringSliceEqualKafkaOrdered 考虑顺序的情况下比较字符串切片是否相等
func isStringSliceEqualKafkaOrdered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
