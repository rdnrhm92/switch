package drivers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/tool"
)

// 客户端记录当前版本号 每次长轮询前都携带版本号 version是最新的
var version atomic.Uint64

const PollingConsumerDriverType driver.DriverType = "polling_consumer"

type PollingMessageHandler func(ctx context.Context, data json.RawMessage) error

// PollingConsumer 长轮询消费者
type PollingConsumer struct {
	PollingConsumerValidator
	client  *http.Client
	config  *PollingConsumerConfig
	ctx     context.Context
	cancel  context.CancelFunc
	mutex   sync.RWMutex
	handler PollingMessageHandler
	running bool

	publicIPs   []string
	internalIPs []string

	driverName   string
	callback     driver.DriverFailureCallback
	failureCount int // 连续失败次数
}

// NewPollingConsumer 创建长轮询消费者驱动
func NewPollingConsumer(c *PollingConsumerConfig, handler PollingMessageHandler) (*PollingConsumer, error) {
	if handler == nil {
		return nil, fmt.Errorf("polling message handler cannot be nil")
	}

	// 创建HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	// 配置TLS
	if c.getSecurity() != nil && c.getSecurity().InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &PollingConsumer{
		client:  &http.Client{Transport: transport, Timeout: c.getRequestTimeout()},
		config:  c,
		ctx:     ctx,
		cancel:  cancel,
		handler: handler,
		running: false,
	}

	// 初始化网络信息
	if err := consumer.initNetworkInfo(); err != nil {
		logger.Logger.Warnf("Failed to initialize network info: %v", err)
	}

	return consumer, nil
}

func (p *PollingConsumer) RecreateFromConfig() (driver.Driver, error) {
	return NewPollingConsumer(p.config, p.handler)
}

func (p *PollingConsumer) GetDriverName() string {
	return p.driverName
}

func (p *PollingConsumer) SetDriverMeta(name string) {
	p.driverName = name
}

func (p *PollingConsumer) SetFailureCallback(callback driver.DriverFailureCallback) {
	p.callback = callback
}

// GetDriverType 获取驱动类型
func (p *PollingConsumer) GetDriverType() driver.DriverType {
	return PollingConsumerDriverType
}

// initNetworkInfo 初始化网络信息
func (p *PollingConsumer) initNetworkInfo() error {
	// 获取内网IP
	internalIPs, err := tool.GetLocalIPs()
	if err != nil {
		logger.Logger.Warnf("Failed to get internal IPs: %v", err)
	} else {
		p.internalIPs = internalIPs
		logger.Logger.Infof("Internal IPs: %v", p.internalIPs)
	}

	// 获取公网IP
	publicIP, err := tool.GetPublicIP()
	if err != nil {
		logger.Logger.Warnf("Failed to get public IP: %v", err)
	} else {
		p.publicIPs = []string{publicIP}
		logger.Logger.Infof("Public IP: %s", publicIP)
	}

	return nil
}

// Start 启动长轮询消费者
func (p *PollingConsumer) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		logger.Logger.Warnf("PollingConsumer already running")
		return fmt.Errorf("PollingConsumer already running")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.running = true

	backoffDuration := p.config.getWorkBackoffDuration()
	recovery.SafeGo(p.ctx, func(ctx context.Context) error {
		return p.startPolling(ctx)
	}, string(PollingConsumerDriverType), recovery.WithRetryInterval(backoffDuration))

	logger.Logger.Infof("Polling consumer started (retry backoff: %v)", backoffDuration)
	return nil
}

// Close 关闭长轮询消费者
func (p *PollingConsumer) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	logger.Logger.Info("Closing polling consumer...")

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	p.running = false
	logger.Logger.Info("Polling consumer closed successfully")

	return nil
}

