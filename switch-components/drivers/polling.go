package drivers

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const DefaultPollingPath = "/switch/polling"
const DefaultPollingPort = "10002"

// PollingConsumerConfig 长轮询配置
type PollingConsumerConfig struct {
	// 服务器URL配置(第一版admin跟server不分离)
	URL string `json:"url,omitempty" mapstructure:"url,omitempty" yaml:"url,omitempty"`

	// 轮询配置
	PollInterval   string `json:"poll_interval" mapstructure:"poll_interval" yaml:"poll_interval"`       // 普通轮询间隔
	RequestTimeout string `json:"request_timeout" mapstructure:"request_timeout" yaml:"request_timeout"` // HTTP请求超时时间，建议配置比服务端LongPollTimeout大

	// HTTP配置
	Headers   map[string]string `json:"headers,omitempty" mapstructure:"headers,omitempty" yaml:"headers,omitempty"`          // 自定义请求头
	UserAgent string            `json:"user_agent,omitempty" mapstructure:"user_agent,omitempty" yaml:"user_agent,omitempty"` // User-Agent

	// 客户端安全配置
	Security *PollingClientSecurityConfig `json:"security,omitempty" mapstructure:"security,omitempty" yaml:"security,omitempty"`

	// 其他配置
	IgnoreExceptions bool `json:"ignore_exceptions" mapstructure:"ignore_exceptions" yaml:"ignore_exceptions"` // 是否忽略异常

	Retry *RetryConfig `json:"retry,omitempty" mapstructure:"retry,omitempty" yaml:"retry,omitempty"`
}

