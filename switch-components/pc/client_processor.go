package pc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// ClientMessageProcessor 客户端消息处理器
type ClientMessageProcessor struct {
	config *PersistentConnectionConfig // 调用回调函数
}

// NewClientMessageProcessor 创建客户端消息处理器
func NewClientMessageProcessor(config *PersistentConnectionConfig) *ClientMessageProcessor {
	return &ClientMessageProcessor{
		config: config,
	}
}

// ProcessMessage 处理消息
func (p *ClientMessageProcessor) ProcessMessage(ctx context.Context, conn *Connection, message []byte) error {
	// 检查是否为服务端响应消息
	var responseMsg ResponseMessage

	err := json.Unmarshal(message, &responseMsg)
	if err != nil {
		return err
	}

	// 获取消息的类别 不同类别决定不同的处理逻辑 比如是否走一阶段确认
	security := GetMessageSecurity(responseMsg.Type)

	// 统一的分层确认处理
	switch security {
	case MessageTypeSpecial:
		// 注册消息特殊处理，不发送一阶段确认
		switch responseMsg.Type {
		case ChangeTypeConnectHello:
			// 客户端跟服务端的连接建立后的say_hello消息
			// 处理Hello消息
			if helloPayload, ok := responseMsg.Data.(map[string]interface{}); ok {
				p.handleHelloMessage(conn, responseMsg.ServerVersion, helloPayload)
			}
			return nil
		case RegisterSignal:
			// 客户端向服务端发送注册消息后的服务端的受信响应(即注册响应)
			// 处理注册响应消息
			p.handleRegisterResponse(conn, &responseMsg)
			return nil
		default:
			logger.Logger.Warnf("Encountered a message that does not require an ack: %v", responseMsg)
			return nil
		}
	case MessageTypePublic:
	default:
		// 公开消息：先发送一阶段确认，然后异步处理
		if err := p.processPublicMessage(conn, &responseMsg); err != nil {
			logger.Logger.Errorf("Failed to process MessageTypePublic message: %v", err)
		}
		// 处理接收到的消息回调
		if p.config.MessageHandler != nil {
			p.config.MessageHandler(ctx, message)
		} else {
			logger.Logger.Warnf("Received message but no handler is set: %s", string(message))
		}
		return nil
	}
	return nil
}

// processPublicMessage 处理公开消息
func (p *ClientMessageProcessor) processPublicMessage(conn *Connection, respMsg *ResponseMessage) error {
	// 一阶段立即发送接收确认
	receiveAck := ReceiveAck{
		RequestID: respMsg.RequestID,
		Status:    "received",
		Message:   "Message received successfully client send ack to server",
		Timestamp: time.Now().Unix(),
	}
	if err := conn.SendJSON(receiveAck); err != nil {
		logger.Logger.Warnf("Failed to send receive ack: %v", err)
		return err
	}

	// 公开消息：返回nil，让后续的MessageHandler处理具体业务逻辑
	// 二阶段的处理完成确认应该在MessageHandler中发送
	return nil
}

func (p *ClientMessageProcessor) OnConnect(ctx context.Context, conn *Connection) ConnectHandler {
	timeout := time.NewTicker(30 * time.Second)
	defer timeout.Stop()

	select {
	case _, ok := <-conn.sendRegister:
		if !ok {
			logger.Logger.Errorf("Registration signal channel closed for connection %s", conn.ID)
			// 发送内部异常信号
			conn.connectionCancel()
			return nil
		}
		p.sendServiceRegister(conn)
	case <-timeout.C:
		// 握手超时，发送内部异常信号
		logger.Logger.Errorf("Handshake timeout for connection %s", conn.ID)
		conn.connectionCancel()
		return nil
	case <-ctx.Done():
		// 外部退出，不重连
		logger.Logger.Infof("OnConnect cancelled by external context for connection %s", conn.ID)
		return nil
	}
	return p.config.OnConnect
}

