package tool

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RandomToolTestSuite struct {
	suite.Suite
}

func TestRandomToolSuite(t *testing.T) {
	suite.Run(t, new(RandomToolTestSuite))
}

// 测试生成指定范围的随机整数
func (s *RandomToolTestSuite) TestGetRandomBetween() {
	testCases := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "normal range",
			min:  0,
			max:  10,
		},
		{
			name: "negative to positive",
			min:  -100,
			max:  100,
		},
		{
			name: "large range",
			min:  1000,
			max:  2000,
		},
		{
			name: "negative range",
			min:  -2000,
			max:  -1000,
		},
		{
			name: "min equals max",
			min:  5,
			max:  5,
		},
		{
			name: "min greater than max",
			min:  10,
			max:  5,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// 多次测试以确保随机性和范围正确性
			for i := 0; i < 100; i++ {
				result := GetRandomBetween(tc.min, tc.max)
				if tc.min <= tc.max {
					s.GreaterOrEqual(result, tc.min, "Result should be greater than or equal to min")
					s.LessOrEqual(result, tc.max, "Result should be less than or equal to max")
				} else {
					// 当min > max时，函数会交换它们
					s.GreaterOrEqual(result, tc.max, "Result should be greater than or equal to max (swapped min)")
					s.LessOrEqual(result, tc.min, "Result should be less than or equal to min (swapped max)")
				}
			}
		})
	}
}
