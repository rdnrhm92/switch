package actuator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// Dispatcher 找到并执行对应的因子执行器
func Dispatcher(ctx context.Context, factorName string, ruleRemark interface{}) (bool, error) {
	info, ok := Get(factorName)
	if !ok {
		return false, fmt.Errorf("no actuator found for factor: %s", factorName)
	}

	ctxVal := reflect.ValueOf(ctx)

	configPtr := reflect.New(info.configType.Elem())

	switch v := ruleRemark.(type) {
	case json.RawMessage:
		// 处理 json.RawMessage 类型
		if err := json.Unmarshal(v, configPtr.Interface()); err != nil {
			return false, fmt.Errorf("json.RawMessage 解析失败: %w 原文: %v", err, string(v))
		}
	case []byte:
		// 处理 []byte 类型（json.RawMessage 的底层类型）
		if err := json.Unmarshal(v, configPtr.Interface()); err != nil {
			return false, fmt.Errorf("[]byte 解析json失败: %w 原文: %v", err, string(v))
		}
	case string:
		if err := json.Unmarshal([]byte(v), configPtr.Interface()); err != nil {
			return false, fmt.Errorf("字符串 ruleRemark 解析json失败: %w 原文: %v", err, ruleRemark)
		}
	case map[string]interface{}:
		if err := mapstructure.Decode(v, configPtr.Interface()); err != nil {
			return false, fmt.Errorf("map ruleRemark 填充%s失败: %w 原文：%v", info.configType.Name(), err, ruleRemark)
		}
	default:
		return false, fmt.Errorf("不支持的 ruleRemark 类型: %T", ruleRemark)
	}

	args := []reflect.Value{ctxVal, configPtr}
	results := info.fn.Call(args)

	resultBool, _ := results[0].Interface().(bool)
	var resultErr error
	if !results[1].IsNil() {
		resultErr, _ = results[1].Interface().(error)
	}

	return resultBool, resultErr
}
