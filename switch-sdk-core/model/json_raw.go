package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JsonRaw json.RawMessage

func (j JsonRaw) Value() (driver.Value, error) {
	if len(j) == 0 || string(j) == "null" {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JsonRaw) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Scan source is not []byte")
	}
	*j = make(JsonRaw, len(bytes))
	copy(*j, bytes)
	return nil
}

func (j JsonRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JsonRaw) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JsonRaw: UnmarshalJSON on nil pointer")
	}
	*j = make(JsonRaw, len(data))
	copy(*j, data)
	return nil
}

func (j JsonRaw) ToMap() (map[string]interface{}, error) {
	var m map[string]interface{}
	if len(j) == 0 {
		return nil, nil
	}
	err := json.Unmarshal(j, &m)
	return m, err
}

func (j JsonRaw) ToArray() ([]interface{}, error) {
	var a []interface{}
	if len(j) == 0 {
		return nil, nil
	}
	err := json.Unmarshal(j, &a)
	return a, err
}
