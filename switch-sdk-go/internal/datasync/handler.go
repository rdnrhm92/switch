package datasync

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gitee.com/fatzeng/switch-components/drivers"
	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	_switch "gitee.com/fatzeng/switch-sdk-go/core/switch"
	"github.com/segmentio/kafka-go"
)

// ConfigHandler 消费者handler
type ConfigHandler func(ctx context.Context, data json.RawMessage) error

var (
	handlerRegistry = make(map[pc.ConfigChangeType]ConfigHandler)
	handlerMu       sync.RWMutex
)

// RegisterConfigHandler 业务方支持其他消费者驱动后可自行注册
func RegisterConfigHandler(configType pc.ConfigChangeType, handler ConfigHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	handlerRegistry[configType] = handler
}

func init() {
	RegisterConfigHandler(pc.KafkaConsumerConfigChange, handleKafkaConsumerChange)
	RegisterConfigHandler(pc.WebhookConsumerConfigChange, handleWebhookConsumerChange)
	RegisterConfigHandler(pc.PollingConsumerConfigChange, handlePollingConsumerChange)
}

// handleChangeConfig 作为一个中央分发器，监听所有增量/全量配置变更
func handleChangeConfig(ctx context.Context, message []byte) {
	// 解析响应消息
	var responseMsg pc.ResponseMessage
	if err := json.Unmarshal(message, &responseMsg); err != nil {
		logger.Logger.Errorf("Failed to unmarshal ResponseMessage: %v", err)
		return
	}

	// 检查消息类型
	if responseMsg.Type != pc.DriverConfigFull && responseMsg.Type != pc.DriverConfigChange {
		logger.Logger.Warnf("Response message type is not config related, got: %s", responseMsg.Type)
		return
	}

	logger.Logger.Infof("Received config response from server, msg: %v", responseMsg)

	// 解析配置数据
	if responseMsg.Data != nil {
		dataBytes, err := json.Marshal(responseMsg.Data)
		if err != nil {
			logger.Logger.Errorf("Failed to marshal response data: %v", err)
			return
		}

		// 解析配置负载数据
		var configPayloads []*pc.DriverConfigPayload
		if err := json.Unmarshal(dataBytes, &configPayloads); err != nil {
			logger.Logger.Errorf("Failed to unmarshal admin_driver config payloads: %v  Original message content: %s", err, string(message))
			return
		}

		// 处理配置
		isFullSync := responseMsg.Type == pc.DriverConfigFull
		configTypeDesc := "configurations"
		if !isFullSync {
			configTypeDesc = "configuration changes"
		}

		logger.Logger.Infof("Processing %d admin_driver %s", len(configPayloads), configTypeDesc)
		for _, configPayload := range configPayloads {
			// 这里不开协程并行启动了
			if err := processDriverConfigWithRegistry(ctx, configPayload); err != nil {
				logger.Logger.Errorf("Error processing admin_driver config type %s: %v", configPayload.Type, err)
			}
		}

		if isFullSync {
			logger.Logger.Info("Driver config full sync completed successfully")
			// 通知配置全量同步完成
			NotifyConfigSyncDone(pc.WsEndpointFullSyncConfig)
		} else {
			logger.Logger.Info("Driver config change processed successfully")
		}
	}
}

// processDriverConfigWithRegistry 使用注册表处理驱动配置
func processDriverConfigWithRegistry(ctx context.Context, configPayload *pc.DriverConfigPayload) error {
	handlerMu.RLock()
	defer handlerMu.RUnlock()

	// 根据驱动类型查找对应的处理器
	var configChangeType pc.ConfigChangeType
	switch configPayload.Type {
	case drivers.KafkaConsumerDriverType:
		configChangeType = pc.KafkaConsumerConfigChange
	case drivers.WebhookConsumerDriverType:
		configChangeType = pc.WebhookConsumerConfigChange
	case drivers.PollingConsumerDriverType:
		configChangeType = pc.PollingConsumerConfigChange
	default:
		logger.Logger.Warnf("Unknown admin_driver type: %s", configPayload.Type)
		return fmt.Errorf("unknown admin_driver type: %s", configPayload.Type)
	}

	// 查找注册的处理器
	handler, exists := handlerRegistry[configChangeType]
	if !exists {
		logger.Logger.Warnf("No handler registered for config type: %s", configChangeType)
		return fmt.Errorf("no handler registered for config type: %s", configChangeType)
	}

	// 调用处理器
	return handler(ctx, configPayload.Config)
}

