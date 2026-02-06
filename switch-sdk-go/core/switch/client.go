package _switch

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/fatzeng/switch-components/pc"
)

// SwitchOptions switch操作项
type SwitchOptions struct {
	// 必填
	// NamespaceTag 命名空间(tag非name)
	NamespaceTag string
	// Env 环境(tag非name)
	EnvTag string
	// Domain 域名(用于监听配置等)
	Domain string

	// 非必填
	// ServiceName 服务名
	ServiceName string
	// Version 版本
	Version string

	// kafka consumer在驱动更新的时候的新驱动的校验超时时间
	KafkaConsumerReplaceDriverValidationTimeout string
	// kafka consumer在驱动更新的时候的新驱动稳定期
	KafkaConsumerReplaceDriverStabilityPeriod string
	// kafka consumer在驱动更新的时候是否校验brokers的位置
	KafkaConsumerVerifyBrokers bool

	// webhook consumer在驱动更新的时候的新驱动的校验超时时间
	WebhookConsumerReplaceDriverValidationTimeout string
	// webhook consumer在驱动更新的时候的新驱动稳定期
	WebhookConsumerReplaceDriverStabilityPeriod string

	// polling consumer在驱动更新的时候的新驱动的校验超时时间
	PollingConsumerReplaceDriverValidationTimeout string
	// polling consumer在驱动更新的时候的新驱动稳定期
	PollingConsumerReplaceDriverStabilityPeriod string

	// WebSocket连接配置 pc配置
	// ClientVersion 客户端协议版本
	ClientVersion string
	// RequestHeader 连接时携带的自定义Header
	RequestHeader http.Header
	// ReconnectStrategy 断线重连配置
	ReconnectStrategy *pc.ReconnectStrategy
	// Heartbeat 心跳发送间隔
	Heartbeat time.Duration
	// WriteTimeout 写入超时时间
	WriteTimeout time.Duration
	// ReadTimeout 读取超时时间
	ReadTimeout time.Duration
	// DialTimeout 连接超时时间
	DialTimeout time.Duration
}

// Option 操作项用于构建SwitchOptions
type Option func(ctx context.Context, c *SwitchClient)

// SwitchClient switch客户端
type SwitchClient struct {
	options     SwitchOptions
	initialized atomic.Bool
	startOnce   sync.Once
	cancelFun   context.CancelFunc
}

func (s *SwitchClient) StartOnce() *sync.Once {
	return &s.startOnce
}

func (s *SwitchClient) Options() SwitchOptions {
	return s.options
}

// NamespaceTag 获取命名空间
func (s *SwitchClient) NamespaceTag() string {
	return s.options.NamespaceTag
}

// EnvTag 获取环境
func (s *SwitchClient) EnvTag() string {
	return s.options.EnvTag
}

// Domain 获取域名
func (s *SwitchClient) Domain() string {
	return s.options.Domain
}

// ServiceName 获取服务名
func (s *SwitchClient) ServiceName() string {
	return s.options.ServiceName
}

// Version 获取版本
func (s *SwitchClient) Version() string {
	return s.options.Version
}

// ClientVersion 获取客户端协议版本
func (s *SwitchClient) ClientVersion() string {
	return s.options.ClientVersion
}

// RequestHeader 获取连接时携带的自定义Header
func (s *SwitchClient) RequestHeader() http.Header {
	return s.options.RequestHeader
}

// ReconnectStrategy 获取断线重连配置
func (s *SwitchClient) ReconnectStrategy() *pc.ReconnectStrategy {
	return s.options.ReconnectStrategy
}

// Heartbeat 获取心跳发送间隔
func (s *SwitchClient) Heartbeat() time.Duration {
	return s.options.Heartbeat
}

// WriteTimeout 获取写入超时时间
func (s *SwitchClient) WriteTimeout() time.Duration {
	return s.options.WriteTimeout
}

// ReadTimeout 获取读取超时时间
func (s *SwitchClient) ReadTimeout() time.Duration {
	return s.options.ReadTimeout
}

// DialTimeout 获取连接超时时间
func (s *SwitchClient) DialTimeout() time.Duration {
	return s.options.DialTimeout
}

