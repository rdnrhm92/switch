package pc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/logger"

	"github.com/gorilla/websocket"
)

var (
	ErrServerNotRunning = errors.New("server is not running")
	ErrClientNotFound   = errors.New("client not found")
	ErrClientNotActive  = errors.New("client is not active")
	ErrMaxConnections   = errors.New("maximum connections reached")
)

// Server WebSocket服务器
type Server struct {
	config *ServerConfig

	upgrader websocket.Upgrader

	// 客户端管理
	clients map[string]*Connection // 所有连接的客户端
	trusted map[string]*Connection // 受信客户端

	// 客户端管理通道
	register   chan *Connection // 注册新客户端
	unregister chan *Connection // 注销客户端

	// 自定义路由处理器
	customHandlers map[string]http.HandlerFunc

	// 服务器状态
	mu        sync.RWMutex
	isRunning bool
	server    *http.Server

	cancel context.CancelFunc
}

// NewServer 创建新的WebSocket服务器
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	server := &Server{
		config:         config,
		clients:        make(map[string]*Connection),
		trusted:        make(map[string]*Connection),
		register:       make(chan *Connection, 10),
		unregister:     make(chan *Connection, 100),
		customHandlers: make(map[string]http.HandlerFunc),
	}

	// 配置WebSocket升级器
	server.upgrader = websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
	}

	return server
}

// RegisterHandler 注册自定义WebSocket升级处理器 handler可以传递 用于在协议升级前的一些前置处理
func (s *Server) RegisterHandler(pattern string, handler http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		logger.Logger.Warnf("Cannot register handler %s: server is already running", pattern)
		return
	}

	// 自动处理WebSocket升级 每一个http端点都将升级为ws
	wrappedHandler := s.wrapWebSocketHandler(pattern, handler)
	s.customHandlers[pattern] = wrappedHandler
	logger.Logger.Infof("Registered custom WebSocket handler for pattern: %s", pattern)
}

// wrapWebSocketHandler 包装处理器，自动处理WebSocket升级
func (s *Server) wrapWebSocketHandler(pattern string, originalHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Infof("WebSocket endpoint %s accessed from %s", r.URL.Path, r.RemoteAddr)

		// 检查是否是WebSocket升级请求
		if r.Header.Get("Upgrade") != "websocket" {
			http.Error(w, "Expected WebSocket connection", http.StatusBadRequest)
			return
		}

		// 检查请求是否携带了客户端ID
		if r.Header.Get("ConnectionId") == "" {
			http.Error(w, "Expected WebSocket connection", http.StatusBadRequest)
			return
		}

		// 执行业务前置处理器
		if originalHandler != nil {
			originalHandler(w, r)
		}

		// 自动执行WebSocket升级
		s.handleWebSocket(w, r)
	}
}

// HandleWebSocketUpgrade 公开的WebSocket升级处理器，供业务层使用
func (s *Server) HandleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
	s.handleWebSocket(w, r)
}

// Start 启动WebSocket服务器
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return errors.New("server is already running")
	}

	ctx, s.cancel = context.WithCancel(ctx)

	// 监听连接的建立跟销毁
	recovery.SafeGo(ctx, func(ctx context.Context) error {
		return s.run(ctx)
	}, ServerStart)

	mux := http.NewServeMux()
	//健康检查端点
	mux.HandleFunc(Health, s.handleHealth)
	//客户端列表端点
	mux.HandleFunc(Clients, s.handleClients)

	// 注册自定义处理器 - 默认都走WebSocket升级逻辑
	for pattern, handler := range s.customHandlers {
		mux.HandleFunc(pattern, handler)
		logger.Logger.Infof("Registered custom WebSocket endpoint: %s", pattern)
	}

	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         s.config.Address,
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.isRunning = true
	logger.Logger.Infof("WebSocket server starting on %s", s.config.Address)

	// 启动HTTP服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Errorf("Server error: %v", err)
		}
	}()

	return nil
}

// Stop 停止WebSocket服务器
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return ErrServerNotRunning
	}

	logger.Logger.Info("Stopping WebSocket server...")

	s.cancel()

	for _, conn := range s.clients {
		conn.Close(DisconnectReasonExternalClose)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Logger.Errorf("Server shutdown error: %v", err)
		return err
	}

	s.isRunning = false
	logger.Logger.Info("WebSocket server stopped")
	return nil
}

