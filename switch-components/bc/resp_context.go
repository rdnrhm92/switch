package bc

import (
	"context"

	"gitee.com/fatzeng/switch-sdk-core/debug"
	"gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-core/resp"
	"gitee.com/fatzeng/switch-sdk-core/resp/proto"
)

// 一个自定义的context，贯穿业务，用来记录执行信息的，非日志(这部分信息会响应给调用方)
type respKey struct{}

func NewRespContext(ctx context.Context) context.Context {
	ctx = SetResp(ctx)
	ctx = SetTrace(ctx)
	ctx = SetDebug(ctx)
	ctx = SetStatistics(ctx)
	return ctx
}

func NewRespContextExtensions(ctx context.Context, securityCfg *resp.SecurityConfig) context.Context {
	respContext := NewRespContext(ctx)
	respWrapper := GetResp(respContext)
	respWrapper.SetConfig(securityCfg)
	return respContext
}

func GetResp(ctx context.Context) *resp.RespWrapper {
	if respVal, ok := ctx.Value(respKey{}).(*resp.RespWrapper); ok {
		return respVal
	}
	return nil
}

func SetResp(ctx context.Context) context.Context {
	if _, ok := ctx.Value(respKey{}).(*resp.RespWrapper); ok {
		return ctx
	}
	return context.WithValue(ctx, respKey{}, &resp.RespWrapper{})
}

func Response(ctx context.Context) *proto.Resp {
	if respVal, ok := ctx.Value(respKey{}).(*resp.RespWrapper); ok {
		protoResp := respVal.ToResp()
		if statisticsWrapper := GetStatistics(ctx); statisticsWrapper != nil {
			protoResp.Statistics = statisticsWrapper.ToStatistics()
		}
		if debugWrapper := GetDebugInfo(ctx); debugWrapper != nil {
			protoResp.DebugInfo = debugWrapper.ToDebugInfo()
		}
		if traceWrapper := GetTrace(ctx); traceWrapper != nil {
			protoResp.Trace = traceWrapper.ToTrace()
		}
		return protoResp
	}
	return nil
}

func SetCode(ctx context.Context, code int64) {
	if respVal := GetResp(ctx); respVal != nil {
		respVal.SetCode(resp.Code(code))
	}
}

func GetCode(ctx context.Context) int64 {
	if respVal := GetResp(ctx); respVal != nil {
		return int64(respVal.Code)
	}
	return 0
}

func SetMessage(ctx context.Context, message string) {
	if respVal := GetResp(ctx); respVal != nil {
		respVal.SetMessage(resp.Message(message))
	}
}

func GetMessage(ctx context.Context) string {
	if respVal := GetResp(ctx); respVal != nil {
		return string(respVal.Message)
	}
	return ""
}

func SetData(ctx context.Context, data interface{}) {
	if respVal := GetResp(ctx); respVal != nil {
		respVal.Data = data
	}
}

func GetData(ctx context.Context) interface{} {
	if respVal := GetResp(ctx); respVal != nil {
		return respVal.Data
	}
	return nil
}

func SetParams(ctx context.Context, params interface{}) {
	if respVal := GetResp(ctx); respVal != nil {
		respVal.Params = params
	}
}

func GetParams(ctx context.Context) interface{} {
	if respVal := GetResp(ctx); respVal != nil {
		return respVal.Params
	}
	return nil
}

func SetExtraData(ctx context.Context, extraData interface{}) {
	if respVal := GetResp(ctx); respVal != nil {
		respVal.ExtraData = extraData
	}
}

func GetExtraData(ctx context.Context) interface{} {
	if respVal := GetResp(ctx); respVal != nil {
		return respVal.ExtraData
	}
	return nil
}

// 串联自定义error
func SetError(ctx context.Context, err *reply.Error) {
	if err == nil {
		return
	}

	if respVal := GetResp(ctx); respVal != nil {
		respVal.SetCode(resp.Code(err.Code))
		respVal.SetMessage(resp.Message(err.Message))

		if debugVal := GetDebugInfo(ctx); debugVal != nil {
			// 设置调试信息
			if len(err.Details) > 0 {
				for _, detail := range err.Details {
					debugVal.AddEntry(debug.DebugLevelError, err.Message, detail)
				}
			}
		}
	}
}
