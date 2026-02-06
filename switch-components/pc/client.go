package pc

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-components/snowflake"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"github.com/gorilla/websocket"
)

var generate *snowflake.BusinessGenerator

func init() {
	generator, err := snowflake.NewBusinessGenerator(snowflake.BusinessConfig{
		BusinessType: "pc",
		MachineID:    1,
		Prefix:       "CONNECTION_ID:",
	})
	if err != nil {
		panic(err)
	}
	generate = generator
}

// Client 基于统一连接的客户端
type Client struct {
	config    *PersistentConnectionConfig
	conn      *Connection
	cancel    context.CancelFunc
	mu        sync.Mutex
	isRunning bool
	isClosed  bool // 防止重复关闭

	// 连接状态变更通道
	stateChangeCh chan *ConnectionStateChange

	// 确保 Start 只执行一次
	startOnce sync.Once
}

// NewClient 创建统一客户端
func NewClient(config *PersistentConnectionConfig) *Client {
	return &Client{
		config:        config,
		stateChangeCh: make(chan *ConnectionStateChange, 1),
	}
}

// Start 启动客户端并开始连接
func (c *Client) Start(ctx context.Context) {
	c.startOnce.Do(func() {
		logger.Logger.Info("Starting client!!!")

		ctx, c.cancel = context.WithCancel(ctx)

		c.mu.Lock()
		c.isRunning = true
		c.mu.Unlock()

		go func(ctx context.Context) {
			c.run(ctx)
		}(ctx)
	})
}

// Close 停止客户端
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 防止重复关闭
	if c.isClosed {
		logger.Logger.Info("Client already closed")
		return nil
	}
	c.isClosed = true

	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		c.conn.Close(DisconnectReasonExternalClose)
		c.conn = nil
	}

	// 清理状态变更通道
	select {
	case <-c.stateChangeCh:
	default:
	}
	close(c.stateChangeCh)

	c.isRunning = false
	logger.Logger.Info("Unified WebSocket client stopped")
	return nil
}

// Send 发送数据
func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return errors.New("connection is not established")
	}

	return conn.Send(data)
}

// SendWithoutRetry 发送数据 不带重试 自主构建消息体
func (c *Client) SendWithoutRetry(msgType MessageType, reqMsg interface{}) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return errors.New("connection is not established")
	}

	return conn.SendRequest(msgType, reqMsg)
}

func (c *Client) SendRequest(msgType MessageType, reqMsg interface{}, timeout time.Duration) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return errors.New("client is not connected")
	}

	// 使用Connection的统一重试机制，默认参数：3次重试，2秒间隔
	return conn.SendRequestWithRetry(msgType, reqMsg, timeout, 3, 2*time.Second)
}

// SendRequestWithCustomRetry 发送请求并自定义重试参数
func (c *Client) SendRequestWithCustomRetry(msgType MessageType, reqMsg interface{}, timeout time.Duration, maxRetries int, retryDelay time.Duration) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return errors.New("client is not connected")
	}

	return conn.SendRequestWithRetry(msgType, reqMsg, timeout, maxRetries, retryDelay)
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && c.conn.IsActive()
}

// GetServerClientID 获取服务端分配的客户端ID
func (c *Client) GetServerClientID() string {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return ""
	}

	return conn.ID
}

// GetServerVersion 获取服务端版本
func (c *Client) GetServerVersion() string {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return ""
	}

	clientData := conn.GetClientData()
	if clientData == nil {
		return ""
	}

	return conn.GetServerVersion()
}

// GetServerInfo 获取服务端信息
func (c *Client) GetServerInfo() *ServerInfo {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil
	}

	return conn.GetServerInfo()
}

// run 核心循环，管理连接和重连
func (c *Client) run(ctx context.Context) {
	strategy := c.config.GetReconnectStrategy()
	state := &ReconnectState{}

	for {
		// 每次循环开始时检查外部退出，确保快速响应
		select {
		case <-ctx.Done():
			logger.Logger.Info("client run stopped context done %v", ctx.Err())
			return
		default:
			// 检查是否需要重置重连计数器
			if state.ShouldReset(strategy) {
				logger.Logger.Infof("Resetting reconnection counter after %v", strategy.ResetInterval)
				state.Reset()
			}

			// 尝试连接
			if err := c.connect(ctx); err != nil {
				// 连接失败，检查是否超过最大重试次数
				if strategy.MaxRetries > 0 && state.GetAttempts() >= strategy.MaxRetries {
					logger.Logger.Errorf("Connection failed after %d retries: %v", strategy.MaxRetries, err)
					return
				}

				// 计算下次重连延迟
				delay := state.CalculateNextDelay(strategy)
				state.IncrementAttempts()

				ticker := time.NewTicker(delay)

				logger.Logger.Warnf("Connection attempt %d failed: %v, retrying in %v", state.GetAttempts(), err, delay)

				// 等待重连延迟
				select {
				case <-ticker.C:
					continue
				case <-ctx.Done():
					logger.Logger.Info("client run stopped context done %v", ctx.Err())
					return
				}
			}

			logger.Logger.Infof("Successfully connected to %s", c.config.Address)

			// 连接成功，重置重连状态
			state.Reset()

			// 等待连接状态变更
			select {
			case stateChange, ok := <-c.stateChangeCh:
				if !ok {
					logger.Logger.Info("State change channel closed, client is shutting down")
					return
				}

				logger.Logger.Infof("Connection state changed: event=%v, reason=%v",
					stateChange.Event, stateChange.DisconnectReason)

				// 如果是断开事件，检查是否需要重连
				if stateChange.Event == ConnectionEventDisconnected {
					if !stateChange.DisconnectReason.ShouldReconnect() {
						logger.Logger.Infof("Disconnect reason [%s] indicates no reconnection needed",
							stateChange.DisconnectReason.String())
						return
					}
					logger.Logger.Warnf("Connection lost due to [%s]. Attempting to reconnect...",
						stateChange.DisconnectReason.String())
				}

			case <-ctx.Done():
				logger.Logger.Info("Context cancelled while waiting for disconnection")
				return
			}
		}
	}
}

// connect 建立WebSocket连接
func (c *Client) connect(ctx context.Context) error {
	dialer := websocket.Dialer{HandshakeTimeout: c.config.DialTimeout}

	// 生成客户端连接ID(不走服务端生成策略)
	clientId := generate.GeneratePrefixID()

	headers := make(http.Header)
	for k, v := range c.config.RequestHeader {
		headers[k] = v
	}
	headers.Set("ConnectionId", clientId)

	logger.Logger.Infof("Generated connection ID: %s for address: %s", clientId, c.config.Address)

	wsConn, _, err := dialer.DialContext(ctx, c.config.Address, headers)
	if err != nil {
		return err
	}

	// 创建消息处理器
	processor := NewClientMessageProcessor(c.config)

	conn := NewConnection(clientId, wsConn, RoleClient, c.config, processor)

	// 设置连接状态变更回调
	var stateChangeOnce sync.Once
	conn.SetStateChangeCallback(func(change *ConnectionStateChange) {
		stateChangeOnce.Do(func() {
			select {
			case c.stateChangeCh <- change:
				logger.Logger.Infof("Connection state change sent: event=%v, reason=%v",
					change.Event, change.DisconnectReason)
			default:
				logger.Logger.Warn("State change signal already pending")
			}
		})
	})

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	conn.Start(ctx)

	return nil
}
