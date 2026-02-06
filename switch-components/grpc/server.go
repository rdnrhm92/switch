package grpc

import (
	"crypto/tls"
	"fmt"
	"net"

	"gitee.com/fatzeng/switch-sdk-core/invoke/rpc"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	//封装一下grpc
	*grpc.Server
	config *rpc.ServerConfig
}

func NewServer(config *rpc.ServerConfig) (*Server, error) {
	if config == nil {
		config = rpc.DefaultServerConfig()
	} else {
		config.Initial()
	}
	opts := make([]grpc.ServerOption, 0)
	//流控相关的
	if config.MaxConcurrentStreams != 0 {
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams)
	}
	if config.MaxRecvMsgSize != 0 {
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize)
	}
	if config.MaxSendMsgSize != 0 {
		grpc.MaxSendMsgSize(config.MaxSendMsgSize)
	}
	if config.InitialWindowSize != 0 {
		grpc.InitialWindowSize(config.InitialWindowSize)
	}
	if config.InitialConnWindowSize != 0 {
		grpc.InitialConnWindowSize(config.InitialConnWindowSize)
	}

	//缓冲区默认用的64 业务可设置
	if config.WriteBufferSize != 0 {
		grpc.WriteBufferSize(config.WriteBufferSize)
	}
	if config.ReadBufferSize != 0 {
		grpc.ReadBufferSize(config.ReadBufferSize)
	}

	if config.ConnectionTimeout > 0 {
		opts = append(opts, grpc.ConnectionTimeout(config.ConnectionTimeout))
	}
	if config.Codec != nil {
		encoding.RegisterCodec(config.Codec)
	}

	// TLS凭证
	if config.KeyFile != "" && config.CertFile != "" {
		creds, err := loadTLSCredentials(config)
		if err != nil {
			return nil, fmt.Errorf("load TLS fail: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}
	if config.KeepaliveParams != nil {
		opts = append(opts, grpc.KeepaliveParams(*config.KeepaliveParams))
	}
	if config.KeepalivePolicy != nil {
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(*config.KeepalivePolicy))
	}
	if config.MaxHeaderListSize != nil {
		opts = append(opts, grpc.MaxHeaderListSize(*config.MaxHeaderListSize))
	}
	if config.HeaderTableSize != nil {
		opts = append(opts, grpc.HeaderTableSize(*config.HeaderTableSize))
	}
	for _, handler := range config.StatsHandlers {
		opts = append(opts, grpc.StatsHandler(handler))
	}
	if config.InTapHandle != nil {
		opts = append(opts, grpc.InTapHandle(config.InTapHandle))
	}

	opts = append(opts, grpc.WaitForHandlers(config.WaitForHandlers))

	// 拦截器设置
	if len(config.UnaryInterceptors) > 0 {
		opts = append(opts, grpc.UnaryInterceptor(UnaryServerInterceptorChain(config.UnaryInterceptors...)))
	} else {
		// 默认拦截器
		opts = append(opts, grpc.UnaryInterceptor(UnaryServerInterceptorChain(
			ResponseSetInterceptor(),
			MetadataInterceptor(),
			LoggingInterceptor(logger.Logger),
			ErrorHandlingInterceptor(),
			RecoveryInterceptor(),
		)))
	}

	if len(config.StreamInterceptors) > 0 {
		opts = append(opts, grpc.StreamInterceptor(StreamServerInterceptorChain(config.StreamInterceptors...)))
	}

	grpcServer := grpc.NewServer(opts...)

	//健康检查
	if config.EnableHealthCheck {
		healthServer := health.NewServer()
		healthpb.RegisterHealthServer(grpcServer, healthServer)
	}

	return &Server{
		Server: grpcServer,
		config: config,
	}, nil
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("server listen port fail: %v", err)
	}

	logger.Logger.Printf("gRPC server starting，listen address: %s", s.config.Address)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server start fail: %v", err)
	}

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() {
	logger.Logger.Info("gRPC server stopping...")
	s.GracefulStop()
	logger.Logger.Info("gRPC server stopped")
}

func loadTLSCredentials(config *rpc.ServerConfig) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("server load tls fail: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return credentials.NewTLS(tlsConfig), nil
}
