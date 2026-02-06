package switch_sdk_go

import (
	"context"
	"errors"
	"net/http"
	"time"

	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	_switch "gitee.com/fatzeng/switch-sdk-go/core/switch"
	"gitee.com/fatzeng/switch-sdk-go/internal/datasync"
)

// setDefaultWebSocketConfig 设置WebSocket连接配置和驱动配置的默认值
func setDefaultWebSocketConfig(ctx context.Context) {
	// WebSocket连接配置默认值
	if _switch.GlobalClient.ClientVersion() == "" {
		_switch.WithClientVersion("1.0.0")(ctx, _switch.GlobalClient)
	}

	if _switch.GlobalClient.ReconnectStrategy() == nil {
		_switch.WithReconnectStrategy(pc.DefaultReconnectStrategy())(ctx, _switch.GlobalClient)
	}

	if _switch.GlobalClient.RequestHeader() == nil {
		_switch.WithRequestHeader(make(http.Header))(ctx, _switch.GlobalClient)
	}

	if _switch.GlobalClient.Heartbeat() == 0 {
		_switch.WithHeartbeat(30*time.Second)(ctx, _switch.GlobalClient)
	}

	if _switch.GlobalClient.WriteTimeout() == 0 {
		_switch.WithWriteTimeout(1000*time.Second)(ctx, _switch.GlobalClient)
	}

	if _switch.GlobalClient.ReadTimeout() == 0 {
		_switch.WithReadTimeout(2000*time.Second)(ctx, _switch.GlobalClient)
	}

	if _switch.GlobalClient.DialTimeout() == 0 {
		_switch.WithDialTimeout(10*time.Second)(ctx, _switch.GlobalClient)
	}

	// Kafka Consumer驱动配置默认值
	setKafkaConsumerDefaults(ctx)

	// Webhook Consumer驱动配置默认值
	setWebhookConsumerDefaults(ctx)

	// Polling Consumer驱动配置默认值
	setPollingConsumerDefaults(ctx)

	logger.Logger.Info("WebSocket connection and driver config defaults applied")
}

// setKafkaConsumerDefaults 设置Kafka Consumer驱动配置默认值
func setKafkaConsumerDefaults(ctx context.Context) {
	// 校验超时时间
	timeout, err := _switch.GlobalClient.KafkaConsumerReplaceDriverValidationTimeout()
	if err != nil || timeout <= 0 {
		if err != nil {
			logger.Logger.Debugf("Kafka consumer timeout not configured, using default: %v", err)
		}
		_switch.WithKafkaConsumerReplaceDriverValidationTimeout("60s")(ctx, _switch.GlobalClient)
	}

	// 稳定期
	period, err := _switch.GlobalClient.KafkaConsumerReplaceDriverStabilityPeriod()
	if err != nil || period <= 0 {
		if err != nil {
			logger.Logger.Debugf("Kafka consumer stability period not configured, using default: %v", err)
		}
		_switch.WithKafkaConsumerReplaceDriverStabilityPeriod("5s")(ctx, _switch.GlobalClient)
	}
}

// setWebhookConsumerDefaults 设置Webhook Consumer驱动配置默认值
func setWebhookConsumerDefaults(ctx context.Context) {
	// 校验超时时间
	timeout, err := _switch.GlobalClient.WebhookConsumerReplaceDriverValidationTimeout()
	if err != nil || timeout <= 0 {
		if err != nil {
			logger.Logger.Debugf("Webhook consumer timeout not configured, using default: %v", err)
		}
		_switch.WithWebhookConsumerReplaceDriverValidationTimeout("60s")(ctx, _switch.GlobalClient)
	}

	// 稳定期
	period, err := _switch.GlobalClient.WebhookConsumerReplaceDriverStabilityPeriod()
	if err != nil || period <= 0 {
		if err != nil {
			logger.Logger.Debugf("Webhook consumer stability period not configured, using default: %v", err)
		}
		_switch.WithWebhookConsumerReplaceDriverStabilityPeriod("5s")(ctx, _switch.GlobalClient)
	}
}

// setPollingConsumerDefaults 设置Polling Consumer驱动配置默认值
func setPollingConsumerDefaults(ctx context.Context) {
	// 校验超时时间
	timeout, err := _switch.GlobalClient.PollingConsumerReplaceDriverValidationTimeout()
	if err != nil || timeout <= 0 {
		if err != nil {
			logger.Logger.Debugf("Polling consumer timeout not configured, using default: %v", err)
		}
		_switch.WithPollingConsumerReplaceDriverValidationTimeout("60s")(ctx, _switch.GlobalClient)
	}

	// 稳定期
	period, err := _switch.GlobalClient.PollingConsumerReplaceDriverStabilityPeriod()
	if err != nil || period <= 0 {
		if err != nil {
			logger.Logger.Debugf("Polling consumer stability period not configured, using default: %v", err)
		}
		_switch.WithPollingConsumerReplaceDriverStabilityPeriod("5s")(ctx, _switch.GlobalClient)
	}
}

// Start 初始化switch
func Start(ctx context.Context, opts ..._switch.Option) error {
	var err error
	_switch.GlobalClient.StartOnce().Do(func() {
		for _, opt := range opts {
			opt(ctx, _switch.GlobalClient)
		}

		if _switch.GlobalClient.Domain() == "" {
			err = errors.New("domain must be provided")
			return
		}

		if _switch.GlobalClient.NamespaceTag() == "" {
			err = errors.New("NamespaceTag must be provided")
			return
		}

		if _switch.GlobalClient.EnvTag() == "" {
			err = errors.New("EnvTag must be provided")
			return
		}

		// 设置WebSocket连接配置的默认值
		setDefaultWebSocketConfig(ctx)

		_, cancelFunc := context.WithCancel(ctx)
		_switch.GlobalClient.CancelFun(cancelFunc)

		logger.Logger.Info("Switch SDK starting initialization...")

		//开始同步，包含配置监听以及开关的全量
		if err = datasync.StartSyncer(ctx); err != nil {
			return
		}

		_switch.GlobalClient.Initialized(true)
		logger.Logger.Info("Switch SDK initialized successfully.")
	})

	return err
}

// Shutdown 停机
func Shutdown() {
	logger.Logger.Info("Shutting down Switch SDK...")
	if _switch.GlobalClient.GetCancelFun() != nil {
		_switch.GlobalClient.GetCancelFun()()
	}
	datasync.StopSyncer()
	//清理掉所有缓存
	_switch.ClearAllRules()
	logger.Logger.Info("Shutting down Switch SDK Over...")
}