// run 处理客户端连接建立/销毁和广播
func (s *Server) run(ctx context.Context) error {
	//健康检查器
	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Warnf("The server has detected an external exit")
			return nil
		case conn := <-s.register:
			s.mu.Lock()
			// 超过最大连接数-拒绝 不加入连接管理
			if len(s.clients) >= s.config.MaxConnections {
				s.mu.Unlock()
				conn.Close(DisconnectReasonServerKick)
				logger.Logger.Warnf("Rejected connection %s: maximum connections reached", conn.ID)
				continue
			}

			// 加入连接管理但不加入受信列表
			s.clients[conn.ID] = conn
			logger.Logger.Infof("%s: Already included in the scope of server management", conn.ID)
			s.mu.Unlock()

			info := conn.GetClientInfo()
			if info != nil {
				logger.Logger.Infof("Connection %s established from %s", conn.ID, info.RemoteAddr)
			}

			// 注册完成后启动连接（解决竞态条件问题）
			conn.Start(ctx)

		case conn := <-s.unregister:
			s.removeConnection(conn)

		case <-ticker.C:
			//健康检查
			s.checkClientHealth()
		}
	}
}

// handleWebSocket 处理WebSocket连接请求
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 将http请求升级为ws协议
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	// 创建客户端信息
	connectionId := r.Header.Get("ConnectionId")
	clientInfo := &ClientProxyInfo{
		ID:            connectionId,
		RemoteAddr:    r.RemoteAddr,
		UserAgent:     r.UserAgent(),
		Headers:       make(map[string]string),
		ServerVersion: s.config.ServerVersion,
		ConnectTime:   time.Now(),
		Endpoint:      r.URL.Path,
	}

	// 提取请求头信息
	for key, values := range r.Header {
		if len(values) > 0 {
			clientInfo.Headers[key] = values[0]
		}
	}

	// 创建消息处理器
	processor := NewProxyMessageProcessor(s)

	// 创建统一连接
	connection := NewConnection(clientInfo.ID, conn, RoleProxy, s.config, processor)

	// 设置客户端信息
	connection.SetClientInfo(clientInfo)

	// 设置连接关闭回调
	connection.SetStateChangeCallback(func(change *ConnectionStateChange) {
		c := change.Conn
		select {
		case s.unregister <- c:
		default:
			// 通道满了 直接移除 不需要关闭因为这个被回调本身就是从关闭过来的
			logger.Logger.Warnf("Unregister channel full, directly removing connection %s", c.ID)
			s.removeConnection(c)
		}
	})

	//注册客户端（连接将在注册完成后启动 必须先注册后启动）
	s.RegisterClient(connection)

}