// KafkaConsumerReplaceDriverValidationTimeout 获取kafka consumer在驱动更新的时候的新驱动的校验超时时间
func (s *SwitchClient) KafkaConsumerReplaceDriverValidationTimeout() (time.Duration, error) {
	duration, err := time.ParseDuration(s.options.KafkaConsumerReplaceDriverValidationTimeout)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// KafkaConsumerReplaceDriverStabilityPeriod 获取kafka consumer在驱动更新的时候的新驱动稳定期
func (s *SwitchClient) KafkaConsumerReplaceDriverStabilityPeriod() (time.Duration, error) {
	duration, err := time.ParseDuration(s.options.KafkaConsumerReplaceDriverStabilityPeriod)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// KafkaConsumerVerifyBrokers 获取kafka consumer在驱动更新的时候是否校验brokers的位置
func (s *SwitchClient) KafkaConsumerVerifyBrokers() bool {
	return s.options.KafkaConsumerVerifyBrokers
}

// WebhookConsumerReplaceDriverValidationTimeout 获取webhook consumer在驱动更新的时候的新驱动的校验超时时间
func (s *SwitchClient) WebhookConsumerReplaceDriverValidationTimeout() (time.Duration, error) {
	duration, err := time.ParseDuration(s.options.WebhookConsumerReplaceDriverValidationTimeout)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// WebhookConsumerReplaceDriverStabilityPeriod 获取webhook consumer在驱动更新的时候的新驱动稳定期
func (s *SwitchClient) WebhookConsumerReplaceDriverStabilityPeriod() (time.Duration, error) {
	duration, err := time.ParseDuration(s.options.WebhookConsumerReplaceDriverStabilityPeriod)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// PollingConsumerReplaceDriverValidationTimeout 获取polling consumer在驱动更新的时候的新驱动的校验超时时间
func (s *SwitchClient) PollingConsumerReplaceDriverValidationTimeout() (time.Duration, error) {
	duration, err := time.ParseDuration(s.options.PollingConsumerReplaceDriverValidationTimeout)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// PollingConsumerReplaceDriverStabilityPeriod 获取polling consumer在驱动更新的时候的新驱动稳定期
func (s *SwitchClient) PollingConsumerReplaceDriverStabilityPeriod() (time.Duration, error) {
	duration, err := time.ParseDuration(s.options.PollingConsumerReplaceDriverStabilityPeriod)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

func (s *SwitchClient) CancelFun(cancelFun context.CancelFunc) {
	s.cancelFun = cancelFun
}

func (s *SwitchClient) GetCancelFun() context.CancelFunc {
	return s.cancelFun
}

func (s *SwitchClient) IsInitialized() bool {
	return s.initialized.Load()
}

func (s *SwitchClient) Initialized(b bool) {
	s.initialized.Store(b)
}

var GlobalClient = &SwitchClient{}

// WithNamespaceTag 设置命名空间
func WithNamespaceTag(namespaceTag string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.NamespaceTag = namespaceTag
	}
}

// WithEnvTag 设置环境
func WithEnvTag(envTag string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.EnvTag = envTag
	}
}

// WithDomain 设置域名
func WithDomain(domain string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.Domain = domain
	}
}

// WithServiceName 设置服务名
func WithServiceName(serviceName string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.ServiceName = serviceName
	}
}

// WithVersion 设置版本
func WithVersion(version string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.Version = version
	}
}

// WithKafkaConsumerReplaceDriverValidationTimeout 设置kafka consumer在驱动更新的时候的新驱动的校验超时时间
func WithKafkaConsumerReplaceDriverValidationTimeout(timeout string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.KafkaConsumerReplaceDriverValidationTimeout = timeout
	}
}

// WithKafkaConsumerReplaceDriverStabilityPeriod 设置kafka consumer在驱动更新的时候的新驱动稳定期
func WithKafkaConsumerReplaceDriverStabilityPeriod(period string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.KafkaConsumerReplaceDriverStabilityPeriod = period
	}
}

// WithKafkaConsumerVerifyBrokers 设置kafka consumer在驱动更新的时候是否校验brokers的位置
func WithKafkaConsumerVerifyBrokers(verify bool) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.KafkaConsumerVerifyBrokers = verify
	}
}

// WithWebhookConsumerReplaceDriverValidationTimeout 设置webhook consumer在驱动更新的时候的新驱动的校验超时时间
func WithWebhookConsumerReplaceDriverValidationTimeout(timeout string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.WebhookConsumerReplaceDriverValidationTimeout = timeout
	}
}

// WithWebhookConsumerReplaceDriverStabilityPeriod 设置webhook consumer在驱动更新的时候的新驱动稳定期
func WithWebhookConsumerReplaceDriverStabilityPeriod(period string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.WebhookConsumerReplaceDriverStabilityPeriod = period
	}
}

// WithPollingConsumerReplaceDriverValidationTimeout 设置polling consumer在驱动更新的时候的新驱动的校验超时时间
func WithPollingConsumerReplaceDriverValidationTimeout(timeout string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.PollingConsumerReplaceDriverValidationTimeout = timeout
	}
}

// WithPollingConsumerReplaceDriverStabilityPeriod 设置polling consumer在驱动更新的时候的新驱动稳定期
func WithPollingConsumerReplaceDriverStabilityPeriod(period string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.PollingConsumerReplaceDriverStabilityPeriod = period
	}
}

// WithClientVersion 设置客户端协议版本
func WithClientVersion(clientVersion string) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.ClientVersion = clientVersion
	}
}

// WithRequestHeader 设置连接时携带的自定义Header
func WithRequestHeader(requestHeader http.Header) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.RequestHeader = requestHeader
	}
}

// WithReconnectStrategy 设置断线重连间隔
func WithReconnectStrategy(reconnectStrategy *pc.ReconnectStrategy) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.ReconnectStrategy = reconnectStrategy
	}
}

// WithHeartbeat 设置心跳发送间隔
func WithHeartbeat(heartbeat time.Duration) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.Heartbeat = heartbeat
	}
}

// WithWriteTimeout 设置写入超时时间
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.WriteTimeout = writeTimeout
	}
}

// WithReadTimeout 设置读取超时时间
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.ReadTimeout = readTimeout
	}
}

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(ctx context.Context, c *SwitchClient) {
		c.options.DialTimeout = dialTimeout
	}
}
