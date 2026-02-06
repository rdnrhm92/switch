package grpc

import (
	"context"

	"gitee.com/fatzeng/switch-components/system"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// GetRequestID 从上下文中获取请求ID
func GetRequestID(ctx context.Context) (context.Context, string) {
	// 如果已经有了就不用再生成了。约定
	var requestID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if requestIDs := md.Get(system.X_Request_Id); len(requestIDs) > 0 {
			return ctx, requestIDs[0]
		}
	}
	requestID = generate.GeneratePrefixID()
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(system.X_Request_Id, requestID)
	//context需要被覆盖 & 构建新的元信息
	return metadata.NewIncomingContext(ctx, md), requestID
}

// GetTraceID 从上下文中获取请求ID
func GetTraceID(ctx context.Context) (context.Context, string) {
	// 如果已经有了就不用再生成了。约定
	var traceID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if traceIDs := md.Get(system.X_Trace_Id); len(traceIDs) > 0 {
			return ctx, traceIDs[0]
		}
	}
	//上游如果不提供traceid 则用requestid代替
	ctx, requestId := GetRequestID(ctx)
	traceID = requestId
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(system.X_Trace_Id, traceID)
	//context需要被覆盖 & 构建新的元信息
	return metadata.NewIncomingContext(ctx, md), traceID
}

// GetServiceName 从上下文中获取请求ID
func GetServiceName(ctx context.Context) (context.Context, string) {
	// 如果已经有了就不用再生成了。约定
	var serviceName string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if serverNames := md.Get(system.X_Service_Name); len(serverNames) > 0 {
			return ctx, serverNames[0]
		}
	}
	//上游如果不提供使用默认的服务名字(业务方可修改)
	serviceName = system.GrpcServiceName
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(system.X_Service_Name, serviceName)
	//context需要被覆盖 & 构建新的元信息
	return metadata.NewIncomingContext(ctx, md), serviceName
}

// GetServiceName 从上下文中获取debug开关
func GetDebugSwitch(ctx context.Context) (context.Context, bool) {
	// 如果已经有了就不用再生成了。约定
	var debug string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if debugs := md.Get(system.X_Debug); len(debugs) > 0 {
			return ctx, debugs[0] == "true"
		}
	}
	debug = system.GrpcDebug
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(system.X_Debug, debug)
	//context需要被覆盖 & 构建新的元信息
	return metadata.NewIncomingContext(ctx, md), debug == "true" || debug == "True"
}

// GetClientPeer 从grpc上下文中获取客户端信息
func GetClientPeer(ctx context.Context) *ClientPeer {
	clientPeer := &ClientPeer{
		Address: "unknown",
	}

	// 从上下文中获取对等方信息
	if p, ok := peer.FromContext(ctx); ok {
		clientPeer.Address = p.Addr.String()
		if p.AuthInfo != nil {
			clientPeer.AuthType = p.AuthInfo.AuthType()
		}
	}

	return clientPeer
}

// ClientPeer 客户端信息结构
type ClientPeer struct {
	Address  string // 客户端地址
	AuthType string // 认证类型
}

// GetHostName 从上下文中获取主机名
func GetHostName(ctx context.Context) (context.Context, string) {
	// 如果已经有了就不用再生成了。约定
	var hostName string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if hostNames := md.Get(system.X_Host_Name); len(hostNames) > 0 {
			return ctx, hostNames[0]
		}
	}
	//上游如果不提供使用默认的主机名
	hostName = system.GrpcHostName
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(system.X_Host_Name, hostName)
	//context需要被覆盖 & 构建新的元信息
	return metadata.NewIncomingContext(ctx, md), hostName
}

// GetVersion 从上下文中获取版本信息
func GetVersion(ctx context.Context) (context.Context, string) {
	// 如果已经有了就不用再生成了。约定
	var version string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if versions := md.Get(system.X_Version); len(versions) > 0 {
			return ctx, versions[0]
		}
	}
	//上游如果不提供使用默认的版本号
	version = system.GrpcVersion
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(system.X_Version, version)
	//context需要被覆盖 & 构建新的元信息
	return metadata.NewIncomingContext(ctx, md), version
}
