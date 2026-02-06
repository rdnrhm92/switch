package pc

const (
	Health  = "/health"
	Clients = "/clients"
)

const (
	// WsEndpointChangeConfig ws增量配置变更监听
	WsEndpointChangeConfig = "/ws/config/change"
	// WsEndpointFullSyncConfig ws全量配置监听
	WsEndpointFullSyncConfig = "/ws/config/fullSyncConfig"
	// WsEndpointFullSync ws全量开关监听
	WsEndpointFullSync = "/ws/config/fullSync"
)
