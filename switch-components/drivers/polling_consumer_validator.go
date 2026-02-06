package drivers

import (
	"fmt"
	"net/http"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// PollingConsumerValidator 长轮询消费者驱动验证器
// 1.配置有效性验证
// 2.服务端的连通性验证
type PollingConsumerValidator struct {
	pollingDriver *PollingConsumer
}

// Validate 验证长轮询消费者驱动
func (p *PollingConsumerValidator) Validate(driver driver.Driver) error {
	pollingDriver, ok := driver.(*PollingConsumer)
	if !ok {
		return fmt.Errorf("admin_driver is not a PollingConsumer")
	}
	p.pollingDriver = pollingDriver

	// 验证配置有效性
	if err := p.pollingDriver.config.isValid(); err != nil {
		return fmt.Errorf("polling consumer config validation failed: %w", err)
	}

	// 验证连接性
	if err := p.validateConnectivity(); err != nil {
		return fmt.Errorf("polling consumer connectivity validation failed: %w", err)
	}

	return nil
}

// validateConnectivity 验证连接性
func (p *PollingConsumerValidator) validateConnectivity() error {
	url, err := p.pollingDriver.config.getValidURL()
	if err != nil {
		return fmt.Errorf("failed to get valid URL: %w", err)
	}

	// 检查连通性
	return p.checkPollingConnectivity(url)
}

// checkPollingConnectivity 检查polling服务器的连通性
func (p *PollingConsumerValidator) checkPollingConnectivity(url string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 检查连通性
	if err := checkPollingURLConnectivity(client, url); err != nil {
		logger.Logger.Warnf("Polling URL unreachable: %s, error: %v", url, err)
		return fmt.Errorf("polling URL unreachable: %w", err)
	}

	logger.Logger.Infof("Polling URL reachable: %s", url)
	return nil
}

// checkPollingURLConnectivity 检查单个polling URL的连通性
func checkPollingURLConnectivity(client *http.Client, url string) error {
	resp, err := client.Head(url)
	if err != nil {
		return fmt.Errorf("connection failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return nil
	}

	return fmt.Errorf("bad status code %d for %s", resp.StatusCode, url)
}
