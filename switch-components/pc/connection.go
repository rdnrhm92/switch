package pc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"
	"github.com/gorilla/websocket"
)

// ConnectionConfig 统一连接配置接口
type ConnectionConfig interface {
	GetHeartbeatInterval() time.Duration
	GetWriteTimeout() time.Duration
	GetReadTimeout() time.Duration
	GetVersion() string
}

// MessageProcessor 统一消息处理
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, conn *Connection, message []byte) error
	OnConnect(ctx context.Context, conn *Connection) ConnectHandler
	OnDisconnect(conn *Connection, err error)
}

// Connection 统一的WebSocket连接抽象
type Connection struct {
	// 基础连接信息
	ID   string
	conn *websocket.Conn
	role ConnectionRole

	// 通道和上下文
	sendCh chan []byte
	mu     sync.RWMutex

	// 配置接口
	config ConnectionConfig

	// 处理器
	process MessageProcessor

	isActive    bool
	connectTime time.Time
	lastPing    time.Time
	lastPong    time.Time

	// 区分客户端跟服务端的客户端代理信息
	clientData *ClientConnectionData
	proxyData  *ProxyConnectionData

	// pending请求管理(超时无响应将重发)
	pendingRequests map[string]*PendingRequest
	requestMu       sync.RWMutex

	// 共用的连接信息 服务端的信息记录
	serverInfo *ServerInfo

	sendRegister chan struct{}

	// 连接内部的上下文
	connectionCtx context.Context
	// 连接内部的退出协定
	connectionCancel context.CancelFunc

	// start执行保证幂等
	startOnce sync.Once
}

// NewConnection 创建统一连接
func NewConnection(id string, conn *websocket.Conn, role ConnectionRole, config ConnectionConfig, process MessageProcessor) *Connection {
	uc := &Connection{
		ID:              id,
		conn:            conn,
		role:            role,
		config:          config,
		process:         process,
		sendCh:          make(chan []byte, 256),
		isActive:        true,
		connectTime:     time.Now(),
		lastPong:        time.Now(),
		pendingRequests: make(map[string]*PendingRequest),
		sendRegister:    make(chan struct{}, 1),
		serverInfo:      &ServerInfo{},
	}

	// 根据角色初始化特定数据
	switch role {
	case RoleClient:
		uc.clientData = &ClientConnectionData{}
	case RoleProxy:
		uc.proxyData = &ProxyConnectionData{}
	}

	return uc
}

// Start 启动连接（带重连机制）
func (uc *Connection) Start(ctx context.Context) {
	// 防止同一connection重复start 比如重连的时候
	uc.startOnce.Do(func() {
		logger.Logger.Infof("Starting connection %s", uc.ID)

		// 用于内部的读写循环退出或者握手退出 区分外界取消退出跟内部异常退出
		// 加锁保证内部在彻底退出前不会尝试新的启动
		uc.mu.Lock()
		uc.connectionCtx, uc.connectionCancel = context.WithCancel(context.Background())
		uc.mu.Unlock()

		// 启动(开始监听请求并处理请求读写)
		go func(ctx context.Context) {
			if uc.conn == nil || uc.isClosed() {
				logger.Logger.Infof("Connection %s needs reconnection", uc.ID)
				// 通知一下上层 该重连了
				uc.connectionCancel()
				return
			}

			// 记录连接建立信息
			logger.Logger.Infof("Connection %s established successfully, starting read/write loops", uc.ID)
			if uc.conn != nil {
				logger.Logger.Infof("Connection %s details - LocalAddr: %v, RemoteAddr: %v",
					uc.ID, uc.conn.LocalAddr(), uc.conn.RemoteAddr())
			}

			// 启动读写循环(阻塞)
			uc.startReadWriteLoops(ctx)
		}(ctx)

		// 启动(开始处理握手动作)
		// 服务端say hello 客户端处理say hello 客户端发送注册请求 服务端处理注册请求(受信) 服务端返回注册响应 客户端处理注册响应
		go func(ctx context.Context) {
			uc.handleConnect(ctx)
		}(ctx)

		// 统一监听退出信号和重连信号
		go func() {
			select {
			case <-ctx.Done():
				// 外部退出 不重连
				logger.Logger.Infof("Connection %s external shutdown, closing without reconnect", uc.ID)
				uc.Close(DisconnectReasonExternalClose)

			case <-uc.connectionCtx.Done():
				// 内部异常 需要重连
				logger.Logger.Infof("Connection %s internal error detected, closing and triggering reconnect", uc.ID)
				// 默认使用网络错误作为断开原因，具体原因应该在触发 connectionCancel 的地方设置
				uc.Close(DisconnectReasonNetworkError)
			}
		}()
	})
}

