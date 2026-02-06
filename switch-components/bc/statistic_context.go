package bc

import (
	"context"

	"gitee.com/fatzeng/switch-sdk-core/statistics"
)

type statisticsKey struct{}

func SetStatistics(ctx context.Context) context.Context {
	if _, ok := ctx.Value(statisticsKey{}).(*statistics.StatisticsWrapper); ok {
		return ctx
	}
	return context.WithValue(ctx, statisticsKey{}, statistics.NewStatisticsWrapper())
}

func GetStatistics(ctx context.Context) *statistics.StatisticsWrapper {
	if statisticsVal, ok := ctx.Value(statisticsKey{}).(*statistics.StatisticsWrapper); ok {
		return statisticsVal
	}
	return nil
}
