package driver

import (
	"fmt"
	"reflect"
)

// WrapCreator 将任意创建函数转换为 Creator 类型
func WrapCreator(createFn interface{}) Creator {
	return func() (Driver, error) {
		fnValue := reflect.ValueOf(createFn)
		if fnValue.Kind() != reflect.Func {
			return nil, fmt.Errorf("createFn must be a function")
		}

		results := fnValue.Call([]reflect.Value{})

		if len(results) != 2 {
			return nil, fmt.Errorf("creator function must return (Driver, error)")
		}

		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}

		return results[0].Interface().(Driver), nil
	}
}