func (p *ClientMessageProcessor) OnDisconnect(conn *Connection, err error) {
	if p.config.OnDisconnect != nil {
		if err != nil {
			logger.Logger.Warnf("Client connection %s disconnected with error: %v", conn.ID, err)
		} else {
			logger.Logger.Infof("Client connection %s disconnected gracefully", conn.ID)
		}
		p.config.OnDisconnect(conn, err)
		return
	}

	if err != nil {
		logger.Logger.Warnf("Client connection %s disconnected with error (no handler): %v", conn.ID, err)
	} else {
		logger.Logger.Infof("Client connection %s disconnected gracefully (no handler)", conn.ID)
	}
}

// sendServiceRegister 发送服务注册信息
func (p *ClientMessageProcessor) sendServiceRegister(conn *Connection) {
	if p.config.OnRegister == nil {
		return
	}
	// 获取注册消息
	registerMsg := p.config.OnRegister(conn)

	// 使用统一的注册消息发送 注册消息不属于一阶段确认消息 没有ack 不走pending
	if err := conn.SendRequest(RegisterSignal, registerMsg); err != nil {
		logger.Logger.Errorf("Service register message sent fail: %s error: %s", registerMsg.ServiceName, err)
	} else {
		logger.Logger.Infof("Service register message sent: %s", registerMsg.ServiceName)
	}
}

// handleHelloMessage 处理服务端的hello消息
func (p *ClientMessageProcessor) handleHelloMessage(conn *Connection, serverVersion string, helloPayload map[string]interface{}) {
	payloadBytes, err := json.Marshal(helloPayload)
	if err != nil {
		logger.Logger.Errorf("Failed to marshal hello payload: %v", err)
		return
	}

	var helloData ConnectHelloPayload
	if err = json.Unmarshal(payloadBytes, &helloData); err != nil {
		logger.Logger.Errorf("Failed to unmarshal hello payload: %v", err)
		return
	}

	conn.mu.Lock()

	// 设置ServerInfo
	serverInfo := &ServerInfo{
		Version:      serverVersion,
		Name:         helloData.ServerInfo,
		Description:  fmt.Sprintf("Max clients: %d, Current: %d", helloData.MaxClients, helloData.CurrentCount),
		Capabilities: p.convertSupportTypesToCapabilities(helloData.SupportTypes),
		Config:       helloData.Metadata,
		StartTime:    helloData.ServerTime,
	}
	conn.serverInfo = serverInfo
	conn.mu.Unlock()

	// 发送客户端注册信号 客户端收到信号执行注册逻辑
	// 使用 sync.Once 或非阻塞方式，防止重复关闭通道
	select {
	case conn.sendRegister <- struct{}{}:
		logger.Logger.Infof("Registration signal sent for connection %s", conn.ID)
	default:
		logger.Logger.Warnf("Registration signal channel is full or closed for connection %s", conn.ID)
	}
}

// handleRegisterResponse 处理注册响应消息
func (p *ClientMessageProcessor) handleRegisterResponse(conn *Connection, responseMsg *ResponseMessage) {
	if responseMsg.Data == nil {
		logger.Logger.Errorf("Client received message with no data: %s", responseMsg.Type)
		// 数据格式错误，触发重连
		conn.connectionCancel()
		return
	}

	dataMap, ok := responseMsg.Data.(map[string]interface{})
	if !ok {
		logger.Logger.Errorf("Client received message data is not map[string]interface{}: %s", responseMsg.Type)
		// 数据格式错误，触发重连
		conn.connectionCancel()
		return
	}

	success, exists := dataMap["success"]
	if !exists {
		logger.Logger.Errorf("Client received message data not has 'success': %s", responseMsg.Type)
		// 数据格式错误，触发重连
		conn.connectionCancel()
		return
	}

	if successBool, ok := success.(bool); ok && successBool {
		logger.Logger.Infof("Client %s registration successful", conn.ID)

		// 调用受信回调
		if p.config.OnTrusted != nil {
			p.config.OnTrusted(conn)
		}
	} else {
		message, _ := dataMap["message"].(string)
		logger.Logger.Errorf("Client %s registration failed: %s", conn.ID, message)
		// 注册失败，触发重连
		conn.connectionCancel()
	}
}

// convertSupportTypesToCapabilities 转换支持列表
func (p *ClientMessageProcessor) convertSupportTypesToCapabilities(supportTypes []MessageType) []string {
	capabilities := make([]string, len(supportTypes))
	for i, msgType := range supportTypes {
		capabilities[i] = string(msgType)
	}
	return capabilities
}
