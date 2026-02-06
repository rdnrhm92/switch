package http

import (
	"context"
	"net/http"

	"gitee.com/fatzeng/switch-components/system"
)

// GetRequestID 从上下文或请求头中获取请求ID
func GetRequestID(r *http.Request) (context.Context, string) {
	ctx := r.Context()
	// 优先从请求头中获取
	if requestID := r.Header.Get(system.X_Request_Id); requestID != "" {
		return context.WithValue(ctx, system.X_Request_Id, requestID), requestID
	}

	// 其次从上下文中获取
	if requestID, ok := ctx.Value(system.X_Request_Id).(string); ok && requestID != "" {
		return ctx, requestID
	}

	// 生成新的请求ID
	requestID := generate.GeneratePrefixID()
	return context.WithValue(ctx, system.X_Request_Id, requestID), requestID
}

// GetDebugSwitch 获取debug开关
func GetDebugSwitch(r *http.Request) (context.Context, bool) {
	ctx := r.Context()
	var debug string
	// 优先从请求头中获取
	if debug = r.Header.Get(system.X_Debug); debug != "" {
		return context.WithValue(ctx, system.X_Debug, debug), debug == "true" || debug == "True"
	}

	// 其次从上下文中获取
	if debug, ok := ctx.Value(system.X_Debug).(string); ok && debug != "" {
		return ctx, debug == "true" || debug == "True"
	}

	debug = system.HttpDebug
	return context.WithValue(ctx, system.X_Debug, debug), debug == "true" || debug == "True"
}

// GetTraceID 从上下文或请求头中获取追踪ID
func GetTraceID(r *http.Request) (context.Context, string) {
	ctx := r.Context()
	// 优先从请求头中获取
	if traceID := r.Header.Get(system.X_Trace_Id); traceID != "" {
		return context.WithValue(ctx, system.X_Trace_Id, traceID), traceID
	}

	// 其次从上下文中获取
	if traceID, ok := ctx.Value(system.X_Trace_Id).(string); ok && traceID != "" {
		return ctx, traceID
	}

	// 如果没有，使用请求ID作为追踪ID
	ctx, requestID := GetRequestID(r)
	return context.WithValue(ctx, system.X_Trace_Id, requestID), requestID
}

// GetServiceName 从上下文或请求头中获取服务名称
func GetServiceName(r *http.Request) (context.Context, string) {
	ctx := r.Context()
	// 优先从请求头中获取
	if serviceName := r.Header.Get(system.X_Service_Name); serviceName != "" {
		return context.WithValue(ctx, system.X_Service_Name, serviceName), serviceName
	}

	// 其次从上下文中获取
	if serviceName, ok := ctx.Value(system.X_Service_Name).(string); ok && serviceName != "" {
		return ctx, serviceName
	}

	return context.WithValue(ctx, system.X_Service_Name, system.HttpServiceName), system.HttpServiceName
}

// GetHostName 从上下文或请求头中获取主机名
func GetHostName(r *http.Request) (context.Context, string) {
	ctx := r.Context()
	// 优先从请求头中获取
	if hostName := r.Header.Get(system.X_Host_Name); hostName != "" {
		return context.WithValue(ctx, system.X_Host_Name, hostName), hostName
	}

	// 其次从上下文中获取
	if hostName, ok := ctx.Value(system.X_Host_Name).(string); ok && hostName != "" {
		return ctx, hostName
	}

	return context.WithValue(ctx, system.X_Host_Name, system.HttpHostName), system.HttpHostName
}

// GetVersion 从上下文或请求头中获取版本信息
func GetVersion(r *http.Request) (context.Context, string) {
	ctx := r.Context()
	// 优先从请求头中获取
	if version := r.Header.Get(system.X_Version); version != "" {
		return context.WithValue(ctx, system.X_Version, version), version
	}

	// 其次从上下文中获取
	if version, ok := ctx.Value(system.X_Version).(string); ok && version != "" {
		return ctx, version
	}

	return context.WithValue(ctx, system.X_Version, system.HttpVersion), system.HttpVersion
}

// GetClientIP 获取客户端IP
func GetClientIP(r *http.Request) string {
	// 尝试从各种头部获取真实IP
	for _, header := range []string{"X-Real-IP", "X-Forwarded-For"} {
		if ip := r.Header.Get(header); ip != "" {
			return ip
		}
	}
	// 如果没有代理头部，则使用远程地址
	return r.RemoteAddr
}