// PollingClientSecurityConfig 客户端安全配置
type PollingClientSecurityConfig struct {
	// 客户端认证token
	Token string `json:"token,omitempty" mapstructure:"token,omitempty" yaml:"token,omitempty"`

	// HTTPS配置
	InsecureSkipVerify bool `json:"insecure_skip_verify" mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// hasURL 检查是否有URL配置
func (p *PollingConsumerConfig) hasURL() bool {
	return strings.TrimSpace(p.URL) != ""
}

// getValidURL 获取有效的URL
func (p *PollingConsumerConfig) getValidURL() (string, error) {
	rawURL := strings.TrimSpace(p.URL)
	if rawURL == "" {
		return "", fmt.Errorf("URL is empty")
	}

	// 解析URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// 验证协议：只支持http和https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("unsupported protocol: %s, only http and https are supported", parsedURL.Scheme)
	}

	// 验证主机部分
	if parsedURL.Host == "" {
		return "", fmt.Errorf("host cannot be empty")
	}

	// 不支持指定路径
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		return "", fmt.Errorf("path is not supported, use host only")
	}

	// 验证主机是否为有效的域名或IP
	hostOnly := parsedURL.Hostname()
	if hostOnly == "" {
		return "", fmt.Errorf("invalid host")
	}

	if err := validateDomainOrIP(hostOnly); err != nil {
		return "", fmt.Errorf("invalid host: %w", err)
	}

	return rawURL, nil
}

// getPollInterval 获取轮询间隔
func (p *PollingConsumerConfig) getPollInterval() time.Duration {
	if interval, err := time.ParseDuration(p.PollInterval); err == nil {
		return interval
	}
	return 30 * time.Second
}

// getRequestTimeout 获取请求超时时间
func (p *PollingConsumerConfig) getRequestTimeout() time.Duration {
	if timeout, err := time.ParseDuration(p.RequestTimeout); err == nil {
		return timeout
	}
	return 30 * time.Second
}

// getHeader 获取请求头
func (p *PollingConsumerConfig) getHeader() map[string]string {
	if p.Headers == nil {
		return make(map[string]string)
	}
	return p.Headers
}

// getUserAgent 获取代理
func (p *PollingConsumerConfig) getUserAgent() string {
	if p.UserAgent == "" {
		return "Switch-SDK-Polling/1.0"
	}
	return p.UserAgent
}

// getSecurity 获取安全配置
func (p *PollingConsumerConfig) getSecurity() *PollingClientSecurityConfig {
	return p.Security
}

// getIgnoreExceptions 获取是否忽略错误
func (p *PollingConsumerConfig) getIgnoreExceptions() bool {
	return p.IgnoreExceptions
}

// getRetry 获取重试配置
func (p *PollingConsumerConfig) getRetry() *RetryConfig {
	if p.Retry == nil {
		return &RetryConfig{
			Count:   5,
			Backoff: "3s",
		}
	}
	return p.Retry
}

// getWorkBackoffDuration 获取运行策略-重试退避时间
func (p *PollingConsumerConfig) getWorkBackoffDuration() time.Duration {
	if p.getRetry() != nil && p.getRetry().Backoff != "" {
		if duration, err := time.ParseDuration(p.getRetry().Backoff); err == nil {
			return duration
		}
	}
	return 3 * time.Second
}

// getWorkMaxRetries 获取运行策略-最大次数
func (p *PollingConsumerConfig) getWorkMaxRetries() int {
	if p.getRetry() != nil && p.getRetry().Count != 0 {
		return p.getRetry().Count
	}
	return 5
}

// isValid 验证配置
func (p *PollingConsumerConfig) isValid() error {
	if p == nil {
		return fmt.Errorf("polling config cannot be nil")
	}

	// 验证基础配置
	if err := p.validateBase(); err != nil {
		return fmt.Errorf("polling config validation failed: %w", err)
	}

	// 验证URL配置
	if err := p.validateURLs(); err != nil {
		return fmt.Errorf("polling config validation failed: %w", err)
	}

	// 验证安全配置
	if err := p.validateSecurity(); err != nil {
		return fmt.Errorf("polling config validation failed: %w", err)
	}

	return nil
}

// validateBase 验证基础配置
func (p *PollingConsumerConfig) validateBase() error {
	if !p.hasURL() {
		return fmt.Errorf("polling must have URL configured")
	}

	if p.Retry != nil {
		if p.Retry.Count < 0 {
			return fmt.Errorf("work retry count cannot be negative")
		}
		if p.Retry.Backoff != "" {
			if _, err := time.ParseDuration(p.Retry.Backoff); err != nil {
				return fmt.Errorf("invalid work retry backoff format '%s': %w", p.Retry.Backoff, err)
			}
		}
	}
	return nil
}

// validateURLs 验证URL配置
func (p *PollingConsumerConfig) validateURLs() error {
	return validateSingleURL(p.URL)
}

// validateSecurity 验证安全配置
func (p *PollingConsumerConfig) validateSecurity() error {
	if p.Security == nil {
		return nil
	}

	sec := p.Security

	// 验证Token
	if sec.Token != "" {
		if err := validateToken(sec.Token); err != nil {
			return fmt.Errorf("invalid client token: %w", err)
		}
	}

	return nil
}

// validateSingleURL 验证单个URL格式并返回主机名
func validateSingleURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("URL is empty")
	}

	// 解析URL
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// 验证协议：只支持http和https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported protocol: %s, only http and https are supported", parsedURL.Scheme)
	}

	// 验证主机部分
	if parsedURL.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// 不支持指定路径
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		return fmt.Errorf("path is not supported, use host only")
	}

	// 提取主机部分（去掉端口）
	hostOnly := parsedURL.Hostname()
	if hostOnly == "" {
		return fmt.Errorf("invalid host")
	}

	// 获取端口
	port := parsedURL.Port()
	if err = validatePort(port); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// 验证主机是否为有效的域名或IP
	if err = validateDomainOrIP(hostOnly); err != nil {
		return fmt.Errorf("invalid host/IP: %s, error: %w", hostOnly, err)
	}

	return nil
}

// validatePort 验证端口号
func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	portNum := 0
	if _, err := fmt.Sscanf(port, "%d", &portNum); err != nil {
		return fmt.Errorf("port must be a valid number: %s", port)
	}

	// 端口范围
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", portNum)
	}

	return nil
}

// validateDomainOrIP 验证域名或IP地址
func validateDomainOrIP(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// 检查是否为有效的IP地址
	if net.ParseIP(host) != nil {
		return nil
	}

	// 验证域名格式
	if len(host) == 0 {
		return fmt.Errorf("hostname cannot be empty")
	}

	// 基本域名格式检查
	if strings.Contains(host, "//") || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return fmt.Errorf("invalid domain format")
	}

	// 域名长度检查
	if len(host) > 253 {
		return fmt.Errorf("domain name too long")
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// PollingProducerConfig 长轮询生产者配置
type PollingProducerConfig struct {
	// 服务器端口配置(polling的server端端口)
	Port string `json:"port" mapstructure:"port" yaml:"port"`

	// 轮询配置
	LongPollTimeout string `json:"long_poll_timeout" mapstructure:"long_poll_timeout" yaml:"long_poll_timeout"` // 长轮询超时

	// 服务器超时配置
	ServerReadTimeout  string `json:"server_read_timeout" mapstructure:"server_read_timeout" yaml:"server_read_timeout"`    // 服务器读取超时
	ServerWriteTimeout string `json:"server_write_timeout" mapstructure:"server_write_timeout" yaml:"server_write_timeout"` // 服务器写入超时
	ServerIdleTimeout  string `json:"server_idle_timeout" mapstructure:"server_idle_timeout" yaml:"server_idle_timeout"`    // 服务器空闲超时

	// 服务端安全配置
	Security *PollingServerSecurityConfig `json:"security,omitempty" mapstructure:"security,omitempty" yaml:"security,omitempty"`

	Retry *RetryConfig `json:"retry,omitempty" mapstructure:"retry,omitempty" yaml:"retry,omitempty"`
}

// PollingServerSecurityConfig 服务端安全配置
type PollingServerSecurityConfig struct {
	// 服务端验证客户端token的配置
	ValidTokens []string `json:"valid_tokens,omitempty" mapstructure:"valid_tokens,omitempty" yaml:"valid_tokens,omitempty"`

	// HTTPS配置
	CertFile string `json:"cert_file,omitempty" mapstructure:"cert_file,omitempty" yaml:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty" mapstructure:"key_file,omitempty" yaml:"key_file,omitempty"`
}

