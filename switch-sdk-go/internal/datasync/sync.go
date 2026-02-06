package datasync

import (
	"time"

	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	_switch "gitee.com/fatzeng/switch-sdk-go/core/switch"
)

// syncFull 申请全量
func syncFull(c *pc.Client) {
	// 构建申请全量开关数据的请求
	requestData := map[string]interface{}{
		"service_name": _switch.GlobalClient.ServiceName(),
		"request_type": "full_sync",
	}

	if err := c.SendRequest(pc.SwitchFull, requestData, 30*time.Second); err != nil {
		logger.Logger.Errorf("Failed to send full sync request: %v", err)
	}
}

// syncConfigFull 拉取全量配置
func syncConfigFull(c *pc.Client) {
	// 构建申请全量驱动配置的请求
	requestData := map[string]interface{}{
		"service_name": _switch.GlobalClient.ServiceName(),
		"request_type": "config_full_sync",
	}

	if err := c.SendRequest(pc.DriverConfigFull, requestData, 30*time.Second); err != nil {
		logger.Logger.Errorf("Failed to send config full sync request: %v", err)
	}
}