// startReadWriteLoops 启动读写循环
func (uc *Connection) startReadWriteLoops(ctx context.Context) {
	// 启动读循环
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Errorf("Read loop panic for connection %s: %v", uc.ID, r)
				// panic 退出，发送内部异常信号
				uc.connectionCancel()
			}
			logger.Logger.Infof("Read loop exited for connection %s", uc.ID)
		}()

		uc.readLoop(ctx)
	}()

	// 启动写循环
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Errorf("Write loop panic for connection %s: %v", uc.ID, r)
				// panic 退出，发送内部异常信号
				uc.connectionCancel()
			}
			logger.Logger.Infof("Write loop exited for connection %s", uc.ID)
		}()

		uc.writeLoop(ctx)
	}()
}

// isClosed 检查连接是否已关闭
func (uc *Connection) isClosed() bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return !uc.isActive
}

// triggerStateChange 触发连接状态变更（客户端和服务端通用）
func (uc *Connection) triggerStateChange(event ConnectionEvent, reason DisconnectReason, err error) {
	change := &ConnectionStateChange{
		Event:            event,
		Conn:             uc,
		DisconnectReason: reason,
		Error:            err,
	}

	logger.Logger.Infof("Triggering state change for connection %s (role=%v): event=%v, reason=%v",
		uc.ID, uc.role, event, reason)

	// 根据角色调用对应的回调
	switch uc.role {
	case RoleClient:
		if uc.clientData != nil && uc.clientData.stateChangeCallback != nil {
			uc.clientData.stateChangeCallback(change)
		}
	case RoleProxy:
		if uc.proxyData != nil && uc.proxyData.stateChangeCallback != nil {
			uc.proxyData.stateChangeCallback(change)
		}
	}
}

// SetStateChangeCallback 设置连接状态变更回调
func (uc *Connection) SetStateChangeCallback(callback func(change *ConnectionStateChange)) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.role == RoleClient && uc.clientData != nil {
		uc.clientData.stateChangeCallback = callback
	}

	if uc.role == RoleProxy && uc.proxyData != nil {
		uc.proxyData.stateChangeCallback = callback
	}
}

// Close 关闭连接
func (uc *Connection) Close(reason DisconnectReason) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if !uc.isActive {
		return nil
	}

	uc.isActive = false

	close(uc.sendCh)

	if uc.conn != nil {
		uc.conn.Close()
	}

	// 连接关闭清空pending请求，避免pending携程泄漏
	uc.requestMu.Lock()
	pendingCount := len(uc.pendingRequests)
	uc.pendingRequests = make(map[string]*PendingRequest)
	uc.requestMu.Unlock()

	if pendingCount > 0 {
		logger.Logger.Infof("Cleared %d pending requests for connection %s", pendingCount, uc.ID)
	}

	// 触发状态变更事件
	uc.triggerStateChange(ConnectionEventDisconnected, reason, nil)

	// 处理断开回调
	uc.handleDisconnect(nil)

	return nil
}

// Send 发送消息
func (uc *Connection) Send(data []byte) error {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	if !uc.isActive {
		return errors.New("connection is not active")
	}

	select {
	case uc.sendCh <- data:
		return nil
	default:
		return errors.New("send channel is full")
	}
}

// SendJSON 发送JSON消息
func (uc *Connection) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return uc.Send(data)
}

// IsActive 检查连接是否活跃
func (uc *Connection) IsActive() bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.isActive
}

// GetRole 获取连接角色
func (uc *Connection) GetRole() ConnectionRole {
	return uc.role
}

// GetClientData 获取客户端数据
func (uc *Connection) GetClientData() *ClientConnectionData {
	if uc.role == RoleClient {
		return uc.clientData
	}
	return nil
}

// GetProxyData 获取代理数据
func (uc *Connection) GetProxyData() *ProxyConnectionData {
	if uc.role == RoleProxy {
		return uc.proxyData
	}
	return nil
}

// UpdateLastPing 更新最后心跳时间
func (uc *Connection) UpdateLastPing() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.lastPing = time.Now()
}