// getPort 获取端口号
func (p *PollingProducerConfig) getPort() string {
	if p.Port == "" {
		return DefaultPollingPort
	}
	return p.Port
}

// getLongPollTimeout 获取长轮询超时时间
func (p *PollingProducerConfig) getLongPollTimeout() time.Duration {
	if timeout, err := time.ParseDuration(p.LongPollTimeout); err == nil {
		return timeout
	}
	return 60 * time.Second
}

// getServerReadTimeout 获取服务器读取超时时间
func (p *PollingProducerConfig) getServerReadTimeout() time.Duration {
	if timeout, err := time.ParseDuration(p.ServerReadTimeout); err == nil {
		return timeout
	}
	return p.getLongPollTimeout() + 30*time.Second
}

// getServerWriteTimeout 获取服务器写入超时时间
func (p *PollingProducerConfig) getServerWriteTimeout() time.Duration {
	if timeout, err := time.ParseDuration(p.ServerWriteTimeout); err == nil {
		return timeout
	}
	return p.getLongPollTimeout() + 10*time.Second
}

// getServerIdleTimeout 获取服务器空闲超时时间
func (p *PollingProducerConfig) getServerIdleTimeout() time.Duration {
	if timeout, err := time.ParseDuration(p.ServerIdleTimeout); err == nil {
		return timeout
	}
	return 120 * time.Second
}

// getSecurity 获取安全配置
func (p *PollingProducerConfig) getSecurity() *PollingServerSecurityConfig {
	return p.Security
}

// getRetry 获取重试配置
func (p *PollingProducerConfig) getRetry() *RetryConfig {
	return p.Retry
}

// getBackoffDuration 获取重试退避时间
func (p *PollingProducerConfig) getBackoffDuration() time.Duration {
	if p.getRetry() != nil && p.getRetry().Backoff != "" {
		if duration, err := time.ParseDuration(p.getRetry().Backoff); err == nil {
			return duration
		}
	}
	return 3 * time.Second
}

