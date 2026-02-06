package pc

import (
	"net/http"
	"time"
)

// ServerConfig WebSocket服务端配置
type ServerConfig struct {
	Address      string        `json:"address" yaml:"address" mapstructure:"address"`                // 监听地址
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`    // 读超时时间
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"` // 写超时时间

	CheckOrigin     func(r *http.Request) bool
	ReadBufferSize  int `json:"readBufferSize" yaml:"readBufferSize" mapstructure:"readBufferSize"`    // 读缓冲区大小
	WriteBufferSize int `json:"writeBufferSize" yaml:"writeBufferSize" mapstructure:"writeBufferSize"` // 写缓冲区大小

	HeartbeatInterval time.Duration `json:"heartbeatInterval" yaml:"heartbeatInterval" mapstructure:"heartbeatInterval"` // 心跳检查间隔
	ClientTimeout     time.Duration `json:"clientTimeout" yaml:"clientTimeout" mapstructure:"clientTimeout"`             // 客户端超时时间
	MaxConnections    int           `json:"maxConnections" yaml:"maxConnections" mapstructure:"maxConnections"`          // 最大连接数

	OnClientTrusted func(clientInfo *ClientProxyInfo) // 受信后的回调通道
	ServerVersion   string                            `json:"serverVersion"  yaml:"serverVersion" mapstructure:"serverVersion"` // 服务端版本号

	OnConnect      ConnectHandler                         // 连接成功后的回调函数
	OnDisconnect   func(conn *Connection, err error)      // 连接断开后的回调函数
	MessageHandler func(conn *Connection, message []byte) // 消息处理器
}

// DefaultServerConfig 返回默认的服务器配置
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:           ":8080",
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		HeartbeatInterval: 30 * time.Second,
		ClientTimeout:     90 * time.Second,
		MaxConnections:    1000,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该更严格
		},
	}
}

// GetVersion 获取服务版本号
func (c *ServerConfig) GetVersion() string {
	return c.ServerVersion
}

// GetHeartbeatInterval 获取心跳间隔
func (c *ServerConfig) GetHeartbeatInterval() time.Duration {
	if c.HeartbeatInterval > 0 {
		return c.HeartbeatInterval
	}
	return 30 * time.Second
}

// GetWriteTimeout 获取写超时
func (c *ServerConfig) GetWriteTimeout() time.Duration {
	if c.WriteTimeout > 0 {
		return c.WriteTimeout
	}
	return 10 * time.Second
}

// GetReadTimeout 获取读超时
func (c *ServerConfig) GetReadTimeout() time.Duration {
	if c.ReadTimeout > 0 {
		return c.ReadTimeout
	}
	return 60 * time.Second
}

// ServerInfo 服务端信息
type ServerInfo struct {
	Version      string            // 服务端版本
	Name         string            // 服务端名称
	Description  string            // 服务端描述
	Capabilities []string          // 服务端能力列表
	Config       map[string]string // 服务端配置信息
	StartTime    int64             // 服务端启动时间
}

// ProxyConnectionData 服务端代理特有数据
type ProxyConnectionData struct {
	// 客户端信息
	Info *ClientProxyInfo

	// 受信状态(注册后受信)
	isTrusted bool

	// 连接状态变更回调
	stateChangeCallback func(change *ConnectionStateChange)
}