// UpdateLastPong 更新最后pong时间
func (uc *Connection) UpdateLastPong() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.lastPong = time.Now()
}

// readLoop 读取循环
func (uc *Connection) readLoop(ctx context.Context) {
	logger.Logger.Infof("Starting read loop for connection %s", uc.ID)

	// 设置读取限制
	if uc.conn != nil {
		uc.conn.SetReadLimit(512 * 1024) // 512KB
	}

	// 设置pong处理器
	if uc.conn != nil {
		uc.conn.SetPongHandler(func(appData string) error {
			uc.UpdateLastPong()
			// 重置读超时时间
			if err := uc.conn.SetReadDeadline(time.Now().Add(uc.config.GetReadTimeout())); err != nil {
				return err
			}
			logger.Logger.Info("Received pong from peer")
			return nil
		})
	}

	// 设置初始读超时
	if uc.conn != nil {
		if err := uc.conn.SetReadDeadline(time.Now().Add(uc.config.GetReadTimeout())); err != nil {
			logger.Logger.Errorf("Failed to set initial read deadline for connection %s: %v", uc.ID, err)
			// 发送内部异常信号
			uc.connectionCancel()
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			// 外界退出
			logger.Logger.Infof("Read loop external context cancelled for connection %s", uc.ID)
			return
		case <-uc.connectionCtx.Done():
			// 内部退出
			logger.Logger.Infof("Read loop internal context cancelled for connection %s", uc.ID)
			return
		default:
			if uc.conn == nil {
				logger.Logger.Errorf("Connection is nil for %s", uc.ID)
				// 发送内部异常信号
				uc.connectionCancel()
				return
			}

			_, message, err := uc.conn.ReadMessage()
			if err != nil {
				if uc.shouldIgnoreReadError(err) {
					logger.Logger.Infof("Ignoring read error for connection %s: %v", uc.ID, err)
					continue
				}

				logger.Logger.Warnf("Read error for connection %s: %v", uc.ID, err)
				// 发送内部异常信号
				uc.connectionCancel()
				return // 严重错误，退出读循环，触发重连
			}

			// 更新最后活跃时间
			uc.UpdateLastPing()

			// 处理消息
			if err = uc.handleMessage(ctx, message); err != nil {
				logger.Logger.Errorf("Message handling error for connection %s: %v", uc.ID, err)
				// 消息处理错误不应该断开连接，继续处理下一条消息
			}
		}
	}
}

// shouldIgnoreReadError 判断读取错误是否可以忽略(先不判断了)
func (uc *Connection) shouldIgnoreReadError(err error) bool {
	if err == nil {
		return false
	}

	logger.Logger.Infof("Connection %s read error, triggering reconnection: [===%v===]", uc.ID, err)
	return false
}

// writeLoop 写入循环
func (uc *Connection) writeLoop(ctx context.Context) {
	logger.Logger.Infof("Starting write loop for connection %s", uc.ID)

	heartbeat := time.NewTicker(uc.config.GetHeartbeatInterval())
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			// 外界退出
			logger.Logger.Infof("Write loop external context cancelled for connection %s", uc.ID)
			return
		case <-uc.connectionCtx.Done():
			// 内部退出
			logger.Logger.Infof("Write loop internal context cancelled for connection %s", uc.ID)
			return
		case message, ok := <-uc.sendCh:
			if uc.conn == nil {
				logger.Logger.Errorf("Connection is nil for %s", uc.ID)
				// 发送内部异常信号
				uc.connectionCancel()
				return
			}

			if err := uc.conn.SetWriteDeadline(time.Now().Add(uc.config.GetWriteTimeout())); err != nil {
				logger.Logger.Errorf("Failed to set write deadline for connection %s: %v", uc.ID, err)
				// 发送内部异常信号
				uc.connectionCancel()
				return
			}

			if !ok {
				// 通道已关闭，发送关闭消息
				if err := uc.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Logger.Warnf("Failed to send close message for connection %s: %v", uc.ID, err)
				}
				// 发送内部异常信号
				uc.connectionCancel()
				return
			}

			if err := uc.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				if uc.shouldIgnoreWriteError(err) {
					logger.Logger.Infof("Ignoring write error for connection %s: %v", uc.ID, err)
					continue
				}
				logger.Logger.Warnf("Write error for connection %s: %v", uc.ID, err)
				// 发送内部异常信号
				uc.connectionCancel()
				return // 严重错误，退出写循环，触发重连
			}

		case <-heartbeat.C:
			if uc.conn == nil {
				logger.Logger.Errorf("Connection is nil for %s during heartbeat", uc.ID)
				// 发送内部异常信号
				uc.connectionCancel()
				return
			}

			// 检查心跳超时
			uc.mu.RLock()
			lastPong := uc.lastPong
			uc.mu.RUnlock()

			// 如果超过两个心跳周期没有收到pong，认为连接异常
			if !lastPong.IsZero() && time.Since(lastPong) > 2*uc.config.GetHeartbeatInterval() {
				logger.Logger.Warnf("Heartbeat timeout for connection %s, connection may be lost", uc.ID)
				// 发送内部异常信号
				uc.connectionCancel()
				return // 心跳超时，触发重连
			}

			// 发送ping
			if err := uc.conn.SetWriteDeadline(time.Now().Add(uc.config.GetWriteTimeout())); err != nil {
				logger.Logger.Errorf("Failed to set write deadline for heartbeat %s: %v", uc.ID, err)
				// 发送内部异常信号
				uc.connectionCancel()
				return
			}

			if err := uc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if uc.shouldIgnoreWriteError(err) {
					logger.Logger.Infof("Ignoring heartbeat write error for connection %s: %v", uc.ID, err)
					continue
				}
				logger.Logger.Warnf("Failed to send heartbeat (ping) for connection %s: %v", uc.ID, err)
				// 发送内部异常信号
				uc.connectionCancel()
				return // 心跳发送失败，触发重连
			}

			// 更新发送ping时间
			uc.UpdateLastPing()
		}
	}
}