// startPolling 开始轮询
func (p *PollingConsumer) startPolling(ctx context.Context) error {
	validUrl, err := p.config.getValidURL()
	if err != nil {
		logger.Logger.Errorf("Failed to get valid URL: %v", err)
		return fmt.Errorf("failed to get valid URL: %w", err)
	}

	validUrl = validUrl + DefaultPollingPath

	errCh := make(chan error, 1)

	// 开启长轮询
	go func(validUrl string) {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Errorf("Polling goroutine panicked: %v", r)
				errCh <- fmt.Errorf("polling panic: %v", r)
			}
		}()

		if err := p.pollURL(validUrl, 0); err != nil {
			errCh <- err
		}
	}(validUrl)

	select {
	case <-ctx.Done():
		logger.Logger.Info("Polling consumer context cancelled, removing from driver pool")

		if err := p.Close(); err != nil {
			logger.Logger.Errorf("Failed to close polling consumer: %v", err)
		}

		p.mutex.Lock()
		p.failureCount = 0
		p.mutex.Unlock()

		// 触发清理回调
		if p.callback != nil {
			p.callback("polling consumer context cancelled", fmt.Errorf("driver context cancelled"))
		}

		return nil

	case err := <-errCh:
		p.mutex.Lock()
		p.failureCount++
		currentFailures := p.failureCount
		p.mutex.Unlock()
		retries := p.config.getWorkMaxRetries()

		logger.Logger.Errorf("Polling consumer error (failure %d/%d): %v", currentFailures, retries, err)

		// 如果超过最大重试次数，触发故障回调
		if currentFailures >= retries {
			logger.Logger.Errorf("Polling consumer reached max retries (%d), triggering failure callback", retries)

			if err := p.Close(); err != nil {
				logger.Logger.Errorf("Failed to close polling consumer: %v", err)
			}

			// 触发回调通知
			if p.callback != nil {
				p.callback("polling consumer retry exhausted", fmt.Errorf("max retries exhausted: %w", err))
			}

			// 返回nil safe go不再重试
			return nil
		}

		// 继续重试
		return err
	}
}

// pollURL 对单个URL进行长轮询，持续轮询直到遇到致命错误或被取消
// 职责：只负责持续轮询，遇到错误立即返回给上层处理，不做重试
func (p *PollingConsumer) pollURL(url string, urlIndex int) error {
	logger.Logger.Infof("Starting long polling for URL[%d]: %s", urlIndex, url)

	ticker := time.NewTicker(p.config.getPollInterval())
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			logger.Logger.Infof("Long polling stopped for URL[%d]: %s", urlIndex, url)
			return nil
		default:
			// 开始长轮询
			err := p.poll(url)

			if err != nil {
				// 检查错误类型
				var pollingErr *PollingError
				if !errors.As(err, &pollingErr) {
					// 未知错误，视为致命错误，返回给上层处理
					logger.Logger.Errorf("Unknown error for URL[%d] %s: %v", urlIndex, url, err)
					return fmt.Errorf("unknown error for URL[%d]: %w", urlIndex, err)
				}

				// 根据错误类型处理
				switch pollingErr.Type {
				case TimeoutError:
					// 超时是正常的长轮询行为，直接继续下一次轮询
					logger.Logger.Debug("Long polling timeout, continuing next poll")
					continue

				case NeedRetry:
					// 可重试错误，返回给上层让 startPolling 决定是否重启整个轮询任务
					logger.Logger.Warnf("Retryable error for URL[%d] %s: %v", urlIndex, url, err)
					return fmt.Errorf("retryable error for URL[%d]: %w", urlIndex, err)
				}

			} else {
				// 成功，使用配置的轮询间隔，避免连续请求
				if p.config.getPollInterval() > 0 {
					logger.Logger.Debugf("Waiting %v before next poll for URL[%d]", p.config.getPollInterval(), urlIndex)

					select {
					case <-ticker.C:
						// 继续下一次轮询
					case <-p.ctx.Done():
						return nil
					}
				}
			}
		}
	}
}

// ErrorType 通信的错误类型
type ErrorType int

const (
	TimeoutError ErrorType = iota // 超时错误 - 正常的长轮询行为不属于错误
	NeedRetry                     // 需要重试的错误 - 不详细区分到底是什么原因了总之要重试
)

// PollingError 轮训过程中的错误信息
type PollingError struct {
	Type ErrorType
	Err  error
}

func (pe *PollingError) Error() string {
	return pe.Err.Error()
}

func (pe *PollingError) Unwrap() error {
	return pe.Err
}

