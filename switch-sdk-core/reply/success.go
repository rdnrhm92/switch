package reply

// SuccessResponse 代表一个成功的 API 响应,新增data字段
type SuccessResponse struct {
	BaseResponse
	Data interface{} `json:"data"`
}