// shouldIgnoreWriteError 判断写入错误是否可以忽略(先不判断了)
func (uc *Connection) shouldIgnoreWriteError(err error) bool {
	if err == nil {
		return false
	}

	logger.Logger.Infof("Connection %s write error, triggering reconnection: %v", uc.ID, err)
	return false
}

// GenerateRequestID 生成唯一的请求ID
func (uc *Connection) GenerateRequestID() string {
	var requestIDPrefix string

	if uc.role == RoleClient {
		// 客户端：使用ID
		requestIDPrefix = fmt.Sprintf("client-%s", uc.ID)
	} else if uc.role == RoleProxy {
		// 服务端：使用连接ID作为前缀
		requestIDPrefix = fmt.Sprintf("server-%s", uc.ID)
	} else {
		requestIDPrefix = "unknown"
	}

	return fmt.Sprintf("%s-%d", requestIDPrefix, time.Now().UnixNano())
}

// buildRequestMessage 构建请求消息
func (uc *Connection) buildRequestMessage(msgType MessageType, reqMsg interface{}) *RequestMessage {
	if uc.role != RoleClient || uc.clientData == nil {
		return nil
	}

	uc.mu.RLock()
	clientID := uc.ID
	// 获取客户端配置中的版本信息
	clientVersion := "1.0.0" // 默认协议版本
	if uc.config != nil && uc.config.GetVersion() != "" {
		clientVersion = uc.config.GetVersion()
	}
	uc.mu.RUnlock()

	requestID := uc.GenerateRequestID()

	return &RequestMessage{
		Type:      msgType,
		RequestID: requestID,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
		Version:   clientVersion,
		Data:      reqMsg,
	}
}

// SendRequest 统一的请求发送和重试机制 不带重试 自主构建消息体
func (uc *Connection) SendRequest(msgType MessageType, reqMsg interface{}) error {
	// 根据角色构建不同的消息
	var message interface{}
	var requestID string

	if uc.role == RoleClient {
		if uc.clientData == nil {
			return errors.New("client data not available")
		}
		// 客户端构建RequestMessage
		requestMsg := uc.buildRequestMessage(msgType, reqMsg)
		if requestMsg == nil {
			return errors.New("failed to build request message")
		}
		message = requestMsg
		requestID = requestMsg.RequestID
	} else if uc.role == RoleProxy {
		// 服务端构建ResponseMessage（主动推送）
		requestID = uc.GenerateRequestID()
		responseMsg := ResponseMessage{
			Type:          msgType,
			RequestID:     requestID,
			ServerVersion: uc.serverInfo.Version,
			Timestamp:     time.Now().Unix(),
			ClientID:      uc.ID,
			Data:          reqMsg,
		}
		message = responseMsg
	} else {
		return errors.New("unsupported connection role for SendRequestWithRetry")
	}

	// 发送消息
	return uc.SendJSON(message)
}

