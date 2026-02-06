package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

const WebhookProducerDriverType driver.DriverType = "webhook_producer"

// WebhookProducer webhook-生产者驱动
type WebhookProducer struct {
	WebhookProducerValidator
	client     *http.Client
	config     *WebhookProducerConfig
	driverName string
}

func NewWebhookDriver(w *WebhookProducerConfig) (d *WebhookProducer, err error) {
	if w == nil {
		return nil, fmt.Errorf("webhook producer Incorrect configuration")
	}
	return &WebhookProducer{
		client: &http.Client{
			Timeout: w.getTimeOut(),
		},
		config: w,
	}, nil
}

func (w *WebhookProducer) RecreateFromConfig() (driver.Driver, error) {
	return NewWebhookDriver(w.config)
}

func (w *WebhookProducer) Start(ctx context.Context) error {
	return nil
}

func (w *WebhookProducer) GetClient() *http.Client {
	return w.client
}

func (w *WebhookProducer) Close() error {
	return nil
}

func (w *WebhookProducer) GetDriverName() string {
	return w.driverName
}

func (w *WebhookProducer) SetDriverMeta(name string) {
	w.driverName = name
}

func (w *WebhookProducer) SetFailureCallback(callback driver.DriverFailureCallback) {
	// 作为调用方 不存在驱动异常关闭 不需要回调
}

func (w *WebhookProducer) GetDriverType() driver.DriverType {
	return WebhookProducerDriverType
}

// getWebhookURLs 获取webhook URL列表
func (w *WebhookProducer) getWebhookURLs() []string {
	reachableIPs := GetReachableIPs()

	var urls []string

	if len(reachableIPs) > 0 {
		// 使用全局IP池中可达的IP
		for _, ip := range reachableIPs {
			// 黑名单过滤
			if w.config.IsIPBlacklisted(ip, w.config.getBlacklistIPs()) {
				logger.Logger.Debugf("Skipping blacklisted IP: %s", ip)
				continue
			}

			url := BuildWebhookUrl(ip, w.config.getPort())
			urls = append(urls, url)
		}
	} else {
		// 如果没有可达IP，回退为全部IP做尝试
		logger.Logger.Warn("No reachable IPs available, using all IPs as fallback")
		allIPs := GetAllIPs()
		for _, ip := range allIPs {
			if w.config.IsIPBlacklisted(ip, w.config.getBlacklistIPs()) {
				logger.Logger.Debugf("Skipping blacklisted IP: %s", ip)
				continue
			}
			url := BuildWebhookUrl(ip, w.config.getPort())
			urls = append(urls, url)
		}
	}

	return urls
}

func (w *WebhookProducer) Notify(ctx context.Context, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialize webhook data: %w", err)
	}

	client := w.GetClient()
	urls := w.getWebhookURLs()

	if len(urls) == 0 {
		logger.Logger.Warn("No webhook URLs available for notification")
		return nil
	}

	logger.Logger.Infof("Sending webhook notifications to %d URLs. Payload size: %d bytes", len(urls), len(jsonData))

	var wg sync.WaitGroup
	errorChan := make(chan error, len(urls))

	for _, url := range urls {
		requestURL := url
		wg.Add(1)
		go func() {
			defer wg.Done()
			var resp *http.Response
			var err error

			//设置重试间隔跟重试次数
			maxAttempts := w.config.getWorkMaxRetries()
			backoff := w.config.getWorkBackoffDuration()

			var lastErr error
			var lastResp *http.Response

			ticker := time.NewTicker(backoff)
			defer ticker.Stop()

			for attempt := 0; attempt < maxAttempts; attempt++ {
				bodyReader := bytes.NewReader(jsonData)
				req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bodyReader)
				if reqErr != nil {
					logger.Logger.Errorf("Failed to create webhook request for URL '%s': %v", requestURL, reqErr)
					lastErr = reqErr
					break
				}
				req.Header.Set("Content-Type", "application/json")

				// 接口安全配置sha256
				if w.config.Security != nil && w.config.Security.Secret != "" {
					headerKey, signature := BuildSignature(w.config.Security.Secret, jsonData)
					req.Header.Set(headerKey, signature)
					timestamp := strconv.FormatInt(time.Now().Unix(), 10)
					req.Header.Set("X-Switch-Timestamp", timestamp)
				}

				resp, err = client.Do(req)
				lastErr = err
				lastResp = resp

				// 判断是否需要重试
				needRetry := false
				if err != nil {
					logger.Logger.Errorf("Failed to send webhook request for URL '%s': %v", requestURL, err)
					lastErr = err
					needRetry = true
				} else if resp.StatusCode >= 500 {
					bodyBytes, _ := io.ReadAll(resp.Body)
					lastErr = fmt.Errorf("send webhook successfully but response status is not ok URL '%s': %v", requestURL, string(bodyBytes))
					logger.Logger.Errorf(lastErr.Error())
					needRetry = true
				} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					logger.Logger.Infof("Successfully sent webhook notification to URL '%s'", requestURL)
					lastErr = nil
					break
				} else {
					// 1xx跟4xx统一算失败，记录一下
					bodyBytes, _ := io.ReadAll(resp.Body)
					lastErr = fmt.Errorf("send webhook successfully but response status is not ok URL '%s': %v", requestURL, string(bodyBytes))
					logger.Logger.Errorf(lastErr.Error())
					needRetry = true
				}

				// 最后一次重试还有错误 进行收集
				if attempt == maxAttempts-1 {
					if lastErr != nil {
						errorChan <- lastErr
					}
					break
				} else {
					if needRetry {
						logger.Logger.Warnf("Attempt %d/%d: Failed to send webhook to URL '%s', retrying in %v... ",
							attempt+1, maxAttempts, requestURL, backoff)

						select {
						case <-ticker.C:
						case <-ctx.Done():
							errorChan <- ctx.Err()
							return
						}
					}
				}
			}

			// 使用最后一次的结果
			if lastResp != nil {
				defer resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	close(errorChan)

	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		combinedError := fmt.Errorf("webhook notifications failed: %s", strings.Join(errorMessages, "; "))

		logger.Logger.Errorf("One or more webhook notifications failed: %v", combinedError)

		if w.config.getIgnoreExceptions() {
			logger.Logger.Infof("Ignoring webhook notification errors due to IgnoreExceptions=true, record error is %v", combinedError)
			return nil
		}
		return combinedError
	} else {
		logger.Logger.Debugf("All webhook notifications sent successfully")
		return nil
	}
}
