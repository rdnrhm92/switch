package system

const (
	X_Request_Id     = "x_request_id"
	X_Trace_Id       = "x_trace_id"
	X_Service_Name   = "x_service_name"
	X_Debug          = "x_debug"
	X_Client_Msg     = "x_client_msg"
	X_Env            = "x_env"
	X_Host_Name      = "x_host_name"
	X_Ip             = "x_ip"
	X_Version        = "x_version"
	X_Span_Id        = "x_span_id"
	X_Logger_Request = "x_logger_request"
)

var (
	GrpcServiceName = "grpc_server"
	GrpcDebug       = "true"
	GrpcVersion     = "1.0.0"
	GrpcHostName    = "grpc_server_pod_01"
)

var (
	HttpDebug       = "true"
	HttpServiceName = "http_server"
	HttpVersion     = "1.0.0"
	HttpHostName    = "http_server_pod_01"
)
