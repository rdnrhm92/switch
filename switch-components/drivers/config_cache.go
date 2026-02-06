package drivers

import (
	"sync"
	"sync/atomic"
	"time"
)

// ConfigVersion 配置版本信息
type ConfigVersion struct {
	Version   uint64      `json:"version"`
	Config    interface{} `json:"config"`
	Timestamp time.Time   `json:"timestamp"`
}

// ConfigCache 配置缓存管理器
type ConfigCache struct {
	mutex sync.RWMutex

	// 按时间顺序存储的配置数组
	configs []*ConfigVersion

	// 全局自增版本号
	versionCounter uint64

	// 配置参数
	maxVersions int           // 缓存的配置最大数量
	maxAge      time.Duration // 缓存的配置存在的最长时间
}

// NewConfigCache 默认的配置
func NewConfigCache() *ConfigCache {
	return NewConfigCacheWithOptions(3000, 1*time.Hour)
}

// NewConfigCacheWithOptions 自定义的配置
func NewConfigCacheWithOptions(maxVersions int, maxAge time.Duration) *ConfigCache {
	return &ConfigCache{
		configs:     make([]*ConfigVersion, 0, maxVersions),
		maxVersions: maxVersions,
		maxAge:      maxAge,
	}
}

// AddConfig 添加一个配置
func (c *ConfigCache) AddConfig(config interface{}) uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 原子递增版本号 是全局的
	atomicVersion := atomic.AddUint64(&c.versionCounter, 1)

	buildConfig := &ConfigVersion{
		Version:   atomicVersion,
		Config:    config,
		Timestamp: time.Now(),
	}

	// 添加到数组尾部
	c.configs = append(c.configs, buildConfig)

	// 清理超量和过期的配置
	c.cleanup()

	return atomicVersion
}

// GetVersion 获取指定版本的配置
func (c *ConfigCache) GetVersion(version uint64) (*ConfigVersion, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 在数组中查找指定版本
	for _, config := range c.configs {
		if config.Version == version {
			return config, true
		}
	}
	return nil, false
}

// GetLatestVersion 获取最新版本号
func (c *ConfigCache) GetLatestVersion() uint64 {
	return atomic.LoadUint64(&c.versionCounter)
}

// GetLatestConfig 获取最新配置
func (c *ConfigCache) GetLatestConfig() (*ConfigVersion, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.configs) == 0 {
		return nil, false
	}

	// 返回最新的数据
	return c.configs[len(c.configs)-1], true
}

// HasNewerVersion 检查是否有比指定版本更新的配置
func (c *ConfigCache) HasNewerVersion(clientVersion uint64) bool {
	return atomic.LoadUint64(&c.versionCounter) > clientVersion
}

// GetVersionsSince 获取指定版本之后的所有版本 客户端请求过来后。服务端响应客户端版本后的所有内容
func (c *ConfigCache) GetVersionsSince(clientVersion uint64) []*ConfigVersion {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var result []*ConfigVersion

	// 从数组中找到客户端版本的位置
	for i, config := range c.configs {
		if config.Version == clientVersion {
			// 返回该位置之后的所有配置
			if i+1 < len(c.configs) {
				result = make([]*ConfigVersion, len(c.configs)-i-1)
				copy(result, c.configs[i+1:])
			}
			return result
		}
	}

	// 如果没找到客户端版本，返回所有比它大的版本(可能是已经删除了或者过期了，暂时不考虑这样的情况了)
	for _, config := range c.configs {
		if config.Version > clientVersion {
			result = append(result, config)
		}
	}

	return result
}

// GetOldestCachedVersion 获取最旧的缓存版本号
func (c *ConfigCache) GetOldestCachedVersion() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.configs) == 0 {
		return 0
	}

	// 最旧的数据
	return c.configs[0].Version
}

// cleanup 清理过期和超量的版本
func (c *ConfigCache) cleanup() {
	now := time.Now()
	cutoffTime := now.Add(-c.maxAge)

	// 从数组头部删除过期的配置 append追加最新的到尾部
	for len(c.configs) > 0 && c.configs[0].Timestamp.Before(cutoffTime) {
		c.configs = c.configs[1:]
	}

	// 从数组头部删除超量的配置 append追加最新的到尾部
	if len(c.configs) > c.maxVersions {
		deleteCount := len(c.configs) - c.maxVersions
		c.configs = c.configs[deleteCount:]
	}
}

// GetCacheStats 获取缓存统计信息
func (c *ConfigCache) GetCacheStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"total_versions":    len(c.configs),
		"current_version":   atomic.LoadUint64(&c.versionCounter),
		"oldest_version":    c.GetOldestCachedVersion(),
		"max_versions":      c.maxVersions,
		"max_age_seconds":   int(c.maxAge.Seconds()),
		"cache_utilization": float64(len(c.configs)) / float64(c.maxVersions),
	}
}

// Clear 清空所有缓存 重置归0
func (c *ConfigCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.configs = c.configs[:0]
	atomic.StoreUint64(&c.versionCounter, 0)
}
