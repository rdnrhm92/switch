package debug

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/reply"
	"gitee.com/fatzeng/switch-sdk-core/resp/proto"
	"gitee.com/fatzeng/switch-sdk-core/tool"
	"google.golang.org/protobuf/types/known/structpb"
)

// DebugLevel debug level
type DebugLevel string

const (
	DebugLevelTrace DebugLevel = "TRACE"
	DebugLevelDebug DebugLevel = "DEBUG"
	DebugLevelInfo  DebugLevel = "INFO"
	DebugLevelWarn  DebugLevel = "WARN"
	DebugLevelError DebugLevel = "ERROR"
)

type DebugEntryWrapper struct {
	*proto.DebugEntry
}

type DebugInfoWrapper struct {
	*proto.DebugInfo
}

func (d *DebugInfoWrapper) ToDebugInfo() *proto.DebugInfo {
	return d.DebugInfo
}

// NewDebugInfoWrapper create new debugInfo
func NewDebugInfoWrapper() *DebugInfoWrapper {
	return &DebugInfoWrapper{
		DebugInfo: &proto.DebugInfo{
			Enabled: true,
			Parts:   make([]*proto.DebugEntry, 0),
		},
	}
}

// AddEntry AddEntry
func (d *DebugInfoWrapper) AddEntry(level DebugLevel, message string, data interface{}) *DebugInfoWrapper {
	if !d.Enabled {
		return d
	}
	valueData, err := tool.ConvertToVal(data)
	if err != nil {
		errMsg := fmt.Sprintf("无法序列化数据: %v, 原始数据: %v", err, data)
		valueData = structpb.NewStringValue(errMsg)
	}
	entry := &proto.DebugEntry{
		Level:     string(level),
		Timestamp: time.Now().Unix(),
		Message:   message,
		Data:      valueData,
	}

	d.Parts = append(d.Parts, entry)
	return d
}

// SetTrace SetTrace kv
func (d *DebugInfoWrapper) SetTrace(message string, data ...interface{}) *DebugInfoWrapper {
	var debugData interface{}
	if len(data) > 0 {
		debugData = data[0]
	}
	return d.AddEntry(DebugLevelTrace, message, debugData)
}

// SetDebug SetDebug kv
func (d *DebugInfoWrapper) SetDebug(message string, data ...interface{}) *DebugInfoWrapper {
	var debugData interface{}
	if len(data) > 0 {
		debugData = data[0]
	}
	return d.AddEntry(DebugLevelDebug, message, debugData)
}

// SetInfo SetInfo kv
func (d *DebugInfoWrapper) SetInfo(message string, data ...interface{}) *DebugInfoWrapper {
	var debugData interface{}
	if len(data) > 0 {
		debugData = data[0]
	}
	return d.AddEntry(DebugLevelInfo, message, debugData)
}

// SetWarn SetWarn kv
func (d *DebugInfoWrapper) SetWarn(message string, data ...interface{}) *DebugInfoWrapper {
	var debugData interface{}
	if len(data) > 0 {
		debugData = data[0]
	}
	return d.AddEntry(DebugLevelWarn, message, debugData)
}

// SetError SetError kv
func (d *DebugInfoWrapper) SetError(err ...error) *DebugInfoWrapper {
	var errorData string
	if len(err) > 0 {
		for _, info := range err {
			errorData += "\r\n" + info.Error()
		}
	}
	return d.AddEntry(DebugLevelError, "", errorData)
}

func (d *DebugInfoWrapper) SetErrorm(message string, err ...error) *DebugInfoWrapper {
	var errorData string
	if len(err) > 0 {
		for _, info := range err {
			errorData += "\r\n" + info.Error()
		}
	}
	return d.AddEntry(DebugLevelError, message, errorData)
}

func (d *DebugInfoWrapper) SetErrors(err *reply.Error) {
	if err == nil {
		return
	}

	if len(err.Details) > 0 {
		if d.DebugInfo == nil {
			d.DebugInfo = &proto.DebugInfo{
				Enabled: true,
				Parts:   make([]*proto.DebugEntry, 0),
			}
		}
		for _, detail := range err.Details {
			d.AddEntry(DebugLevelError, err.Message, detail)
		}
	}
}

// AddTrace SetTrace single
func (d *DebugInfoWrapper) AddTrace(message string, data ...interface{}) *DebugInfoWrapper {
	return d.SetTrace(message, data...)
}

// AddTracef AddTracef
func (d *DebugInfoWrapper) AddTracef(format string, args ...interface{}) *DebugInfoWrapper {
	return d.SetTrace(fmt.Sprintf(format, args...))
}

// AddDebug SetDebug single
func (d *DebugInfoWrapper) AddDebug(message string, data ...interface{}) *DebugInfoWrapper {
	return d.SetDebug(message, data...)
}

// AddDebugf AddDebugf
func (d *DebugInfoWrapper) AddDebugf(format string, args ...interface{}) *DebugInfoWrapper {
	return d.SetDebug(fmt.Sprintf(format, args...))
}

// AddInfo SetInfo single
func (d *DebugInfoWrapper) AddInfo(message string, data ...interface{}) *DebugInfoWrapper {
	return d.SetInfo(message, data...)
}

// AddInfof AddInfof
func (d *DebugInfoWrapper) AddInfof(format string, args ...interface{}) *DebugInfoWrapper {
	return d.SetInfo(fmt.Sprintf(format, args...))
}

// AddWarn SetWarn single
func (d *DebugInfoWrapper) AddWarn(message string, data ...interface{}) *DebugInfoWrapper {
	return d.SetWarn(message, data...)
}

// AddWarnf AddWarnf
func (d *DebugInfoWrapper) AddWarnf(format string, args ...interface{}) *DebugInfoWrapper {
	return d.SetWarn(fmt.Sprintf(format, args...))
}

// AddError SetError single
func (d *DebugInfoWrapper) AddError(err ...error) *DebugInfoWrapper {
	return d.SetErrorm("", err...)
}

// AddErrorf AddErrorf
func (d *DebugInfoWrapper) AddErrorf(format string, args ...interface{}) *DebugInfoWrapper {
	return d.SetErrorm(fmt.Sprintf(format, args...))
}

// Enable 启用调试
func (d *DebugInfoWrapper) Enable() *DebugInfoWrapper {
	d.Enabled = true
	return d
}

// Disable 禁用调试
func (d *DebugInfoWrapper) Disable() *DebugInfoWrapper {
	d.Enabled = false
	return d
}

// Clear 清空所有调试信息
func (d *DebugInfoWrapper) Clear() *DebugInfoWrapper {
	d.Parts = make([]*proto.DebugEntry, 0)
	return d
}

// ToJSON ToJSON
func (d *DebugInfoWrapper) ToJSON() (string, error) {
	jsonData, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ToJSONIndent convert 2 json
func (d *DebugInfoWrapper) ToJSONIndent() (string, error) {
	jsonData, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ToString 转换为可读字符串
func (d *DebugInfoWrapper) ToString() string {
	var parts []string
	return strings.Join(parts, "\n")
}
