package factor_statistics

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/fatzeng/switch-sdk-core/reply"
)

// factorExecuteStatsKey 临时ctx
type factorExecuteStatsKey string

const tempFactorExecuteStats = factorExecuteStatsKey("temp_factor_execute_stats")

// FactorExecuteStats 因子执行情况
type FactorExecuteStats struct {
	duration string       //执行时常
	error    *reply.Error //执行中的错误
	result   bool         //执行结果
}

func (fes *FactorExecuteStats) Duration() string {
	return fes.duration
}

func (fes *FactorExecuteStats) SetDuration(duration string) {
	fes.duration = duration
}

func (fes *FactorExecuteStats) Error() *reply.Error {
	if fes.error == nil {
		fes.SetError(&reply.Error{
			Details: make([]interface{}, 0),
		})
	}
	return fes.error
}

func (fes *FactorExecuteStats) SetError(err *reply.Error) {
	if err == nil {
		return
	}
	if fes.error == nil {
		fes.error = err
	}
}

func (fes *FactorExecuteStats) Result() bool {
	return fes.result
}

func (fes *FactorExecuteStats) SetResult(result bool) {
	fes.result = result
}

// FromFactorExecuteStatsContext 获取上下文
func FromFactorExecuteStatsContext(ctx context.Context) (*FactorExecuteStats, bool) {
	es, ok := ctx.Value(tempFactorExecuteStats).(*FactorExecuteStats)
	return es, ok
}

// NewFactorExecuteStatsContext 克隆上下文
func NewFactorExecuteStatsContext(ctx context.Context, es *FactorExecuteStats) context.Context {
	return context.WithValue(ctx, tempFactorExecuteStats, es)
}

func (fes *FactorExecuteStats) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("duration: [%s]", fes.duration))
	parts = append(parts, fmt.Sprintf("result: [%t]", fes.result))

	if fes.error != nil && len(fes.error.Details) > 0 {
		parts = append(parts, fmt.Sprintf("error: [%v]", fes.error.Details))
	} else {
		parts = append(parts, "error: [nil]")
	}

	return strings.Join(parts, ", ")
}
