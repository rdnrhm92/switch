package cache

import (
	"context"
	"reflect"
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-go/core/middleware"
)

type cacheKey string

const factorCacheKey = cacheKey("FactorCacheKey")

// FactorCacheKey 缓存key
type FactorCacheKey struct {
	SwitchName string
	FactorName string
}

type FactorCache struct {
	mu           sync.RWMutex
	factorsCache map[FactorCacheKey]*factorCacheItem
}

type factorCacheItem struct {
	Condition interface{}
	Result    bool
}

// NewFactorCache 创建新的FactorCache
func NewFactorCache() *FactorCache {
	return &FactorCache{
		factorsCache: make(map[FactorCacheKey]*factorCacheItem),
	}
}

// Set 给缓存中增加一个项目
func (f *FactorCache) Set(switchName, factorName string, condition interface{}, result bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := FactorCacheKey{SwitchName: switchName, FactorName: factorName}
	f.factorsCache[key] = &factorCacheItem{Condition: condition, Result: result}
}

// Get 获取因子的缓存项(bool(缓存项的缓存结果), bool(缓存项是否存在))
func (f *FactorCache) Get(switchName string, factorRule *model.RuleNode) (bool, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	key := FactorCacheKey{SwitchName: switchName, FactorName: factorRule.Factor}
	item, ok := f.factorsCache[key]

	//不存在或者因子执行参数不同都要重新执行因子，则缓存项为false表示不存在
	if !ok || !reflect.DeepEqual(item.Condition, factorRule.Config) {
		return false, false
	}

	return item.Result, true
}

// UseCache 使用缓存功能
func UseCache(ctx context.Context) context.Context {
	middleware.WithMiddleware(FactorCacheMiddleware())
	return context.WithValue(ctx, factorCacheKey, NewFactorCache())
}

// FromContext 从ctx中获取缓存配置
func FromContext(ctx context.Context) (*FactorCache, bool) {
	cache, ok := ctx.Value(factorCacheKey).(*FactorCache)
	return cache, ok
}
