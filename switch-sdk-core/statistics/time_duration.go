package statistics

import (
	"fmt"
	"time"
)

type Granularity time.Duration

const (
	Nanosecond  = Granularity(time.Nanosecond)
	Microsecond = Granularity(time.Microsecond)
	Millisecond = Granularity(time.Millisecond)
	Second      = Granularity(time.Second)
	Minute      = Granularity(time.Minute)
	Hour        = Granularity(time.Hour)
)

type Timer struct {
	// 开始时间
	Begin time.Time
	// 结束时间
	End time.Time
	// 耗时
	ExecuteTime time.Duration
	// 颗粒度
	Granularity Granularity
	// 统计名称
	Name string
}

// TimerBegin 开始计时
func TimerBegin(g ...Granularity) *Timer {
	granularity := Millisecond
	if len(g) > 0 {
		granularity = g[0]
	}
	return &Timer{
		Begin:       time.Now(),
		Granularity: granularity,
	}
}

// TimerBeginWithName 开始计时
func TimerBeginWithName(name string, g ...Granularity) *Timer {
	granularity := Millisecond
	if len(g) > 0 {
		granularity = g[0]
	}
	return &Timer{
		Begin:       time.Now(),
		Granularity: granularity,
		Name:        name,
	}
}

// Complete 完成计时
func (t *Timer) Complete() *Timer {
	t.End = time.Now()
	t.ExecuteTime = t.End.Sub(t.Begin)
	return t
}

// Duration 获取执行时间
func (t *Timer) Duration() time.Duration {
	if t.End.IsZero() {
		return time.Since(t.Begin)
	}
	return t.End.Sub(t.Begin)
}

// Duration 格式化执行时间
func (t *Timer) Format() string {
	executeTime := t.Duration()
	name := t.Name
	if name == "" {
		name = "default_timer"
	}
	return fmt.Sprintf("%s: %s", name, t.formatDuration(executeTime))
}

// formatDuration 根据粒度格式化时间
func (t *Timer) formatDuration(d time.Duration) string {
	switch t.Granularity {
	case Nanosecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case Microsecond:
		return fmt.Sprintf("%.3fμs", float64(d.Nanoseconds())/1000)
	case Millisecond:
		return fmt.Sprintf("%.3fms", float64(d.Nanoseconds())/1000000)
	case Second:
		return fmt.Sprintf("%.3fs", d.Seconds())
	case Minute:
		return fmt.Sprintf("%.3fm", d.Minutes())
	case Hour:
		return fmt.Sprintf("%.3fh", d.Hours())
	default:
		return d.String()
	}
}

// Reset 重置计时器
func (t *Timer) Reset() {
	t.Begin = time.Now()
	t.End = time.Time{}
	t.ExecuteTime = 0
}

// IsCompleted is over?
func (t *Timer) IsCompleted() bool {
	return !t.End.IsZero()
}