// poll 执行轮询
func (p *PollingConsumer) poll(url string) error {
	req, err := p.createRequest(url)
	if err != nil {
		return &PollingError{
			Type: NeedRetry,
			Err:  fmt.Errorf("failed to create request: %w", err),
		}
	}

	// 设置请求超时
	ctx, cancel := context.WithTimeout(p.ctx, p.config.getRequestTimeout())
	defer cancel()
	req = req.WithContext(ctx)

	// 发送请求 服务端夯住
	resp, err := p.client.Do(req)
	if err != nil {
		// 区分是超时错误还是异常
		errorType := p.classifyRequestError(err, resp)
		return &PollingError{
			Type: errorType,
			Err:  fmt.Errorf("failed to send request: %w", err),
		}
	}

	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			return &PollingError{
				Type: TimeoutError,
				Err:  fmt.Errorf("long poll request timed out"),
			}
		}

		// 处理响应
		err = p.handleResponse(resp)
		if err != nil {
			// 能正常拿到结果就不关心处理的异常了，记录一下错误后 继续等待下一次的轮训
			logger.Logger.Errorf("Polling consumer handleResponse error for URL[%v]: %v", url, err)
			return nil
		}
	}
	return nil
}

// classifyRequestError 分类请求错误
func (p *PollingConsumer) classifyRequestError(err error, resp *http.Response) ErrorType {
	if resp != nil && resp.StatusCode == http.StatusNoContent {
		return TimeoutError
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return TimeoutError
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return TimeoutError
	}

	if urlErr, ok := err.(*url.Error); ok {
		if netErr, ok := urlErr.Err.(net.Error); ok && netErr.Timeout() {
			return TimeoutError
		}
		return NeedRetry
	}

	if _, ok := err.(net.Error); ok {
		return NeedRetry
	}

	return NeedRetry
}

// createRequest 创建HTTP请求
func (p *PollingConsumer) createRequest(url string) (*http.Request, error) {
	latestVersion := version.Load()
	url = url + "?version=" + strconv.FormatUint(latestVersion, 10)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置基础头部
	req.Header.Set("User-Agent", p.config.getUserAgent())
	req.Header.Set("Accept", "application/json")

	// 添加客户端网络信息(服务端需要用来受信判断)
	if len(p.publicIPs) > 0 {
		req.Header.Set("X-Public-IPs", strings.Join(p.publicIPs, ","))
	}
	if len(p.internalIPs) > 0 {
		req.Header.Set("X-Internal-IPs", strings.Join(p.internalIPs, ","))
	}

	// 自定义头部
	for key, value := range p.config.getHeader() {
		req.Header.Set(key, value)
	}

	// 安全配置
	if p.config.getSecurity() != nil {
		p.setAuthHeaders(req)
	}

	return req, nil
}

// handleResponse 处理HTTP响应
func (p *PollingConsumer) handleResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// 处理消息
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		logger.Logger.Debug("Received empty response")
		return nil
	}

	return p.processMessage(body)
}

// processMessage 处理消息
func (p *PollingConsumer) processMessage(body []byte) error {
	logger.Logger.Debugf("Processing message: %d bytes", len(body))

	var configs []*ConfigVersion
	if err := json.Unmarshal(body, &configs); err != nil {
		return fmt.Errorf("failed to parse JSON message to ConfigVersions: %w", err)
	}

	if len(configs) == 0 {
		logger.Logger.Warnf("No ConfigVersions found in message")
		return nil
	}

	// 最大的版本号 暂存一下 防止频繁的cas
	var maxVersion uint64 = 0
	for _, config := range configs {
		bytes, err := json.Marshal(config.Config)
		if err != nil {
			return fmt.Errorf("failed to Marshal config: %v", err)
		}

		// 处理消息
		if err = p.handler(p.ctx, bytes); err != nil {
			if p.config.getIgnoreExceptions() {
				logger.Logger.Warnf("Message handler error (ignored): %v", err)
				return nil
			}
			return fmt.Errorf("message handler error: %w", err)
		}

		if config.Version > maxVersion {
			maxVersion = config.Version
		}
	}

	// 更新版本号 必须是最大的更新 防止版本号回退
	if maxVersion > 0 {
		for {
			currentVersion := version.Load()
			if maxVersion <= currentVersion {
				break
			}
			if version.CompareAndSwap(currentVersion, maxVersion) {
				logger.Logger.Debugf("Updated client version from %d to %d", currentVersion, maxVersion)
				break
			}
		}
	}

	logger.Logger.Debug("Message processed successfully")
	return nil
}

// setAuthHeaders 设置认证头部
func (p *PollingConsumer) setAuthHeaders(req *http.Request) {
	sec := p.config.getSecurity()

	// Token认证
	if sec.Token != "" {
		req.Header.Set("Authorization", "Bearer "+sec.Token)
	}
}

// IsRunning 检查是否正在运行
func (p *PollingConsumer) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}
