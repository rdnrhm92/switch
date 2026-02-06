package config

import (
	"fmt"
	"os"
	"strings"
)

const EnvVarName = "APP_ENV"

// Environment 环境类型
type Environment string

// String 方便打印
func (e Environment) String() string {
	return string(e)
}

// ParseEnvironment 解析环境字符串
func ParseEnvironment(env string) Environment {
	return Environment(strings.ToLower(env))
}

// Resolver 解析环境
type Resolver interface {
	Resolve() (Environment, error)
}

// envResolver 解析器来初始化全局环境
func envResolver(resolver Resolver) error {
	env, err := resolver.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve environment: %w", err)
	}
	currentEnvironment = env
	return nil
}

// SetOsEnvironment 设置环境变量
func SetOsEnvironment(e Environment) error {
	if e == "" {
		return fmt.Errorf("invalid environment: %s", e)
	}
	return os.Setenv(EnvVarName, e.String())
}

// GetOsEnvironment 获取环境变量
func GetOsEnvironment() Environment {
	return currentEnvironment
}

var currentEnvironment = func() Environment {
	environment := Environment(os.Getenv(EnvVarName))
	if environment == "" {
		return "<UNKNOWN>"
	}
	return environment
}()
