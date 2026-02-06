package resp

import (
	"encoding/json"
	"fmt"

	"gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-core/resp/proto"
	"gitee.com/fatzeng/switch-sdk-core/tool"
	"google.golang.org/protobuf/types/known/structpb"
)

type RespWrapper struct {
	config    *SecurityConfig
	Code      Code        `json:"code"`
	Message   Message     `json:"message"`
	Data      interface{} `json:"data"`
	Params    interface{} `json:"params"`
	ExtraData interface{} `json:"extraData"`
}

// IsSuccess is ok?
func (r *RespWrapper) IsSuccess(cc CheckCode) bool {
	return cc(r.Code)
}

// IsError is error?
func (r *RespWrapper) IsError(cc CheckCode) bool {
	return cc(r.Code)
}

// SetCode set code
func (r *RespWrapper) SetCode(code Code) *RespWrapper {
	r.Code = code
	return r
}

// SetMessage set message
func (r *RespWrapper) SetMessage(msg Message) *RespWrapper {
	r.Message = msg
	return r
}

// SetData set data
func (r *RespWrapper) SetData(data interface{}) *RespWrapper {
	r.Data = data
	return r
}

func (r *RespWrapper) SetErrorm(message string, err ...error) *RespWrapper {
	var errorData string
	if len(err) > 0 {
		for _, info := range err {
			errorData += "\r\n" + info.Error()
		}
	}
	r.Message = Message(fmt.Sprintf(message, errorData))
	r.Code = Code(400)
	return r
}

func (r *RespWrapper) SetError(err ...error) *RespWrapper {
	return r.SetErrorm("", err...)
}

func (r *RespWrapper) SetErrors(err *reply.Error) *RespWrapper {
	if err == nil {
		return r
	}
	r.Code = Code(err.Code)
	r.Message = Message(err.Message)
	return r
}

func (r *RespWrapper) SetSuccess(success *reply.SuccessResponse) *RespWrapper {
	if success == nil {
		return r
	}
	r.Code = Code(success.Code)
	r.Message = Message(success.Message)
	r.Data = success.Data
	return r
}

// ToJSON convert 2 json
func (r *RespWrapper) ToJSON() (string, error) {
	if r.config != nil {
		FilterFields(r.Data, r.config)
		MaskFields(r.Data, r.config)
	}
	jsonData, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ToJSONIndent convert 2 json
func (r *RespWrapper) ToJSONIndent() (string, error) {
	if r.config != nil {
		FilterFields(r.Data, r.config)
		MaskFields(r.Data, r.config)
	}
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// Clone clone
func (r *RespWrapper) Clone() (*RespWrapper, error) {
	jsonData, _ := json.Marshal(r)
	var cloned RespWrapper
	err := json.Unmarshal(jsonData, &cloned)
	if err != nil {
		return nil, err
	}
	return &cloned, nil
}

// CalculateResponseSize resp的大小
func (r *RespWrapper) CalculateResponseSize() int {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return 0
	}
	return len(jsonData)
}

// SetConfig SetConfig
func (r *RespWrapper) SetConfig(config *SecurityConfig) *RespWrapper {
	r.config = config
	return r
}

func (r *RespWrapper) ToResp() *proto.Resp {
	code := int64(r.Code)
	resp := &proto.Resp{
		Code:    &code,
		Message: string(r.Message),
	}

	if r.Data != nil {
		if val, err := tool.ConvertToVal(r.Data); err != nil {
			resp.Data = structpb.NewStringValue(fmt.Sprintf("convert error: [%v] ", err.Error()))
		} else {
			resp.Data = val
		}
	} else {
		r.Data = new(interface{})
		resp.Data = structpb.NewNullValue()
	}

	if r.Params != nil {
		if val, err := tool.ConvertToVal(r.Params); err != nil {
			resp.Params = structpb.NewStringValue("")
		} else {
			resp.Params = val
		}
	} else {
		r.Params = new(interface{})
		resp.Params = structpb.NewNullValue()
	}

	if r.ExtraData != nil {
		if val, err := tool.ConvertToVal(r.ExtraData); err != nil {
			resp.ExtraData = structpb.NewStringValue("")
		} else {
			resp.ExtraData = val
		}
	} else {
		r.ExtraData = new(interface{})
		resp.ExtraData = structpb.NewNullValue()
	}

	return resp
}
