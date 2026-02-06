package config

import (
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-components/drivers"
	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

type Config struct {
	Server      *ServerConfig        `yaml:"server"`
	MySQL       *MySQLConfig         `yaml:"mysql"`
	LogConfig   *logger.LoggerConfig `yaml:"log_config"`
	JWT         *JWTConfig           `yaml:"jwt"`
	Pc          *pc.ServerConfig     `yaml:"pc"`
	Retry       *RetryConfig         `yaml:"retry"`
	Replacement *ReplacementConfig   `yaml:"replacement"`
	Cache       *Cache               `yaml:"cache"`
}

type JWTConfig struct {
	Secret                   string `yaml:"secret"`
	ExpirationHours          int    `yaml:"expiration_hours"`
	AutoLoginExpirationHours int    `yaml:"auto_login_expiration_hours"`
}

type ServerConfig struct {
	Port    int    `yaml:"port"`
	Version string `yaml:"version"`
	Name    string `yaml:"name"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`

	Charset   string `yaml:"charset"`
	ParseTime bool   `yaml:"parseTime"`
	Loc       string `yaml:"loc"`
	Logger    struct {
		SlowThreshold             string `yaml:"slow_threshold"`   // 慢SQL阈值
		LogLevel                  string `yaml:"log_level"`        // 日志级别
		IgnoreRecordNotFoundError bool   `yaml:"ignore_not_found"` // 是否忽略记录未找到错误
		ParameterizedQueries      bool   `yaml:"parameterized"`    // 是否在SQL日志中包含参数
		Colorful                  bool   `yaml:"colorful"`         // 是否启用彩色日志
	} `yaml:"logging"`
	Pool struct {
		MinConnections int   `yaml:"min_connections"`
		MaxConnections int   `yaml:"max_connections"`
		ConnectTimeout int64 `yaml:"connect_timeout"`
		MaxIdleTime    int64 `yaml:"max_idle_time"`
	} `yaml:"pool"`
}

func (c *MySQLConfig) ToDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.Charset,
		c.ParseTime,
		c.Loc,
	)
}

// RetryConfig WebSocket消息重试配置
type RetryConfig struct {
	// 普通消息重试配置
	Default *RetryParams `yaml:"default"`
	// 配置变更消息重试配置
	ConfigChange *RetryParams `yaml:"config_change"`
	// 全量配置消息重试配置
	FullConfig *RetryParams `yaml:"full_config"`
	// 开关数据消息重试配置
	SwitchData *RetryParams `yaml:"switch_data"`
	// IP连通性检查配置
	IPConnectivity *drivers.IPConnectivityConfig `yaml:"ip_connectivity"`
}

// ReplacementConfig 驱动替换的配置
type ReplacementConfig struct {
	// kafka producer在驱动更新的时候的新驱动的校验超时时间
	KafkaProducerReplaceDriverValidationTimeout string `yaml:"kafka_producer_replace_driver_validation_timeout"`
	// kafka producer在驱动更新的时候的新驱动稳定期
	KafkaProducerReplaceDriverStabilityPeriod string `yaml:"kafka_producer_replace_driver_stability_period"`
	// kafka producer在驱动更新的时候是否校验brokers的位置
	KafkaProducerVerifyBrokers bool `yaml:"kafka_producer_verify_brokers"`

	// webhook producer在驱动更新的时候的新驱动的校验超时时间
	WebhookProducerReplaceDriverValidationTimeout string `yaml:"webhook_producer_replace_driver_validation_timeout"`
	// webhook producer在驱动更新的时候的新驱动稳定期
	WebhookProducerReplaceDriverStabilityPeriod string `yaml:"webhook_producer_replace_driver_stability_period"`

	// polling producer在驱动更新的时候的新驱动的校验超时时间
	PollingProducerReplaceDriverValidationTimeout string `yaml:"polling_producer_replace_driver_validation_timeout"`
	// polling producer在驱动更新的时候的新驱动稳定期
	PollingProducerReplaceDriverStabilityPeriod string `yaml:"polling_producer_replace_driver_stability_period"`
}

