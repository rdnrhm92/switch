package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitee.com/fatzeng/switch-components/bc"
	"gitee.com/fatzeng/switch-components/logging/request"
	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-components/snowflake"
	"gitee.com/fatzeng/switch-components/system"
	"gitee.com/fatzeng/switch-sdk-core/config"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-core/resp"
	"gitee.com/fatzeng/switch-sdk-core/statistics"
	"gitee.com/fatzeng/switch-sdk-core/tool"
	"github.com/zeromicro/go-zero/rest"
)

// http的中间件使用go-http原生，如果需要嵌套使用web框架需要做一次转换
// http中间件规则跟grpc基本保持一致
var generate *snowflake.BusinessGenerator

func init() {
	generator, err := snowflake.NewBusinessGenerator(snowflake.BusinessConfig{
		BusinessType: "http",
		MachineID:    1,
		Prefix:       "ID:",
	})
	if err != nil {
		panic(err)
	}
	generate = generator
}

type BaseMiddleware struct {
	handler func(*responseWriter, *http.Request, http.Handler)
	rw      *responseWriter
}

// GoZeroMiddleware 转换为 go-zero 中间件
// 参见Middleware func(next http.HandlerFunc) http.HandlerFunc
func (m *BaseMiddleware) GoZeroMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			//多个中间件需要使用同一个rw所以传递
			type responseWriterKey struct{}
			var rw *responseWriter
			if existingRw, ok := r.Context().Value(responseWriterKey{}).(*responseWriter); ok {
				rw = existingRw
			} else {
				rw = &responseWriter{
					ResponseWriter: w,
					statusCode:     http.StatusOK,
				}
				//重新更新一下request
				r = r.WithContext(context.WithValue(r.Context(), responseWriterKey{}, rw))
			}

			wrapper := bc.GetResp(r.Context())
			if wrapper == nil {
				//同样对响应结果也初始化一下
				ctx := bc.NewRespContext(r.Context())
				r = r.WithContext(ctx)
			}

			m.handler(rw, r, next)
		}
	}
}

// MetadataMiddleware 创建元数据中间件
func MetadataMiddleware() *BaseMiddleware {
	return &BaseMiddleware{
		handler: func(w *responseWriter, r *http.Request, next http.Handler) {
			// 设置请求相关的元数据
			newCtx, _ := GetRequestID(r)
			newCtx, _ = GetTraceID(r)
			newCtx, _ = GetServiceName(r)
			newCtx, open := GetDebugSwitch(r)
			bc.GetDebugInfo(newCtx).Enabled = open

			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		},
	}
}

func LoggingMiddleware(logger logger.ILogger) *BaseMiddleware {
	return &BaseMiddleware{
		handler: func(w *responseWriter, r *http.Request, next http.Handler) {
			ctx := r.Context()

			ctx, requestID := GetRequestID(r)
			ctx, traceID := GetTraceID(r)
			ctx, serviceName := GetServiceName(r)
			ip, _ := tool.GetLocalIP()
			env := config.GetOsEnvironment()
			ctx, version := GetVersion(r)
			ctx, hostName := GetHostName(r)
			spanID := tool.GenerateSpanId()
			clientIP := r.RemoteAddr

			// 创建带上下文的logger
			ctxLogger := logger.With(
				map[string]interface{}{
					system.X_Request_Id:   requestID,
					system.X_Trace_Id:     traceID,
					system.X_Client_Msg:   clientIP,
					system.X_Service_Name: serviceName,
					system.X_Ip:           ip,
					system.X_Env:          env,
					system.X_Version:      version,
					system.X_Host_Name:    hostName,
					system.X_Span_Id:      spanID,
				},
			)
			ctx = context.WithValue(ctx, system.X_Logger_Request, ctxLogger)

			// 维护响应上游的信息
			traceMember := bc.GetTrace(ctx)
			traceMember.SetRequestId(requestID)
			traceMember.SetTraceId(traceID)
			traceMember.SetServiceName(serviceName)
			traceMember.SetSpanId(spanID)
			traceMember.SetEnv(string(env))
			traceMember.SetVersion(version)
			traceMember.SetHostName(hostName)
			traceMember.SetIp(ip)

			// 耗时统计信息
			statisticsWrapper := bc.GetStatistics(ctx)
			statisticsWrapper.From = clientIP
			statisticsWrapper.To = ip

			// 记录请求开始日志
			ctxLogger.Infof("[HTTP Request Start] Method: %s, Path: %s, RequestID: %s, Client: %s, Request: %+v",
				r.Method, r.URL.Path, requestID, clientIP, r)

			// 耗时统计
			start := statistics.TimerBeginWithName(r.URL.Path, statistics.Millisecond)
			statisticsWrapper.RequestTime = start.Begin.UnixMilli()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

			// 计算耗时
			duration := start.Complete()
			formatRes := duration.Format()
			statisticsWrapper.ResponseTime = duration.End.UnixMilli()
			executeTime := duration.ExecuteTime.Milliseconds()
			statisticsWrapper.ExecuteTime = &executeTime

			// 记录响应日志
			ctxLogger.Infof("[HTTP Request End] Method: %s, Path: %s, RequestID: %s, Client: %s, Duration: %s, StatusCode: %+v",
				r.Method, r.URL.Path, requestID, clientIP, formatRes, w.statusCode)
		},
	}
}

