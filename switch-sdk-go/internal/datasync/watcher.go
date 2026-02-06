package datasync

import (
	"context"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/tool"
	_ "gitee.com/fatzeng/switch-sdk-go/core/factor"
	_switch "gitee.com/fatzeng/switch-sdk-go/core/switch"
)

// 全局通道，用于接收配置拉取完成信号
var (
	configSyncDone = make(chan string, 10) // 带缓冲的通道，避免阻塞
	syncOnce       = make(map[string]*sync.Once)
	syncMutex      sync.Mutex
)

// NotifyConfigSyncDone 通知指定端点的配置同步完成(只有当某些特殊的消息回调回来后，再执行对应的客户端启动逻辑)
// sync.once保证只通知一次
func NotifyConfigSyncDone(endpointPath string) {
	syncMutex.Lock()
	once, exists := syncOnce[endpointPath]
	syncMutex.Unlock()

	if exists && once != nil {
		once.Do(func() {
			select {
			case configSyncDone <- endpointPath:
				logger.Logger.Infof("Notified config sync done for endpoint: %s", endpointPath)
			default:
				logger.Logger.Warnf("Config sync done channel is full, skipping notification for: %s", endpointPath)
			}
		})
	}
}

// EndpointConfig 端点配置
type EndpointConfig struct {
	Path       string
	Handler    pc.MessageHandler
	OnTrusted  func(client *pc.Client) func(conn *pc.Connection)
	OnRegister func(client *pc.Client) func(conn *pc.Connection) *pc.RegisterPayload
}

// buildDefaultConfig 构建默认的WebSocket连接配置 start中已经初始化
func buildDefaultConfig() *pc.PersistentConnectionConfig {
	return &pc.PersistentConnectionConfig{
		ClientVersion:     _switch.GlobalClient.ClientVersion(),
		RequestHeader:     _switch.GlobalClient.RequestHeader(),
		ReconnectStrategy: _switch.GlobalClient.ReconnectStrategy(),
		Heartbeat:         _switch.GlobalClient.Heartbeat(),
		WriteTimeout:      _switch.GlobalClient.WriteTimeout(),
		ReadTimeout:       _switch.GlobalClient.ReadTimeout(),
		DialTimeout:       _switch.GlobalClient.DialTimeout(),
	}
}

// StartSyncer 启动所有数据同步连接
func StartSyncer(ctx context.Context) error {
	// 准备基础连接配置
	baseConfig := buildDefaultConfig()

	endpointConfigs := []EndpointConfig{
		createConfigChangeEndpoint(),
		createFullSyncConfigEndpoint(),
		createFullSyncEndpoint(),
	}

	var wg sync.WaitGroup
	wg.Add(len(endpointConfigs))

	for _, endpointConfig := range endpointConfigs {
		cfg := *baseConfig
		cfg.Address = _switch.GlobalClient.Domain() + endpointConfig.Path

		logger.Logger.Infof("Starting data syncer for endpoint: %s", cfg.Address)

		syncMutex.Lock()
		syncOnce[endpointConfig.Path] = &sync.Once{}
		syncMutex.Unlock()

		// 消息处理逻辑
		cfg.MessageHandler = endpointConfig.Handler

		// 创建客户端
		clientCfg := cfg
		client := pc.NewClient(&clientCfg)

		// 握手过程中需要提供客户端信息
		clientCfg.OnRegister = endpointConfig.OnRegister(client)
		// 受信成功后的逻辑
		clientCfg.OnTrusted = endpointConfig.OnTrusted(client)

		client.Start(ctx)
	}

	// 等待所有端点的配置拉取完成信号
	go func() {
		completedEndpoints := make(map[string]bool)
		for len(completedEndpoints) < len(endpointConfigs) {
			select {
			case endpointPath := <-configSyncDone:
				if !completedEndpoints[endpointPath] {
					completedEndpoints[endpointPath] = true
					logger.Logger.Infof("Release %s signal", endpointPath)
					wg.Done()
				}
			case <-ctx.Done():
				logger.Logger.Warn("Context cancelled while waiting for release signal")
				return
			}
		}
	}()

	// 不直接使用 wg.Wait() 确保外界的ctx可以正常退出
	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		wg.Wait()
	}()

	// 因为服务端如果没传递全量驱动或者全量开关(认为是一种危险操作)将不允许启动客户端，方便告知用户为何会阻塞，这里给个提示
	promptDelay := 5 * time.Second
	ticker := time.NewTicker(promptDelay)
	defer ticker.Stop()

	for {
		select {
		case <-waitDone:
			logger.Logger.Info("All endpoints sync completed")
			return nil
		case <-ctx.Done():
			logger.Logger.Warn("Context cancelled while waiting for config sync")
			return ctx.Err()
		case <-ticker.C:
			logger.Logger.Info("Blocked! insufficient startup data has not been obtained, SWITCH_FULL/DRIVER_CONFIG_FULL")
			continue
		}
	}
}

