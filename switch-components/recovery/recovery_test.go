package recovery

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RecoveryTestSuite struct {
	suite.Suite
}

func TestRecoverySuite(t *testing.T) {
	suite.Run(t, new(RecoveryTestSuite))
}

// 测试基本panic恢复
func (s *RecoveryTestSuite) TestBasicRecovery() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	var taskRun bool
	task := func(ctx context.Context) error {
		defer wg.Done()
		taskRun = true
		panic("test panic")
	}

	SafeGo(ctx, task, "test_panic")
	wg.Wait()

	s.True(taskRun, "Task should have run")
}

// 测试正常执行（无panic）
func (s *RecoveryTestSuite) TestNormalExecution() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	var taskRun bool
	task := func(ctx context.Context) error {
		defer wg.Done()
		taskRun = true
		return nil
	}

	SafeGo(ctx, task, "test_normal")
	wg.Wait()

	s.True(taskRun, "Task should have run")
}

// 测试返回错误（无panic）
func (s *RecoveryTestSuite) TestErrorReturn() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	retryCount := 0
	task := func(ctx context.Context) error {
		retryCount++
		if retryCount == 1 {
			return errors.New("test error")
		}
		wg.Done()
		return nil
	}

	SafeGo(ctx, task, "test_error", WithRetryInterval(time.Millisecond))
	wg.Wait()

	s.Equal(2, retryCount, "Task should have retried once")
}

// 测试上下文取消
func (s *RecoveryTestSuite) TestContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)

	task := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			wg.Done()
			return nil
		default:
			return errors.New("retry me")
		}
	}

	SafeGo(ctx, task, "test_cancel", WithRetryInterval(time.Millisecond))
	time.Sleep(time.Millisecond * 10) // 让任务重试几次
	cancel()
	wg.Wait()
}

// 测试重试间隔
func (s *RecoveryTestSuite) TestRetryInterval() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	retryCount := 0
	start := time.Now()
	task := func(ctx context.Context) error {
		retryCount++
		if retryCount == 1 {
			return errors.New("retry me")
		}
		wg.Done()
		return nil
	}

	interval := time.Millisecond * 100
	SafeGo(ctx, task, "test_interval", WithRetryInterval(interval))
	wg.Wait()

	elapsed := time.Since(start)
	s.GreaterOrEqual(elapsed, interval, "Should have waited for the retry interval")
}

// 测试多个panic
func (s *RecoveryTestSuite) TestMultiplePanics() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	panicCount := 0
	task := func(ctx context.Context) error {
		panicCount++
		if panicCount < 3 {
			panic("test panic")
		}
		wg.Done()
		return nil
	}

	SafeGo(ctx, task, "test_multiple_panics", WithRetryInterval(time.Millisecond))
	wg.Wait()

	s.Equal(3, panicCount, "Task should have panicked twice and succeeded on the third try")
}

// 测试不同类型的panic
func (s *RecoveryTestSuite) TestDifferentPanicTypes() {
	testCases := []struct {
		name  string
		panic interface{}
	}{
		{"String", "panic string"},
		{"Integer", 42},
		{"Error", errors.New("error panic")},
		{"Nil", nil},
		{"Boolean", true},
		{"Struct", struct{ msg string }{"panic struct"}},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			wg := sync.WaitGroup{}
			wg.Add(1)

			var recovered bool
			task := func(ctx context.Context) error {
				if !recovered {
					recovered = true
					panic(tc.panic)
				}
				wg.Done()
				return nil
			}

			SafeGo(ctx, task, "test_panic_types", WithRetryInterval(time.Millisecond))
			wg.Wait()
		})
	}
}

// 测试默认重试间隔
func (s *RecoveryTestSuite) TestDefaultRetryInterval() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)

	retryCount := 0
	task := func(ctx context.Context) error {
		retryCount++
		if retryCount == 1 {
			return errors.New("retry me")
		}
		wg.Done()
		return nil
	}

	SafeGo(ctx, task, "test_default_interval")
	wg.Wait()

	s.Equal(2, retryCount, "Task should have retried once")
}
