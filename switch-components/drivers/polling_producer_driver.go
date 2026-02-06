package drivers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

const PollingProducerDriverType driver.DriverType = "polling_producer"

// PendingConnection 等待中的长轮询连接
type PendingConnection struct {
	ResponseWriter http.ResponseWriter
	Context        context.Context
	Cancel         context.CancelFunc
	ConnectedAt    time.Time
	ClientID       string

	// 客户端网络信息
	PublicIPs   []string
	InternalIPs []string
}

// PollingProducer 长轮询生产者
type PollingProducer struct {
	PollingProducerValidator
	server             *http.Server
	config             *PollingProducerConfig
	mutex              sync.RWMutex
	running            bool
	pendingConnections map[string]*PendingConnection
	messageQueue       chan []byte
	ctx                context.Context
	cancel             context.CancelFunc

	// 配置版本缓存
	configCache *ConfigCache

	driverName   string
	callback     driver.DriverFailureCallback
	failureCount int // 连续失败次数
}

// NewPollingProducer 创建长轮询生产者驱动
// 使用http + 长轮询的模式
func NewPollingProducer(c *PollingProducerConfig) (*PollingProducer, error) {
	producer := &PollingProducer{
		config:             c,
		pendingConnections: make(map[string]*PendingConnection),
		messageQueue:       make(chan []byte, 1000),
		configCache:        NewConfigCache(),
		failureCount:       0,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(DefaultPollingPath, producer.handleLongPoll)
	mux.HandleFunc("/health", producer.handleHealth)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  c.getServerReadTimeout(),
		WriteTimeout: c.getServerWriteTimeout(),
		IdleTimeout:  c.getServerIdleTimeout(),
	}

	// 配置HTTPS
	if c.getSecurity() != nil && c.getSecurity().CertFile != "" && c.getSecurity().KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.getSecurity().CertFile, c.getSecurity().KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	producer.server = server

	return producer, nil
}

func (p *PollingProducer) RecreateFromConfig() (driver.Driver, error) {
	return NewPollingProducer(p.config)
}

func (p *PollingProducer) GetDriverName() string {
	return p.driverName
}

func (p *PollingProducer) SetDriverMeta(name string) {
	p.driverName = name
}

func (p *PollingProducer) SetFailureCallback(callback driver.DriverFailureCallback) {
	p.callback = callback
}

func (p *PollingProducer) GetDriverType() driver.DriverType {
	return PollingProducerDriverType
}

// Start 启动长轮询生产者服务器
func (p *PollingProducer) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		logger.Logger.Errorf("polling producer is already running")
		return fmt.Errorf("polling producer is already running")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.running = true

	// 启动消息分发协程
	go p.startMessageDispatcher()

	// 启动清理协程
	go p.startConnectionCleaner()

	backoffDuration := p.config.getBackoffDuration()
	recovery.SafeGo(p.ctx, func(ctx context.Context) error {
		return p.startHTTPServer(ctx)
	}, string(PollingProducerDriverType), recovery.WithRetryInterval(backoffDuration))

	logger.Logger.Infof("Polling producer server starting on port %s", p.config.getPort())
	return nil
}