// handleHealth 健康检查端点
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	clientCount := len(s.clients)
	isRunning := s.isRunning
	s.mu.RUnlock()

	status := map[string]interface{}{
		"status":      "ok",
		"running":     isRunning,
		"clients":     clientCount,
		"max_clients": s.config.MaxConnections,
		"timestamp":   time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleClients 获取客户端列表端点
func (s *Server) handleClients(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	clients := make([]ClientProxyInfo, 0, len(s.clients))
	for _, conn := range s.clients {
		if conn.IsActive() {
			info := conn.GetClientInfo()
			if info != nil {
				clients = append(clients, *info)
			}
		}
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

// checkClientHealth 检查客户端健康状态
// 使用客户端代理的最后Ping的时间跟超时时间做对比
func (s *Server) checkClientHealth() {
	now := time.Now()
	timeout := s.config.ClientTimeout

	// 先获取所有客户端的快照，避免长时间持有锁
	s.mu.RLock()
	clientsSnapshot := make([]*Connection, 0, len(s.clients))
	for _, conn := range s.clients {
		clientsSnapshot = append(clientsSnapshot, conn)
	}
	s.mu.RUnlock()

	// 检查了直接关闭 然后通过关闭的回调移除这个连接
	for _, conn := range clientsSnapshot {
		conn.mu.RLock()
		lastPing := conn.lastPing
		conn.mu.RUnlock()

		if now.Sub(lastPing) > timeout {
			logger.Logger.Infof("Connection %s timed out, disconnecting", conn.ID)
			conn.Close(DisconnectReasonHeartbeatTimeout)
		}
	}
}

// removeConnection 从服务器中移除连接
// 只负责从列表中删除 不负责关闭连接 谁发现谁关闭 否则关闭 删除 关闭可能会有死循环
func (s *Server) removeConnection(conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[conn.ID]; ok {
		delete(s.clients, conn.ID)
		delete(s.trusted, conn.ID)
		logger.Logger.Infof("Connection %s removed from server", conn.ID)
	}
}

// RegisterClient 注册客户端
func (s *Server) RegisterClient(connection *Connection) {
	s.register <- connection
}

// TrustedClient 信任客户端
func (s *Server) TrustedClient(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	connection, ok := s.clients[id]
	if !ok {
		return fmt.Errorf("connection %s is not found from server.clients", id)
	}
	s.trusted[id] = connection
	return nil
}

// IsTrustedClient 检查是否为受信客户端
func (s *Server) IsTrustedClient(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.trusted[id]
	return exists
}

// getTrustedActiveConnections 获取所有活跃的受信连接（支持过滤器）
func (s *Server) getTrustedActiveConnections(filter func(*Connection) bool) []*Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var trustedConnections []*Connection
	for _, conn := range s.trusted {
		if conn.IsActive() && (filter == nil || filter(conn)) {
			trustedConnections = append(trustedConnections, conn)
		}
	}
	return trustedConnections
}

// getUniqueDeviceConnections 获取去重后的设备连接（每台设备只保留一个连接）
// 去重规则：相同 IP + ServiceName 的设备只保留最新的连接
func (s *Server) getUniqueDeviceConnections(filter func(*Connection) bool) []*Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 使用 map 去重：key = 设备唯一标识
	deviceMap := make(map[string]*Connection)

	for _, conn := range s.trusted {
		if !conn.IsActive() {
			continue
		}

		if filter != nil && !filter(conn) {
			continue
		}

		info := conn.GetClientInfo()
		if info == nil {
			continue
		}

		// 生成设备唯一标识
		deviceKey := s.generateDeviceKey(info)

		// 如果该设备还没有连接，或者当前连接更新，则使用当前连接
		if existing, exists := deviceMap[deviceKey]; !exists || conn.connectTime.After(existing.connectTime) {
			deviceMap[deviceKey] = conn
		}
	}

	// 转换为切片
	result := make([]*Connection, 0, len(deviceMap))
	for _, conn := range deviceMap {
		result = append(result, conn)
	}

	logger.Logger.Infof("Unique device connections: %d (from %d total trusted connections)", len(result), len(s.trusted))
	return result
}

// generateDeviceKey 生成设备唯一标识
// 优先使用内网IP，其次公网IP，最后使用RemoteAddr
func (s *Server) generateDeviceKey(info *ClientProxyInfo) string {
	// 优先使用第一个内网IP + ServiceName
	if len(info.InternalIP) > 0 && info.InternalIP[0] != "" {
		return fmt.Sprintf("%s:%s:%s:%s", info.InternalIP[0], info.ServiceName, info.NamespaceTag, info.EnvTag)
	}

	// 其次使用第一个公网IP + ServiceName
	if len(info.PublicIP) > 0 && info.PublicIP[0] != "" {
		return fmt.Sprintf("%s:%s:%s:%s", info.PublicIP[0], info.ServiceName, info.NamespaceTag, info.EnvTag)
	}

	// 兜底使用 RemoteAddr + ServiceName
	return fmt.Sprintf("%s:%s:%s:%s", info.RemoteAddr, info.ServiceName, info.NamespaceTag, info.EnvTag)
}

// BroadcastToTrusted 只向受信客户端(不带过滤)广播
func (s *Server) BroadcastToTrusted(message []byte) {
	trustedConnections := s.getTrustedActiveConnections(nil)

	if len(trustedConnections) == 0 {
		return
	}

	logger.Logger.Infof("Broadcasting to %d trusted connections", len(trustedConnections))

	// 并发发送
	var wg sync.WaitGroup
	for _, conn := range trustedConnections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.Send(message); err != nil {
				logger.Logger.Errorf("Broadcasting to %v trusted has error: %v", conn.ID, err.Error())
			}
		}(conn)
	}
	wg.Wait()
}

