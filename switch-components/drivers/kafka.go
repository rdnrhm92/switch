package drivers

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducerConfig producer config
type KafkaProducerConfig struct {
	RequiredAcks    string `yaml:"requiredAcks,omitempty" json:"requiredAcks,omitempty" mapstructure:"requiredAcks,omitempty"`
	Timeout         string `yaml:"timeout,omitempty" json:"timeout,omitempty" mapstructure:"timeout,omitempty"`
	BatchTimeout    string `yaml:"batchTimeout,omitempty" json:"batchTimeout,omitempty" mapstructure:"batchTimeout,omitempty"`
	BatchBytes      int64  `yaml:"batchBytes,omitempty" json:"batchBytes,omitempty" mapstructure:"batchBytes,omitempty"`
	BatchSize       int    `yaml:"batchSize,omitempty" json:"batchSize,omitempty" mapstructure:"batchSize,omitempty"`
	Retries         int    `yaml:"retries,omitempty" json:"retries,omitempty" mapstructure:"retries,omitempty"`
	RetryBackoffMin string `yaml:"retryBackoffMin,omitempty" json:"retryBackoffMin,omitempty" mapstructure:"retryBackoffMin,omitempty"`
	RetryBackoffMax string `yaml:"retryBackoffMax,omitempty" json:"retryBackoffMax,omitempty" mapstructure:"retryBackoffMax,omitempty"`
	Compression     string `yaml:"compression,omitempty" json:"compression,omitempty" mapstructure:"compression,omitempty"`

	// 连接和验证超时配置(不需要做same config的比较替换)
	ConnectTimeout  string `yaml:"connectTimeout,omitempty" json:"connectTimeout,omitempty" mapstructure:"connectTimeout,omitempty"`
	ValidateTimeout string `yaml:"validateTimeout,omitempty" json:"validateTimeout,omitempty" mapstructure:"validateTimeout,omitempty"`

	Brokers  []string        `yaml:"brokers" json:"brokers" mapstructure:"brokers"`
	Topic    string          `yaml:"topic" json:"topic" mapstructure:"topic"`
	Security *SecurityConfig `yaml:"security,omitempty" json:"security,omitempty" mapstructure:"security,omitempty"`
}

func (k *KafkaProducerConfig) parseRequiredAcks() kafka.RequiredAcks {
	switch k.RequiredAcks {
	case "all":
		return kafka.RequireAll
	case "one":
		return kafka.RequireOne
	case "none":
		return kafka.RequireNone
	default:
		return kafka.RequireAll
	}
}

// getTimeout 获取等待broker回应超时时间
func (k *KafkaProducerConfig) getTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.Timeout); err == nil {
		return timeout
	}
	return 5 * time.Second
}

// getBatchTimeout 获取发送前等待时间
func (k *KafkaProducerConfig) getBatchTimeout() time.Duration {
	if batchTimeout, err := time.ParseDuration(k.BatchTimeout); err == nil {
		return batchTimeout
	}
	return 1 * time.Second
}

// getBatchBytes 获取最大bytes
func (k *KafkaProducerConfig) getBatchBytes() int64 {
	if k.BatchBytes != 0 {
		return k.BatchBytes
	}
	return 1048576
}

// getBatchSize 获取最大size
func (k *KafkaProducerConfig) getBatchSize() int {
	if k.BatchSize != 0 {
		return k.BatchSize
	}
	return 50
}

// getRetries 获取最大重试次数
func (k *KafkaProducerConfig) getRetries() int {
	if k.Retries != 0 {
		return k.Retries
	}
	return 3
}

// getRetryBackoffMin 获取写超时时间(最小)
func (k *KafkaProducerConfig) getRetryBackoffMin() time.Duration {
	if timeout, err := time.ParseDuration(k.RetryBackoffMin); err == nil {
		return timeout
	}
	return 100 * time.Millisecond
}

// getRetryBackoffMax 获取写超时时间(最大)
func (k *KafkaProducerConfig) getRetryBackoffMax() time.Duration {
	if timeout, err := time.ParseDuration(k.RetryBackoffMax); err == nil {
		return timeout
	}
	return 1 * time.Second
}

