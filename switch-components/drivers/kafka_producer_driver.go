package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const KafkaProducerDriverType driver.DriverType = "kafka_producer"

// KafkaProducer 生产者配置
type KafkaProducer struct {
	KafkaProducerValidator
	Writer     *kafka.Writer
	config     *KafkaProducerConfig
	mutex      sync.RWMutex
	driverName string
	callback   driver.DriverFailureCallback
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Start 启动 Kafka Producer 驱动
func (k *KafkaProducer) Start(ctx context.Context) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.Writer != nil {
		logger.Logger.Warnf("Kafka producer already started for topic: %s", k.config.getTopic())
		return nil
	}

	w, err := createKafkaWriter(k.config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka writer: %w", err)
	}

	k.Writer = w
	k.ctx, k.cancelFunc = context.WithCancel(ctx)

	// 准备关闭
	go k.closeMonitor()

	logger.Logger.Infof("Kafka producer started successfully for topic: %s", k.config.getTopic())
	return nil
}

func (k *KafkaProducer) closeMonitor() {
	<-k.ctx.Done()
	logger.Logger.Infof("Kafka producer is shutting down")
	k.Close()
}

// NewKafkaProducer 创建一个生产者驱动
func NewKafkaProducer(c *KafkaProducerConfig) (*KafkaProducer, error) {
	if c == nil {
		return nil, errors.New("kafka config must contain producer configuration")
	}

	return &KafkaProducer{config: c, Writer: nil}, nil
}

func (k *KafkaProducer) RecreateFromConfig() (driver.Driver, error) {
	return NewKafkaProducer(k.config)
}

func (k *KafkaProducer) GetDriverType() driver.DriverType {
	return KafkaProducerDriverType
}

func (k *KafkaProducer) GetDriverName() string {
	return k.driverName
}

func (k *KafkaProducer) SetDriverMeta(name string) {
	k.driverName = name
}

func (k *KafkaProducer) SetFailureCallback(callback driver.DriverFailureCallback) {
	k.callback = callback
}

// Notify 推送消息
func (k *KafkaProducer) Notify(ctx context.Context, data interface{}) error {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	if k.Writer == nil {
		if k.cancelFunc != nil {
			k.cancelFunc()
		}
		return errors.New("kafka producer not started, please call Start() first")
	}

	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialize switch data to JSON: %w", err)
	}

	// 按照要求如果有一批消息，某条大小超出了1M 则剔除这条消息，不影响剩余消息的发送 同时，重试也不要所有消息都重试
	// github.com/segmentio/kafka-go@v0.4.49/writer.go:634
	messageSize := len(value)
	const maxMessageSize = 1048576
	if messageSize > maxMessageSize {
		logger.Logger.Errorf("Message size %d bytes exceeds maximum allowed size %d bytes (1MB), message will be discarded", messageSize, maxMessageSize)
		return fmt.Errorf("message too large: %d bytes (max: %d bytes)", messageSize, maxMessageSize)
	}

	err = k.Writer.WriteMessages(ctx, kafka.Message{Value: value})

	if err != nil {
		// 检查是否是消息太大的错误
		var msgTooLarge kafka.MessageTooLargeError
		if errors.As(err, &msgTooLarge) {
			logger.Logger.Errorf("Message too large error from broker: message size %d bytes exceeds broker's limit, message discarded", messageSize)
			return fmt.Errorf("message exceeds broker's max size limit: %w", err)
		}

		// 其他错误
		logger.Logger.Errorf("Failed to send notification for Kafka '%v'", err)
	} else {
		logger.Logger.Debugf("Successfully sent notification for Kafka, message size: %d bytes", messageSize)
	}
	return err
}

// Close 关闭生产者
func (k *KafkaProducer) Close() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.Writer != nil {
		logger.Logger.Infof("Shutting down Kafka writer for topic: %s", k.Writer.Topic)
		err := k.Writer.Close()
		k.Writer = nil

		if k.cancelFunc != nil {
			k.cancelFunc()
			k.cancelFunc = nil
		}

		return err
	}
	return nil
}

// createKafkaWriter 生产者的写操作
func createKafkaWriter(cfg *KafkaProducerConfig) (*kafka.Writer, error) {
	// transport使用的conn是default Transport的策略 够用了
	transport := &kafka.Transport{
		Dial: (&net.Dialer{
			Timeout:   cfg.getConnectTimeout(),
			DualStack: true,
		}).DialContext,
	}

	if cfg.getSecurity() != nil {
		if cfg.getSecurity().TLS != nil && cfg.getSecurity().TLS.Enabled {
			tlsConfig, err := createTLSConfig(cfg.getSecurity().TLS)
			if err != nil {
				return nil, err
			}
			transport.TLS = tlsConfig
		}

		if cfg.getSecurity().SASL != nil && cfg.getSecurity().SASL.Enabled {
			mechanism, err := scram.Mechanism(scram.SHA512, cfg.getSecurity().SASL.Username, cfg.getSecurity().SASL.Password)
			if err != nil {
				return nil, err
			}
			transport.SASL = mechanism
		}
	}

	w := &kafka.Writer{
		Addr:      kafka.TCP(cfg.getBrokers()...),
		Topic:     cfg.getTopic(),
		Balancer:  &kafka.LeastBytes{},
		Transport: transport,
		//默认所有副本集必须全部同步
		RequiredAcks: kafka.RequireAll,
		Compression:  cfg.parseCompression(),
		Async:        false,
	}

	w.WriteTimeout = cfg.getTimeout()
	w.BatchTimeout = cfg.getBatchTimeout()
	w.BatchSize = cfg.getBatchSize()
	w.BatchBytes = cfg.getBatchBytes()
	w.RequiredAcks = cfg.parseRequiredAcks()
	w.MaxAttempts = cfg.getRetries()
	w.WriteBackoffMin = cfg.getRetryBackoffMin()
	w.WriteBackoffMax = cfg.getRetryBackoffMax()
	return w, nil
}
