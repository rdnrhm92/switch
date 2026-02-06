package config

import (
	"context"

	_ "gitee.com/fatzeng/switch-components/config"
	"gitee.com/fatzeng/switch-sdk-core/config"
)

var GlobalContext context.Context

var GlobalConfig = new(Config)

func init() {
	loader := config.NewConfigLoader()
	err := loader.From("configs/switch-config.yaml").
		Bind(GlobalConfig).
		Load()
	if err != nil {
		panic(err)
	}
}
