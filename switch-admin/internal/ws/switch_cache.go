package ws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// SwitchCacheItem 缓存项
type SwitchCacheItem struct {
	Data      []*model.SwitchModel
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired 检查缓存是否过期
func (item *SwitchCacheItem) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// SwitchCache 开关数据缓存管理器
type SwitchCache struct {
	cache    map[string]*SwitchCacheItem
	mutex    sync.RWMutex
	ttl      time.Duration
	stopChan chan struct{}
	started  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSwitchCache 创建缓存管理器
func NewSwitchCache(ctx context.Context, ttl time.Duration) *SwitchCache {
	childCtx, cancel := context.WithCancel(ctx)
	return &SwitchCache{
		cache:    make(map[string]*SwitchCacheItem),
		ttl:      ttl,
		stopChan: make(chan struct{}),
		ctx:      childCtx,
		cancel:   cancel,
	}
}

// start 启动缓存清理协程
func (c *SwitchCache) start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.started {
		return
	}

	c.started = true
	go c.cleanupExpired()
	logger.Logger.Infof("Switch cache started with TTL: %v", c.ttl)
}

// Stop 停止缓存管理器
func (c *SwitchCache) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.started {
		return
	}

	if c.cancel != nil {
		c.cancel()
	}

	close(c.stopChan)
	c.started = false
	logger.Logger.Infof("Switch cache stopped")
}

// buildKey 构建缓存键
func (c *SwitchCache) buildKey(namespaceTag, envTag string) string {
	return fmt.Sprintf("switch:%s:%s", namespaceTag, envTag)
}

// Get 获取缓存数据
func (c *SwitchCache) Get(namespaceTag, envTag string) ([]*model.SwitchModel, bool) {
	key := c.buildKey(namespaceTag, envTag)

	c.mutex.RLock()
	item, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		logger.Logger.Debugf("Cache miss for key: %s", key)
		return nil, false
	}

	if item.IsExpired() {
		logger.Logger.Debugf("Cache expired for key: %s", key)
		c.Delete(namespaceTag, envTag)
		return nil, false
	}

	logger.Logger.Debugf("Cache hit for key: %s, data count: %d, age: %v",
		key, len(item.Data), time.Since(item.CreatedAt))
	return item.Data, true
}

// Set 设置缓存数据
func (c *SwitchCache) Set(namespaceTag, envTag string, data []*model.SwitchModel) {
	key := c.buildKey(namespaceTag, envTag)
	now := time.Now()

	item := &SwitchCacheItem{
		Data:      data,
		ExpiresAt: now.Add(c.ttl),
		CreatedAt: now,
	}

	c.mutex.Lock()
	c.cache[key] = item
	c.mutex.Unlock()

	logger.Logger.Infof("Cached switch data for key: %s, count: %d, expires at: %v",
		key, len(data), item.ExpiresAt.Format("15:04:05"))
}

// Delete 删除缓存数据
func (c *SwitchCache) Delete(namespaceTag, envTag string) {
	key := c.buildKey(namespaceTag, envTag)

	c.mutex.Lock()
	delete(c.cache, key)
	c.mutex.Unlock()

	logger.Logger.Debugf("Deleted cache for key: %s", key)
}

// Clear 清空所有缓存
func (c *SwitchCache) Clear() {
	c.mutex.Lock()
	c.cache = make(map[string]*SwitchCacheItem)
	c.mutex.Unlock()

	logger.Logger.Infof("Cleared all switch cache")
}

// GetStats 获取缓存统计信息
func (c *SwitchCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	totalItems := len(c.cache)
	expiredItems := 0
	totalSwitches := 0

	for _, item := range c.cache {
		if item.IsExpired() {
			expiredItems++
		}
		totalSwitches += len(item.Data)
	}

	return map[string]interface{}{
		"total_keys":     totalItems,
		"expired_keys":   expiredItems,
		"total_switches": totalSwitches,
		"ttl_seconds":    int(c.ttl.Seconds()),
	}
}

// cleanupExpired 定期清理过期缓存
func (c *SwitchCache) cleanupExpired() {
	// 每分钟执行一次
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.doCleanup()
		case <-c.ctx.Done():
			logger.Logger.Debugf("Switch cache cleanup stopped by context")
			return
		case <-c.stopChan:
			logger.Logger.Debugf("Switch cache cleanup stopped by stop channel")
			return
		}
	}
}

// doCleanup 执行清理操作
func (c *SwitchCache) doCleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiredKeys := make([]string, 0)
	for key, item := range c.cache {
		if item.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.cache, key)
	}

	if len(expiredKeys) > 0 {
		logger.Logger.Debugf("Cleaned up %d expired cache entries", len(expiredKeys))
	}
}

// 全局缓存实例
var switchCache *SwitchCache

// StartSwitchCache 初始化全局缓存
func StartSwitchCache(ctx context.Context, cache *config.Cache) error {
	ttl, err := cache.GetCacheTime()
	if err != nil {
		return err
	}
	switchCache = NewSwitchCache(ctx, ttl)
	switchCache.start()
	return nil
}

// GetSwitchCache 获取全局缓存实例
func GetSwitchCache() *SwitchCache {
	return switchCache
}

// StopSwitchCache 停止全局缓存
func StopSwitchCache() {
	if switchCache != nil {
		switchCache.Stop()
	}
}
