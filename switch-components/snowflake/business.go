package snowflake

// BusinessType 业务code
type BusinessType string

// BusinessConfig 业务配置
type BusinessConfig struct {
	BusinessType BusinessType `json:"business_type" yaml:"business_type"`
	MachineID    int64        `json:"machine_id" yaml:"machine_id"`
	// ID前缀 比如  order-
	Prefix string `json:"prefix" yaml:"prefix"`
}
