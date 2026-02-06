package statistics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TimerTestSuite struct {
	suite.Suite
}

func TestTimerSuite(t *testing.T) {
	suite.Run(t, new(TimerTestSuite))
}

// 测试基本计时功能
func (s *TimerTestSuite) TestBasicTiming() {
	timer := TimerBeginWithName("test_operation")

	// 模拟操作耗时
	time.Sleep(time.Millisecond * 100)

	timer.Complete()

	s.Greater(timer.Duration().Milliseconds(), int64(90)) // 允许一些误差
	s.Less(timer.Duration().Milliseconds(), int64(150))   // 允许一些误差

	// 验证名称
	s.Equal("test_operation", timer.Name)
}

// 测试多次计时
func (s *TimerTestSuite) TestMultipleTiming() {
	timer := TimerBeginWithName("multi_test")

	// 第一次计时
	time.Sleep(time.Millisecond * 50)
	timer.Complete()
	duration1 := timer.Duration()

	// 第二次计时
	timer.Reset()
	time.Sleep(time.Millisecond * 100)
	timer.Complete()
	duration2 := timer.Duration()

	s.Greater(duration1.Milliseconds(), int64(40))
	s.Less(duration1.Milliseconds(), int64(80))

	s.Greater(duration2.Milliseconds(), int64(90))
	s.Less(duration2.Milliseconds(), int64(150))
}

// 测试嵌套计时
func (s *TimerTestSuite) TestNestedTiming() {
	outerTimer := TimerBeginWithName("outer")

	time.Sleep(time.Millisecond * 50)

	innerTimer := TimerBeginWithName("inner")
	time.Sleep(time.Millisecond * 100)
	innerTimer.Complete()
	innerDuration := innerTimer.Duration()

	time.Sleep(time.Millisecond * 50)
	outerTimer.Complete()
	outerDuration := outerTimer.Duration()

	s.Greater(innerDuration.Milliseconds(), int64(90))
	s.Less(innerDuration.Milliseconds(), int64(150))

	s.Greater(outerDuration.Milliseconds(), int64(190))
	s.Less(outerDuration.Milliseconds(), int64(250))
}

// 测试零持续时间
func (s *TimerTestSuite) TestZeroDuration() {
	timer := TimerBegin()
	timer.Complete()
	duration := timer.Duration()

	s.GreaterOrEqual(duration.Nanoseconds(), int64(0))
	s.Less(duration.Milliseconds(), int64(1))
}

// 测试重置功能
func (s *TimerTestSuite) TestReset() {
	timer := TimerBeginWithName("reset_test")

	time.Sleep(time.Millisecond * 100)
	timer.Complete()
	duration1 := timer.Duration()

	timer.Reset()
	timer.Complete()
	duration2 := timer.Duration()

	s.Greater(duration1.Milliseconds(), int64(90))
	s.Less(duration2.Milliseconds(), int64(10)) // 重置后应该接近0
}

// 测试并发使用
func (s *TimerTestSuite) TestConcurrentUsage() {
	timer1 := TimerBeginWithName("concurrent1")
	timer2 := TimerBeginWithName("concurrent2")

	go func() {
		time.Sleep(time.Millisecond * 100)
		timer1.Complete()
		duration := timer1.Duration()
		s.Greater(duration.Milliseconds(), int64(90))
	}()

	time.Sleep(time.Millisecond * 50)
	timer2.Complete()
	duration := timer2.Duration()
	s.Greater(duration.Milliseconds(), int64(40))
}

// 测试不同粒度
func (s *TimerTestSuite) TestGranularity() {
	// 测试纳秒粒度
	timer := TimerBeginWithName("nano_test", Nanosecond)
	timer.Complete()
	s.Contains(timer.Format(), "ns")

	// 测试微秒粒度
	timer = TimerBeginWithName("micro_test", Microsecond)
	timer.Complete()
	s.Contains(timer.Format(), "μs")

	// 测试毫秒粒度
	timer = TimerBeginWithName("milli_test", Millisecond)
	timer.Complete()
	s.Contains(timer.Format(), "ms")

	// 测试秒粒度
	timer = TimerBeginWithName("second_test", Second)
	timer.Complete()
	s.Contains(timer.Format(), "s")

	// 测试分钟粒度
	timer = TimerBeginWithName("minute_test", Minute)
	timer.Complete()
	s.Contains(timer.Format(), "m")

	// 测试小时粒度
	timer = TimerBeginWithName("hour_test", Hour)
	timer.Complete()
	s.Contains(timer.Format(), "h")
}

// 测试格式化输出
func (s *TimerTestSuite) TestFormattedOutput() {
	timer := TimerBeginWithName("format_test")

	time.Sleep(time.Millisecond * 1500)
	timer.Complete()

	formatted := timer.Format()
	s.Contains(formatted, "format_test")
	s.Contains(formatted, "ms")
}

// 测试边界情况
func (s *TimerTestSuite) TestEdgeCases() {
	// 测试空名称
	timer := TimerBegin()
	timer.Complete()
	s.Contains(timer.Format(), "default_timer")

	// 测试特殊字符名称
	timer = TimerBeginWithName("!@#$%^&*()")
	timer.Complete()
	s.Contains(timer.Format(), "!@#$%^&*()")

	// 测试未完成的计时器
	timer = TimerBegin()
	s.False(timer.IsCompleted())
	s.Greater(timer.Duration(), time.Duration(0))

	// 测试完成的计时器
	timer.Complete()
	s.True(timer.IsCompleted())
}
