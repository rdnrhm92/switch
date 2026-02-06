package recovery

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
)

// PanicError 自定义error类型
type PanicError struct {
	Err           interface{}
	Stack         string
	Message       string
	Filename      string
	Line          int
	Function      string
	Additional    interface{}
	CancelMessage string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v\nfile: %s:%d\nfunc: %s\nmessage: %s\nadditional: %+v\nstack:\n%s",
		e.Err, e.Filename, e.Line, e.Function, e.Message, e.Additional, e.Stack)
}

type RecoverOptionFun = func(*RecoverOption)

// RecoverOption 记录项目
type RecoverOption struct {
	ErrorHandler func(error)
	PrintStack   bool
	Additional   interface{}
	Context      context.Context
}

// DefaultRecoverOption 默认的记录选项
var DefaultRecoverOption = RecoverOption{
	ErrorHandler: func(err error) {
		fmt.Printf("recovered from panic: %+v\n", err)
	},
	PrintStack: true,
}

// WrapRecover 对recover函数的一个封装提供可多个RecoverOption可扩展的逻辑
func WrapRecover(message string, opts ...RecoverOptionFun) {
	var opt RecoverOption
	if len(opts) > 0 {
		for _, option := range opts {
			option(&opt)
		}
	} else {
		opt = DefaultRecoverOption
	}

	if r := recover(); r != nil {
		var filename, function string
		var line int
		pc, file, l, ok := runtime.Caller(2)
		if ok {
			filename = file
			line = l
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				function = fn.Name()
			}
		}

		err := &PanicError{
			Err:        r,
			Stack:      string(debug.Stack()),
			Message:    message,
			Filename:   filename,
			Line:       line,
			Function:   function,
			Additional: opt.Additional,
		}

		if !opt.PrintStack {
			err.Stack = ""
		}

		if opt.ErrorHandler != nil {
			opt.ErrorHandler(err)
		}

		if opt.Context != nil {
			select {
			case <-opt.Context.Done():
				err.CancelMessage = fmt.Sprintf("context cancelled during panic recovery: %v\n", opt.Context.Err())
			default:
			}
		}
	}
}

// WithContext 添加context到恢复选项
func WithContext(ctx context.Context) func(*RecoverOption) {
	return func(opt *RecoverOption) {
		opt.Context = ctx
	}
}

// WithErrorHandler 添加自定义错误处理函数
func WithErrorHandler(handler func(error)) func(*RecoverOption) {
	return func(opt *RecoverOption) {
		opt.ErrorHandler = handler
	}
}

// WithAdditional 添加附加信息
func WithAdditional(additional interface{}) func(*RecoverOption) {
	return func(opt *RecoverOption) {
		opt.Additional = additional
	}
}