// SendRequestWithRetry 统一的请求发送和重试机制
func (uc *Connection) SendRequestWithRetry(msgType MessageType, reqMsg interface{}, timeout time.Duration, maxRetries int, retryDelay time.Duration) error {
	// 根据角色构建不同的消息
	var message interface{}
	var requestID string

	if uc.role == RoleClient {
		if uc.clientData == nil {
			return errors.New("client data not available")
		}
		// 客户端构建RequestMessage
		requestMsg := uc.buildRequestMessage(msgType, reqMsg)
		if requestMsg == nil {
			return errors.New("failed to build request message")
		}
		message = requestMsg
		requestID = requestMsg.RequestID
	} else if uc.role == RoleProxy {
		// 服务端构建ResponseMessage（主动推送）
		clientInfo := uc.GetClientInfo()
		if clientInfo == nil {
			return errors.New("failed to get client info")
		}
		requestID = uc.GenerateRequestID()
		responseMsg := ResponseMessage{
			Type:          msgType,
			RequestID:     requestID,
			ServerVersion: uc.serverInfo.Version,
			Timestamp:     time.Now().Unix(),
			ClientID:      uc.ID,
			ServiceName:   clientInfo.ServiceName,
			NamespaceTag:  clientInfo.NamespaceTag,
			EnvTag:        clientInfo.EnvTag,
			Data:          reqMsg,
		}
		message = responseMsg
	} else {
		return errors.New("unsupported connection role for SendRequestWithRetry")
	}

	// 添加到pending列表
	pendingReq := &PendingRequest{
		RequestID:    requestID,
		MessageType:  msgType,
		SendTime:     time.Now(),
		Timeout:      timeout,
		MaxRetries:   maxRetries,
		CurrentRetry: 0,
		RetryDelay:   retryDelay,
		OriginalMsg:  message, // 用于重新发送
	}

	// 添加到pending列表
	uc.addPendingRequest(requestID, pendingReq)

	// 启动超时和重试处理(pending列表)
	go uc.handleRequestTimeout(requestID)

	// 发送消息
	return uc.SendJSON(message)
}

// handleRequestTimeout 统一的请求超时和重试逻辑
func (uc *Connection) handleRequestTimeout(requestID string) {

	for {
		req, exists := uc.getPendingRequest(requestID)
		if !exists {
			return // pending消息被处理了
		}

		// 等待超时
		time.Sleep(req.Timeout)

		req, exists = uc.getPendingRequest(requestID)
		if !exists {
			return // pending消息被处理了
		}

		// 检查是否还有重试次数
		if req.CurrentRetry >= req.MaxRetries {
			uc.removePendingRequest(requestID)
			logger.Logger.Errorf("Request %s failed after %d retries, giving up", requestID, req.MaxRetries)
			return
		}

		// 检查连接状态
		if !uc.IsActive() {
			uc.removePendingRequest(requestID)
			logger.Logger.Warnf("Request %s stopped retrying due to connection lost", requestID)
			return
		}

		// 增加重试次数
		req.CurrentRetry++
		uc.addPendingRequest(requestID, req)

		logger.Logger.Warnf("Request %s timeout, retrying (%d/%d)", requestID, req.CurrentRetry, req.MaxRetries)

		// 重新发送消息
		if err := uc.SendJSON(req.OriginalMsg); err != nil {
			uc.removePendingRequest(requestID)
			logger.Logger.Errorf("Request %s send error on retry: %v", requestID, err)
			return
		}

		// 等待重试间隔
		time.Sleep(req.RetryDelay)
	}
}

// handlePendingResponse 处理pending请求的响应
func (uc *Connection) handlePendingResponse(requestID string, response interface{}) {
	uc.requestMu.Lock()
	req, exists := uc.pendingRequests[requestID]
	if exists {
		delete(uc.pendingRequests, requestID)
		uc.requestMu.Unlock()
		logger.Logger.Infof("Request %s completed successfully (retry: %d/%d) received: %v", requestID, req.CurrentRetry, req.MaxRetries, response)
	} else {
		uc.requestMu.Unlock()
		logger.Logger.Infof("Received response for unknown request: %s", requestID)
	}
}

func (uc *Connection) addPendingRequest(requestID string, req *PendingRequest) {
	uc.requestMu.Lock()
	uc.pendingRequests[requestID] = req
	uc.requestMu.Unlock()
}

