package trace

import (
	"gitee.com/fatzeng/switch-sdk-core/resp/proto"
)

type TraceWrapper struct {
	*proto.Trace
}

func NewTraceWrapper() *TraceWrapper {
	return &TraceWrapper{
		Trace: &proto.Trace{},
	}
}

func (t *TraceWrapper) ToTrace() *proto.Trace {
	return t.Trace
}

func (t *TraceWrapper) SetRequestId(requestId string) *TraceWrapper {
	t.RequestId = requestId
	return t
}

func (t *TraceWrapper) GetRequestId() string {
	if t.Trace == nil {
		return ""
	}
	return t.RequestId
}

func (t *TraceWrapper) SetTraceId(traceId string) *TraceWrapper {
	t.TraceId = traceId
	return t
}

func (t *TraceWrapper) GetTraceId() string {
	if t.Trace == nil {
		return ""
	}
	return t.TraceId
}

func (t *TraceWrapper) SetSpanId(spanId string) *TraceWrapper {
	t.SpanId = spanId
	return t
}

func (t *TraceWrapper) GetSpanId() string {
	if t.Trace == nil {
		return ""
	}
	return t.SpanId
}

func (t *TraceWrapper) SetVersion(version string) *TraceWrapper {
	t.Version = version
	return t
}

func (t *TraceWrapper) GetVersion() string {
	if t.Trace == nil {
		return ""
	}
	return t.Version
}

func (t *TraceWrapper) SetServiceName(serviceName string) *TraceWrapper {
	t.ServiceName = serviceName
	return t
}

func (t *TraceWrapper) GetServiceName() string {
	if t.Trace == nil {
		return ""
	}
	return t.ServiceName
}

func (t *TraceWrapper) SetHostName(hostName string) *TraceWrapper {
	t.HostName = hostName
	return t
}

func (t *TraceWrapper) GetHostName() string {
	if t.Trace == nil {
		return ""
	}
	return t.HostName
}

func (t *TraceWrapper) SetEnv(env string) *TraceWrapper {
	t.Env = env
	return t
}

func (t *TraceWrapper) GetEnv() string {
	if t.Trace == nil {
		return ""
	}
	return t.Env
}

func (t *TraceWrapper) SetIp(ip string) *TraceWrapper {
	t.Ip = ip
	return t
}

func (t *TraceWrapper) GetIp() string {
	if t.Trace == nil {
		return ""
	}
	return t.Ip
}