// GetKafkaProducerValidationTimeout 获取kafka producer在驱动更新的时候的新驱动的校验超时时间
func (r *ReplacementConfig) GetKafkaProducerValidationTimeout() (time.Duration, error) {
	duration, err := time.ParseDuration(r.KafkaProducerReplaceDriverValidationTimeout)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// GetKafkaProducerStabilityPeriod 获取kafka producer在驱动更新的时候的新驱动稳定期
func (r *ReplacementConfig) GetKafkaProducerStabilityPeriod() (time.Duration, error) {
	duration, err := time.ParseDuration(r.KafkaProducerReplaceDriverStabilityPeriod)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// GetKafkaProducerVerifyBrokers 获取kafka producer在驱动更新的时候是否校验brokers的位置
func (r *ReplacementConfig) GetKafkaProducerVerifyBrokers() bool {
	return r.KafkaProducerVerifyBrokers
}

// GetWebhookProducerValidationTimeout 获取webhook producer在驱动更新的时候的新驱动的校验超时时间
func (r *ReplacementConfig) GetWebhookProducerValidationTimeout() (time.Duration, error) {
	duration, err := time.ParseDuration(r.WebhookProducerReplaceDriverValidationTimeout)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// GetWebhookProducerStabilityPeriod 获取webhook producer在驱动更新的时候的新驱动稳定期
func (r *ReplacementConfig) GetWebhookProducerStabilityPeriod() (time.Duration, error) {
	duration, err := time.ParseDuration(r.WebhookProducerReplaceDriverStabilityPeriod)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// GetPollingProducerValidationTimeout 获取polling producer在驱动更新的时候的新驱动的校验超时时间
func (r *ReplacementConfig) GetPollingProducerValidationTimeout() (time.Duration, error) {
	duration, err := time.ParseDuration(r.PollingProducerReplaceDriverValidationTimeout)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// GetPollingProducerStabilityPeriod 获取polling producer在驱动更新的时候的新驱动稳定期
func (r *ReplacementConfig) GetPollingProducerStabilityPeriod() (time.Duration, error) {
	duration, err := time.ParseDuration(r.PollingProducerReplaceDriverStabilityPeriod)
	if err != nil {
		return time.Duration(0), err
	}
	return duration, nil
}

// RetryParams 重试参数
type RetryParams struct {
	Timeout    time.Duration `yaml:"timeout"`     // 超时时间
	MaxRetries int           `yaml:"max_retries"` // 最大重试次数
	RetryDelay time.Duration `yaml:"retry_delay"` // 重试间隔
}

// DefaultRetryConfig 返回默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		Default: &RetryParams{
			Timeout:    10 * time.Second,
			MaxRetries: 3,
			RetryDelay: 2 * time.Second,
		},
		ConfigChange: &RetryParams{
			Timeout:    30 * time.Second,
			MaxRetries: 5,
			RetryDelay: 3 * time.Second,
		},
		FullConfig: &RetryParams{
			Timeout:    60 * time.Second,
			MaxRetries: 3,
			RetryDelay: 5 * time.Second,
		},
		SwitchData: &RetryParams{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			RetryDelay: 2 * time.Second,
		},
		IPConnectivity: &drivers.IPConnectivityConfig{
			CheckInterval:  5 * time.Minute, // 存量IP每5分钟检查一次
			CheckTimeout:   5 * time.Second, // 单个IP检查超时5秒
			NewIPQueueSize: 100,             // 新IP队列大小100
		},
	}
}

// GetRetryConfig 获取重试配置
func GetRetryConfig() *RetryConfig {
	if GlobalConfig.Retry != nil {
		return GlobalConfig.Retry
	}
	return DefaultRetryConfig()
}

// Cache 缓存配置
type Cache struct {
	// CacheTime 缓存时间
	CacheTime string `yaml:"cache_time"`
}

// GetCacheTime 获取缓存过期时间
func (c *Cache) GetCacheTime() (time.Duration, error) {
	if c.CacheTime == "" {
		return 3 * time.Minute, nil
	}
	cacheTime, err := time.ParseDuration(c.CacheTime)
	if err != nil {
		return time.Duration(0), err
	}
	return cacheTime, nil
}
