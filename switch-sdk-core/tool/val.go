package tool

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func ConvertToAny(data interface{}) (*anypb.Any, error) {
	if msg, ok := data.(proto.Message); ok {
		return anypb.New(msg)
	}

	value, err := structpb.NewValue(data)
	if err != nil {
		return nil, err
	}

	return anypb.New(value)
}

// 适配多数据类型
func ConvertToVal(data interface{}) (*structpb.Value, error) {
	if data == nil {
		return structpb.NewNullValue(), nil
	}

	val, err := structpb.NewValue(data)
	if err == nil {
		return val, nil
	}
	var marshaledData interface{}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(jsonBytes, &marshaledData); err != nil {
		return nil, err
	}

	return structpb.NewValue(marshaledData)
}

func ConvertToMap(req map[string]*structpb.Value) map[interface{}]interface{} {
	data := make(map[interface{}]interface{})
	for k, v := range req {
		data[k] = getValueFromStructpb(v)
	}
	return data
}

func getValueFromStructpb(v *structpb.Value) interface{} {
	switch v.GetKind().(type) {
	case *structpb.Value_StringValue:
		return v.GetStringValue()
	case *structpb.Value_NumberValue:
		return v.GetNumberValue()
	case *structpb.Value_BoolValue:
		return v.GetBoolValue()
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_StructValue:
		return v.GetStructValue().AsMap()
	case *structpb.Value_ListValue:
		return v.GetListValue().AsSlice()
	default:
		return v.AsInterface()
	}
}

func ConvertToMapS(req map[string]*structpb.Value) map[string]interface{} {
	data := make(map[string]interface{})
	for k, v := range req {
		data[k] = v.AsInterface()
	}
	return data
}

func ToString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case fmt.Stringer:
		return val.String()
	default:
		if reflect.TypeOf(v).Kind() == reflect.Struct ||
			reflect.TypeOf(v).Kind() == reflect.Map ||
			reflect.TypeOf(v).Kind() == reflect.Slice ||
			reflect.TypeOf(v).Kind() == reflect.Array {

			jsonBytes, err := json.Marshal(v)
			if err == nil {
				return string(jsonBytes)
			}
		}
		// 最后的兜底方案
		return fmt.Sprintf("%v", v)
	}
}

func ToJSONString(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}

	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v), err
	}

	return string(jsonBytes), nil
}

func InArray[K comparable](target K, arrays *[]K) bool {
	if len(*arrays) == 0 {
		return false
	}
	for _, val := range *arrays {
		if val == target {
			return true
		}
	}
	return false
}

func Trim(s *string) string {
	return strings.Trim(*s, " \n\r\t\v\u0000")
}

func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int63n(max-min) + min
}

func String2Bytes(s *string) []byte {
	if *s == "" {
		return nil
	}
	return []byte(*s)
}

func Bytes2String(b []byte) string {
	if b == nil {
		return ""
	}
	return string(b)
}

func ToFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, fmt.Errorf("cannot convert nil to float64")
	}

	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil

	case string:
		if val == "" {
			return 0, nil
		}
		return strconv.ParseFloat(val, 64)
	case bool:
		if val {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Float32, reflect.Float64:
			return rv.Float(), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), nil
		case reflect.String:
			str := rv.String()
			if str == "" {
				return 0, nil
			}
			return strconv.ParseFloat(str, 64)
		case reflect.Bool:
			if rv.Bool() {
				return 1.0, nil
			}
			return 0.0, nil
		default:
			return 0, fmt.Errorf("cannot convert %T to float64", v)
		}
	}
}

func ToInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("cannot convert nil to int64")
	}

	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil

	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil

	case string:
		if val == "" {
			return 0, nil
		}
		return strconv.ParseInt(val, 10, 64)

	case bool:
		if val {
			return 1, nil
		}
		return 0, nil

	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(rv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return int64(rv.Float()), nil
		case reflect.String:
			str := rv.String()
			if str == "" {
				return 0, nil
			}
			return strconv.ParseInt(str, 10, 64)
		case reflect.Bool:
			if rv.Bool() {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("cannot convert %T to int64", v)
		}
	}
}

// ToInt 将 interface{} 转换为 int
// 支持多种类型的转换：int系列、float系列、string、bool等
func ToInt(v interface{}) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("cannot convert nil to int")
	}

	switch val := v.(type) {
	case int:
		return val, nil
	case int8:
		return int(val), nil
	case int16:
		return int(val), nil
	case int32:
		return int(val), nil
	case int64:
		return int(val), nil
	case uint:
		return int(val), nil
	case uint8:
		return int(val), nil
	case uint16:
		return int(val), nil
	case uint32:
		return int(val), nil
	case uint64:
		return int(val), nil
	case float32:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		if val == "" {
			return 0, nil
		}
		return strconv.Atoi(val)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		// 尝试通过反射处理其他类型
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(rv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(rv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return int(rv.Float()), nil
		case reflect.String:
			str := rv.String()
			if str == "" {
				return 0, nil
			}
			return strconv.Atoi(str)
		case reflect.Bool:
			if rv.Bool() {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("cannot convert %T to int", v)
		}
	}
}

// ToBool 将任意类型转换为布尔值
func ToBool(value interface{}) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "yes"
	case int, int64:
		return ToString(v) != "0"
	case float64:
		return v != 0.0
	default:
		return ToString(v) != ""
	}
}

// ToIntWithDefault 将 interface{} 转换为 int，如果转换失败则返回默认值
func ToIntWithDefault(v interface{}, defaultValue int) int {
	result, err := ToInt(v)
	if err != nil {
		return defaultValue
	}
	return result
}

func Bool(b bool) *bool {
	return &b
}