// startHTTPServer 启动 HTTP 服务器
func (p *PollingProducer) startHTTPServer(ctx context.Context) error {
	// 确定监听地址
	addr := ":" + p.config.getPort()
	retries := p.config.getMaxRetries()

	// 尝试监听端口
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Logger.Errorf("Failed to listen on %s: %v", addr, err)

		// 统计失败次数
		p.mutex.Lock()
		p.failureCount++
		currentFailures := p.failureCount
		p.mutex.Unlock()

		logger.Logger.Errorf("Polling producer listen error (failure %d/%d): %v", currentFailures, retries, err)

		// 如果超过最大重试次数，触发故障回调
		if currentFailures >= retries {
			logger.Logger.Errorf("Polling producer reached max retries (%d), triggering failure callback", retries)

			// 主动关闭驱动，确保资源清理
			if err := p.Close(); err != nil {
				logger.Logger.Errorf("Failed to close polling producer: %v", err)
			}

			// 触发回调通知 Manager
			if p.callback != nil {
				p.callback("PollingProducer retry exhausted", fmt.Errorf("max retries exhausted: %w", err))
			}
			// 不再重试
			return nil
		}

		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	logger.Logger.Infof("Polling producer HTTP server listening on %s", addr)

	errCh := make(chan error, 1)

	// 启动 HTTP 服务器
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Errorf("Polling producer HTTP server goroutine panicked: %v", r)
				errCh <- fmt.Errorf("polling producer server panic: %v", r)
			}
		}()

		if err := p.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Logger.Errorf("Polling producer HTTP server failed: %v", err)
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Logger.Info("Polling producer context cancelled, removing from driver pool")

		if err := p.Close(); err != nil {
			logger.Logger.Errorf("Failed to close polling producer: %v", err)
		}

		p.mutex.Lock()
		p.failureCount = 0
		p.mutex.Unlock()

		// 触发清理回调
		if p.callback != nil {
			p.callback("PollingProducer context cancelled", fmt.Errorf("driver context cancelled"))
		}

		return nil

	case err := <-errCh:
		p.mutex.Lock()
		p.failureCount++
		currentFailures := p.failureCount
		p.mutex.Unlock()

		logger.Logger.Errorf("Polling producer HTTP server error (failure %d/%d): %v", currentFailures, retries, err)

		// 如果超过最大重试次数，触发故障回调
		if currentFailures >= retries {
			logger.Logger.Errorf("Polling producer reached max retries (%d), triggering failure callback", retries)

			if err := p.Close(); err != nil {
				logger.Logger.Errorf("Failed to close polling producer: %v", err)
			}

			// 触发回调通知
			if p.callback != nil {
				p.callback("PollingProducer retry exhausted", fmt.Errorf("max retries exhausted: %w", err))
			}
			// 不再重试
			return nil
		}

		return err
	}
}

func (p *PollingProducer) Notify(ctx context.Context, data interface{}) error {
	return p.PushMessage(data)
}

// PushMessage 推送消息给所有等待的长轮询连接
func (p *PollingProducer) PushMessage(message interface{}) error {
	// 添加一个缓存项
	configVersion := p.configCache.AddConfig(message)

	// 获取刚添加的配置
	config, ok := p.configCache.GetVersion(configVersion)
	if !ok {
		return fmt.Errorf("failed to retrieve added config version %d", configVersion)
	}

	// 获取该版本之后的所有配置 可能在并发场景下有其他配置被添加
	laterConfigs := p.configCache.GetVersionsSince(configVersion)

	// 合并当前配置和后续配置
	configs := make([]*ConfigVersion, 0, 1+len(laterConfigs))
	configs = append(configs, config)
	configs = append(configs, laterConfigs...)

	// 序列化消息
	data, err := json.Marshal(configs)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.PushRawMessage(data)
}

// GetCacheStats 获取缓存统计信息
func (p *PollingProducer) GetCacheStats() map[string]interface{} {
	return p.configCache.GetCacheStats()
}

// PushRawMessage 推送原始消息给所有等待的长轮询连接
func (p *PollingProducer) PushRawMessage(data []byte) error {
	if !p.running {
		return fmt.Errorf("polling producer is not running")
	}

	select {
	case p.messageQueue <- data:
		logger.Logger.Debugf("Message queued for push: %d bytes", len(data))
		return nil
	default:
		return fmt.Errorf("message queue is full")
	}
}

