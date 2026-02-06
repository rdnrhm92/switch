package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type CommonModel struct {
	ID         uint       `gorm:"primarykey;comment:主键ID" json:"id"`
	CreatedBy  string     `gorm:"default:'';comment:创建人" json:"createdBy"`
	UpdateBy   string     `gorm:"default:'';comment:修改人" json:"updateBy"`
	CreateTime *time.Time `gorm:"autoCreateTime;<-:create;comment:创建时间" json:"createTime"`
	UpdateTime *time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updateTime"`
}
type StringArray []string

func (a *StringArray) Value() (driver.Value, error) {
	if len(*a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan StringArray: invalid value type")
	}

	return json.Unmarshal(bytes, a)
}

type Version struct {
	Version int64 `gorm:"default:0;comment:版本号" json:"version"`
}
