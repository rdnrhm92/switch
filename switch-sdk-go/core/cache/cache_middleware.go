package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"

	"gitee.com/fatzeng/switch-sdk-core/model"
	"gitee.com/fatzeng/switch-sdk-go/core/factor_statistics"
	"gitee.com/fatzeng/switch-sdk-go/core/middleware"
	"golang.org/x/sync/singleflight"
)

// FactorCacheMiddleware 因子执行缓存中间件，应对两种场景：
// 1.单次请求，一个开关内配置了多个相同的因子(包括因子配置也相同)尽管这个场景并不常见
// 2.单次请求，内部开启协程并发执行一个开关，开关内部的因子执行结果将会缓存，
// 条件是开关名(一次请求执行了多个开关,开关内部有同名因子,所以要加开关名)因子名和执行条件
// 如果开关只是简单的context获取值并匹配，那么不管场景一还是场景二都不要开启这个中间件
func FactorCacheMiddleware() middleware.Middleware {
	var sf singleflight.Group
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, switchRule *model.SwitchModel, factorRule *model.RuleNode) bool {
			// 判断是否开启了缓存(配置判断)
			if !switchRule.UseCache {
				return next(ctx, switchRule, factorRule)
			}
			// 判断是否开启了缓存(请求判断)
			fc, fcOK := FromContext(ctx)
			if !fcOK {
				return next(ctx, switchRule, factorRule)
			}

			statsContext, esOK := factor_statistics.FromFactorExecuteStatsContext(ctx)

			// 生成单飞用的key
			key, err := generateSingleFlightKey(switchRule.Name, factorRule.Factor, factorRule.Config)
			if err != nil {
				if esOK {
					statsContext.Error().WithDetails(fmt.Sprintf("FactorCacheMiddleware: failed to generate key: %v", err))
				}
				return false
			}

			v, doErr, _ := sf.Do(key, func() (interface{}, error) {
				//缓存项检查
				if result, found := fc.Get(switchRule.Name, factorRule); found {
					return result, nil
				}

				//缓存没有命中执行因子
				result := next(ctx, switchRule, factorRule)

				//设置因子执行结果
				fc.Set(switchRule.Name, factorRule.Factor, factorRule.Config, result)

				return result, nil
			})

			if doErr != nil {
				if esOK {
					statsContext.Error().WithDetails(fmt.Sprintf("FactorCacheMiddleware: singleflight error: %v", doErr))
				}
				return false
			}

			result, ok := v.(bool)
			if !ok {
				if esOK {
					statsContext.Error().WithDetails(fmt.Sprintf("FactorCacheMiddleware: unexpected type from singleflight: got %T", v))
				}
				return false
			}

			return result
		}
	}
}

// generateSingleFlightKey 构建一个单飞key
func generateSingleFlightKey(switchName, factorName string, condition interface{}) (string, error) {
	h := fnv.New64a()
	var conditionStr string
	switch v := condition.(type) {
	case string:
		conditionStr = v
	case []byte:
		conditionStr = string(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		conditionStr = fmt.Sprintf("%d", v)
	case float32, float64:
		conditionStr = fmt.Sprintf("%f", v)
	case bool:
		conditionStr = strconv.FormatBool(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal condition to JSON: %w", err)
		}
		conditionStr = string(b)
	}

	key := fmt.Sprintf("%s#%s#%s", switchName, factorName, conditionStr)
	_, err := h.Write([]byte(key))
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(h.Sum64(), 10), nil
}
