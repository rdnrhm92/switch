package drivers

import (
	"context"
	"net"
	"os/exec"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-components/pc"
	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// IPConnectivityConfig IP连通性检查配置
type IPConnectivityConfig struct {
	CheckInterval  time.Duration `yaml:"check_interval"`    // 存量IP定时检查间隔
	CheckTimeout   time.Duration `yaml:"check_timeout"`     // 单个IP检查超时时间
	NewIPQueueSize int           `yaml:"new_ip_queue_size"` // 新IP检查队列大小
}

var ipManager *IPPoolManager

// StartIPPoolManager 启动
func StartIPPoolManager(ctx context.Context, config *IPConnectivityConfig) error {
	if ipManager != nil {
		logger.Logger.Warnf("IP Pool Manager already started")
		return nil
	}

	ipManager = newIPPoolManager(ctx, config)
	ipManager.start(ctx)

	logger.Logger.Infof("Global IP Pool Manager started successfully")
	return nil
}

// StopIPPoolManager 停止
func StopIPPoolManager() error {
	if ipManager == nil {
		logger.Logger.Warnf("IP Pool Manager not started")
		return nil
	}

	ipManager.stop()
	ipManager = nil

	logger.Logger.Infof("Global IP Pool Manager stopped successfully")
	return nil
}

// GetIPPoolManager 获取全局IP池管理器
func GetIPPoolManager() *IPPoolManager {
	return ipManager
}

// AddClientIPsToPool 添加客户端IP到全局池
func AddClientIPsToPool(clientInfo *pc.ClientProxyInfo) {
	if ipManager == nil || clientInfo == nil {
		return
	}

	newIPs := ipManager.AddIPs(clientInfo.PublicIP, clientInfo.InternalIP)

	// 异步检查新IP
	if len(newIPs) > 0 {
		select {
		case ipManager.newIPQueue <- newIPs:
			logger.Logger.Infof("Queued %d new IPs for connectivity check", len(newIPs))
		default:
			logger.Logger.Warnf("New IP queue is full, skipping connectivity check")
		}
	}
}

// RemoveClientIPsFromPool 从全局池中移除客户端IP
func RemoveClientIPsFromPool(clientInfo *pc.ClientProxyInfo) {
	if ipManager == nil || clientInfo == nil {
		return
	}

	removedCount := ipManager.RemoveIPs(clientInfo.PublicIP, clientInfo.InternalIP)
	logger.Logger.Infof("Removed %d IPs from pool for client %s", removedCount, clientInfo.ID)
}

// IPPoolManager 全局IP池 维护长连接回调中的内网公网IP
// 只针对webhook 驱动中的连接做检查,比IP检查更严苛
type IPPoolManager struct {
	// 公网IP
	publicIPPool map[string]*IPStatus
	publicMutex  sync.RWMutex

	// 内网IP
	internalIPPool map[string]*IPStatus
	internalMutex  sync.RWMutex

	// 异步任务
	newIPQueue chan []string
	ctx        context.Context
	cancel     context.CancelFunc

	// 配置
	config *IPConnectivityConfig
}

// IPStatus IP状态信息
type IPStatus struct {
	IP string
	// 是否可达
	IsReachable bool
	// 最近一次的检查时间
	LastChecked time.Time
}

// newIPPoolManager 创建IP池管理器
func newIPPoolManager(ctx context.Context, config *IPConnectivityConfig) *IPPoolManager {
	childCtx, cancel := context.WithCancel(ctx)

	queueSize := config.NewIPQueueSize

	return &IPPoolManager{
		publicIPPool:   make(map[string]*IPStatus),
		internalIPPool: make(map[string]*IPStatus),
		newIPQueue:     make(chan []string, queueSize),
		ctx:            childCtx,
		cancel:         cancel,
		config:         config,
	}
}

// start 启动
func (m *IPPoolManager) start(ctx context.Context) {

	// 启动新IP检查工作协程
	recovery.SafeGo(ctx, func(ctx context.Context) error {
		m.newIPWorker()
		return nil
	}, "increment_check")

	// 启动定时检查连通性
	recovery.SafeGo(ctx, func(ctx context.Context) error {
		m.periodicChecker()
		return nil
	}, "inventory_check")

	logger.Logger.Infof("IP Pool Manager started")
}

// stop 停止
func (m *IPPoolManager) stop() {
	if m.cancel != nil {
		m.cancel()
	}
	logger.Logger.Infof("IP Pool Manager stopped")
}

// AddIPs 添加IP到池中
// 新添加的IP都设置成不可达等待检查
func (m *IPPoolManager) AddIPs(publicIPs, internalIPs []string) []string {
	var newIPs []string

	// 添加公网IP
	m.publicMutex.Lock()
	for _, ip := range publicIPs {
		if ip == "" {
			continue
		}
		if _, exists := m.publicIPPool[ip]; !exists {
			m.publicIPPool[ip] = &IPStatus{
				IP:          ip,
				IsReachable: false,
				LastChecked: time.Time{},
			}
			newIPs = append(newIPs, ip)
		}
	}
	m.publicMutex.Unlock()

	// 添加内网IP
	m.internalMutex.Lock()
	for _, ip := range internalIPs {
		if ip == "" {
			continue
		}
		if _, exists := m.internalIPPool[ip]; !exists {
			m.internalIPPool[ip] = &IPStatus{
				IP:          ip,
				IsReachable: false,
				LastChecked: time.Time{},
			}
			newIPs = append(newIPs, ip)
		}
	}
	m.internalMutex.Unlock()

	logger.Logger.Infof("Added IPs (public: %d, internal: %d), %d are new",
		len(publicIPs), len(internalIPs), len(newIPs))
	return newIPs
}

// RemoveIPs 从池中移除IP
func (m *IPPoolManager) RemoveIPs(publicIPs, internalIPs []string) int {
	var removedCount int

	// 移除公网IP
	m.publicMutex.Lock()
	for _, ip := range publicIPs {
		if ip == "" {
			continue
		}
		if _, exists := m.publicIPPool[ip]; exists {
			delete(m.publicIPPool, ip)
			removedCount++
			logger.Logger.Debugf("Removed public IP: %s", ip)
		}
	}
	m.publicMutex.Unlock()

	// 移除内网IP
	m.internalMutex.Lock()
	for _, ip := range internalIPs {
		if ip == "" {
			continue
		}
		if _, exists := m.internalIPPool[ip]; exists {
			delete(m.internalIPPool, ip)
			removedCount++
			logger.Logger.Debugf("Removed internal IP: %s", ip)
		}
	}
	m.internalMutex.Unlock()

	logger.Logger.Infof("Removed IPs (public: %d, internal: %d), total removed: %d",
		len(publicIPs), len(internalIPs), removedCount)
	return removedCount
}

// getAllIPs 获取所有IP列表(分析前的结果)
func (m *IPPoolManager) getAllIPs() []string {
	var ips []string

	m.publicMutex.RLock()
	for ip := range m.publicIPPool {
		ips = append(ips, ip)
	}
	m.publicMutex.RUnlock()

	m.internalMutex.RLock()
	for ip := range m.internalIPPool {
		ips = append(ips, ip)
	}
	m.internalMutex.RUnlock()

	return ips
}

// getReachableIPs 获取可达的IP列表(分析后的结果)
func (m *IPPoolManager) getReachableIPs() []string {
	var reachableIPs []string

	m.publicMutex.RLock()
	for _, status := range m.publicIPPool {
		if status.IsReachable {
			reachableIPs = append(reachableIPs, status.IP)
		}
	}
	m.publicMutex.RUnlock()

	m.internalMutex.RLock()
	for _, status := range m.internalIPPool {
		if status.IsReachable {
			reachableIPs = append(reachableIPs, status.IP)
		}
	}
	m.internalMutex.RUnlock()

	return reachableIPs
}

// getUnreachableIPs 获取不可达的IP列表(分析后的结果)
func (m *IPPoolManager) getUnreachableIPs() []string {
	var unreachableIPs []string

	m.publicMutex.RLock()
	for _, status := range m.publicIPPool {
		if !status.IsReachable && !status.LastChecked.IsZero() {
			unreachableIPs = append(unreachableIPs, status.IP)
		}
	}
	m.publicMutex.RUnlock()

	m.internalMutex.RLock()
	for _, status := range m.internalIPPool {
		if !status.IsReachable && !status.LastChecked.IsZero() {
			unreachableIPs = append(unreachableIPs, status.IP)
		}
	}
	m.internalMutex.RUnlock()

	return unreachableIPs
}

// GetReachableIPs 获取可达的IP列表
func GetReachableIPs() []string {
	if ipManager == nil {
		return []string{}
	}
	return ipManager.getReachableIPs()
}

// GetUnreachableIPs 获取不可达的IP列表
func GetUnreachableIPs() []string {
	if ipManager == nil {
		return []string{}
	}
	return ipManager.getUnreachableIPs()
}

// GetAllIPs 获取所有IP列表
func GetAllIPs() []string {
	if ipManager == nil {
		return []string{}
	}
	return ipManager.getAllIPs()
}

// newIPWorker 新IP检查工作协程
func (m *IPPoolManager) newIPWorker() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case ips := <-m.newIPQueue:
			m.checkIPsConnectivity(ips)
		}
	}
}

