package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// RuleNode 业务执行开关的单元，or或者and关系的基础单元
type RuleNode struct {
	Id uint `json:"id"`
	// 逻辑节点属性 and都为true则为true or相反
	NodeType string      `json:"nodeType,omitempty"` // and, or
	Children []*RuleNode `json:"children,omitempty"`

	// 因子，业务执行的具体逻辑
	Factor string          `json:"factor,omitempty"`
	Config json.RawMessage `json:"config,omitempty"` // 因子所需的具体配置 比如定义一下操作符等

	Description string `json:"description,omitempty"` // 对该节点或因子的描述
}

func (n *RuleNode) Value() (driver.Value, error) {
	return json.Marshal(n)
}

func (n *RuleNode) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &n)
}

// SwitchModel 定义了一个不可拆分的开关，原子的单位
type SwitchModel struct {
	CommonModel
	Version
	NamespaceTag   string    `gorm:"size:100;not null;uniqueIndex:idx_namespace_switch_name;comment:命名空间标签" json:"namespaceTag"`
	Name           string    `gorm:"size:255;not null;uniqueIndex:idx_namespace_switch_name;comment:开关名字" json:"name"`
	CurrentEnvTag  string    `gorm:"size:50;comment:开关当前所处的环境" json:"currentEnvTag"`
	Rules          *RuleNode `gorm:"type:json;not null;comment:开关具体的执行计划" json:"rules"`
	Description    string    `gorm:"size:1024;comment:开关描述" json:"description"`
	UseCache       bool      `gorm:"default:false;comment:是否使用缓存" json:"useCache"`
	ApproverStatus string    `gorm:"size:50;not null;default:'';comment:当前开关的审批状态" json:"approver_status"`
}

func (SwitchModel) TableName() string {
	return "switches"
}
