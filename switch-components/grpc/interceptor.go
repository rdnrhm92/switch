package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-components/bc"
	"gitee.com/fatzeng/switch-components/logging/request"
	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-components/snowflake"
	"gitee.com/fatzeng/switch-components/system"
	"gitee.com/fatzeng/switch-sdk-core/config"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	cerr "gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-core/resp"
	"gitee.com/fatzeng/switch-sdk-core/statistics"
	"gitee.com/fatzeng/switch-sdk-core/tool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var generate *snowflake.BusinessGenerator

func init() {
	generator, err := snowflake.NewBusinessGenerator(snowflake.BusinessConfig{
		BusinessType: "rpc",
		MachineID:    1,
		Prefix:       "ID:",
	})
	if err != nil {
		panic(err)
	}
	generate = generator
}

// 工具函数合并多个拦截器
func UnaryServerInterceptorChain(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildChain(interceptors[i], chain, info)
		}
		return chain(ctx, req)
	}
}

// buildChain 构建拦截器链
func buildChain(interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler, info *grpc.UnaryServerInfo) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptor(ctx, req, info, handler)
	}
}

// ResponseSetInterceptor 公共的响应处理
func ResponseSetInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//初始化一个公共的响应结果
		ctx = bc.NewRespContext(ctx)
		return handler(ctx, req)
	}
}

// MetadataInterceptor 元信息拦截器 这个放在拦截器最前面，可以辅助构建一些元信息，方便排查问题
func MetadataInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//设置一个请求ID，可以根据这个ID进行请求内链路的追踪
		//后续可以搭配开关系统，在此处设置一些定制化的(请求级别)参数，用开关做控制
		newCtx, _ := GetRequestID(ctx)
		newCtx, _ = GetTraceID(newCtx)
		newCtx, _ = GetServiceName(newCtx)
		newCtx, open := GetDebugSwitch(newCtx)
		bc.GetDebugInfo(newCtx).Enabled = open
		return handler(newCtx, req)
	}
}

// LoggingInterceptor 日志拦截器 会执行一些统计信息
// 继续封装请求级别的logger 携带通用信息便于排查问题
func LoggingInterceptor(logger logger.ILogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//请求ID
		ctx, requestID := GetRequestID(ctx)
		//traceID
		ctx, traceID := GetTraceID(ctx)
		//用户信息
		clientPeer := GetClientPeer(ctx)
		//服务信息
		ctx, serviceName := GetServiceName(ctx)
		//ip信息
		ip, _ := tool.GetLocalIP()
		//环境信息
		env := config.GetOsEnvironment()
		//版本信息
		ctx, version := GetVersion(ctx)
		//hostname信息
		ctx, hostName := GetHostName(ctx)
		//span信息
		spanId := tool.GenerateSpanId()

		// 创建新的logger，这个logger会在每条日志中都带上这些信息
		ctxLogger := logger.With(
			map[string]interface{}{
				system.X_Request_Id:   requestID,
				system.X_Trace_Id:     traceID,
				system.X_Client_Msg:   clientPeer,
				system.X_Service_Name: serviceName,
				system.X_Ip:           ip,
				system.X_Env:          env,
				system.X_Version:      version,
				system.X_Host_Name:    hostName,
				system.X_Span_Id:      spanId,
			},
		)
		ctx = context.WithValue(ctx, system.X_Logger_Request, ctxLogger)

		//维护响应上游的信息
		ctx = bc.NewRespContext(ctx)

		//trace信息
		traceMember := bc.GetTrace(ctx)
		traceMember.SetRequestId(requestID)
		traceMember.SetTraceId(traceID)
		traceMember.SetServiceName(serviceName)
		traceMember.SetSpanId(spanId)
		traceMember.SetEnv(string(env))
		traceMember.SetVersion(version)
		traceMember.SetHostName(hostName)
		traceMember.SetIp(ip)

		//耗时统计信息
		statisticsWrapper := bc.GetStatistics(ctx)
		statisticsWrapper.From = clientPeer.Address
		statisticsWrapper.To = ip

		//日志落库可以回溯
		ctxLogger.Infof("[gRPC Request Start] Method: %s, RequestID: %s, Client: %+v, Request: %+v",
			info.FullMethod, requestID, clientPeer, req)

		// 耗时统计(毫秒级别)
		start := statistics.TimerBeginWithName(info.FullMethod, statistics.Millisecond)
		statisticsWrapper.RequestTime = start.Begin.UnixMilli()

		handleResp, err := handler(ctx, req)

		respMember := bc.GetResp(ctx)

		// 计算耗时
		duration := start.Complete()
		formatRes := duration.Format()
		statisticsWrapper.ResponseTime = duration.End.UnixMilli()
		executeTime := duration.ExecuteTime.Milliseconds()
		statisticsWrapper.ExecuteTime = &executeTime

		// 获取错误状态
		var cErr *cerr.Error
		statusCode := codes.OK
		statusMsg := "success"
		var data interface{}
		if err != nil {
			//包装成grpc error
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
				statusMsg = st.Message()
				data = new(interface{})
			} else if errors.As(err, &cErr) {
				statusCode = codes.Code(cErr.Code)
				statusMsg = err.Error()
			} else {
				statusCode = codes.Code(400)
				statusMsg = err.Error()
				data = new(interface{})
			}
		} else {
			data = handleResp
		}

		respMember.SetData(data)
		respMember.SetCode(resp.Code(statusCode))
		respMember.SetMessage(resp.Message(statusMsg))

		// 记录普通日志
		ctxLogger.Infof("[gRPC Request End] Method: %s, RequestID: %s, Client: %+v, Duration: %s, StatusCode: %s, StatusMsg: %s, Response: %+v",
			info.FullMethod, requestID, clientPeer, formatRes, statusCode, statusMsg, respMember)

		// 记录错误日志
		if err != nil {
			ctxLogger.Errorf("[gRPC Error] Method: %s, RequestID: %s, Client: %+v, Error: %v",
				info.FullMethod, requestID, clientPeer, err)
		}

		return bc.Response(ctx), nil
	}
}