// removePendingRequest 移除待处理请求
func (uc *Connection) removePendingRequest(requestID string) {
	uc.requestMu.Lock()
	delete(uc.pendingRequests, requestID)
	uc.requestMu.Unlock()
}

// getPendingRequest 获取待处理请求
func (uc *Connection) getPendingRequest(requestID string) (*PendingRequest, bool) {
	uc.requestMu.RLock()
	req, exists := uc.pendingRequests[requestID]
	uc.requestMu.RUnlock()
	return req, exists
}

// SetTrusted 设置受信状态
func (uc *Connection) SetTrusted(trusted bool) {
	if uc.role == RoleProxy && uc.proxyData != nil {
		uc.proxyData.isTrusted = trusted
	}
}

// IsTrusted 检查是否受信
func (uc *Connection) IsTrusted() bool {
	if uc.role == RoleProxy && uc.proxyData != nil {
		return uc.proxyData.isTrusted
	}
	return false
}

// SetClientInfo 设置客户端信息
func (uc *Connection) SetClientInfo(info *ClientProxyInfo) {
	if uc.role == RoleProxy && uc.proxyData != nil {
		uc.proxyData.Info = info
	}
}

// GetClientInfo 获取客户端信息
func (uc *Connection) GetClientInfo() *ClientProxyInfo {
	if uc.role == RoleProxy && uc.proxyData != nil {
		return uc.proxyData.Info
	}
	return nil
}

// GetServerVersion 获取服务端版本
func (uc *Connection) GetServerVersion() string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.serverInfo.Version
}

// SetServerVersion 设置服务端版本
func (uc *Connection) SetServerVersion(version string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.serverInfo.Version = version
}

// GetServerInfo 获取服务端信息
func (uc *Connection) GetServerInfo() *ServerInfo {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.serverInfo
}

// SetServerInfo 设置服务端信息
func (uc *Connection) SetServerInfo(info *ServerInfo) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.serverInfo = info
}

// handleConnect 处理连接建立
func (uc *Connection) handleConnect(ctx context.Context) {
	logger.Logger.Infof("Connection %s Start executing connection callback", uc.ID)

	// 连接建立回调
	if onConnect := uc.process.OnConnect(ctx, uc); onConnect != nil {
		onConnect(uc)
	}
}

// handleDisconnect 处理连接断开
func (uc *Connection) handleDisconnect(err error) {
	logger.Logger.Infof("Connection %s permanently disconnected: %v", uc.ID, err)

	// 连接断开回调
	uc.process.OnDisconnect(uc, err)
}

// handleMessage 消息的处理逻辑，区分客户端跟服务端的客户端代理
func (uc *Connection) handleMessage(ctx context.Context, message []byte) error {
	receivedStr := ""
	switch uc.GetRole() {
	case RoleProxy:
		receivedStr = "[SERVER]"
		break
	case RoleClient:
		receivedStr = "[CLIENT]"
		break
	default:
		receivedStr = "unknown role"
	}

	logger.Logger.Infof("Connection %s "+receivedStr+"received message: %s", uc.ID, string(message))

	// 先解析消息类型来区分ACK和业务响应
	var msgType struct {
		Type      string `json:"type,omitempty"`       // 业务响应独有字段
		Status    string `json:"status,omitempty"`     // ReceiveAck响应独有字段
		RequestID string `json:"request_id,omitempty"` // 两者都有的字段
	}

	if err := json.Unmarshal(message, &msgType); err != nil {
		return fmt.Errorf("failed to parse message structure: %v", err)
	}

	// 通过status字段存在且type字段为空来判断是ReceiveAck
	if msgType.Status != "" && msgType.Type == "" && msgType.RequestID != "" {
		var receiveAck ReceiveAck
		if err := json.Unmarshal(message, &receiveAck); err != nil {
			return fmt.Errorf("failed to parse receive ack: %v", err)
		}
		uc.handleReceiveAck(&receiveAck)
		return nil
	}

	// 否则当作业务响应消息处理
	return uc.process.ProcessMessage(ctx, uc, message)
}

// handleReceiveAck 统一处理ReceiveAck消息
func (uc *Connection) handleReceiveAck(ack *ReceiveAck) {
	logger.Logger.Infof("Connection %s received ack for request %s: %s",
		uc.ID, ack.RequestID, ack.Message)

	// 删除pending请求
	uc.handlePendingResponse(ack.RequestID, ack)
}
