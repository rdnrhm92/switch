package actuator

import (
	"context"
	"fmt"
	"reflect"

	"gitee.com/fatzeng/switch-sdk-core/factor"
)

// actuatorInfo 执行因子缓存项，提供缓存的能力
type actuatorInfo struct {
	//因子执行器
	fn reflect.Value
	//因子执行器的配置
	configType reflect.Type
}

// registry 执行器因子存储器
var registry = make(map[string]actuatorInfo)

// Register 注册一个因子执行器 强制一个meta信息，业务侧增加可读性
func Register(meta *factor.SwitchFactor, actuatorFunc interface{}) {
	if meta == nil || meta.Factor == "" || meta.Description == "" {
		panic(fmt.Sprintf("注册无效：必须指定具体的因子名称跟描述"))
	}
	funcVal := reflect.ValueOf(actuatorFunc)
	funcType := funcVal.Type()

	//参数1 context 参数2 config(任意类型any) 返回值1 bool(是否通过) 返回值2 error(错误)
	if funcType.Kind() != reflect.Func || funcType.NumIn() != 2 || funcType.NumOut() != 2 {
		panic(fmt.Sprintf("函数 %s 签名无效：必须有2个输入参数和2个返回值", meta.Factor))
	}

	// 校验输入参数类型
	if funcType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		panic(fmt.Sprintf("函数 %s 的第一个参数必须是 context.Context", meta.Factor))
	}

	// 校验返回值类型
	boolType := reflect.TypeOf(true)
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if funcType.Out(0) != boolType || funcType.Out(1) != errorType {
		panic(fmt.Sprintf("函数 %s 的返回值必须是 (bool, error)", meta.Factor))
	}

	info := actuatorInfo{
		fn: funcVal,
		//因为第二个参数是配置参数，只缓存第二个就可以,避免dispatcher频繁检查配置类型
		configType: funcType.In(1),
	}

	//存在则覆盖
	registry[meta.Factor] = info
}

// Get 获取一个因子执行器信息
func Get(name string) (actuatorInfo, bool) {
	info, found := registry[name]
	return info, found
}
