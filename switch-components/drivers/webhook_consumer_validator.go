package drivers

import (
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/tool"
)

// WebhookConsumerValidator Webhook驱动验证器
// 1.配置有效性的检查
// 2.黑名单检查
type WebhookConsumerValidator struct {
	webhookDriver *WebhookConsumer
}

func (w *WebhookConsumerValidator) Validate(driver driver.Driver) error {
	webhookDriver, ok := driver.(*WebhookConsumer)
	if !ok {
		return fmt.Errorf("admin_driver is not a WebhookConsumer")
	}
	w.webhookDriver = webhookDriver

	// 校验配置有效性
	if err := w.webhookDriver.config.isValid(); err != nil {
		return fmt.Errorf("webhook config validation failed: %w", err)
	}

	// 检查本机IP是否在黑名单中，如果在则关闭驱动
	if err := w.checkLocalIPAgainstBlacklist(); err != nil {
		return fmt.Errorf("local IP blacklist check failed: %w", err)
	}

	return nil
}

// checkLocalIPAgainstBlacklist 检查本机IP是否在黑名单中
func (w *WebhookConsumerValidator) checkLocalIPAgainstBlacklist() error {
	if !w.webhookDriver.config.HasBlacklistIPs(w.webhookDriver.config.getBlacklistIPs()) {
		logger.Logger.Debug("No blacklist IPs configured, skipping blacklist check")
		return nil
	}

	blacklistIPs := w.webhookDriver.config.getBlacklistIPs()
	logger.Logger.Infof("Checking local IPs against blacklist: %v", blacklistIPs)

	// 网络信息获取
	networkInfo, err := tool.GetNetworkInfo()
	if err != nil {
		logger.Logger.Warnf("Failed to get network info: %v", err)
		return nil
	}

	// 检查本机内网IP
	if len(networkInfo.LocalIPs) > 0 {
		logger.Logger.Infof("Local IPs detected: %v", networkInfo.LocalIPs)
		for _, localIP := range networkInfo.LocalIPs {
			if w.webhookDriver.config.IsIPBlacklisted(localIP, w.webhookDriver.config.getBlacklistIPs()) {
				logger.Logger.Infof("Local IP %s is in blacklist, closing existing webhook drivers", localIP)
				w.closeExistingWebhookDrivers()
				return fmt.Errorf("local IP %s is blacklisted, webhook consumer disabled", localIP)
			}
		}
	}

	// 检查本机外网IP
	if len(networkInfo.PublicIPs) > 0 {
		logger.Logger.Infof("Public IPs detected: %v", networkInfo.PublicIPs)
		for _, publicIP := range networkInfo.PublicIPs {
			if w.webhookDriver.config.IsIPBlacklisted(publicIP, w.webhookDriver.config.getBlacklistIPs()) {
				logger.Logger.Infof("Public IP %s is in blacklist, closing existing webhook drivers", publicIP)
				w.closeExistingWebhookDrivers()
				return fmt.Errorf("public IP %s is blacklisted, webhook consumer disabled", publicIP)
			}
		}
	}

	logger.Logger.Info("Local IPs are not in blacklist, webhook consumer validation passed")
	return nil
}

// closeExistingWebhookDrivers 关闭现有的webhook驱动
func (w *WebhookConsumerValidator) closeExistingWebhookDrivers() {
	// 关闭所有webhook consumer类型的驱动
	if err := driver.GetManager().CloseByType(WebhookConsumerDriverType); err != nil {
		logger.Logger.Errorf("Failed to close existing webhook drivers: %v", err)
	} else {
		logger.Logger.Info("Successfully closed all existing webhook drivers due to blacklist match")
	}
}
