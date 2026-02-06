package pc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// ProxyMessageProcessor 服务端代理消息处理器
type ProxyMessageProcessor struct {
	config *ServerConfig
	server *Server
}

// NewProxyMessageProcessor 创建服务端代理消息处理器
func NewProxyMessageProcessor(server *Server) *ProxyMessageProcessor {
	return &ProxyMessageProcessor{
		server: server,
		config: server.config,
	}
}

// ProcessMessage 处理消息（实现分层确认机制）
func (p *ProxyMessageProcessor) ProcessMessage(ctx context.Context, conn *Connection, message []byte) error {
	// 然后处理RequestMessage（客户端发送的请求）
	var reqMsg RequestMessage
	if err := json.Unmarshal(message, &reqMsg); err != nil {
		logger.Logger.Errorf("Failed to unmarshal request message from client %s: %v, original message: %s", conn.ID, err, string(message))
		return err
	}

	// 获取消息的类别 不同类别决定不同的处理逻辑 比如是否走一阶段确认
	security := GetMessageSecurity(reqMsg.Type)

	// 统一的分层确认处理
	switch security {
	case MessageTypeSpecial:
		// 注册消息特殊处理，不发送一阶段确认
		return p.processRegisterMessage(conn, &reqMsg)
	case MessageTypeTrusted:
		// 受信消息：先发送一阶段确认，检查受信状态，然后异步处理
		if err := p.processTrustedMessage(conn, &reqMsg); err != nil {
			logger.Logger.Errorf("Failed to process MessageTypeTrusted message: %v", err)
			return err
		}
		// 处理接收到的消息回调
		if p.config.MessageHandler != nil {
			p.config.MessageHandler(conn, message)
		} else {
			logger.Logger.Warnf("Received message but no handler is set: %s", string(message))
		}
		return nil
	default:
		// 公开消息：先发送一阶段确认，然后异步处理
		if err := p.processPublicMessage(conn, &reqMsg); err != nil {
			logger.Logger.Errorf("Failed to process MessageTypePublic message: %v", err)
		}
		// 处理接收到的消息回调
		if p.config.MessageHandler != nil {
			p.config.MessageHandler(conn, message)
		} else {
			logger.Logger.Warnf("Received message but no handler is set: %s", string(message))
		}
		return nil
	}
}

func (p *ProxyMessageProcessor) OnConnect(ctx context.Context, conn *Connection) ConnectHandler {
	logger.Logger.Infof("Proxy connection %s established", conn.ID)

	// 服务端收到ws连接建立后主动发送hello消息
	p.sendConnectHello(conn)

	return func(c *Connection) {
		p.config.OnConnect(c)
	}
}

func (p *ProxyMessageProcessor) OnDisconnect(conn *Connection, err error) {
	logger.Logger.Infof("Proxy connection %s disconnected: %v", conn.ID, err)
	if p.config.MessageHandler != nil {
		p.config.OnDisconnect(conn, err)
	}
}

// processRegisterMessage 处理注册消息（不发送一阶段确认）
func (p *ProxyMessageProcessor) processRegisterMessage(conn *Connection, reqMsg *RequestMessage) error {
	// 解析注册数据
	var registerPayload RegisterPayload
	if payloadBytes, err := json.Marshal(reqMsg.Data); err == nil {
		if err = json.Unmarshal(payloadBytes, &registerPayload); err != nil {
			logger.Logger.Errorf("Failed to unmarshal register payload from client %s: %v", conn.ID, err)
			return err
		}

		// 更新客户端信息
		proxyData := conn.GetProxyData()
		if proxyData != nil && proxyData.Info != nil {
			proxyData.Info.ServiceName = registerPayload.ServiceName
			proxyData.Info.InternalIP = registerPayload.InternalIPs
			proxyData.Info.PublicIP = registerPayload.PublicIPs
			proxyData.Info.NamespaceTag = registerPayload.NamespaceTag
			proxyData.Info.EnvTag = registerPayload.EnvTag
		}
	}

	// 信任客户端
	var registerSuccess = true
	var errorMsg string

	if err := p.server.TrustedClient(conn.ID); err != nil {
		logger.Logger.Warnf("Failed to trust client %s: %v", conn.ID, err)
		registerSuccess = false
		errorMsg = err.Error()
	} else {
		// 设置连接的受信状态
		conn.SetTrusted(true)

		// 执行受信后的回调
		if p.server.config.OnClientTrusted != nil {
			clientInfo := conn.GetClientInfo()
			if clientInfo != nil {
				logger.Logger.Debugf("Executing OnClientTrusted callback for client %s", conn.ID)
				p.server.config.OnClientTrusted(clientInfo)
			}
		}
	}

	// 发送注册响应（不走受信列表）
	response := ResponseMessage{
		Type:          RegisterSignal,
		RequestID:     reqMsg.RequestID,
		ServerVersion: p.server.config.ServerVersion,
		Timestamp:     time.Now().Unix(),
		ClientID:      conn.ID,
		ServiceName:   registerPayload.ServiceName,
		NamespaceTag:  registerPayload.NamespaceTag,
		EnvTag:        registerPayload.EnvTag,
		Data: map[string]interface{}{
			"success": registerSuccess,
			"message": func() string {
				if registerSuccess {
					return "Registration successful"
				}
				return errorMsg
			}(),
		},
	}

	// 直接发送，不检查受信状态
	if err := conn.SendJSON(response); err != nil {
		logger.Logger.Errorf("Failed to send register response: %v", err)
	}

	return nil
}

