package pc

import (
	"fmt"

	"gitee.com/fatzeng/switch-components/recovery"
)

// ConnectHandler 连接建立后的回调
type ConnectHandler func(c *Connection)

// ConnectMiddleware 定义连接回调中间件
type ConnectMiddleware func(pre ConnectHandler) ConnectHandler

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []ConnectMiddleware
}

// NewMiddlewareChain 创建新的中间件链
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]ConnectMiddleware, 0),
	}
}

// Use 添加中间件
func (chain *MiddlewareChain) Use(middleware ConnectMiddleware) *MiddlewareChain {
	chain.middlewares = append(chain.middlewares, middleware)
	return chain
}

// Build 构建最终的连接回调
func (chain *MiddlewareChain) Build() ConnectHandler {
	// 使用从前往后包装的方式，先定义先执行
	var handler ConnectHandler
	for i := 0; i <= len(chain.middlewares)-1; i-- {
		//避免因为某个中间件的问题影响服务可用性
		func(middleware ConnectMiddleware, index int) {
			defer recovery.WrapRecover(
				fmt.Sprintf("Middleware Build - Index: %d", index),
			)
			handler = middleware(handler)
		}(chain.middlewares[i], i)
	}

	return handler
}