// handleKafkaConsumerChange 监听kafka配置变更
func handleKafkaConsumerChange(ctx context.Context, data json.RawMessage) error {
	driverModel := &model.Driver{}
	if err := json.Unmarshal(data, driverModel); err != nil {
		logger.Logger.Errorf("Failed to unmarshal Driver: %v", err)
		return err
	}

	kafkaConfig := new(drivers.KafkaConsumerConfig)
	if err := json.Unmarshal(driverModel.DriverConfig, kafkaConfig); err != nil {
		logger.Logger.Errorf("Failed to unmarshal KafkaConsumerConfig from DriverConfig: %v", err)
		return err
	}

	// 旧驱动名字
	driverName := fmt.Sprintf("%s_%s_%s_%d", _switch.GlobalClient.NamespaceTag(), _switch.GlobalClient.EnvTag(), driverModel.DriverType, driverModel.ID)
	replacer := driver.NewGracefulDriverReplacer()

	timeout, _ := _switch.GlobalClient.KafkaConsumerReplaceDriverValidationTimeout()
	period, _ := _switch.GlobalClient.KafkaConsumerReplaceDriverStabilityPeriod()

	err, _ := replacer.ReplaceDriver(ctx, &driver.DriverReplacementConfig{
		DriverType:        drivers.KafkaConsumerDriverType,
		DriverName:        driverName,
		ValidationTimeout: timeout,
		StabilityPeriod:   period,
		ConfigComparator: func(oldDriver, newDriver driver.Driver) bool {
			return drivers.KafkaConsumerConfigComparatorWithOptions(oldDriver, newDriver, drivers.KafkaConfigCompareOptions{
				StrictBrokerOrder: _switch.GlobalClient.KafkaConsumerVerifyBrokers(),
			})
		},
		SkipIfSameConfig: true,
		SupportedTypes:   drivers.SupportedTypes,
	}, func() (driver.Driver, error) {
		kafkaConsumer, err := drivers.NewKafkaConsumer(kafkaConfig, consumeKafkaMessage)
		if err != nil {
			return nil, err
		}

		// 不在这里启动，由 ReplaceDriver 在合适的时机启动
		// kafkaConsumer.Start(ctx)

		return kafkaConsumer, nil
	})
	return err
}

// handleWebhookConsumerChange 监听webhook配置变更
func handleWebhookConsumerChange(ctx context.Context, data json.RawMessage) error {
	// 获取驱动数据
	driverModel := &model.Driver{}
	if err := json.Unmarshal(data, driverModel); err != nil {
		logger.Logger.Errorf("Failed to unmarshal Driver: %v", err)
		return err
	}

	// 获取驱动配置
	webhookConfig := new(drivers.WebhookConsumerConfig)
	if err := json.Unmarshal(driverModel.DriverConfig, webhookConfig); err != nil {
		logger.Logger.Errorf("Failed to unmarshal webhookConsumerConfig from DriverConfig: %v", err)
		return err
	}

	// 旧驱动名字
	driverName := fmt.Sprintf("%s_%s_%s_%d", _switch.GlobalClient.NamespaceTag(), _switch.GlobalClient.EnvTag(), driverModel.DriverType, driverModel.ID)
	replacer := driver.NewGracefulDriverReplacer()

	timeout, _ := _switch.GlobalClient.WebhookConsumerReplaceDriverValidationTimeout()
	period, _ := _switch.GlobalClient.WebhookConsumerReplaceDriverStabilityPeriod()

	err, _ := replacer.ReplaceDriver(ctx, &driver.DriverReplacementConfig{
		DriverType:        drivers.WebhookConsumerDriverType,
		DriverName:        driverName,
		ValidationTimeout: timeout,
		StabilityPeriod:   period,
		ConfigComparator:  drivers.WebhookConsumerConfigComparator,
		SkipIfSameConfig:  true,
		SupportedTypes:    drivers.SupportedTypes,
	}, func() (driver.Driver, error) {
		webhookConsumer, err := drivers.NewWebhookConsumer(webhookConfig, consumeWebhookMessage)
		if err != nil {
			return nil, err
		}
		// 不在这里启动，由 ReplaceDriver 在合适的时机启动
		// webhookConsumer.Start(ctx)

		return webhookConsumer, nil
	})
	return err
}

