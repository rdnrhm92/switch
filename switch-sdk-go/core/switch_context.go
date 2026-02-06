package core

import (
	"context"
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-go/core/factor_statistics"
)

// contextKey 开关内部SwitchContext不用于函数传递
type contextKey string

const switchContextKey = contextKey("SwitchContext")

// FactorExecutionRecord 因子执行情况
type FactorExecutionRecord struct {
	Name string
	Stat *factor_statistics.FactorExecuteStats
}

// SwitchContext 一个开关执行的上下文
type SwitchContext struct {
	context.Context
	mu                      sync.RWMutex
	requestID               string                   // 请求ID
	executionRecords        []*FactorExecutionRecord // 因子的执行记录序列(为了可以记录一个开关内的多个相同因子的执行情况使用slice)
	switchExecutionDuration string                   // 开关的执行耗时
	switchExecutionResult   bool                     // 开关的执行结果
	error                   *reply.Error             // 执行过程中遇到的错误
	executionStatsOpen      bool                     // 是否开启统计项
}

func (sc *SwitchContext) RequestID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.requestID
}

func (sc *SwitchContext) SetRequestID(requestID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.requestID = requestID
}

func (sc *SwitchContext) ExecutionRecords() []*FactorExecutionRecord {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	recordsCopy := make([]*FactorExecutionRecord, len(sc.executionRecords))
	copy(recordsCopy, sc.executionRecords)
	return recordsCopy
}

// UseFactorExecutionRecord 在因子执行前使用，每次调用都会新增一个因子执行分析
func (sc *SwitchContext) UseFactorExecutionRecord(factorName string) *factor_statistics.FactorExecuteStats {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.executionRecords == nil {
		sc.executionRecords = make([]*FactorExecutionRecord, 0)
	}
	stat := &factor_statistics.FactorExecuteStats{}
	stat.SetError(&reply.Error{
		Details: []interface{}{},
	})
	sc.executionRecords = append(sc.executionRecords, &FactorExecutionRecord{Name: factorName, Stat: stat})
	return stat
}

func (sc *SwitchContext) SwitchExecutionDuration() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.switchExecutionDuration
}

func (sc *SwitchContext) SetSwitchExecutionDuration(duration string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.switchExecutionDuration = duration
}

func (sc *SwitchContext) SwitchExecutionResult() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.switchExecutionResult
}

func (sc *SwitchContext) SetSwitchExecutionResult(switchExecutionResult bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.switchExecutionResult = switchExecutionResult
}

func (sc *SwitchContext) Error() *reply.Error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.error == nil {
		sc.error = &reply.Error{
			Details: make([]interface{}, 0),
		}
	}
	return sc.error
}

func (sc *SwitchContext) SetError(err *reply.Error) {
	if err == nil {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.error == nil {
		sc.error = err
	}
}

func (sc *SwitchContext) ExecutionStatsOpen() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.executionStatsOpen
}

func (sc *SwitchContext) SetExecutionStatsOpen(executionStatsOpen bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.executionStatsOpen = executionStatsOpen
}

func NewSwitchContext(parent context.Context) *SwitchContext {
	if parent == nil {
		parent = context.Background()
	}
	sc := &SwitchContext{
		Context:          parent,
		executionRecords: make([]*FactorExecutionRecord, 0),
	}
	sc.Context = context.WithValue(parent, switchContextKey, sc)
	return sc
}

// FromContext 获取上下文
func FromContext(ctx context.Context) (*SwitchContext, bool) {
	sc, ok := ctx.Value(switchContextKey).(*SwitchContext)
	return sc, ok
}
