package bc

import (
	"context"

	"gitee.com/fatzeng/switch-sdk-core/trace"
)

type tracerKey struct{}

func SetTrace(ctx context.Context) context.Context {
	if _, ok := ctx.Value(tracerKey{}).(*trace.TraceWrapper); ok {
		return ctx
	}
	return context.WithValue(ctx, tracerKey{}, trace.NewTraceWrapper())
}

func GetTrace(ctx context.Context) *trace.TraceWrapper {
	if traceVal, ok := ctx.Value(tracerKey{}).(*trace.TraceWrapper); ok {
		return traceVal
	}
	return nil
}
