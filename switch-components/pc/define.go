package pc

import (
	"context"
	"encoding/json"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/driver"
)

// MessageType 定义了ws通讯的消息类型
type MessageType string

const (
	// SwitchFull 启动后的全量开关推送(运行中的增量开关推送不走ws)
	SwitchFull MessageType = "SWITCH_FULL"

	// DriverConfigFull 启动后的驱动配置全量推送
	DriverConfigFull MessageType = "DRIVER_CONFIG_FULL"
	// DriverConfigChange 运行中的驱动配置增量推送
	DriverConfigChange MessageType = "DRIVER_CONFIG_CHANGE"

	// ChangeTypeConnectHello say hello连接建立后服务端推送的消息
	ChangeTypeConnectHello MessageType = "CONNECT_HELLO"

	// RegisterSignal 注册消息
	RegisterSignal MessageType = "REGISTER_SIGNAL"
)

// ConfigChangeType 定义了配置变更的类型
type ConfigChangeType string

const (
	// KafkaConsumerConfigChange kafka的配置变更
	KafkaConsumerConfigChange ConfigChangeType = "kafka_consumer_config_change"
	// WebhookConsumerConfigChange webhook的配置变更
	WebhookConsumerConfigChange ConfigChangeType = "webhook_consumer_config_change"
	// PollingConsumerConfigChange polling的配置变更
	PollingConsumerConfigChange ConfigChangeType = "polling_consumer_config_change"
)

// DriverConfigPayload 驱动配置数据
type DriverConfigPayload struct {
	Type   driver.DriverType `json:"type"`   // 驱动类型
	Config json.RawMessage   `json:"config"` // 具体配置
}

// IncrementConfigChangeType 增量配置变更类型
type IncrementConfigChangeType string

const (
	// UPDATE 重启驱动
	UPDATE IncrementConfigChangeType = "UPDATE"
	// ADD 新增驱动
	ADD IncrementConfigChangeType = "ADD"
	// DELETE 停止驱动
	DELETE IncrementConfigChangeType = "DELETE"
)

// MessageSecurity 消息安全级别
type MessageSecurity int

const (
	// MessageTypePublic 公开消息，不需要受信
	MessageTypePublic MessageSecurity = iota
	// MessageTypeTrusted 受信消息，需要受信客户端
	MessageTypeTrusted
	// MessageTypeSpecial 特殊消息，需要特殊处理
	MessageTypeSpecial
)

// GetMessageSecurity 获取消息安全级别
func GetMessageSecurity(msgType MessageType) MessageSecurity {
	switch msgType {
	case RegisterSignal, ChangeTypeConnectHello:
		return MessageTypeSpecial // 注册消息特殊处理 不需要走一阶段确认
	case DriverConfigFull, DriverConfigChange, SwitchFull:
		return MessageTypeTrusted // 配置和开关消息需要受信 需要一阶段确认
	default:
		return MessageTypePublic // 其他消息为公开消息 不需要一阶段确认
	}
}

const (
	// ServerStart 服务端启动
	ServerStart = "WS_SERVER_START"
)

// ConnectionRole 连接角色类型
type ConnectionRole int

const (
	RoleClient ConnectionRole = iota // 客户端角色
	RoleProxy                        // 服务端的客户端代理角色
)

// ConnectionEvent 连接事件类型
type ConnectionEvent int

const (
	ConnectionEventConnected    ConnectionEvent = iota // 连接建立
	ConnectionEventDisconnected                        // 连接断开
	ConnectionEventReconnecting                        // 正在重连
)

// DisconnectReason 断开原因
type DisconnectReason int

const (
	DisconnectReasonUnknown          DisconnectReason = iota
	DisconnectReasonNetworkError                      // 网络错误 -> 应该重连
	DisconnectReasonReadTimeout                       // 读超时 -> 应该重连
	DisconnectReasonWriteError                        // 写错误 -> 应该重连
	DisconnectReasonHeartbeatTimeout                  // 心跳超时 -> 应该重连
	DisconnectReasonHandshakeFailed                   // 握手失败 -> 应该重连
	DisconnectReasonAuthFailed                        // 认证失败 -> 不应该重连
	DisconnectReasonExternalClose                     // 外部主动关闭 -> 不应该重连
	DisconnectReasonServerKick                        // 服务端踢出 -> 不应该重连
)

// ShouldReconnect 判断是否应该重连
func (r DisconnectReason) ShouldReconnect() bool {
	switch r {
	case DisconnectReasonNetworkError,
		DisconnectReasonReadTimeout,
		DisconnectReasonWriteError,
		DisconnectReasonHeartbeatTimeout,
		DisconnectReasonHandshakeFailed:
		return true
	default:
		return false
	}
}

// String 返回断开原因的字符串描述
func (r DisconnectReason) String() string {
	switch r {
	case DisconnectReasonNetworkError:
		return "NetworkError"
	case DisconnectReasonReadTimeout:
		return "ReadTimeout"
	case DisconnectReasonWriteError:
		return "WriteError"
	case DisconnectReasonHeartbeatTimeout:
		return "HeartbeatTimeout"
	case DisconnectReasonHandshakeFailed:
		return "HandshakeFailed"
	case DisconnectReasonAuthFailed:
		return "AuthFailed"
	case DisconnectReasonExternalClose:
		return "ExternalClose"
	case DisconnectReasonServerKick:
		return "ServerKick"
	default:
		return "Unknown"
	}
}

// ConnectionStateChange 连接状态变更事件
type ConnectionStateChange struct {
	Event            ConnectionEvent  // 事件类型
	Conn             *Connection      // 连接对象
	DisconnectReason DisconnectReason // 断开原因
	Error            error            // 错误信息
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, message []byte)

// PendingRequest 待处理请求
type PendingRequest struct {
	RequestID    string
	MessageType  MessageType
	SendTime     time.Time
	Timeout      time.Duration
	MaxRetries   int           // 最大重试次数
	CurrentRetry int           // 当前重试次数
	RetryDelay   time.Duration // 重试间隔
	OriginalMsg  interface{}   // 原始消息，用于重发
}
