package dto

import (
	"encoding/json"
	"reflect"

	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/model"
)

// CreateUpdateDriverReq 创建编辑的api request
type CreateUpdateDriverReq struct {
	Id           uint              `json:"id"`
	Name         string            `json:"name"`
	Usage        model.UsageType   `json:"usage"`
	DriverType   driver.DriverType `json:"driverType"`
	DriverConfig json.RawMessage   `json:"driverConfig"`
	CreatedBy    string            //创建人
	UpdatedBy    string            //修改人
}

// IsEqual 判断两套配置是否一致
func IsEqual(config1, config2 json.RawMessage) (bool, error) {
	if (len(config1) == 0 || string(config1) == "null") && (len(config2) == 0 || string(config2) == "null") {
		return true, nil
	}

	var i1, i2 interface{}

	if err := json.Unmarshal(config1, &i1); err != nil {
		return false, err
	}
	if err := json.Unmarshal(config2, &i2); err != nil {
		return false, err
	}

	return reflect.DeepEqual(i1, i2), nil
}

// DriverListReq 驱动列表api request
type DriverListReq struct {
	*PageLimit
	Name       string            `json:"name"`
	Usage      model.UsageType   `json:"usage"`
	DriverType driver.DriverType `json:"driver_type"`
}
