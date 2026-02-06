package rpc

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/tap"
)

// ServerConfig gRPC服务器配置 以下配置粘贴自/google.golang.org/rpc@v1.73.0/server.go:153
// 这是grpc服务端的options配置
type ServerConfig struct {
	// 基本配置
	Address string `json:"address" yaml:"address"` // 服务器地址

	//证书密钥都不为空则开启tls
	CertFile string `json:"cert_file" yaml:"cert_file"` // 证书文件路径
	KeyFile  string `json:"key_file" yaml:"key_file"`   // 密钥文件路径

	// 是否健康检查
	EnableHealthCheck bool `json:"enable_health_check" yaml:"enable_health_check"`

	// 编解码器
	Codec encoding.Codec

	// 拦截器
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor

	// 监控和统计
	BinaryLog     bool `json:"binary_log" yaml:"binary_log"`
	InTapHandle   tap.ServerInHandle
	StatsHandlers []stats.Handler

	// 流控部分的配置
	MaxConcurrentStreams  uint32 `json:"max_concurrent_streams" yaml:"max_concurrent_streams"`     // 最大并发流数
	MaxRecvMsgSize        int    `json:"max_recv_msg_size" yaml:"max_recv_msg_size"`               // 最大接收消息大小
	MaxSendMsgSize        int    `json:"max_send_msg_size" yaml:"max_send_msg_size"`               // 最大发送消息大小
	InitialWindowSize     int32  `json:"initial_window_size" yaml:"initial_window_size"`           // 初始窗口大小
	InitialConnWindowSize int32  `json:"initial_conn_window_size" yaml:"initial_conn_window_size"` // 初始连接窗口大小

	// 缓冲区设置
	WriteBufferSize int `json:"write_buffer_size" yaml:"write_buffer_size"` // 写缓冲区大小
	ReadBufferSize  int `json:"read_buffer_size" yaml:"read_buffer_size"`   // 读缓冲区大小

	// 连接和超时
	ConnectionTimeout time.Duration                `json:"connection_timeout" yaml:"connection_timeout"`
	KeepaliveParams   *keepalive.ServerParameters  `json:"keepalive_params" yaml:"keepalive_params"`
	KeepalivePolicy   *keepalive.EnforcementPolicy `json:"keepalive_policy" yaml:"keepalive_policy"`

	// 头部设置
	MaxHeaderListSize *uint32 `json:"max_header_list_size" yaml:"max_header_list_size"` // 最大头部列表大小
	HeaderTableSize   *uint32 `json:"header_table_size" yaml:"header_table_size"`       // 头部表大小
	WaitForHandlers   bool    `json:"wait_for_handlers" yaml:"wait_for_handlers"`       // 是否等待所有处理程序完成（用于优雅关机）
}

// DefaultServerConfig 返回默认服务器配置
func DefaultServerConfig() *ServerConfig {
	cfg := new(ServerConfig)
	cfg.Initial()
	return cfg
}

// Initial 对一些配置做一个初始化
func (s *ServerConfig) Initial() {
	//默认监听本地端口10001
	if s.Address == "" {
		s.Address = ":10001"
	}
	//使用http2的默认配置
	//下面的初始化逻辑是grpc server端的官方默认配置
	//if s.MaxConcurrentStreams <= 0 {
	//	s.MaxConcurrentStreams = math.MaxUint32
	//}
	//if s.MaxRecvMsgSize <= 0 {
	//	s.MaxRecvMsgSize = 1024 * 1024 * 4
	//}
	//if s.MaxSendMsgSize <= 0 {
	//	s.MaxSendMsgSize = math.MaxInt32
	//}
	//官方默认的超时时间设置的是120秒，此处设置为180秒
	if s.ConnectionTimeout <= 0 {
		s.ConnectionTimeout = 180 * time.Second
	} else {
		timeout := s.ConnectionTimeout
		if timeout <= 1000 {
			//秒
			s.ConnectionTimeout = timeout * time.Second
		} else if timeout <= 1000000 {
			//毫秒
			s.ConnectionTimeout = timeout / 1000 * time.Second
		} else {
			//默认值
			s.ConnectionTimeout = 180 * time.Second
		}
	}

	//使用http2没有对此做默认设置，设置的稍微大一点，缓存多帧
	if s.WriteBufferSize <= 0 {
		s.WriteBufferSize = 64 * 1024
	}
	if s.ReadBufferSize <= 0 {
		s.ReadBufferSize = 64 * 1024
	}
	//使用http2的默认配置
	//grpc中默认最小秒级 此处增大但是比长连接服务要短
	//if s.KeepaliveParams == nil {
	//	s.KeepaliveParams = &keepalive.ServerParameters{
	//		MaxConnectionIdle:     20 * time.Minute,
	//		MaxConnectionAge:      30 * time.Minute,
	//		MaxConnectionAgeGrace: time.Minute,
	//		Time:                  30 * time.Minute,
	//		Timeout:               10 * time.Second,
	//	}
	//}
	//if s.KeepalivePolicy == nil {
	//	s.KeepalivePolicy = &keepalive.EnforcementPolicy{
	//		MinTime:             time.Minute,
	//		PermitWithoutStream: true,
	//	}
	//}
	//强制更正为true，优雅关机
	if !s.WaitForHandlers {
		s.WaitForHandlers = true
	}
	//使用http2的默认配置
	//if s.MaxHeaderListSize == nil {
	//	u := uint32(16 * 1024)
	//	s.MaxHeaderListSize = &u
	//}
	//if s.HeaderTableSize == nil {
	//	u := uint32(4 * 1024)
	//	s.HeaderTableSize = &u
	//}
}