// RecoveryInterceptor recover拦截器
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//请求ID
		ctx, requestID := GetRequestID(ctx)
		//用户信息
		clientPeer := GetClientPeer(ctx)

		defer recovery.WrapRecover(
			fmt.Sprintf("Method: %s, RequestID: %s", info.FullMethod, requestID),
			recovery.WithErrorHandler(func(recoverErr error) {
				sprintfErr := fmt.Sprintf("[gRPC Panic] Method: %s, RequestID: %s, Client: %+v, Request: %+v, Error: %v",
					info.FullMethod, requestID, clientPeer, req, recoverErr)
				request.Logger(ctx).Errorf(sprintfErr)
				bc.GetDebugInfo(ctx).AddErrorf(sprintfErr)
				err = status.Errorf(codes.Internal, sprintfErr)
			}),
			recovery.WithAdditional(map[string]interface{}{
				"method":    info.FullMethod,
				"requestID": requestID,
				"client":    clientPeer,
				"request":   req,
				"err":       err,
			}),
		)
		return handler(ctx, req)
	}
}

// ErrorHandlingInterceptor 创建一个用于统一错误处理的服务器拦截器
func ErrorHandlingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//请求ID
		ctx, requestID := GetRequestID(ctx)
		//用户信息
		clientPeer := GetClientPeer(ctx)

		respMember, err := handler(ctx, req)
		if err != nil {
			//封装grpc的错误
			if st, ok := status.FromError(err); ok {
				errDetails := fmt.Sprintf("[gRPC Error] Method: %s, RequestID: %s, Client: %+v, Request: %+v, Code: %s, Message: %s",
					info.FullMethod, requestID, clientPeer, req, st.Code(), st.Message())
				bc.GetDebugInfo(ctx).AddErrorf(errDetails)
				request.Logger(ctx).Errorf(errDetails)
				return respMember, err
			} else {
				// 其他错误转换为 gRPC 错误
				errDetails := fmt.Sprintf("[gRPC Error] Method: %s, RequestID: %s, Client: %+v, Request: %+v, Error: %v",
					info.FullMethod, requestID, clientPeer, req, err)
				bc.GetDebugInfo(ctx).AddErrorf(errDetails)
				request.Logger(ctx).Errorf(errDetails)
				return nil, status.Error(codes.Internal, errDetails)
			}
		}
		return respMember, nil
	}
}

// TimeoutInterceptor 超时拦截器
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if timeout <= 0 {
			return handler(ctx, req)
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		//请求ID
		ctx, requestID := GetRequestID(ctx)
		//用户信息
		clientPeer := GetClientPeer(ctx)

		defer recovery.WrapRecover(
			fmt.Sprintf("Method: %s, RequestID: %s", info.FullMethod, requestID),
			recovery.WithErrorHandler(func(recoverErr error) {
				sprintfErr := fmt.Sprintf("[gRPC Timeout Panic] Method: %s, RequestID: %s, Client: %+v, Request: %+v, Error: %v",
					info.FullMethod, requestID, clientPeer, req, recoverErr)
				request.Logger(ctx).Errorf(sprintfErr)
				err = status.Errorf(codes.Internal, sprintfErr)
			}),
			recovery.WithAdditional(map[string]interface{}{
				"method":    info.FullMethod,
				"requestID": requestID,
				"client":    clientPeer,
				"request":   req,
				"timeout":   timeout,
			}),
		)

		resp, err = handler(timeoutCtx, req)

		// 如果超时了做超时的处理
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			timeoutErr := fmt.Sprintf("[gRPC Timeout] Method: %s, RequestID: %s, Client: %+v, Request: %+v, Timeout: %v",
				info.FullMethod, requestID, clientPeer, req, timeout)
			request.Logger(ctx).Warnf(timeoutErr)
			return nil, status.Error(codes.DeadlineExceeded, timeoutErr)
		}

		return resp, err
	}
}

// StreamServerInterceptorChain 合并多个流式拦截器
func StreamServerInterceptorChain(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildStreamChain(interceptors[i], chain, info)
		}
		return chain(srv, ss)
	}
}

// buildStreamChain 构建流式拦截器链
func buildStreamChain(interceptor grpc.StreamServerInterceptor, handler grpc.StreamHandler, info *grpc.StreamServerInfo) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		return interceptor(srv, stream, info, handler)
	}
}
