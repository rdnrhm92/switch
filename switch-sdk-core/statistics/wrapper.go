package statistics

import (
	"gitee.com/fatzeng/switch-sdk-core/resp/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type StatisticsWrapper struct {
	*proto.Statistics
}

func NewStatisticsWrapper() *StatisticsWrapper {
	return &StatisticsWrapper{
		Statistics: &proto.Statistics{},
	}
}

func (s *StatisticsWrapper) ToStatistics() *proto.Statistics {
	return s.Statistics
}

func (s *StatisticsWrapper) MarshalJSON() ([]byte, error) {
	// 使用 protojson.MarshalOptions 确保所有字段都被序列化
	m := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	return m.Marshal(s.Statistics)
}

// SetExecuteTime SetExecuteTime
func (s *StatisticsWrapper) SetExecuteTime(executeTime int64) *StatisticsWrapper {
	s.Statistics.ExecuteTime = &executeTime
	return s
}

// SetRequestTime SetRequestTime
func (s *StatisticsWrapper) SetRequestTime(requestTime int64) *StatisticsWrapper {
	s.Statistics.RequestTime = requestTime
	return s
}

// SetResponseTime SetResponseTime
func (s *StatisticsWrapper) SetResponseTime(responseTime int64) *StatisticsWrapper {
	s.Statistics.ResponseTime = responseTime
	return s
}

// SetFrom from
func (s *StatisticsWrapper) SetFrom(from string) *StatisticsWrapper {
	s.Statistics.From = from
	return s
}

// SetTo to
func (s *StatisticsWrapper) SetTo(to string) *StatisticsWrapper {
	s.Statistics.To = to
	return s
}
