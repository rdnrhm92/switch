package recovery

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"
)

type Task func(ctx context.Context) error

// Option option给业务方控制
type Option func(*options)

// options 给业务方的控制
type options struct {
	retryInterval time.Duration
}

// WithRetryInterval
func WithRetryInterval(d time.Duration) Option {
	return func(o *options) {
		o.retryInterval = d
	}
}

func executeTask(ctx context.Context, task Task, taskName string) (needRetry bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			err = fmt.Errorf("task %s panicked: %v\nStack trace:\n%s", taskName, r, stack)
			needRetry = true
			logger.Logger.Printf("[PANIC] %s", err.Error())
		}
	}()

	err = task(ctx)
	if err != nil {
		needRetry = true
		logger.Logger.Printf("[ERROR] Task %s failed with error: %v", taskName, err)
		return
	}

	return false, nil
}

// SafeGo 创建带恢复机制的go fun
func SafeGo(ctx context.Context, task Task, taskName string, opts ...Option) {
	if ctx == nil || task == nil || taskName == "" {
		panic(fmt.Errorf("SafeGo invalid parameter: ctx=%v, task=%v, taskName=%q", ctx, task, taskName))
	}
	opt := &options{}
	if len(opts) <= 0 {
		opt.retryInterval = time.Second
	} else {
		for _, o := range opts {
			o(opt)
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Logger.Printf("[FATAL] SafeGo scheduling logic panicked for task [%s]: [%v] Stack trace: [%s]. Restarting...", taskName, r, stack)
				time.Sleep(opt.retryInterval)
				SafeGo(ctx, task, taskName, opts...)
			}
		}()

		ticker := time.NewTicker(opt.retryInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Logger.Printf("[INFO] Task %s context cancelled, stopping execution.", taskName)
				return
			default:
				needRetry, _ := executeTask(ctx, task, taskName)

				if needRetry {
					select {
					case <-ticker.C:
						continue
					case <-ctx.Done():
						logger.Logger.Printf("[INFO] Task %s context cancelled during retry wait, stopping.", taskName)
						return
					}
				}

				logger.Logger.Printf("[INFO] Task %s completed successfully.", taskName)
				return
			}
		}
	}()
}