// periodicChecker 定时检查存量IP
func (m *IPPoolManager) periodicChecker() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkExistingIPs()
		}
	}
}

// checkExistingIPs 检查存量IP的连通性
func (m *IPPoolManager) checkExistingIPs() {
	var ipsToCheck []string
	now := time.Now()

	// 检查公网IP池
	m.publicMutex.RLock()
	for _, status := range m.publicIPPool {
		// 如果从未检查过，或者距离上次检查超过间隔时间
		if status.LastChecked.IsZero() || now.Sub(status.LastChecked) >= m.config.CheckInterval {
			ipsToCheck = append(ipsToCheck, status.IP)
		}
	}
	m.publicMutex.RUnlock()

	// 检查内网IP池
	m.internalMutex.RLock()
	for _, status := range m.internalIPPool {
		// 如果从未检查过，或者距离上次检查超过间隔时间
		if status.LastChecked.IsZero() || now.Sub(status.LastChecked) >= m.config.CheckInterval {
			ipsToCheck = append(ipsToCheck, status.IP)
		}
	}
	m.internalMutex.RUnlock()

	if len(ipsToCheck) > 0 {
		logger.Logger.Infof("Periodic check: checking %d existing IPs", len(ipsToCheck))

		// 开始检查
		go m.checkIPsConnectivity(ipsToCheck)
	}
}

