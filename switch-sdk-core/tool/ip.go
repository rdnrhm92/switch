package tool

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GetLocalIP 获取本机IP地址(内网地址)
func GetLocalIP() (string, error) {
	ips, err := GetLocalIPs()
	if err != nil {
		return "", err
	}
	return ips[0], nil
}

// GetLocalIPs 获取本机IP地址(内网地址)
func GetLocalIPs() ([]string, error) {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}

	for _, iface := range interfaces {
		//判断网口是否是打开的
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no available IP address found")
	}

	return ips, nil
}

// ResolveDomainToIP 根据域名解析IP地址(返回第一个ipv4地址，没有则返回ipv6)
func ResolveDomainToIP(domain string) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", fmt.Errorf("failed to resolve domain %s: %v", domain, err)
	}

	// 优先返回IPv4地址
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}

	if len(ips) > 0 {
		return ips[0].String(), nil
	}

	return "", fmt.Errorf("no IP address found for domain: %s", domain)
}

// ResolveDomainToIPs 根据域名解析所有IP地址(ipv4 ipv6全部地址)
func ResolveDomainToIPs(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve domain %s: %v", domain, err)
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}

	if len(ipStrings) == 0 {
		return nil, fmt.Errorf("no IP address found for domain: %s", domain)
	}

	return ipStrings, nil
}

// ResolveDomainToIPv4 根据域名解析IPv4地址(全部地址)
func ResolveDomainToIPv4(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve domain %s: %v", domain, err)
	}

	var ipv4s []string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip.String())
		}
	}

	if len(ipv4s) == 0 {
		return nil, fmt.Errorf("no IPv4 address found for domain: %s", domain)
	}

	return ipv4s, nil
}

// ResolveDomainWithTimeout 带超时的域名解析 优先返回ipv4
func ResolveDomainWithTimeout(domain string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resolver := &net.Resolver{}
	ips, err := resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return "", fmt.Errorf("failed to resolve domain %s with timeout: %v", domain, err)
	}

	for _, ip := range ips {
		if ip.IP.To4() != nil {
			return ip.IP.String(), nil
		}
	}

	if len(ips) > 0 {
		return ips[0].IP.String(), nil
	}

	return "", fmt.Errorf("no IP address found for domain: %s", domain)
}

// GetPublicIP 获取外网IP地址
func GetPublicIP() (string, error) {
	return GetPublicIPWithTimeout(10 * time.Second)
}

// GetPublicIPWithTimeout 获取本机的外网IP
func GetPublicIPWithTimeout(timeout time.Duration) (string, error) {
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ipinfo.io/ip",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{
		Timeout: timeout,
	}

	for _, service := range services {
		if ip, err := fetchIPFromService(client, service); err == nil {
			return ip, nil
		}
	}

	return "", fmt.Errorf("failed to get public IP from all services")
}

// fetchIPFromService 从指定服务获取IP地址
func fetchIPFromService(client *http.Client, serviceURL string) (string, error) {
	resp, err := client.Get(serviceURL)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body failed: %w", err)
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP format: %s", ip)
	}

	return ip, nil
}

// IsPrivateIP 判断是否为私有IP地址
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 定义私有地址段
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// IsPublicIP 判断是否为公网IP
func IsPublicIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为内网IP
	return !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast()
}

// IsIntranetIP 判断是否为内网IP
func IsIntranetIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	return ip.IsPrivate()
}

// HasIntranetIPs 检查IP列表中是否有内网IP
func HasIntranetIPs(ips []string) bool {
	for _, ip := range ips {
		if IsIntranetIP(ip) {
			return true
		}
	}
	return false
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	LocalIPs  []string // 本机内网IP列表
	PublicIPs []string // 本机外网IP列表
}

var (
	networkInfo     *NetworkInfo
	networkInfoOnce sync.Once
	networkInfoErr  error
)

// GetNetworkInfo 获取网络信息
func GetNetworkInfo() (*NetworkInfo, error) {
	networkInfoOnce.Do(func() {
		networkInfo, networkInfoErr = discoverNetworkInfo()
	})
	return networkInfo, networkInfoErr
}

// discoverNetworkInfo 发现网络信息
func discoverNetworkInfo() (*NetworkInfo, error) {
	info := &NetworkInfo{}

	// 获取本机内网IP列表
	localIPs, err := GetLocalIPs()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IPs: %w", err)
	}
	info.LocalIPs = localIPs

	// 获取本机外网IP
	publicIP, err := GetPublicIP()
	if err != nil {
		info.PublicIPs = []string{}
	} else {
		info.PublicIPs = []string{publicIP}
	}

	return info, nil
}

// ResetNetworkInfo 重置网络信息缓存(慎用)
func ResetNetworkInfo() {
	networkInfoOnce = sync.Once{}
	networkInfo = nil
	networkInfoErr = nil
}

// GetCachedLocalIPs 获取缓存的本机IP列表
func GetCachedLocalIPs() ([]string, error) {
	info, err := GetNetworkInfo()
	if err != nil {
		return nil, err
	}
	return info.LocalIPs, nil
}

// GetCachedPublicIP 获取缓存的外网IP
func GetCachedPublicIP() (string, error) {
	info, err := GetNetworkInfo()
	if err != nil {
		return "", err
	}
	if len(info.PublicIPs) == 0 {
		return "", fmt.Errorf("public IP not available")
	}
	return info.PublicIPs[0], nil
}

// GetCachedPublicIPs 获取缓存的外网IP列表
func GetCachedPublicIPs() ([]string, error) {
	info, err := GetNetworkInfo()
	if err != nil {
		return nil, err
	}
	return info.PublicIPs, nil
}

// GetAllCachedIPs 获取所有缓存的IP地址（内网+外网）
func GetAllCachedIPs() ([]string, error) {
	info, err := GetNetworkInfo()
	if err != nil {
		return nil, err
	}

	allIPs := make([]string, 0, len(info.LocalIPs)+len(info.PublicIPs))
	allIPs = append(allIPs, info.LocalIPs...)
	allIPs = append(allIPs, info.PublicIPs...)

	return allIPs, nil
}
