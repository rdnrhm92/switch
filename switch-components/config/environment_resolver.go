package config

import (
	"fmt"
	"os"

	"gitee.com/fatzeng/switch-sdk-core/config"
)

// OSResolver 系统读取环境变量
type OSResolver struct {
	defaultSpecify config.Environment
	specify        config.Environment
}

func NewOSResolver(defaultSpecify config.Environment) config.Resolver {
	resolver := new(OSResolver)
	resolver.defaultSpecify = defaultSpecify
	return resolver
}

func (r *OSResolver) Resolve() (config.Environment, error) {
	// 从环境变量读取
	if envStr := os.Getenv(config.EnvVarName); envStr != "" {
		environment := config.ParseEnvironment(envStr)
		r.specify = environment
		return environment, nil
	}
	// 其次使用默认值
	if r.defaultSpecify != "" {
		return r.defaultSpecify, nil
	}

	return "", fmt.Errorf("environment variable %q not set and no valid default was provided", config.EnvVarName)
}