// handlePollingConsumerChange 监听polling配置变更
func handlePollingConsumerChange(ctx context.Context, data json.RawMessage) error {
	// 获取驱动数据
	driverModel := &model.Driver{}
	if err := json.Unmarshal(data, driverModel); err != nil {
		logger.Logger.Errorf("Failed to unmarshal Driver: %v", err)
		return err
	}

	// 获取驱动配置
	pollingConfig := new(drivers.PollingConsumerConfig)
	if err := json.Unmarshal(driverModel.DriverConfig, pollingConfig); err != nil {
		logger.Logger.Errorf("Failed to unmarshal pollingConsumerConfig from DriverConfig: %v", err)
		return err
	}

	// 旧驱动名字
	driverName := fmt.Sprintf("%s_%s_%s_%d", _switch.GlobalClient.NamespaceTag(), _switch.GlobalClient.EnvTag(), driverModel.DriverType, driverModel.ID)
	replacer := driver.NewGracefulDriverReplacer()

	timeout, _ := _switch.GlobalClient.PollingConsumerReplaceDriverValidationTimeout()
	period, _ := _switch.GlobalClient.PollingConsumerReplaceDriverStabilityPeriod()

	err, _ := replacer.ReplaceDriver(ctx, &driver.DriverReplacementConfig{
		DriverType:        drivers.PollingConsumerDriverType,
		DriverName:        driverName,
		ValidationTimeout: timeout,
		StabilityPeriod:   period,
		ConfigComparator:  drivers.PollingConsumerConfigComparator,
		SkipIfSameConfig:  true,
		SupportedTypes:    drivers.SupportedTypes,
	}, func() (driver.Driver, error) {
		pollingConsumer, err := drivers.NewPollingConsumer(pollingConfig, consumePollingMessage)
		if err != nil {
			return nil, err
		}

		// 不在这里启动，由 ReplaceDriver 在合适的时机启动
		// pollingConsumer.Start(ctx)

		return pollingConsumer, nil
	})
	return err
}

// handleFullSync 服务启动后的开关全量
func handleFullSync(ctx context.Context, message []byte) {
	// 解析响应消息
	var responseMsg pc.ResponseMessage
	if err := json.Unmarshal(message, &responseMsg); err != nil {
		logger.Logger.Errorf("Failed to unmarshal ResponseMessage: %v", err)
		return
	}

	// 检查消息类型
	if responseMsg.Type != pc.SwitchFull {
		logger.Logger.Warnf("Response message type is not SwitchFull, got: %s", responseMsg.Type)
		return
	}

	logger.Logger.Infof("Received switch full sync response from server, request_id: %s", responseMsg.RequestID)

	// 解析开关数据
	var switches []*model.SwitchModel
	if responseMsg.Data != nil {
		dataBytes, err := json.Marshal(responseMsg.Data)
		if err != nil {
			logger.Logger.Errorf("Failed to marshal response data: %v", err)
			return
		}

		if err = json.Unmarshal(dataBytes, &switches); err != nil {
			logger.Logger.Errorf("Failed to unmarshal switches data: %v  Original message content: %s", err, string(message))
			return
		}
	}

	// 注册开关规则
	logger.Logger.Infof("Registering %d switch rules", len(switches))
	for _, s := range switches {
		if s == nil {
			continue
		}
		_switch.RegisterRule(s.Name, s)
		logger.Logger.Debugf("Registered switch rule: %s", s.Name)
	}

	logger.Logger.Info("Switch full sync completed successfully")

	// 通知开关全量同步完成
	NotifyConfigSyncDone(pc.WsEndpointFullSync)
}

// consumeKafkaMessage 消费kafka消息
func consumeKafkaMessage(ctx context.Context, msg *kafka.Message) error {
	s := new(model.SwitchModel)
	if err := json.Unmarshal(msg.Value, s); err != nil {
		logger.Logger.Errorf("Failed to unmarshal kafka message into switch admin_model: %v", err)
		return nil
	}

	_switch.RegisterRule(s.Name, s)
	logger.Logger.Infof("Consumed and registered switch update from Kafka for: %s, version: %d", s.Name, s.Version.Version)
	return nil
}

// consumeWebhookMessage 消费Webhook消息
func consumeWebhookMessage(ctx context.Context, msg json.RawMessage) error {
	s := new(model.SwitchModel)
	if err := json.Unmarshal(msg, s); err != nil {
		logger.Logger.Errorf("Failed to unmarshal Webhook message into switch admin_model: %v", err)
		return nil
	}

	_switch.RegisterRule(s.Name, s)
	logger.Logger.Infof("Consumed and registered switch update from Webhook for: %s, version: %d", s.Name, s.Version.Version)
	return nil
}

// consumePollingMessage 消费Polling消息
func consumePollingMessage(ctx context.Context, msg json.RawMessage) error {
	s := new(model.SwitchModel)
	if err := json.Unmarshal(msg, s); err != nil {
		logger.Logger.Errorf("Failed to unmarshal Polling message into switch admin_model: %v", err)
		return nil
	}

	_switch.RegisterRule(s.Name, s)
	logger.Logger.Infof("Consumed and registered switch update from Polling for: %s, version: %d", s.Name, s.Version.Version)
	return nil
}
