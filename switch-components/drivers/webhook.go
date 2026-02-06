package drivers

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const WebhookReceivePoint = "/switch/webhook"
const DefaultProto = "http://"
const DefaultWebhookPort = "20002"

// BuildWebhookUrl 构建webhook地址(客户端使用配置中的端口避免端口冲突)
func BuildWebhookUrl(ip, port string) string {
	return DefaultProto + ip + ":" + port + WebhookReceivePoint
}

type BlacklistConfigAssistance struct {
}

// HasBlacklistIPs 检查是否有黑名单IP配置
func (b *BlacklistConfigAssistance) HasBlacklistIPs(blacklistIPs []string) bool {
	return len(blacklistIPs) > 0
}

// IsIPBlacklisted 检查IP是否在黑名单中
func (b *BlacklistConfigAssistance) IsIPBlacklisted(ip string, blacklistIPs []string) bool {
	for _, blacklistIP := range blacklistIPs {
		if blacklistIP == ip {
			return true
		}
	}
	return false
}

// WebhookProducerConfig webhook producer配置
type WebhookProducerConfig struct {
	BlacklistConfigAssistance
	BlacklistIPs []string `json:"blacklistIPs,omitempty" mapstructure:"blacklistIPs,omitempty" yaml:"blacklistIPs,omitempty"`
	// 客户端端口配置
	Port string `json:"port" mapstructure:"port" yaml:"port"`
	// 基础配置
	IgnoreExceptions bool                   `json:"ignoreExceptions" mapstructure:"ignoreExceptions" yaml:"ignoreExceptions"`
	TimeOut          string                 `json:"timeOut" mapstructure:"timeOut" yaml:"timeOut"`
	Retry            *RetryConfig           `json:"retry,omitempty" mapstructure:"retry,omitempty" yaml:"retry,omitempty"`
	Security         *WebhookSecurityConfig `json:"security,omitempty" mapstructure:"security,omitempty" yaml:"security,omitempty"`
}

func (w *WebhookProducerConfig) getBlacklistIPs() []string {
	if w.BlacklistIPs == nil {
		return make([]string, 0)
	}
	return w.BlacklistIPs
}

func (w *WebhookProducerConfig) getPort() string {
	if w.Port == "" {
		return DefaultWebhookPort
	}
	return w.Port
}

func (w *WebhookProducerConfig) getIgnoreExceptions() bool {
	return w.IgnoreExceptions
}

func (w *WebhookProducerConfig) getTimeOut() time.Duration {
	if pTimeout, err := time.ParseDuration(w.TimeOut); err == nil {
		return pTimeout
	}
	return 10 * time.Second
}

func (w *WebhookProducerConfig) getRetry() *RetryConfig {
	if w.Retry == nil {
		return &RetryConfig{
			Count:   5,
			Backoff: "3s",
		}
	}
	return w.Retry
}

// getWorkBackoffDuration 获取运行策略-重试退避时间
func (w *WebhookProducerConfig) getWorkBackoffDuration() time.Duration {
	if w.getRetry() != nil && w.getRetry().Backoff != "" {
		if duration, err := time.ParseDuration(w.getRetry().Backoff); err == nil {
			return duration
		}
	}
	return 3 * time.Second
}

// getWorkMaxRetries 获取运行策略-最大次数
func (w *WebhookProducerConfig) getWorkMaxRetries() int {
	if w.getRetry() != nil && w.getRetry().Count != 0 {
		return w.getRetry().Count
	}
	return 5
}

// WebhookSecurityConfig 安全配置
type WebhookSecurityConfig struct {
	Secret string `json:"secret" mapstructure:"secret" yaml:"secret"`
}

// isValid 验证config配置
func (w *WebhookProducerConfig) isValid() error {
	if w == nil {
		return fmt.Errorf("webhook config cannot be nil")
	}

	if err := w.isValidBase(); err != nil {
		return fmt.Errorf("webhook producer config validation failed: %w", err)
	}

	if err := isValidBlacklistIPs(w.BlacklistIPs); err != nil {
		return fmt.Errorf("webhook producer config validation failed: %w", err)
	}

	if err := isValidSecurity(w.Security); err != nil {
		return fmt.Errorf("webhook producer config validation failed: %w", err)
	}

	return nil
}

// isValidBase 验证基础配置
func (w *WebhookProducerConfig) isValidBase() error {
	// 验证端口配置
	if err := validateWebhookPort(w.getPort()); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	if w.TimeOut != "" {
		if _, err := time.ParseDuration(w.TimeOut); err != nil {
			return fmt.Errorf("invalid timeout format '%s': %w", w.TimeOut, err)
		}
	}

	if w.Retry != nil {
		if w.Retry.Count < 0 {
			return fmt.Errorf("retry count cannot be negative")
		}
		if w.Retry.Backoff != "" {
			if _, err := time.ParseDuration(w.Retry.Backoff); err != nil {
				return fmt.Errorf("invalid retry backoff format '%s': %w", w.Retry.Backoff, err)
			}
		}
	}
	return nil
}

