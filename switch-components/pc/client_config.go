package pc

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// PersistentConnectionConfig WebSocket长链接配置
type PersistentConnectionConfig struct {
	OnConnect         ConnectHandler                            // 连接成功后的回调函数
	OnDisconnect      func(conn *Connection, err error)         // 连接断开后的回调函数
	OnTrusted         func(conn *Connection)                    // 受信后的回调函数
	OnRegister        func(conn *Connection) *RegisterPayload   // 注册信息函数(使用方需要提供注册信息)
	MessageHandler    func(ctx context.Context, message []byte) // 消息处理器
	Address           string                                    `json:"address" yaml:"address" mapstructure:"address"`                   // 服务端地址
	ClientVersion     string                                    `json:"clientVersion" yaml:"clientVersion" mapstructure:"clientVersion"` // 客户端协议版本
	RequestHeader     http.Header                               // 连接时携带的自定义Header
	ReconnectStrategy *ReconnectStrategy                        `json:"reconnectStrategy" yaml:"reconnectStrategy" mapstructure:"reconnectStrategy"` // 重连策略配置
	Heartbeat         time.Duration                             `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`                         // 心跳发送间隔
	WriteTimeout      time.Duration                             `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`                // 写入超时时间
	ReadTimeout       time.Duration                             `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`                   // 读取超时时间
	DialTimeout       time.Duration                             `json:"dialTimeout" yaml:"dialTimeout" mapstructure:"dialTimeout"`                   // 连接超时时间
}

// ClientProxyInfo 客户端信息
type ClientProxyInfo struct {
	ID            string            `json:"id"`            // 客户端唯一ID
	RemoteAddr    string            `json:"remote_addr"`   // 客户端地址
	PublicIP      []string          `json:"public_ip"`     // 客户端公网IP
	InternalIP    []string          `json:"internal_ip"`   // 客户端局域网IP
	UserAgent     string            `json:"user_agent"`    // 用户代理
	Headers       map[string]string `json:"headers"`       // 连接时的头信息
	ConnectTime   time.Time         `json:"connect_time"`  // 连接时间
	ServiceName   string            `json:"service_name"`  // 客户端服务名
	ServerVersion string            `json:"serverVersion"` // 服务端版本号
	NamespaceTag  string            `json:"namespace_tag"` // 命名空间
	EnvTag        string            `json:"env_tag"`       // 环境
	Endpoint      string            `json:"endpoint"`      // 连接的端点路径
}

// GetVersion 获取服务版本号
func (c *PersistentConnectionConfig) GetVersion() string {
	return c.ClientVersion
}

// GetHeartbeatInterval 获取心跳间隔
func (c *PersistentConnectionConfig) GetHeartbeatInterval() time.Duration {
	if c.Heartbeat > 0 {
		return c.Heartbeat
	}
	return 30 * time.Second
}

// GetWriteTimeout 获取写超时
func (c *PersistentConnectionConfig) GetWriteTimeout() time.Duration {
	if c.WriteTimeout > 0 {
		return c.WriteTimeout
	}
	return 10 * time.Second
}

// GetReadTimeout 获取读超时
func (c *PersistentConnectionConfig) GetReadTimeout() time.Duration {
	if c.ReadTimeout > 0 {
		return c.ReadTimeout
	}
	return 60 * time.Second
}

// ReconnectStrategy 重连策略配置
type ReconnectStrategy struct {
	MaxRetries    int           `json:"maxRetries" yaml:"maxRetries" mapstructure:"maxRetries"`          // 最大重连次数，-1表示无限重连
	InitialDelay  time.Duration `json:"initialDelay" yaml:"initialDelay" mapstructure:"initialDelay"`    // 初始重连延迟
	MaxDelay      time.Duration `json:"maxDelay" yaml:"maxDelay" mapstructure:"maxDelay"`                // 最大重连延迟
	BackoffFactor float64       `json:"backoffFactor" yaml:"backoffFactor" mapstructure:"backoffFactor"` // 退避因子（指数退避）
	ResetInterval time.Duration `json:"resetInterval" yaml:"resetInterval" mapstructure:"resetInterval"` // 重连计数器重置间隔
	EnableJitter  bool          `json:"enableJitter" yaml:"enableJitter" mapstructure:"enableJitter"`    // 启用抖动
}

// ReconnectState 重连状态
type ReconnectState struct {
	Attempts    int           // 当前重连次数
	LastAttempt time.Time     // 上次重连时间
	NextDelay   time.Duration // 下次重连延迟
	mu          sync.RWMutex  // 锁
}

// ClientConnectionData 客户端特有数据
type ClientConnectionData struct {
	// 重连配置
	reconnectStrategy *ReconnectStrategy
	// 重连状态
	reconnectState *ReconnectState

	// 连接状态变更回调
	stateChangeCallback func(change *ConnectionStateChange)
}

// DefaultReconnectStrategy 默认重连策略
func DefaultReconnectStrategy() *ReconnectStrategy {
	return &ReconnectStrategy{
		MaxRetries:    -1,               // 无限重连
		InitialDelay:  1 * time.Second,  // 初始延迟1秒
		MaxDelay:      10 * time.Second, // 最大延迟10秒
		BackoffFactor: 2.0,              // 指数退避因子
		ResetInterval: 5 * time.Minute,  // 5分钟后重置计数器
		EnableJitter:  true,             // 启用抖动
	}
}

// GetReconnectStrategy 获取重连策略，如果未配置则返回默认策略
func (c *PersistentConnectionConfig) GetReconnectStrategy() *ReconnectStrategy {
	if c.ReconnectStrategy != nil {
		return c.ReconnectStrategy
	}
	return DefaultReconnectStrategy()
}

// Reset 重置重连状态
func (rs *ReconnectState) Reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Attempts = 0
	rs.LastAttempt = time.Now()
	rs.NextDelay = 0
}

// ShouldReset 检查是否应该重置重连计数器
func (rs *ReconnectState) ShouldReset(strategy *ReconnectStrategy) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	// 重置策略
	// 如果距离上次尝试超过了重置间隔，且当前有重连次数，则重置
	// 比如，连接稳定运行几小时后连接断开，此时要对状态做重置
	return rs.Attempts > 0 && !rs.LastAttempt.IsZero() && time.Since(rs.LastAttempt) > strategy.ResetInterval
}

// CalculateNextDelay 计算下次重连延迟
func (rs *ReconnectState) CalculateNextDelay(strategy *ReconnectStrategy) time.Duration {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// 指数退避计算
	delay := time.Duration(float64(strategy.InitialDelay) * math.Pow(strategy.BackoffFactor, float64(rs.Attempts)))

	// 限制最大延迟
	if delay > strategy.MaxDelay {
		delay = strategy.MaxDelay
	}

	// 添加抖动 正负0.25
	if strategy.EnableJitter {
		jitter := float64(delay) * 0.25 * (rand.Float64()*2 - 1)
		delay = time.Duration(float64(delay) + jitter)
	}

	rs.NextDelay = delay
	return delay
}

// IncrementAttempts 增加重连尝试次数
func (rs *ReconnectState) IncrementAttempts() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Attempts++
	rs.LastAttempt = time.Now()
}

// GetAttempts 获取重连尝试次数
func (rs *ReconnectState) GetAttempts() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.Attempts
}