func RecoveryMiddleware() *BaseMiddleware {
	return &BaseMiddleware{
		handler: func(w *responseWriter, r *http.Request, next http.Handler) {
			ctx := r.Context()
			ctx, requestID := GetRequestID(r)
			clientIP := r.RemoteAddr

			defer recovery.WrapRecover(
				fmt.Sprintf("Path: %s, RequestID: %s", r.URL.Path, requestID),
				recovery.WithErrorHandler(func(recoverErr error) {
					sprintfErr := fmt.Errorf("[HTTP Panic] Path: %s, RequestID: %s, Client: %s, Error: %v",
						r.URL.Path, requestID, clientIP, recoverErr)
					request.Logger(ctx).Error(sprintfErr)
					bc.GetDebugInfo(ctx).AddError(sprintfErr)
					w.code = resp.Code(400)
					w.message = resp.Message(sprintfErr.Error())
					w.data = ""
					w.statusCode = http.StatusInternalServerError
				}),
				recovery.WithAdditional(map[string]interface{}{
					"path":      r.URL.Path,
					"requestID": requestID,
					"client":    clientIP,
					"method":    r.Method,
				}),
			)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		},
	}
}

func TimeoutMiddleware(timeout time.Duration) *BaseMiddleware {
	return &BaseMiddleware{
		handler: func(w *responseWriter, r *http.Request, next http.Handler) {
			if timeout <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx, requestID := GetRequestID(r)
			clientIP := r.RemoteAddr

			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(timeoutCtx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-timeoutCtx.Done():
				if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
					timeoutErr := fmt.Sprintf("[HTTP Timeout] Path: %s, RequestID: %s, Client: %s, Timeout: %v",
						r.URL.Path, requestID, clientIP, timeout)
					request.Logger(ctx).Warnf(timeoutErr)
					http.Error(w, timeoutErr, http.StatusGatewayTimeout)
				}
			}
		},
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	code       resp.Code
	message    resp.Message
	data       interface{}
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	//这里先缓存一下服务端返回的结果
	if !rw.written {
		var errObj reply.Error
		statusCode := resp.Code(0)
		statusMsg := "success"
		var data interface{}

		updateResponse := func(code resp.Code, msg string, responseData interface{}) {
			rw.code = code
			rw.message = resp.Message(msg)
			rw.data = responseData
		}

		if json.Unmarshal(b, &errObj) == nil {
			statusCode = resp.Code(errObj.Code)
			statusMsg = errObj.Message
		} else {
			var respData map[string]interface{}
			if json.Unmarshal(b, &respData) == nil {
				data = respData
			} else {
				data = string(b)
			}
			if rw.statusCode == 0 {
				rw.statusCode = http.StatusOK
			}
			statusCode = resp.Code(rw.statusCode)
			statusMsg = "success"
		}
		if data == nil {
			data = ""
		}
		if statusMsg == "" {
			statusMsg = ""
		}
		updateResponse(statusCode, statusMsg, data)
		return len(b), nil
	}
	return len(b), nil
}

func (rw *responseWriter) WriteResponse(ctx context.Context) {
	if !rw.written {
		rw.Header().Set("Content-Type", "application/json")
		respMember := bc.GetResp(ctx).SetData(rw.data).SetCode(rw.code).SetMessage(rw.message).ToResp()
		rw.ResponseWriter.WriteHeader(rw.statusCode)
		json.NewEncoder(rw.ResponseWriter).Encode(respMember)
		rw.written = true
	}
}

// 一个专门用于response的拦截器
func ResponseWriterMiddleware() *BaseMiddleware {
	return &BaseMiddleware{
		handler: func(w *responseWriter, r *http.Request, next http.Handler) {
			next.ServeHTTP(w, r)
			w.WriteResponse(r.Context())
		},
	}
}