// getMaxRetries 获取最大重试次数
func (p *PollingProducerConfig) getMaxRetries() int {
	if p.getRetry() != nil && p.getRetry().Count != 0 {
		return p.getRetry().Count
	}
	return 5
}

// isValid 验证配置
func (p *PollingProducerConfig) isValid() error {
	if p == nil {
		return fmt.Errorf("polling producer config cannot be nil")
	}

	// 验证基础配置
	if err := p.validateBase(); err != nil {
		return fmt.Errorf("polling producer config validation failed: %w", err)
	}

	// 验证安全配置
	if err := p.validateSecurity(); err != nil {
		return fmt.Errorf("polling producer config validation failed: %w", err)
	}

	return nil
}

// validateBase 验证基础配置
func (p *PollingProducerConfig) validateBase() error {
	// 验证端口配置
	if err := validatePort(p.getPort()); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// 验证长轮询超时
	if p.LongPollTimeout != "" {
		if _, err := time.ParseDuration(p.LongPollTimeout); err != nil {
			return err
		}
	}

	// 验证服务器超时配置
	if p.ServerReadTimeout != "" {
		if _, err := time.ParseDuration(p.ServerReadTimeout); err != nil {
			return err
		}
	}

	if p.ServerWriteTimeout != "" {
		if timeout, err := time.ParseDuration(p.ServerWriteTimeout); err == nil {
			if timeout <= 0 {
				return fmt.Errorf("server write timeout must be positive")
			}
		}
	}

	if p.ServerIdleTimeout != "" {
		if _, err := time.ParseDuration(p.ServerIdleTimeout); err != nil {
			return err
		}
	}

	// 验证超时配置的合理性 读超时不能比轮询的超时还大
	if p.getServerReadTimeout() <= p.getLongPollTimeout() {
		return fmt.Errorf("server read timeout should be greater than long poll timeout")
	}

	if p.getServerWriteTimeout() <= p.getLongPollTimeout() {
		return fmt.Errorf("server write timeout should be greater than long poll timeout")
	}

	if p.Retry != nil {
		if p.Retry.Count < 0 {
			return fmt.Errorf("retry count cannot be negative")
		}
		if p.Retry.Backoff != "" {
			if _, err := time.ParseDuration(p.Retry.Backoff); err != nil {
				return fmt.Errorf("invalid retry backoff format '%s': %w", p.Retry.Backoff, err)
			}
		}
	}

	return nil
}

// validateSecurity 验证安全配置
func (p *PollingProducerConfig) validateSecurity() error {
	if p.Security == nil {
		return nil
	}

	sec := p.Security

	// 验证ValidTokens
	if sec.ValidTokens != nil && len(sec.ValidTokens) > 0 {
		for i, token := range sec.ValidTokens {
			if err := validateToken(token); err != nil {
				return fmt.Errorf("invalid token at index %d: %w", i, err)
			}
		}
	}

	// 验证HTTPS配置
	if err := p.validateHTTPS(); err != nil {
		return fmt.Errorf("HTTPS config validation failed: %w", err)
	}

	return nil
}

// validateToken 验证单个token
func validateToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token cannot be empty or whitespace only")
	}

	if len(token) < 8 {
		return fmt.Errorf("token must be at least 8 characters long")
	}

	if len(token) > 512 {
		return fmt.Errorf("token cannot exceed 512 characters")
	}

	// 检查token格式：不能包含空格、换行符等
	if strings.ContainsAny(token, " \t\n\r") {
		return fmt.Errorf("token cannot contain whitespace characters")
	}

	return nil
}

// validateHTTPS 验证HTTPS配置
func (p *PollingProducerConfig) validateHTTPS() error {
	if p.Security == nil {
		return nil
	}

	sec := p.Security

	if sec.CertFile != "" && sec.KeyFile == "" {
		return fmt.Errorf("key_file is required when cert_file is specified")
	}

	if sec.KeyFile != "" && sec.CertFile == "" {
		return fmt.Errorf("cert_file is required when key_file is specified")
	}

	return nil
}
