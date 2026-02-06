package drivers

import (
	"context"
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"github.com/segmentio/kafka-go"
)

// KafkaProducerValidator
// 1.生产者 是否具有对topic的写权限
type KafkaProducerValidator struct {
	kafkaDriver *KafkaProducer
}

// Validate kafka的可用性校验
// 1.能否连接到broker
// 2.配置的topic是否存在
// 3.能否向topic写消息
func (v *KafkaProducerValidator) Validate(driver driver.Driver) error {
	kafkaDriver, ok := driver.(*KafkaProducer)
	if !ok {
		return fmt.Errorf("admin_driver is not a KafkaProducer")
	}
	v.kafkaDriver = kafkaDriver

	if kafkaDriver.config == nil {
		return fmt.Errorf("kafka driver configuration cannot be nil")
	}

	if err := kafkaDriver.config.isValid(); err != nil {
		return fmt.Errorf("kafka driver is not valid: %v", err)
	}

	if err := checkProducerTopicExists(kafkaDriver.config); err != nil {
		return fmt.Errorf("topic check failed: %w", err)
	}

	return checkProducerPermissions(kafkaDriver.config)
}

// checkProducerPermissions 检查生产者权限
func checkProducerPermissions(config *KafkaProducerConfig) error {
	writer, err := createKafkaWriter(config)
	if err != nil {
		return fmt.Errorf("producer permission test failed: failed to create writer: %w", err)
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.getValidateTimeout())
	defer cancel()

	originalAcks := writer.RequiredAcks
	writer.RequiredAcks = kafka.RequireNone

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("__health_check__"),
		Value: []byte("ping"),
	})

	// 恢复原始设置
	writer.RequiredAcks = originalAcks

	if err != nil {
		return fmt.Errorf("producer permission test failed: %w", err)
	}

	return nil
}