// checkIPsConnectivity 检查IP连通性（只检查网络可达性，不检查具体服务）
func (m *IPPoolManager) checkIPsConnectivity(ips []string) {
	if len(ips) == 0 {
		return
	}

	var reachableIPs []string
	var unreachableIPs []string

	// 并发检查每个IP
	var wg sync.WaitGroup
	var reachableMutex sync.Mutex
	var unreachableMutex sync.Mutex

	for _, ip := range ips {
		wg.Add(1)
		go func(currentIP string) {
			defer func() {
				if r := recover(); r != nil {
					logger.Logger.Errorf("Panic in IP connectivity check for %s: %v", currentIP, r)
					unreachableMutex.Lock()
					unreachableIPs = append(unreachableIPs, currentIP)
					unreachableMutex.Unlock()
				}
				wg.Done()
			}()

			// 只检查IP的网络可达性
			if isIPReachable(currentIP, m.config.CheckTimeout) {
				reachableMutex.Lock()
				reachableIPs = append(reachableIPs, currentIP)
				reachableMutex.Unlock()
				logger.Logger.Debugf("IP %s is reachable", currentIP)
			} else {
				unreachableMutex.Lock()
				unreachableIPs = append(unreachableIPs, currentIP)
				unreachableMutex.Unlock()
				logger.Logger.Debugf("IP %s is unreachable", currentIP)
			}
		}(ip)
	}

	wg.Wait()

	// 只更新IP池管理器内部的状态
	m.updateIPStatus(reachableIPs, unreachableIPs)

	logger.Logger.Infof("IP Pool connectivity check completed: %d reachable, %d unreachable",
		len(reachableIPs), len(unreachableIPs))
}

// updateIPStatus 更新IP状态
func (m *IPPoolManager) updateIPStatus(reachableIPs, unreachableIPs []string) {
	now := time.Now()

	// 更新可达IP状态
	for _, ip := range reachableIPs {
		m.publicMutex.Lock()
		if status, exists := m.publicIPPool[ip]; exists {
			status.IsReachable = true
			status.LastChecked = now
		}
		m.publicMutex.Unlock()

		m.internalMutex.Lock()
		if status, exists := m.internalIPPool[ip]; exists {
			status.IsReachable = true
			status.LastChecked = now
		}
		m.internalMutex.Unlock()
	}

	// 更新不可达IP状态
	for _, ip := range unreachableIPs {
		m.publicMutex.Lock()
		if status, exists := m.publicIPPool[ip]; exists {
			status.IsReachable = false
			status.LastChecked = now
		}
		m.publicMutex.Unlock()

		m.internalMutex.Lock()
		if status, exists := m.internalIPPool[ip]; exists {
			status.IsReachable = false
			status.LastChecked = now
		}
		m.internalMutex.Unlock()
	}
}

// isIPReachable 检查IP是否可达 使用ping ip + 常用端口的方式进行探测
func isIPReachable(ip string, timeout time.Duration) bool {
	if pingIP(ip, timeout) {
		logger.Logger.Debugf("IP %s is reachable via ping", ip)
		return true
	}

	commonPorts := []string{
		"80", "443",
		"8080", "8443",
		"3000", "8000", "9000",
		"22", "21",
		"25", "53",
	}

	for _, port := range commonPorts {
		if tcpConnect(ip, port, timeout) {
			logger.Logger.Debugf("IP %s is reachable on port %s", ip, port)
			return true
		}
	}

	logger.Logger.Debugf("IP %s is not reachable via ping or any common ports", ip)
	return false
}

// pingIP 使用系统ping命令检查IP可达性
func pingIP(ip string, timeout time.Duration) bool {
	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用系统ping命令，发送1个包，超时1秒
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1000", ip)
	err := cmd.Run()
	return err == nil
}

// tcpConnect 尝试TCP连接指定端口
func tcpConnect(ip, port string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", ip+":"+port, timeout)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}