// handleLongPoll 处理长轮询请求
func (p *PollingProducer) handleLongPoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证客户端token
	if !p.validateClientToken(r) {
		logger.Logger.Warnf("Unauthorized request from %s: invalid or missing token", r.RemoteAddr)
		http.Error(w, "Unauthorized: Invalid or missing token", http.StatusUnauthorized)
		return
	}

	// 解析客户端版本号
	clientVersionStr := r.URL.Query().Get("version")
	var clientVersion uint64 = 0
	if clientVersionStr != "" {
		if v, err := strconv.ParseUint(clientVersionStr, 10, 64); err == nil {
			clientVersion = v
		}
	}

	// 检查是否有更新的配置
	if p.configCache.HasNewerVersion(clientVersion) {
		sinceConfigs := p.configCache.GetVersionsSince(clientVersion)
		logger.Logger.Infof("Client %s (version %d) received %d newer configs, latest version: %d",
			r.RemoteAddr, clientVersion, len(sinceConfigs), p.configCache.GetLatestVersion())
		p.sendConfigResponse(w, sinceConfigs)
		return
	}

	// 获取客户端提供的IP信息
	publicIPsHeader := r.Header.Get("X-Public-IPs")
	internalIPsHeader := r.Header.Get("X-Internal-IPs")

	var publicIPs, internalIPs []string
	if publicIPsHeader != "" {
		publicIPs = strings.Split(publicIPsHeader, ",")
		for i, ip := range publicIPs {
			publicIPs[i] = strings.TrimSpace(ip)
		}
	}
	if internalIPsHeader != "" {
		internalIPs = strings.Split(internalIPsHeader, ",")
		for i, ip := range internalIPs {
			internalIPs[i] = strings.TrimSpace(ip)
		}
	}

	// 生成客户端ID
	clientID := fmt.Sprintf("%s-%d", r.RemoteAddr, time.Now().UnixNano())

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(r.Context(), p.config.getLongPollTimeout())
	defer cancel()

	// 创建等待连接
	conn := &PendingConnection{
		ResponseWriter: w,
		Context:        ctx,
		Cancel:         cancel,
		ConnectedAt:    time.Now(),
		ClientID:       clientID,
		PublicIPs:      publicIPs,
		InternalIPs:    internalIPs,
	}

	// 添加到等待列表
	p.mutex.Lock()
	p.pendingConnections[clientID] = conn
	connectionCount := len(p.pendingConnections)
	p.mutex.Unlock()

	logger.Logger.Debugf("New long poll connection: %s, total connections: %d", clientID, connectionCount)

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	<-ctx.Done()

	// 移除连接
	p.removePendingConnection(clientID)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		w.WriteHeader(http.StatusNoContent)
		logger.Logger.Debugf("Long poll connection timeout: %s", clientID)
	} else {
		// 服务端正常响应了或者是客户端断开连接了
		logger.Logger.Debugf("Long poll connection ended: %s", clientID)
	}
}

// startMessageDispatcher 启动消息分发协程
func (p *PollingProducer) startMessageDispatcher() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Errorf("Message dispatcher panic: %v", r)
			// 触发故障回调
			if p.callback != nil {
				p.callback("PollingProducer dispatcher panic", fmt.Errorf("message dispatcher panic: %v", r))
			}
		}
	}()

	logger.Logger.Info("Message dispatcher started")

	for {
		select {
		case <-p.ctx.Done():
			logger.Logger.Info("Message dispatcher stopped")
			return
		case message := <-p.messageQueue:
			p.dispatchMessage(message)
		}
	}
}

// dispatchMessage 分发消息给所有等待的连接
func (p *PollingProducer) dispatchMessage(data []byte) {
	configs := make([]*ConfigVersion, 0)
	err := json.Unmarshal(data, &configs)
	if err != nil {
		logger.Logger.Errorf("Unable to unmarshal config json: %v", err)
		return
	}
	p.mutex.Lock()
	connections := make([]*PendingConnection, 0, len(p.pendingConnections))
	connectionsToRemove := make([]string, 0, len(p.pendingConnections))

	for clientID, conn := range p.pendingConnections {
		// 检查连接是否仍然有效
		select {
		case <-conn.Context.Done():
			connectionsToRemove = append(connectionsToRemove, clientID)
		default:
			connections = append(connections, conn)
			connectionsToRemove = append(connectionsToRemove, clientID)
		}
	}
	// 移除将要处理的连接 客户端ID带时间戳的并不会导致aba问题
	for _, clientID := range connectionsToRemove {
		delete(p.pendingConnections, clientID)
	}
	p.mutex.Unlock()

	logger.Logger.Infof("Dispatching message to %d connections", len(connections))

	// 并发发送消息给所有等待的连接
	for _, conn := range connections {
		go p.sendMessageToConnection(conn, configs)
	}
}

// sendMessageToConnection 发送消息给单个连接
func (p *PollingProducer) sendMessageToConnection(conn *PendingConnection, configs []*ConfigVersion) {
	defer conn.Cancel()

	// 连接有效性的二次检查
	select {
	case <-conn.Context.Done():
		logger.Logger.Debugf("Connection already closed: %s", conn.ClientID)
		return
	default:
	}

	// 发送消息
	p.sendConfigResponse(conn.ResponseWriter, configs)
}