func (s *Server) BroadcastToGroup(message []byte, filter func(*Connection) bool) {
	targetConnections := s.getTrustedActiveConnections(filter)

	if len(targetConnections) == 0 {
		return
	}

	logger.Logger.Infof("Broadcasting to %d filtered connections", len(targetConnections))

	// 并发发送
	var wg sync.WaitGroup
	for _, conn := range targetConnections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.Send(message); err != nil {
				logger.Logger.Errorf("Broadcasting to %v has error: %v", conn.ID, err.Error())
			}
		}(conn)
	}
	wg.Wait()
}

// SendToClient 发送消息给指定受信客户端
func (s *Server) SendToClient(clientID string, message []byte) error {
	s.mu.RLock()
	conn, exists := s.trusted[clientID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client %s is not trusted or not found", clientID)
	}

	if !conn.IsActive() {
		return ErrClientNotActive
	}

	return conn.Send(message)
}

// SendJSONToClient 发送JSON消息给指定客户端
func (s *Server) SendJSONToClient(clientID string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.SendToClient(clientID, data)
}

// SendReliableMessage 可靠发送消息给指定受信客户端（带重试机制）
func (s *Server) SendReliableMessage(clientID string, msgType MessageType, data interface{}, timeout time.Duration) error {
	return s.SendReliableMessageWithRetry(clientID, msgType, data, timeout, 3, 2*time.Second)
}

// SendReliableMessageWithRetry 可靠发送消息给 指定 受信客户端（自定义重试参数）
func (s *Server) SendReliableMessageWithRetry(clientID string, msgType MessageType, data interface{}, timeout time.Duration, maxRetries int, retryDelay time.Duration) error {
	s.mu.RLock()
	conn, exists := s.trusted[clientID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client %s is not trusted or not found", clientID)
	}

	if !conn.IsActive() {
		return ErrClientNotActive
	}

	// 使用Connection的统一重试机制
	return conn.SendRequestWithRetry(msgType, data, timeout, maxRetries, retryDelay)
}

// ConnectionFilter 发送时的过滤器
type ConnectionFilter func(*Connection) bool

// AndFilter 组合多个过滤器 AND 逻辑
func AndFilter(filters ...ConnectionFilter) ConnectionFilter {
	return func(connection *Connection) bool {
		for _, filter := range filters {
			if !filter(connection) {
				return false
			}
		}
		return true
	}
}

// BroadcastReliableMessageToGroup 可靠广播消息给受信客户端（支持过滤器）
func (s *Server) BroadcastReliableMessageToGroup(msgType MessageType, data interface{}, timeout time.Duration, maxRetries int, retryDelay time.Duration, filter ConnectionFilter) error {
	trustedConnections := s.getTrustedActiveConnections(filter)

	if len(trustedConnections) == 0 {
		logger.Logger.Warn("No trusted connections available for reliable broadcast")
		return nil
	}

	logger.Logger.Infof("Broadcasting reliable message to %d trusted connections with retry", len(trustedConnections))

	// 并发发送给所有受信客户端，每个都使用重试机制
	var wg sync.WaitGroup
	var allError []error
	var errorMu sync.Mutex

	for _, conn := range trustedConnections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.SendRequestWithRetry(msgType, data, timeout, maxRetries, retryDelay); err != nil {
				errorMu.Lock()
				allError = append(allError, fmt.Errorf("failed to send to client %s: %v", c.ID, err))
				errorMu.Unlock()
			}
		}(conn)
	}

	wg.Wait()

	// 如果有错误，返回合并的错误信息
	if len(allError) > 0 {
		var errorMsgs []string
		for _, err := range allError {
			errorMsgs = append(errorMsgs, err.Error())
		}
		return fmt.Errorf("broadcast failed for some clients: %s", strings.Join(errorMsgs, "; "))
	}

	return nil
}

// GetClientCount 获取当前连接的客户端数量
func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// GetClient 获取指定ID的客户端
func (s *Server) GetClient(clientID string) (*Connection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, exists := s.clients[clientID]
	return conn, exists
}

// IsRunning 检查服务器是否正在运行
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}
