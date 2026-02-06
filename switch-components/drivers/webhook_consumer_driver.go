package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

const WebhookConsumerDriverType driver.DriverType = "webhook_consumer"

// WebhookMessageHandler webhook消息处理函数
type WebhookMessageHandler func(ctx context.Context, body json.RawMessage) error

// WebhookConsumer webhook消费者
type WebhookConsumer struct {
	WebhookConsumerValidator
	server       *http.Server
	config       *WebhookConsumerConfig
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
	handler      WebhookMessageHandler
	driverName   string
	callback     driver.DriverFailureCallback
	failureCount int // 连续失败次数
}

// NewWebhookConsumer 创建webhook消费者驱动
func NewWebhookConsumer(c *WebhookConsumerConfig, handler WebhookMessageHandler) (*WebhookConsumer, error) {
	if handler == nil {
		return nil, fmt.Errorf("webhook message handler cannot be nil")
	}

	consumer := &WebhookConsumer{
		config:       c,
		handler:      handler,
		failureCount: 0,
	}

	return consumer, nil
}

func (w *WebhookConsumer) RecreateFromConfig() (driver.Driver, error) {
	return NewWebhookConsumer(w.config, w.handler)
}

// Close 关闭webhook消费者
func (w *WebhookConsumer) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	logger.Logger.Info("Closing webhook consumer")

	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}

	if w.server != nil {
		logger.Logger.Info("Shutting down webhook HTTP server gracefully")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := w.server.Shutdown(shutdownCtx); err != nil {
			logger.Logger.Errorf("Failed to shutdown HTTP server gracefully: %v", err)

			// 优雅关闭失败，强制关闭
			logger.Logger.Warn("Forcing HTTP server to close")
			if closeErr := w.server.Close(); closeErr != nil {
				logger.Logger.Errorf("Failed to force close HTTP server: %v", closeErr)
				return closeErr
			}
			return err
		}

		logger.Logger.Info("Webhook HTTP server shutdown successfully")
		w.server = nil
	}

	return nil
}

// Start 启动webhook消费者
func (w *WebhookConsumer) Start(ctx context.Context) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.startInternal(ctx)
	return nil
}

// startInternal 内部启动逻辑
func (w *WebhookConsumer) startInternal(ctx context.Context) {
	// 保存原始 context，用于重启
	if w.ctx == nil {
		w.ctx = ctx
	}

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	backoffDuration := w.config.getBackoffDuration()
	recovery.SafeGo(ctx, func(ctx context.Context) error {
		return w.startHTTPServer(ctx)
	}, string(WebhookConsumerDriverType), recovery.WithRetryInterval(backoffDuration))

	logger.Logger.Infof("Webhook consumer started on address: %s (retry backoff: %v)", w.getListenAddress(), backoffDuration)
}

func (w *WebhookConsumer) GetDriverName() string {
	return w.driverName
}

func (w *WebhookConsumer) SetDriverMeta(name string) {
	w.driverName = name
}

func (w *WebhookConsumer) SetFailureCallback(callback driver.DriverFailureCallback) {
	w.callback = callback
}

func (w *WebhookConsumer) GetDriverType() driver.DriverType {
	return WebhookConsumerDriverType
}

// startHTTPServer 启动webhook
func (w *WebhookConsumer) startHTTPServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(WebhookReceivePoint, w.handleWebhookRequest)

	w.server = &http.Server{
		Addr:    w.getListenAddress(),
		Handler: mux,
	}

	logger.Logger.Infof("Starting webhook HTTP server on %s", w.server.Addr)

	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Errorf("Webhook HTTP server goroutine panicked: %v", r)
				// 不调用 stopService，让 SafeGo 处理重启
				errCh <- fmt.Errorf("webhook http server panic: %v", r)
			}
		}()

		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Errorf("Webhook HTTP server failed: %v", err)
			// 不调用 stopService，让 SafeGo 处理重启
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Logger.Info("Webhook consumer context cancelled, removing from driver pool")

		if err := w.Close(); err != nil {
			logger.Logger.Errorf("Failed to close webhook consumer: %v", err)
		}

		// 重置失败计数
		w.mutex.Lock()
		w.failureCount = 0
		w.mutex.Unlock()

		// 触发清理回调
		if w.callback != nil {
			w.callback("webhook consumer context cancelled", fmt.Errorf("driver context cancelled"))
		}

		return nil
	case err := <-errCh:
		w.mutex.Lock()
		w.failureCount++
		currentFailures := w.failureCount
		w.mutex.Unlock()

		retries := w.config.getMaxRetries()

		logger.Logger.Errorf("Webhook HTTP server error (failure %d/%d): %v", currentFailures, retries, err)

		// 如果超过最大重试次数，触发故障回调
		if currentFailures >= retries {
			logger.Logger.Errorf("Webhook consumer reached max retries (%d), triggering failure callback", retries)

			if err := w.Close(); err != nil {
				logger.Logger.Errorf("Failed to close webhook consumer: %v", err)
			}

			// 触发回调通知
			if w.callback != nil {
				w.callback("retry_exhausted", fmt.Errorf("max retries exhausted: %w", err))
			}
			return nil
		}

		return err
	}
}

// handleWebhookRequest 处理webhook请求
func (w *WebhookConsumer) handleWebhookRequest(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Logger.Errorf("Failed to read webhook request body: %v", err)
		http.Error(writer, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	// 验证安全配置
	if w.config.getSecurity() != nil && w.config.getSecurity().Secret != "" {
		if !w.validateSecret(request, body) {
			logger.Logger.Warn("Webhook request failed security validation")
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	logger.Logger.Debugf("Received webhook request: body_size=%d", len(body))

	// 处理消息
	ctx := request.Context()
	if err := w.handler(ctx, body); err != nil {
		logger.Logger.Errorf("Failed to process webhook message: %v", err)
	}

	// 返回成功响应
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Webhook successfully received",
		"time":    time.Now(),
	})
}

// validateSecret 验证安全密钥
// 使用HMAC-SHA256签名实现验证
// Producer发送时需要在请求头中添加X-Webhook-Signature
func (w *WebhookConsumer) validateSecret(request *http.Request, body []byte) bool {
	if w.config.getSecurity() == nil || w.config.getSecurity().Secret == "" {
		return true
	}

	// 获取签名头
	signature := GetSignature(request.Header)
	if signature == "" {
		logger.Logger.Warn("Missing X-Webhook-Signature header in webhook request")
		return false
	}

	//使用HMAC校验签名
	return ValidateHMACSignature(signature, body, w.config.getSecurity().Secret)
}

// getListenAddress 获取监听地址
func (w *WebhookConsumer) getListenAddress() string {
	if strings.Contains(w.config.getPort(), ":") {
		return w.config.getPort()
	}
	return ":" + w.config.getPort()
}
