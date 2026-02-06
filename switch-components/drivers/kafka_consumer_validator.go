package drivers

import (
	"context"
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"github.com/segmentio/kafka-go"
)

// KafkaConsumerValidator
// 1.消费者 是否具有对topic的读权限
// 2.消费者 是否可以加入消费者组(如果有组)
// 3.消费者 是否可以正确的提交偏移量
type KafkaConsumerValidator struct {
	kafkaDriver *KafkaConsumer
}

// Validate 验证Kafka消费者驱动
func (k *KafkaConsumerValidator) Validate(driver driver.Driver) error {
	kafkaDriver, ok := driver.(*KafkaConsumer)
	if !ok {
		return fmt.Errorf("admin_driver is not a KafkaConsumer")
	}
	k.kafkaDriver = kafkaDriver

	// 验证配置有效性
	if err := kafkaDriver.config.isValid(); err != nil {
		return fmt.Errorf("kafka consumer config validation failed: %w", err)
	}

	if err := checkConsumerTopicExists(kafkaDriver.config); err != nil {
		return fmt.Errorf("topic check failed: %w", err)
	}

	return checkConsumerPermissions(kafkaDriver.config)
}

// checkConsumerPermissions 检查消费者权限
func checkConsumerPermissions(config *KafkaConsumerConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.getValidateTimeout())
	defer cancel()

	// 测试组ID
	testGroupID := "__health_check_" + fmt.Sprintf("%d", time.Now().UnixNano())

	// 创建带认证配置的reader
	readerConfig := kafka.ReaderConfig{
		Brokers: config.Brokers,
		Topic:   config.Topic,
		GroupID: testGroupID,
	}

	// 配置认证
	dialer, err := createDialer(config.Security, config.getConnectTimeout())
	if err != nil {
		return fmt.Errorf("failed to configure reader auth: %w", err)
	}

	readerConfig.Dialer = dialer

	// 默认是消息开始位置开始读 生产者那边校验也会写一个测试消息，非开关消息，测试用的
	reader := kafka.NewReader(readerConfig)
	defer reader.Close()

	// 尝试读取消息(用于测试读权限) - 使用配置的读超时
	readCtx, readCancel := context.WithTimeout(ctx, config.getReadTimeout())
	defer readCancel()

	_, err = reader.ReadMessage(readCtx)
	//因为没有消息，读取失败如果不是超时，可能是其他的异常
	if err != nil && err != context.DeadlineExceeded {
		return fmt.Errorf("consumer read permission test failed: %w", err)
	}

	commitCtx, commitCancel := context.WithTimeout(ctx, config.getCommitTimeout())
	defer commitCancel()

	// 提交偏移量（测试提交权限）
	if err = reader.CommitMessages(commitCtx); err != nil {
		// 提交的时候没有指定消息，非指定消息的错误都要返回
		if err.Error() != "no messages to commit" {
			return fmt.Errorf("failed to commit offset (commit permission): %w", err)
		}
	}

	// 检查是否能获取统计信息（验证消费组权限）
	stats := reader.Stats()
	if stats.Topic == "" {
		return fmt.Errorf("failed to access consumer group for offset management")
	}

	return nil
}