func (k *KafkaProducerConfig) parseCompression() kafka.Compression {
	switch k.Compression {
	case "gzip":
		return 1
	case "snappy":
		return 2
	case "lz4":
		return 3
	case "zstd":
		return 4
	default:
		return 0
	}
}

// getConnectTimeout 获取连接超时时间
func (k *KafkaProducerConfig) getConnectTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.ConnectTimeout); err == nil {
		return timeout
	}
	return 10 * time.Second
}

// getValidateTimeout 获取验证超时时间
func (k *KafkaProducerConfig) getValidateTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.ValidateTimeout); err == nil {
		return timeout
	}
	return 10 * time.Second
}

// getBrokers 地址获取
func (k *KafkaProducerConfig) getBrokers() []string {
	return k.Brokers
}

// getTopic 主题获取
func (k *KafkaProducerConfig) getTopic() string {
	return k.Topic
}

// getSecurity 获取安全配置
func (k *KafkaProducerConfig) getSecurity() *SecurityConfig {
	return k.Security
}

// isValid 验证配置
func (k *KafkaProducerConfig) isValid() error {
	if k == nil {
		return fmt.Errorf("kafka producer config cannot be nil")
	}

	// 验证基础配置
	if err := k.validateBase(); err != nil {
		return fmt.Errorf("kafka producer config validation failed: %w", err)
	}

	// 验证安全配置
	if err := k.validateSecurity(); err != nil {
		return fmt.Errorf("kafka producer config validation failed: %w", err)
	}

	return nil
}

// validateBase 验证基础配置
func (k *KafkaProducerConfig) validateBase() error {
	// 验证Brokers配置
	if k.Brokers == nil || len(k.Brokers) == 0 {
		return fmt.Errorf("brokers must be configured")
	}

	for i, broker := range k.Brokers {
		if broker == "" {
			return fmt.Errorf("broker[%d] cannot be empty", i)
		}
		if err := validateBrokerAddress(broker); err != nil {
			return fmt.Errorf("invalid broker[%d] address '%s': %w", i, broker, err)
		}
	}

	// 验证Topic配置
	if k.Topic == "" {
		return fmt.Errorf("topic must be configured")
	}

	// 验证RequiredAcks
	if k.RequiredAcks != "" {
		validAcks := []string{"all", "one", "none"}
		if !contains(validAcks, k.RequiredAcks) {
			return fmt.Errorf("invalid requiredAcks '%s', must be one of: %v", k.RequiredAcks, validAcks)
		}
	}

	// 验证Compression
	if k.Compression != "" {
		validCompressions := []string{"gzip", "snappy", "lz4", "zstd"}
		if !contains(validCompressions, k.Compression) {
			return fmt.Errorf("invalid Compression '%s', must be one of: %v", k.Compression, validCompressions)
		}
	}

	// 验证超时配置
	if k.Timeout != "" {
		if _, err := time.ParseDuration(k.Timeout); err != nil {
			return fmt.Errorf("invalid timeout format '%s': %w", k.Timeout, err)
		}
	}

	if k.BatchTimeout != "" {
		if _, err := time.ParseDuration(k.BatchTimeout); err != nil {
			return fmt.Errorf("invalid batchTimeout format '%s': %w", k.BatchTimeout, err)
		}
	}

	if k.RetryBackoffMin != "" {
		if _, err := time.ParseDuration(k.RetryBackoffMin); err != nil {
			return fmt.Errorf("invalid retryBackoffMin format '%s': %w", k.RetryBackoffMin, err)
		}
	}

	if k.RetryBackoffMax != "" {
		if _, err := time.ParseDuration(k.RetryBackoffMax); err != nil {
			return fmt.Errorf("invalid retryBackoffMax format '%s': %w", k.RetryBackoffMax, err)
		}
	}

	if k.ConnectTimeout != "" {
		if _, err := time.ParseDuration(k.ConnectTimeout); err != nil {
			return fmt.Errorf("invalid connectTimeout format '%s': %w", k.ConnectTimeout, err)
		}
	}

	if k.ValidateTimeout != "" {
		if _, err := time.ParseDuration(k.ValidateTimeout); err != nil {
			return fmt.Errorf("invalid validateTimeout format '%s': %w", k.ValidateTimeout, err)
		}
	}

	return nil
}

