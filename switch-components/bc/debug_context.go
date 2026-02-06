package bc

import (
	"context"

	"gitee.com/fatzeng/switch-sdk-core/debug"
)

type debugKey struct{}

func SetDebug(ctx context.Context) context.Context {
	if _, ok := ctx.Value(debugKey{}).(*debug.DebugInfoWrapper); ok {
		return ctx
	}
	return context.WithValue(ctx, debugKey{}, debug.NewDebugInfoWrapper())
}

func GetDebugInfo(ctx context.Context) *debug.DebugInfoWrapper {
	if debugVal, ok := ctx.Value(debugKey{}).(*debug.DebugInfoWrapper); ok {
		return debugVal
	}
	return nil
}
