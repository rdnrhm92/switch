package pc

// Server端

// ResponseMessage ws通用响应消息协议格式
type ResponseMessage struct {
	Type          MessageType `json:"type"`           // 消息类型
	RequestID     string      `json:"request_id"`     // 对应的请求ID
	ServerVersion string      `json:"server_version"` // 服务端版本
	Timestamp     int64       `json:"timestamp"`      // 推送时间
	ClientID      string      `json:"client_id"`      // 分配给客户端的ID
	ServiceName   string      `json:"service_name"`   // 目标(sdk)服务名
	NamespaceTag  string      `json:"namespace_tag"`  // 命名空间
	EnvTag        string      `json:"env_tag"`        // 环境
	Data          interface{} `json:"data,omitempty"` // 响应数据
}

// ConnectHelloPayload 连接建立后服务端推送的欢迎消息
type ConnectHelloPayload struct {
	ServerInfo   string            `json:"server_info"`   // 服务端信息
	SupportTypes []MessageType     `json:"support_types"` // 支持的消息类型
	ServerTime   int64             `json:"server_time"`   // 服务端时间
	MaxClients   int               `json:"max_clients"`   // 最大客户端数
	CurrentCount int               `json:"current_count"` // 当前连接数
	Metadata     map[string]string `json:"metadata"`      // 附加元数据
}

// ReceiveAck 接收确认消息
type ReceiveAck struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"` // "received", "processing"
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Client端

// RequestMessage ws通用消息协议格式
type RequestMessage struct {
	Type      MessageType `json:"type"`           // 消息类型
	RequestID string      `json:"request_id"`     // 请求ID
	ClientID  string      `json:"client_id"`      // 客户端ID(say hello时候携带)
	Timestamp int64       `json:"timestamp"`      // 时间戳
	Version   string      `json:"version"`        // 协议版本
	Data      interface{} `json:"data,omitempty"` // 请求数据
}

// RegisterPayload SDK注册时的消息体
type RegisterPayload struct {
	ServiceName string `json:"service_name"` // 服务名
	ClientID    string `json:"client_id"`    // 客户端ID(say hello时候携带)
	SDKVersion  string `json:"sdk_version"`  // SDK版本

	// IP列表
	InternalIPs []string `json:"internal_ips"` // 内网IP列表
	PublicIPs   []string `json:"public_ips"`   // 公网IP列表

	// 环境跟空间
	NamespaceTag string `json:"namespace_tag"` // 命名空间
	EnvTag       string `json:"env_tag"`       // 环境

	// 附加数据
	Metadata map[string]string `json:"metadata,omitempty"`
}