// validateSecurity 验证安全配置
func (k *KafkaProducerConfig) validateSecurity() error {
	if k.Security == nil {
		return nil
	}

	return validateSecurityConfig(k.Security)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// KafkaConsumerConfig consumer config
type KafkaConsumerConfig struct {
	GroupID            string `yaml:"groupId" json:"groupId" mapstructure:"groupId"`
	AutoOffsetReset    string `yaml:"autoOffsetReset,omitempty" json:"autoOffsetReset,omitempty" mapstructure:"autoOffsetReset,omitempty"`
	EnableAutoCommit   bool   `yaml:"enableAutoCommit,omitempty" json:"enableAutoCommit,omitempty" mapstructure:"enableAutoCommit,omitempty"`
	AutoCommitInterval string `yaml:"autoCommitInterval,omitempty" json:"autoCommitInterval,omitempty" mapstructure:"autoCommitInterval,omitempty"`

	// 连接和验证超时配置(这两项配置变更不需要做same config做新老驱动的替换)
	ConnectTimeout  string `yaml:"connectTimeout,omitempty" json:"connectTimeout,omitempty" mapstructure:"connectTimeout,omitempty"`
	ValidateTimeout string `yaml:"validateTimeout,omitempty" json:"validateTimeout,omitempty" mapstructure:"validateTimeout,omitempty"`

	// 读消息和提交偏移量超时配置
	ReadTimeout   string `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" mapstructure:"readTimeout,omitempty"`       // 读消息超时
	CommitTimeout string `yaml:"commitTimeout,omitempty" json:"commitTimeout,omitempty" mapstructure:"commitTimeout,omitempty"` // 提交偏移量超时

	Brokers  []string        `yaml:"brokers" json:"brokers" mapstructure:"brokers"`
	Topic    string          `yaml:"topic" json:"topic" mapstructure:"topic"`
	Security *SecurityConfig `yaml:"security,omitempty" json:"security,omitempty" mapstructure:"security,omitempty"`
	Retry    *RetryConfig    `yaml:"retry,omitempty" json:"retry,omitempty" mapstructure:"retry,omitempty"`
}

// getGroupId 获取组ID
func (k *KafkaConsumerConfig) getGroupId() string {
	if k.GroupID != "" {
		return k.GroupID
	}
	return fmt.Sprintf("switch-consumer-%s", k.Topic)
}

// getAutoOffsetReset 从哪里开始消费
func (k *KafkaConsumerConfig) getAutoOffsetReset() string {
	if k.AutoOffsetReset == "" {
		return "latest"
	}
	return k.AutoOffsetReset
}

// getEnableAutoCommit 是否自动提交
func (k *KafkaConsumerConfig) getEnableAutoCommit() bool {
	return k.EnableAutoCommit
}

// getAutoCommitInterval 获取自动提交间隔
func (k *KafkaConsumerConfig) getAutoCommitInterval() time.Duration {
	if interval, err := time.ParseDuration(k.AutoCommitInterval); err == nil {
		return interval
	}
	return 1 * time.Second
}

// getConnectTimeout 获取连接超时时间
func (k *KafkaConsumerConfig) getConnectTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.ConnectTimeout); err == nil {
		return timeout
	}
	return 10 * time.Second
}

// getValidateTimeout 获取验证超时时间
func (k *KafkaConsumerConfig) getValidateTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.ValidateTimeout); err == nil {
		return timeout
	}
	return 10 * time.Second
}

// getReadTimeout 获取读消息超时时间
func (k *KafkaConsumerConfig) getReadTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.ReadTimeout); err == nil {
		return timeout
	}
	return 30 * time.Second
}

// getCommitTimeout 获取提交偏移量超时时间
func (k *KafkaConsumerConfig) getCommitTimeout() time.Duration {
	if timeout, err := time.ParseDuration(k.CommitTimeout); err == nil {
		return timeout
	}
	return 5 * time.Second
}

// getBrokers 地址获取
func (k *KafkaConsumerConfig) getBrokers() []string {
	return k.Brokers
}

