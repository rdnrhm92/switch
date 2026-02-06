package tool

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// 测试用的结构体
type TestStruct struct {
	Name string
}

type StructDefineTestSuite struct {
	suite.Suite
}

func TestStructDefineSuite(t *testing.T) {
	suite.Run(t, new(StructDefineTestSuite))
}

// 测试获取结构体名称
func (s *StructDefineTestSuite) TestGetStructName() {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "struct instance",
			input:    TestStruct{},
			expected: "TestStruct",
		},
		{
			name:     "struct pointer",
			input:    &TestStruct{},
			expected: "TestStruct",
		},
		{
			name:     "non-struct type",
			input:    "not a struct",
			expected: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := GetStructName(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

// 测试获取结构体完整名称
func (s *StructDefineTestSuite) TestGetStructFullName() {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "struct instance",
			input:    TestStruct{},
			expected: "tool.TestStruct",
		},
		{
			name:     "struct pointer",
			input:    &TestStruct{},
			expected: "tool.TestStruct",
		},
		{
			name:     "non-struct type",
			input:    "not a struct",
			expected: "string",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := GetStructFullName(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}