// // ClientConfig 表示 gRPC 客户端的配置
//
//	type ClientConfig struct {
//		// 目标服务器地址
//		Target string `json:"target" yaml:"target"`
//		// 连接超时时间
//		DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
//		// 是否启用 TLS
//		EnableTLS bool `json:"enable_tls" yaml:"enable_tls"`
//		// TLS 验证的服务器名称
//		ServerName string `json:"server_name" yaml:"server_name"`
//		// 客户端 mTLS 证书文件路径
//		CertFile string `json:"cert_file" yaml:"cert_file"`
//		// 客户端 mTLS 密钥文件路径
//		KeyFile string `json:"key_file" yaml:"key_file"`
//		// 客户端可以接收的最大消息大小（字节）
//		MaxRecvMsgSize int `json:"max_recv_msg_size" yaml:"max_recv_msg_size"`
//		// 客户端可以发送的最大消息大小（字节）
//		MaxSendMsgSize int `json:"max_send_msg_size" yaml:"max_send_msg_size"`
//	}
func (s *ServerConfig) String() string {
	var builder strings.Builder

	builder.WriteString("gRPC Server Configuration:")
	builder.WriteString(fmt.Sprintf("  Address: %s", s.Address))
	builder.WriteString(fmt.Sprintf("  TLS_KeyFile: %s", s.KeyFile))
	builder.WriteString(fmt.Sprintf("  TLS_CertFile: %s", s.CertFile))
	builder.WriteString(fmt.Sprintf("  Health Check: %t", s.EnableHealthCheck))
	builder.WriteString(fmt.Sprintf("  Connection Timeout: %v", s.ConnectionTimeout))
	builder.WriteString(fmt.Sprintf("  Buffer Sizes: Write=%d, Read=%d", s.WriteBufferSize, s.ReadBufferSize))
	builder.WriteString(fmt.Sprintf("  Message Limits: MaxRecv=%d, MaxSend=%d", s.MaxRecvMsgSize, s.MaxSendMsgSize))
	builder.WriteString(fmt.Sprintf("  Concurrent Streams: %d", s.MaxConcurrentStreams))
	builder.WriteString(fmt.Sprintf("  Window Sizes: Initial=%d, InitialConn=%d", s.InitialWindowSize, s.InitialConnWindowSize))
	builder.WriteString(fmt.Sprintf("  Interceptors: Unary=%d, Stream=%d", len(s.UnaryInterceptors), len(s.StreamInterceptors)))
	builder.WriteString(fmt.Sprintf("  Wait For Handlers: %t", s.WaitForHandlers))

	if s.KeepaliveParams != nil {
		builder.WriteString(fmt.Sprintf("  Keepalive: MaxIdle=%v, MaxAge=%v, Time=%v, Timeout=%v",
			s.KeepaliveParams.MaxConnectionIdle,
			s.KeepaliveParams.MaxConnectionAge,
			s.KeepaliveParams.Time,
			s.KeepaliveParams.Timeout))
	}

	return builder.String()
}