// getTopic 主题获取
func (k *KafkaConsumerConfig) getTopic() string {
	return k.Topic
}

// getSecurity 安全配置获取
func (k *KafkaConsumerConfig) getSecurity() *SecurityConfig {
	return k.Security
}

// getRetry 重试配置获取
func (k *KafkaConsumerConfig) getRetry() *RetryConfig {
	if k.Retry == nil {
		return &RetryConfig{
			Count:   5,
			Backoff: "3s",
		}
	}
	return k.Retry
}

// getMaxRetries 获取验证超时时间
func (k *KafkaConsumerConfig) getMaxRetries() int {
	if k.getRetry() != nil && k.getRetry().Count != 0 {
		return k.getRetry().Count
	}
	return 5
}

// getBackoffDuration 获取重试退避时间
func (k *KafkaConsumerConfig) getBackoffDuration() time.Duration {
	if k.getRetry() != nil && k.getRetry().Backoff != "" {
		if duration, err := time.ParseDuration(k.getRetry().Backoff); err == nil {
			return duration
		}
	}
	return 3 * time.Second
}

// isValid 验证配置
func (k *KafkaConsumerConfig) isValid() error {
	if k == nil {
		return fmt.Errorf("kafka consumer config cannot be nil")
	}

	// 验证基础配置
	if err := k.validateBase(); err != nil {
		return fmt.Errorf("kafka consumer config validation failed: %w", err)
	}

	// 验证安全配置
	if err := k.validateSecurity(); err != nil {
		return fmt.Errorf("kafka consumer config validation failed: %w", err)
	}

	return nil
}

// validateBase 验证基础配置
func (k *KafkaConsumerConfig) validateBase() error {
	// 验证Brokers配置
	if k.Brokers == nil || len(k.Brokers) == 0 {
		return fmt.Errorf("brokers must be configured")
	}

	for i, broker := range k.Brokers {
		if broker == "" {
			return fmt.Errorf("broker[%d] cannot be empty", i)
		}
		if err := validateBrokerAddress(broker); err != nil {
			return fmt.Errorf("invalid broker[%d] address '%s': %w", i, broker, err)
		}
	}

	// 验证Topic配置
	if k.Topic == "" {
		return fmt.Errorf("topic must be configured")
	}

	// 验证AutoOffsetReset
	if k.AutoOffsetReset != "" {
		validOffsets := []string{"earliest", "latest"}
		if !contains(validOffsets, k.AutoOffsetReset) {
			return fmt.Errorf("invalid autoOffsetReset '%s', must be one of: %v", k.AutoOffsetReset, validOffsets)
		}
	}

	// 验证AutoCommitInterval
	if k.AutoCommitInterval != "" {
		if _, err := time.ParseDuration(k.AutoCommitInterval); err != nil {
			return fmt.Errorf("invalid autoCommitInterval format '%s': %w", k.AutoCommitInterval, err)
		}
	}

	// 验证超时配置
	if k.ConnectTimeout != "" {
		if _, err := time.ParseDuration(k.ConnectTimeout); err != nil {
			return fmt.Errorf("invalid connectTimeout format '%s': %w", k.ConnectTimeout, err)
		}
	}

	if k.ValidateTimeout != "" {
		if _, err := time.ParseDuration(k.ValidateTimeout); err != nil {
			return fmt.Errorf("invalid validateTimeout format '%s': %w", k.ValidateTimeout, err)
		}
	}

	if k.ReadTimeout != "" {
		if _, err := time.ParseDuration(k.ReadTimeout); err != nil {
			return fmt.Errorf("invalid readTimeout format '%s': %w", k.ReadTimeout, err)
		}
	}

	if k.CommitTimeout != "" {
		if _, err := time.ParseDuration(k.CommitTimeout); err != nil {
			return fmt.Errorf("invalid commitTimeout format '%s': %w", k.CommitTimeout, err)
		}
	}

	if k.getRetry() != nil {
		if k.getRetry().Count < 0 {
			return fmt.Errorf("retry count cannot be negative")
		}
		if k.getRetry().Backoff != "" {
			if _, err := time.ParseDuration(k.getRetry().Backoff); err != nil {
				return fmt.Errorf("invalid retry backoff format '%s': %w", k.getRetry().Backoff, err)
			}
		}
	}

	return nil
}

