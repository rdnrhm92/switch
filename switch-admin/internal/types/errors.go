package types

import (
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/reply"
)

// 错误码定义
const (
	// 参数不合法
	invalidArgument int32 = 1

	// 不存在该业务逻辑
	methodNotFound int32 = 2

	// 数据为空
	empty int32 = 4

	// 无权限
	noPermissions int32 = 26

	// 认证错误
	noAuth int32 = 27

	// 系统繁忙
	busy int32 = 39

	// 账号密码错误
	login int32 = 40

	// 注册失败
	register int32 = 41
)

func formatMessage(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}
	return ": " + fmt.Sprint(args...)
}

// 提供一些工具函数可以快速的构建异常 对标php
var (
	// InvalidArgument 参数不合法
	InvalidArgument = func(args ...interface{}) *reply.Error {
		return reply.New(invalidArgument, "参数不合法"+formatMessage(args...))
	}

	// MethodNotFound 不存在该业务逻辑
	MethodNotFound = func(args ...interface{}) *reply.Error {
		return reply.New(methodNotFound, "不存在该业务逻辑"+formatMessage(args...))
	}

	// Empty 数据为空
	Empty = func(args ...interface{}) *reply.Error {
		return reply.New(empty, "数据为空"+formatMessage(args...))
	}

	// Busy 系统繁忙
	Busy = func(args ...interface{}) *reply.Error {
		return reply.New(busy, "系统繁忙"+formatMessage(args...))
	}

	// NoPermissions 无权限
	NoPermissions = func(args ...interface{}) *reply.Error {
		return reply.New(noPermissions, "无权限"+formatMessage(args...))
	}

	// NoAuth 无认证
	NoAuth = func(args ...interface{}) *reply.Error {
		return reply.New(noPermissions, "无认证"+formatMessage(args...))
	}

	// Login 账号密码错误
	Login = func(args ...interface{}) *reply.Error {
		return reply.New(login, "账号或密码错误"+formatMessage(args...))
	}

	// Register 注册失败
	Register = func(args ...interface{}) *reply.Error {
		return reply.New(register, "注册失败"+formatMessage(args...))
	}
)

const SuccessCode = 0

// Success 成功
func Success(data interface{}) *reply.SuccessResponse {
	return &reply.SuccessResponse{
		BaseResponse: reply.BaseResponse{
			Code:    SuccessCode,
			Message: "Success",
		},
		Data: data,
	}
}
