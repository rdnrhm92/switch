package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type JsonArray []interface{}

func (a JsonArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *JsonArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JsonArray value:", value))
	}

	return json.Unmarshal(bytes, a)
}