// isValidBlacklistIPs 验证黑名单IP配置，严格要求只能是IP地址
func isValidBlacklistIPs(blacklistIPs []string) error {
	if blacklistIPs == nil || len(blacklistIPs) == 0 {
		return nil
	}
	for i, ip := range blacklistIPs {
		if strings.TrimSpace(ip) == "" {
			return fmt.Errorf("blacklist IP at index %d is empty", i)
		}

		// 严格验证必须是IP地址，不能是域名
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("blacklist IP at index %d must be a valid IP address: %s", i, ip)
		}
	}

	return nil
}

// isValidSecurity 验证安全配置
func isValidSecurity(security *WebhookSecurityConfig) error {
	if security == nil {
		return nil
	}

	if security.Secret != "" {
		if len(security.Secret) < 8 {
			return fmt.Errorf("webhook secret must be at least 8 characters long")
		}
		if len(security.Secret) > 256 {
			return fmt.Errorf("webhook secret cannot exceed 256 characters")
		}
	}

	return nil
}

// validateWebhookPort 验证端口号
func validateWebhookPort(port string) error {
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// WebhookConsumerConfig webhook consumer配置
type WebhookConsumerConfig struct {
	BlacklistConfigAssistance
	BlacklistIPs []string `json:"blacklistIPs,omitempty" mapstructure:"blacklistIPs,omitempty" yaml:"blacklistIPs,omitempty"`

	// 客户端端口配置
	Port     string                 `json:"port" mapstructure:"port" yaml:"port"`
	Retry    *RetryConfig           `json:"retry,omitempty" mapstructure:"retry,omitempty" yaml:"retry,omitempty"`
	Security *WebhookSecurityConfig `json:"security,omitempty" mapstructure:"security,omitempty" yaml:"security,omitempty"`
}

func (w *WebhookConsumerConfig) getBlacklistIPs() []string {
	if w.BlacklistIPs == nil {
		return make([]string, 0)
	}
	return w.BlacklistIPs
}

func (w *WebhookConsumerConfig) getPort() string {
	if w.Port == "" {
		return DefaultWebhookPort
	}
	return w.Port
}

func (w *WebhookConsumerConfig) getRetry() *RetryConfig {
	if w.Retry == nil {
		return &RetryConfig{
			Count:   5,
			Backoff: "3s",
		}
	}
	return w.Retry
}

func (w *WebhookConsumerConfig) getSecurity() *WebhookSecurityConfig {
	return w.Security
}

// isValid 验证config配置
func (w *WebhookConsumerConfig) isValid() error {
	if w == nil {
		return fmt.Errorf("webhook consumer config cannot be nil")
	}

	if err := w.isValidBase(); err != nil {
		return fmt.Errorf("webhook consumer config validation failed: %w", err)
	}

	if err := isValidBlacklistIPs(w.BlacklistIPs); err != nil {
		return fmt.Errorf("webhook consumer config validation failed: %w", err)
	}

	if err := isValidSecurity(w.Security); err != nil {
		return fmt.Errorf("webhook consumer config validation failed: %w", err)
	}

	return nil
}

// isValidBase 验证基础配置
func (w *WebhookConsumerConfig) isValidBase() error {
	// 验证端口配置
	if w.Port == "" {
		return fmt.Errorf("port must be configured")
	}
	if err := validateWebhookPort(w.Port); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	if w.Retry != nil {
		if w.Retry.Count < 0 {
			return fmt.Errorf("retry count cannot be negative")
		}
		if w.Retry.Backoff != "" {
			if _, err := time.ParseDuration(w.Retry.Backoff); err != nil {
				return fmt.Errorf("invalid retry backoff format '%s': %w", w.Retry.Backoff, err)
			}
		}
	}
	return nil
}

// getBackoffDuration 获取重试退避时间
func (w *WebhookConsumerConfig) getBackoffDuration() time.Duration {
	if w.getRetry() != nil && w.getRetry().Backoff != "" {
		if duration, err := time.ParseDuration(w.getRetry().Backoff); err == nil {
			return duration
		}
	}
	return 3 * time.Second
}

// getMaxRetries 获取最大重试次数
func (w *WebhookConsumerConfig) getMaxRetries() int {
	if w.getRetry() != nil && w.getRetry().Count != 0 {
		return w.getRetry().Count
	}
	return 5
}
