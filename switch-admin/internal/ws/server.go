package ws

import (
	"context"
	"encoding/json"
	"time"

	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-admin/internal/utils"
	"gitee.com/fatzeng/switch-components/drivers"
	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

var wsServer *pc.Server

func GetWsServer() *pc.Server {
	return wsServer
}

// StartWebSocketServer 启动WebSocket服务器
func StartWebSocketServer(ctx context.Context) error {
	wsConfig := config.GlobalConfig.Pc
	if wsConfig == nil {
		// 默认配置
		wsConfig = pc.DefaultServerConfig()
	}
	// 连接回调
	wsConfig.OnConnect = onClientConnect
	// 断连回调
	wsConfig.OnDisconnect = onClientDisconnect
	// 消息处理回调
	wsConfig.MessageHandler = onMessage
	// 受信回调
	wsConfig.OnClientTrusted = onClientTrusted

	wsServer = pc.NewServer(wsConfig)

	// 注册配置变更业务端点
	wsServer.RegisterHandler(pc.WsEndpointChangeConfig, nil)
	// 注册全量配置同步业务端点
	wsServer.RegisterHandler(pc.WsEndpointFullSyncConfig, nil)
	// 注册全量开关同步业务端点
	wsServer.RegisterHandler(pc.WsEndpointFullSync, nil)

	return wsServer.Start(ctx)
}

// StopWebSocketServer 停止WebSocket服务器
func StopWebSocketServer() error {
	if wsServer == nil {
		return nil
	}
	logger.Logger.Infof("Stopping WebSocket server...")
	return wsServer.Stop()
}

// onClientTrusted 客户端受信回调
func onClientTrusted(clientInfo *pc.ClientProxyInfo) {
	if clientInfo == nil {
		logger.Logger.Warnf("Client trusted callback received nil client info")
		return
	}

	logger.Logger.Infof("Client trusted: %s (service: %s) from %s",
		clientInfo.ID, clientInfo.ServiceName, clientInfo.RemoteAddr)

	if len(clientInfo.PublicIP) > 0 {
		logger.Logger.Infof("Client %s public IPs: %v", clientInfo.ID, clientInfo.PublicIP)
	}
	if len(clientInfo.InternalIP) > 0 {
		logger.Logger.Infof("Client %s internal IPs: %v", clientInfo.ID, clientInfo.InternalIP)
	}

	drivers.AddClientIPsToPool(clientInfo)
	logger.Logger.Infof("Added client %s IPs to global pool", clientInfo.ID)
}

// onClientConnect 客户端连接回调函数
func onClientConnect(conn *pc.Connection) {
	info := conn.GetClientInfo()
	if info != nil {
		logger.Logger.Infof("Client connected: %s from %s", info.ID, info.RemoteAddr)
	} else {
		logger.Logger.Infof("Client connected: %s", conn.GetClientInfo().ID)
	}
}

// onClientDisconnect 客户端断开回调函数
func onClientDisconnect(conn *pc.Connection, err error) {
	clientInfo := conn.GetClientInfo()
	if clientInfo != nil {
		logger.Logger.Infof("Client disconnected: %s from %s, reason: %v", clientInfo.ID, clientInfo.RemoteAddr, err)
		drivers.RemoveClientIPsFromPool(clientInfo)
	} else {
		logger.Logger.Infof("Client disconnected: unknown client, reason: %v", err)
	}
}

// onMessage 处理客户端消息回调函数
func onMessage(conn *pc.Connection, message []byte) {
	// 解析消息
	var requestMsg pc.RequestMessage
	if err := json.Unmarshal(message, &requestMsg); err != nil {
		logger.Logger.Errorf("Failed to unmarshal request message: %v", err)
		return
	}

	businessProcessing(conn, &requestMsg)
}

// businessProcessing 业务处理(针对不同的请求类型)
func businessProcessing(conn *pc.Connection, message *pc.RequestMessage) {
	clientInfo := conn.GetClientInfo()
	if clientInfo == nil {
		logger.Logger.Errorf("Client info is nil for request %s", message.RequestID)
		return
	}

	namespaceTag := clientInfo.NamespaceTag
	envTag := clientInfo.EnvTag
	ctx := context.Background()

	logger.Logger.Infof("Processing request type %s from client %s (namespace: %s, env: %s)",
		message.Type, clientInfo.ID, namespaceTag, envTag)

	switch message.Type {
	case pc.SwitchFull:
		// 处理开关全量数据请求
		switches, err := switchFull(ctx, namespaceTag, envTag)
		if err != nil {
			logger.Logger.Errorf("Failed to get switch full data: %v", err)
			// 发送错误响应
			sendErrorResponse(conn, message.RequestID, pc.SwitchFull, err.Error())
			return
		}

		// 发送成功响应
		// TODO 可以走分批发送根据数据量大小进行优化
		err = sendSwitchFullToClient(clientInfo.ID, switches)
		if err != nil {
			logger.Logger.Errorf("Failed to send switch full response: %v", err)
		}

	case pc.DriverConfigFull:
		// 处理驱动配置全量数据请求
		allDrivers, err := driverConfigFull(ctx, namespaceTag, envTag)
		if err != nil {
			logger.Logger.Errorf("Failed to get driver config full data: %v", err)
			// 发送错误响应
			sendErrorResponse(conn, message.RequestID, pc.DriverConfigFull, err.Error())
			return
		}

		configPayloads := utils.BuildDriverConfigPayloads(allDrivers)

		// 发送成功响应
		err = sendDriverConfigFullToClient(clientInfo.ID, configPayloads)
		if err != nil {
			logger.Logger.Errorf("Failed to send driver config full response: %v", err)
		}

	default:
		logger.Logger.Warnf("Unsupported request type %s from client %s", message.Type, clientInfo.ID)
		sendErrorResponse(conn, message.RequestID, message.Type, "Unsupported request type")
	}
}

// sendErrorResponse 发送错误响应
func sendErrorResponse(conn *pc.Connection, requestID string, msgType pc.MessageType, errorMsg string) {
	clientInfo := conn.GetClientInfo()
	if clientInfo == nil {
		logger.Logger.Errorf("Client info is nil when sending error response")
		return
	}

	response := &pc.ResponseMessage{
		Type:          msgType,
		RequestID:     requestID,
		ServerVersion: "1.0.0",
		Timestamp:     time.Now().Unix(),
		ClientID:      clientInfo.ID,
		ServiceName:   clientInfo.ServiceName,
		NamespaceTag:  clientInfo.NamespaceTag,
		EnvTag:        clientInfo.EnvTag,
		Data: map[string]interface{}{
			"error":   true,
			"message": errorMsg,
		},
	}

	err := conn.SendJSON(response)
	if err != nil {
		logger.Logger.Errorf("Failed to send error response to client %s: %v", clientInfo.ID, err)
	} else {
		logger.Logger.Infof("Sent error response to client %s: %s", clientInfo.ID, errorMsg)
	}
}

// BroadcastMessage 广播消息给所有客户端
func BroadcastMessage(msgType pc.MessageType, message interface{}, endpoint string) error {
	if wsServer == nil || message == nil || endpoint == "" {
		logger.Logger.Infof("WebSocket server not started or config not ready, skip BroadcastMessage")
		return nil
	}

	// 获取重试配置
	retryConfig := config.GetRetryConfig()
	params := retryConfig.Default

	return wsServer.BroadcastReliableMessageToGroup(msgType, message, params.Timeout, params.MaxRetries, params.RetryDelay, func(connection *pc.Connection) bool {
		if connection == nil {
			return false
		}
		pd := connection.GetProxyData()
		if pd == nil {
			return false
		}
		pi := pd.Info
		if pi == nil {
			return false
		}
		return pi.Endpoint == endpoint
	})
}

// GetClientCount 获取当前连接的客户端数量
func GetClientCount() int {
	if wsServer == nil {
		return 0
	}
	return wsServer.GetClientCount()
}

// EndpointMatch 端点匹配
var EndpointMatch = func(endpoint string) pc.ConnectionFilter {
	return func(connection *pc.Connection) bool {
		if connection == nil {
			return false
		}
		pd := connection.GetProxyData()
		if pd == nil {
			return false
		}
		pi := pd.Info
		if pi == nil {
			return false
		}
		return pi.Endpoint == endpoint
	}
}

// EnvMatch 环境匹配
var EnvMatch = func(envTag string) pc.ConnectionFilter {
	return func(connection *pc.Connection) bool {
		if connection == nil {
			return false
		}
		info := connection.GetClientInfo()
		if info == nil {
			return false
		}
		return info.EnvTag == envTag
	}
}

// NSMatch 空间匹配
var NSMatch = func(nsTag string) pc.ConnectionFilter {
	return func(connection *pc.Connection) bool {
		if connection == nil {
			return false
		}
		info := connection.GetClientInfo()
		if info == nil {
			return false
		}
		return info.NamespaceTag == nsTag
	}
}

// BroadcastDriverConfigIncrementChange 广播驱动配置增量变更给所有客户端
func BroadcastDriverConfigIncrementChange(changes []*pc.DriverConfigPayload, endpoint string, filter pc.ConnectionFilter) error {
	if wsServer == nil || changes == nil || len(changes) == 0 || endpoint == "" {
		logger.Logger.Infof("WebSocket server not started or config not ready, skip BroadcastDriverConfigIncrementChange")
		return nil
	}

	// 获取配置变更重试配置
	retryConfig := config.GetRetryConfig()
	params := retryConfig.ConfigChange

	return wsServer.BroadcastReliableMessageToGroup(pc.DriverConfigChange, changes, params.Timeout, params.MaxRetries, params.RetryDelay, filter)
}

// mustMarshal 序列化
func mustMarshal(data interface{}) json.RawMessage {
	if data == nil {
		return json.RawMessage("{}")
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Logger.Errorf("Failed to marshal data: %v", err)
		return json.RawMessage("{}")
	}
	return bytes
}

// sendDriverConfigFullToClient 发送全量驱动配置给指定客户端
func sendDriverConfigFullToClient(clientID string, data interface{}) error {
	if wsServer == nil {
		logger.Logger.Infof("WebSocket server not started, skip send")
		return nil
	}

	// 获取全量配置重试配置
	retryConfig := config.GetRetryConfig()
	params := retryConfig.FullConfig

	logger.Logger.Infof("Sending full config to client %s with retry", clientID)
	return wsServer.SendReliableMessageWithRetry(clientID, pc.DriverConfigFull, data, params.Timeout, params.MaxRetries, params.RetryDelay)
}

// sendSwitchFullToClient 发送全量开关数据给指定客户端
func sendSwitchFullToClient(clientID string, data interface{}) error {
	if wsServer == nil {
		logger.Logger.Infof("WebSocket server not started, skip send")
		return nil
	}

	// 获取开关数据重试配置
	retryConfig := config.GetRetryConfig()
	params := retryConfig.SwitchData

	logger.Logger.Infof("Sending switch full data to client %s with retry", clientID)
	return wsServer.SendReliableMessageWithRetry(clientID, pc.SwitchFull, data, params.Timeout, params.MaxRetries, params.RetryDelay)
}
