package middleware

import (
	"context"
	"sync"

	"gitee.com/fatzeng/switch-sdk-core/model"
)

// Handler 中间件处理器
type Handler func(ctx context.Context, switchRule *model.SwitchModel, factorRule *model.RuleNode) bool

// Middleware 中间件
type Middleware func(next Handler) Handler

// middlewareManager 中间件集合
var middlewareManager []Middleware
var mu sync.Mutex

// WithMiddleware 设置中间件
func WithMiddleware(middleware Middleware) {
	mu.Lock()
	defer mu.Unlock()
	middlewareManager = append(middlewareManager, middleware)
}

// Middlewares 获取中间件
func Middlewares() []Middleware {
	mu.Lock()
	defer mu.Unlock()
	dstMiddlewares := make([]Middleware, len(middlewareManager))
	copy(dstMiddlewares, middlewareManager)
	return dstMiddlewares
}