// removePendingConnection 移除等待中的连接
func (p *PollingProducer) removePendingConnection(clientID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if conn, exists := p.pendingConnections[clientID]; exists {
		conn.Cancel()
		delete(p.pendingConnections, clientID)
		logger.Logger.Debugf("Removed pending connection: %s", clientID)
	}
}

// startConnectionCleaner 启动连接清理协程
func (p *PollingProducer) startConnectionCleaner() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Errorf("Connection cleaner panic: %v", r)
			// 触发故障回调
			if p.callback != nil {
				p.callback("polling producer cleaner panic", fmt.Errorf("connection cleaner panic: %v", r))
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second) // 每30秒清理一次
	defer ticker.Stop()

	logger.Logger.Info("Connection cleaner started")

	for {
		select {
		case <-p.ctx.Done():
			logger.Logger.Info("Connection cleaner stopped")
			return
		case <-ticker.C:
			p.cleanExpiredConnections()
		}
	}
}

// cleanExpiredConnections 清理过期的连接
func (p *PollingProducer) cleanExpiredConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	expiredConnections := make([]string, 0)

	for clientID, conn := range p.pendingConnections {
		// 检查连接是否已过期或已关闭
		select {
		case <-conn.Context.Done():
			expiredConnections = append(expiredConnections, clientID)
		default:
			// 检查是否超过最大等待时间
			if now.Sub(conn.ConnectedAt) > p.config.getLongPollTimeout()+10*time.Second {
				expiredConnections = append(expiredConnections, clientID)
			}
		}
	}

	// 移除过期连接
	for _, clientID := range expiredConnections {
		if conn, exists := p.pendingConnections[clientID]; exists {
			conn.Cancel()
			delete(p.pendingConnections, clientID)
		}
	}

	if len(expiredConnections) > 0 {
		logger.Logger.Debugf("Cleaned %d expired connections", len(expiredConnections))
	}
}

// handleHealth 处理健康检查请求
func (p *PollingProducer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	p.mutex.RLock()
	connectionCount := len(p.pendingConnections)
	p.mutex.RUnlock()

	response := map[string]interface{}{
		"status":      "healthy",
		"connections": connectionCount,
		"timestamp":   time.Now().Unix(),
		"cache":       p.GetCacheStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Close 关闭生产者
func (p *PollingProducer) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	logger.Logger.Info("Closing polling producer...")

	// 取消上下文，停止所有协程
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	// 关闭所有等待的连接
	for clientID, conn := range p.pendingConnections {
		conn.Cancel()
		logger.Logger.Debugf("Closed pending connection: %s", clientID)
	}
	p.pendingConnections = make(map[string]*PendingConnection)

	// 关闭HTTP服务器
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := p.server.Shutdown(ctx); err != nil {
			logger.Logger.Warnf("Failed to gracefully shutdown server: %v", err)
		}
		p.server = nil
	}

	// 关闭消息队列
	if p.messageQueue != nil {
		close(p.messageQueue)
		p.messageQueue = nil
	}

	p.running = false
	logger.Logger.Info("Polling producer closed successfully")
	return nil
}

// GetConfig 获取配置
func (p *PollingProducer) GetConfig() *PollingProducerConfig {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.config
}

// IsRunning 检查是否正在运行
func (p *PollingProducer) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

// GetConnectionCount 获取当前等待连接数
func (p *PollingProducer) GetConnectionCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.pendingConnections)
}

// sendConfigResponse 发送配置响应
func (p *PollingProducer) sendConfigResponse(w http.ResponseWriter, configs []*ConfigVersion) {
	if configs == nil || len(configs) == 0 {
		logger.Logger.Warnf("Empty config list")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(configs)
	if err != nil {
		logger.Logger.Errorf("Failed to marshal config response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		logger.Logger.Errorf("Failed to write config response: %v", err)
	}
}

// validateClientToken 验证客户端token
func (p *PollingProducer) validateClientToken(r *http.Request) bool {
	// 如果没有配置安全设置或没有配置有效token，则跳过验证
	if p.config.getSecurity() == nil || p.config.getSecurity().ValidTokens == nil || len(p.config.getSecurity().ValidTokens) == 0 {
		return true
	}

	// 获取Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// 检查Bearer token格式
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return false
	}

	// 提取token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return false
	}

	// 验证token是否在有效列表中
	for _, validToken := range p.config.getSecurity().ValidTokens {
		if token == validToken {
			return true
		}
	}

	return false
}