// StopSyncer 优雅地关闭所有连接
func StopSyncer() {
	logger.Logger.Info("Stopping all data syncers...")
	//ws关闭 各种驱动关闭
	_ = driver.GetManager().Close()
	//开关驱动关闭
	logger.Logger.Info("All data syncers stopped.")
}

// buildServiceRegister 构建服务注册信息
func buildServiceRegister(c *pc.Client) *pc.RegisterPayload {
	// 获取网络信息
	networkInfo, err := tool.GetNetworkInfo()
	if err != nil {
		logger.Logger.Warnf("Failed to get network info for registration: %v", err)
		networkInfo = &tool.NetworkInfo{
			LocalIPs:  []string{},
			PublicIPs: []string{},
		}
	}

	return &pc.RegisterPayload{
		ServiceName:  _switch.GlobalClient.ServiceName(),
		ClientID:     c.GetServerClientID(),
		SDKVersion:   _switch.GlobalClient.Version(),
		InternalIPs:  networkInfo.LocalIPs,
		PublicIPs:    networkInfo.PublicIPs,
		NamespaceTag: _switch.GlobalClient.NamespaceTag(),
		EnvTag:       _switch.GlobalClient.EnvTag(),
		Metadata: map[string]string{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
}

// createConfigChangeEndpoint 创建配置变更端点配置
func createConfigChangeEndpoint() EndpointConfig {
	return EndpointConfig{
		Path:    pc.WsEndpointChangeConfig,
		Handler: handleChangeConfig,
		OnRegister: func(client *pc.Client) func(conn *pc.Connection) *pc.RegisterPayload {
			return func(conn *pc.Connection) *pc.RegisterPayload {
				return buildServiceRegister(client)
			}
		},
		OnTrusted: func(c *pc.Client) func(conn *pc.Connection) {
			return func(conn *pc.Connection) {
				logger.Logger.Infof("Config change endpoint trusted, ready to receive config changes")
				// 配置的增量变更不走受信后的消息,走server端的推送

				// 配置变更的监听不影响主服务启动,可以在连接建立后立刻响应
				NotifyConfigSyncDone(pc.WsEndpointChangeConfig)
			}
		},
	}
}

// createFullSyncEndpoint 创建全量同步端点配置
func createFullSyncEndpoint() EndpointConfig {
	return EndpointConfig{
		Path:    pc.WsEndpointFullSync,
		Handler: handleFullSync,
		OnRegister: func(client *pc.Client) func(conn *pc.Connection) *pc.RegisterPayload {
			return func(conn *pc.Connection) *pc.RegisterPayload {
				return buildServiceRegister(client)
			}
		},
		OnTrusted: func(c *pc.Client) func(conn *pc.Connection) {
			return func(conn *pc.Connection) {
				logger.Logger.Infof("Full sync endpoint trusted, starting full sync...")
				// 全量同步端点的受信后逻辑：执行全量开关拉取
				syncFull(c)
			}
		},
	}
}

// createFullSyncConfigEndpoint 创建全量配置监听端点
func createFullSyncConfigEndpoint() EndpointConfig {
	return EndpointConfig{
		Path:    pc.WsEndpointFullSyncConfig,
		Handler: handleChangeConfig,
		OnRegister: func(client *pc.Client) func(conn *pc.Connection) *pc.RegisterPayload {
			return func(conn *pc.Connection) *pc.RegisterPayload {
				return buildServiceRegister(client)
			}
		},
		OnTrusted: func(c *pc.Client) func(conn *pc.Connection) {
			return func(conn *pc.Connection) {
				logger.Logger.Infof("Full sync config endpoint trusted, starting full sync...")
				// 全量配置端点的受信后逻辑：执行全量配置拉取
				syncConfigFull(c)
			}
		},
	}
}