// processTrustedMessage 处理受信消息（统一分层确认）
func (p *ProxyMessageProcessor) processTrustedMessage(conn *Connection, reqMsg *RequestMessage) error {
	// 一阶段立即发送接收确认
	receiveAck := ReceiveAck{
		RequestID: reqMsg.RequestID,
		Status:    "received",
		Message:   "Message received successfully server send ack to client",
		Timestamp: time.Now().Unix(),
	}
	if err := conn.SendJSON(receiveAck); err != nil {
		logger.Logger.Warnf("Failed to send receive ack: %v", err)
		return err
	}

	// 检查是否为受信客户端
	if !conn.IsTrusted() {
		// 直接用 ResponseMessage 返回错误
		errorResponse := ResponseMessage{
			Type:          reqMsg.Type,
			RequestID:     reqMsg.RequestID,
			ServerVersion: p.server.config.ServerVersion,
			Timestamp:     time.Now().Unix(),
			ClientID:      conn.ID,
			Data:          map[string]interface{}{"error": "Client not trusted for this operation"},
		}
		if err := conn.SendJSON(errorResponse); err != nil {
			logger.Logger.Warnf("Failed to send error response: %v", err)
			return err
		}
		return errors.New("client not trusted") // 返回错误，阻止后续MessageHandler执行
	}

	// 受信客户端：返回nil，让后续的MessageHandler处理具体业务逻辑
	// 二阶段的处理完成确认应该在MessageHandler中发送
	return nil
}

// processPublicMessage 处理公开消息
func (p *ProxyMessageProcessor) processPublicMessage(conn *Connection, reqMsg *RequestMessage) error {
	// 一阶段立即发送接收确认
	receiveAck := ReceiveAck{
		RequestID: reqMsg.RequestID,
		Status:    "received",
		Message:   "Message received successfully server send ack to client",
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

// sendConnectHello 发送连接欢迎消息
func (p *ProxyMessageProcessor) sendConnectHello(conn *Connection) {
	now := time.Now()

	// 获取当前连接数
	currentCount := p.server.GetClientCount()
	maxClients := p.server.config.MaxConnections

	if err := conn.SendRequest(ChangeTypeConnectHello, &ConnectHelloPayload{
		ServerInfo: "WebSocket Persistent Connection Server",
		SupportTypes: []MessageType{
			SwitchFull,
			DriverConfigFull,
			DriverConfigChange,
			RegisterSignal,
			ChangeTypeConnectHello,
		},
		ServerTime:   now.Unix(),
		MaxClients:   maxClients,
		CurrentCount: currentCount,
		Metadata: map[string]string{
			"server_address": p.server.config.Address,
			"protocol":       "websocket",
			"version":        p.server.config.ServerVersion,
		},
	}); err != nil {
		logger.Logger.Errorf("Failed to send hello message to client %s: %v", conn.ID, err)
	} else {
		logger.Logger.Infof("Sent hello message to client %s", conn.ID)
	}
}