// validateSecurity 验证安全配置
func (k *KafkaConsumerConfig) validateSecurity() error {
	if k.getSecurity() == nil {
		return nil
	}

	return validateSecurityConfig(k.getSecurity())
}

// validateBrokerAddress 验证Broker地址格式
func validateBrokerAddress(broker string) error {
	if broker == "" {
		return fmt.Errorf("broker address cannot be empty")
	}

	// 检查是否包含端口
	if !strings.Contains(broker, ":") {
		return fmt.Errorf("broker address must include port (e.g., 'localhost:9092')")
	}

	// 分离主机和端口
	host, port, err := net.SplitHostPort(broker)
	if err != nil {
		return fmt.Errorf("invalid broker address format: %w", err)
	}

	if host == "" {
		return fmt.Errorf("broker host cannot be empty")
	}

	if port == "" {
		return fmt.Errorf("broker port cannot be empty")
	}

	return nil
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	SASL *SASLConfig `yaml:"sasl,omitempty" json:"sasl,omitempty" mapstructure:"sasl,omitempty"`
	TLS  *TLSConfig  `yaml:"tls,omitempty" json:"tls,omitempty" mapstructure:"tls,omitempty"`
}

// SASLConfig sasl配置
type SASLConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Mechanism string `yaml:"mechanism,omitempty" json:"mechanism,omitempty" mapstructure:"mechanism,omitempty"`
	Username  string `yaml:"username,omitempty" json:"username,omitempty" mapstructure:"username,omitempty"`
	Password  string `yaml:"password,omitempty" json:"password,omitempty" mapstructure:"password,omitempty"`
}

// TLSConfig tls配置
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	CaFile             string `yaml:"caFile,omitempty" json:"caFile,omitempty" mapstructure:"caFile,omitempty"`
	CertFile           string `yaml:"certFile,omitempty" json:"certFile,omitempty" mapstructure:"certFile,omitempty"`
	KeyFile            string `yaml:"keyFile,omitempty" json:"keyFile,omitempty" mapstructure:"keyFile,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify" json:"insecureSkipVerify" mapstructure:"insecureSkipVerify"`
}

// validateSecurityConfig 验证安全配置
func validateSecurityConfig(security *SecurityConfig) error {
	if security == nil {
		return nil
	}

	// 验证SASL配置
	if security.SASL != nil && security.SASL.Enabled {
		if err := validateSASLConfig(security.SASL); err != nil {
			return fmt.Errorf("invalid SASL config: %w", err)
		}
	}

	// 验证TLS配置
	if security.TLS != nil && security.TLS.Enabled {
		if err := validateTLSConfig(security.TLS); err != nil {
			return fmt.Errorf("invalid TLS config: %w", err)
		}
	}

	return nil
}

// validateSASLConfig 验证SASL配置
func validateSASLConfig(sasl *SASLConfig) error {
	if sasl == nil {
		return nil
	}

	if !sasl.Enabled {
		return nil
	}

	// kafka支持的验证机制
	if sasl.Mechanism != "" {
		validMechanisms := []string{
			"PLAIN",
			"SCRAM-SHA-256",
			"SCRAM-SHA-512",
		}
		if !contains(validMechanisms, sasl.Mechanism) {
			return fmt.Errorf("invalid SASL mechanism '%s', must be one of: %v", sasl.Mechanism, validMechanisms)
		}
	}

	// 验证用户名和密码允许为空 不做强制的要求了

	return nil
}

// validateTLSConfig 验证TLS配置
func validateTLSConfig(tls *TLSConfig) error {
	if tls == nil {
		return nil
	}

	if !tls.Enabled {
		return nil
	}

	if tls.CertFile != "" && tls.KeyFile == "" {
		return fmt.Errorf("TLS key file is required when cert file is provided")
	}

	if tls.KeyFile != "" && tls.CertFile == "" {
		return fmt.Errorf("TLS cert file is required when key file is provided")
	}

	return nil
}

// contains 检查字符串切片是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
