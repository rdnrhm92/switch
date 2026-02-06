package request

import (
	"context"

	"gitee.com/fatzeng/switch-components/system"
	"gitee.com/fatzeng/switch-sdk-core/logger"
)

func Logger(ctx context.Context) logger.ILogger {
	if ctx == nil {
		return nil
	}

	if ctxLogger, ok := ctx.Value(system.X_Logger_Request).(logger.ILogger); ok {
		return ctxLogger
	}
	//使用默认的logger实现，不是request级别的日志实现
	return logger.Logger
}
