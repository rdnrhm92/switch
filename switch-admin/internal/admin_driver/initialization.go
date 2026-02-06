package admin_driver

import (
	"fmt"

	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-components/drivers"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func InitializeDriver() error {
	if config.GlobalConfig.MySQL == nil {
		return errors.New("mysql config is nil")
	}
	if err := CreateMysql(config.GlobalConfig.MySQL, "default"); err != nil {
		logger.Logger.Panicf("failed to create mysql driver: %v", err)
		return err
	}
	return nil
}

func WebhookProducerDriver(properties map[string]interface{}) (driver.Driver, error) {
	var webhookCfg drivers.WebhookProducerConfig
	if err := mapstructure.Decode(properties, &webhookCfg); err != nil {
		return nil, fmt.Errorf("decoding Webhook producer Configuration Failed: %w", err)
	}
	return drivers.NewWebhookDriver(&webhookCfg)
}

func KafkaProducerDriver(properties map[string]interface{}) (driver.Driver, error) {
	var kafkaCfg drivers.KafkaProducerConfig
	if err := mapstructure.Decode(properties, &kafkaCfg); err != nil {
		return nil, fmt.Errorf("decoding kafka Configuration Failed: %w", err)
	}
	return drivers.NewKafkaProducer(&kafkaCfg)
}

func PollingProducerDriver(properties map[string]interface{}) (driver.Driver, error) {
	var pollingCfg drivers.PollingProducerConfig
	if err := mapstructure.Decode(properties, &pollingCfg); err != nil {
		return nil, fmt.Errorf("decoding polling producer Configuration Failed: %w", err)
	}
	return drivers.NewPollingProducer(&pollingCfg)
}
