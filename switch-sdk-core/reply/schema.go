package reply

// BaseResponse 包含了所有响应都必须有的 Code 和 Message 字段
type BaseResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}
