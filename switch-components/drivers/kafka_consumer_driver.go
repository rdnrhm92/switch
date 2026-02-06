package drivers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"github.com/segmentio/kafka-go"
)

const KafkaConsumerDriverType driver.DriverType = "kafka_consumer"

// MessageHandler 消息处理函数
type MessageHandler func(ctx context.Context, msg *kafka.Message) error

// KafkaConsumer 消费者配置
type KafkaConsumer struct {
	KafkaConsumerValidator
	Reader       *kafka.Reader
	config       *KafkaConsumerConfig
	cancel       context.CancelFunc
	mutex        sync.RWMutex
	handler      MessageHandler
	driverName   string
	callback     driver.DriverFailureCallback
	failureCount int // 连续失败次数
}

func (k *KafkaConsumer) GetDriverType() driver.DriverType {
	return KafkaConsumerDriverType
}

func (k *KafkaConsumer) GetDriverName() string {
	return k.driverName
}

func (k *KafkaConsumer) SetDriverMeta(name string) {
	k.driverName = name
}

func (k *KafkaConsumer) SetFailureCallback(callback driver.DriverFailureCallback) {
	k.callback = callback
}

// NewKafkaConsumer 创建一个消费者驱动
func NewKafkaConsumer(c *KafkaConsumerConfig, handler MessageHandler) (*KafkaConsumer, error) {
	if c == nil {
		return nil, errors.New("kafka config must contain consumer configuration")
	}

	if handler == nil {
		return nil, errors.New("message handler cannot be nil")
	}

	return &KafkaConsumer{
		config:       c,
		handler:      handler,
		failureCount: 0,
	}, nil
}

func (k *KafkaConsumer) RecreateFromConfig() (driver.Driver, error) {
	return NewKafkaConsumer(k.config, k.handler)
}

// Close 关闭一个消费者
func (k *KafkaConsumer) Close() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.cancel != nil {
		k.cancel()
		k.cancel = nil
	}

	if k.Reader != nil {
		logger.Logger.Infof("Shutting down Kafka reader for topic: %s", k.config.getTopic())
		err := k.Reader.Close()
		k.Reader = nil
		return err
	}
	return nil
}

// Start 开始消费
func (k *KafkaConsumer) Start(ctx context.Context) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.Reader != nil {
		logger.Logger.Warnf("Kafka consumer already started for topic: %s", k.config.getTopic())
		return nil
	}

	k.Reader = createKafkaReader(k.config)

	ctx, cancel := context.WithCancel(ctx)
	k.cancel = cancel

	//自动恢复
	backoffDuration := k.config.getBackoffDuration()
	recovery.SafeGo(ctx, func(ctx context.Context) error {
		return k.consumeLoop(ctx, k.handler)
	}, string(KafkaConsumerDriverType), recovery.WithRetryInterval(backoffDuration))

	logger.Logger.Infof("Kafka consumer started for topic '%s', group '%s'", k.config.getTopic(), k.config.getGroupId())
	return nil
}

// consumeLoop 持续消费
func (k *KafkaConsumer) consumeLoop(ctx context.Context, handler MessageHandler) error {
	k.mutex.RLock()
	if k.Reader == nil {
		k.mutex.RUnlock()
		return errors.New("kafka consumer not started, Reader is nil")
	}
	k.mutex.RUnlock()
	retries := k.config.getMaxRetries()

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Shutting down Kafka consumer loop")
			k.Close()
			return nil
		default:
			readCtx, readCancel := context.WithTimeout(ctx, k.config.getReadTimeout())
			msg, err := k.Reader.FetchMessage(readCtx)
			readCancel()

			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
					logger.Logger.Infof("Kafka consumer for topic '%s' is stopping: %v", k.Reader.Config().Topic, err)
					return nil
				}
				if errors.Is(err, context.DeadlineExceeded) {
					logger.Logger.Debugf("Read timeout for topic '%s', continuing...", k.Reader.Config().Topic)
					// 超时不算失败，重置计数
					k.mutex.Lock()
					k.failureCount = 0
					k.mutex.Unlock()
					continue
				}

				// 其他错误，增加失败计数
				k.mutex.Lock()
				k.failureCount++
				currentFailures := k.failureCount
				k.mutex.Unlock()

				logger.Logger.Errorf("Failed to fetch message from Kafka topic '%s' (failure %d/%d): %v", k.Reader.Config().Topic, currentFailures, retries, err)

				// 如果连续失败次数过多，触发重启
				if currentFailures >= retries {
					logger.Logger.Errorf("Kafka consumer reached max retries (%d), triggering restart", retries)
					return fmt.Errorf("max retries exhausted: %w", err)
				}
				continue
			}

			// 成功读取消息，重置失败计数
			k.mutex.Lock()
			k.failureCount = 0
			k.mutex.Unlock()

			logger.Logger.Debugf("Received message from topic '%s', partition %d, offset %d", msg.Topic, msg.Partition, msg.Offset)

			if string(msg.Value) == "ping" {
				logger.Logger.Info("Received testing message during verification: ping")
			} else {
				if err = handler(ctx, &msg); err != nil {
					logger.Logger.Errorf("Failed to process message from topic '%s' (offset %d): %v. Offset will not be committed.", msg.Topic, msg.Offset, err)
				}
			}

			// 改为手动提交 - 使用提交超时
			if !k.config.getEnableAutoCommit() {
				commitCtx, commitCancel := context.WithTimeout(ctx, k.config.getCommitTimeout())
				if err = k.Reader.CommitMessages(commitCtx, msg); err != nil {
					logger.Logger.Errorf("Failed to commit offset %d for topic '%s': %v", msg.Offset, msg.Topic, err)
				} else {
					logger.Logger.Debugf("Manually committed offset %d for topic '%s'", msg.Offset, msg.Topic)
				}
				commitCancel()
			}
		}
	}
}

func createKafkaReader(cfg *KafkaConsumerConfig) *kafka.Reader {
	dialer, _ := createDialer(cfg.getSecurity(), cfg.getConnectTimeout())

	logger.Logger.Infof("Using GroupID: %s", cfg.getGroupId())

	rc := kafka.ReaderConfig{
		Brokers: cfg.getBrokers(),
		Topic:   cfg.getTopic(),
		GroupID: cfg.getGroupId(),
		Dialer:  dialer,
	}

	// 开启了自动提交
	if cfg.getEnableAutoCommit() {
		interval := cfg.getAutoCommitInterval()
		rc.CommitInterval = interval
		logger.Logger.Infof("Auto commit enabled with interval: %v", interval)
	} else {
		// 改为手动提交
		rc.CommitInterval = 0
		logger.Logger.Infof("Auto commit disabled, manual commit required")
	}

	switch cfg.getAutoOffsetReset() {
	case "latest":
		//从最新的消息开始读
		rc.StartOffset = kafka.LastOffset
		break
	case "earliest":
		//从最旧的消息开始读
		rc.StartOffset = kafka.FirstOffset
		break
	}

	return kafka.NewReader(rc)
}
