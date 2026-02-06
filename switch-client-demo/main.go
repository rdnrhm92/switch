package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/fatzeng/switch-components/logging"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-go/core/cache"

	switchsdk "gitee.com/fatzeng/switch-sdk-go"
	"gitee.com/fatzeng/switch-sdk-go/core/switch"
)

const env = "pre"

func main() {
	// 配置日志
	log, err := logging.New(&logger.LoggerConfig{
		Level:            "info",                     // 日志级别
		OutputDir:        "./logs",                   // 日志输出目录
		FileNameFormat:   "switch-demo_%Y-%m-%d.log", // 日志文件名格式
		MaxSize:          50,                         // 单个日志文件最大大小(MB)
		MaxBackups:       3,                          // 保留的旧日志文件数量
		MaxAge:           7,                          // 日志文件保留天数
		Compress:         false,                      // 是否压缩旧日志文件
		ShowCaller:       true,                       // 是否显示调用者信息
		EnableConsole:    true,                       // 是否启用控制台输出
		EnableJSON:       false,                      // 是否使用JSON格式(demo用false便于阅读)
		EnableStackTrace: true,                       // 是否启用堆栈跟踪
		StackTraceLevel:  "error",                    // 堆栈跟踪级别
		TimeFormat:       "2006-01-02 15:04:05",      // 时间格式
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger.Logger = log

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 缓存项测试
	ctx = cache.UseCache(ctx)

	// 初始化Switch SDK
	err = switchsdk.Start(ctx,
		_switch.WithDomain("ws://localhost:8081"),
		_switch.WithNamespaceTag("test-ns"),
		_switch.WithEnvTag(env),
		_switch.WithServiceName("simple-demo"),
		_switch.WithVersion("1.0.0"),
	)
	if err != nil {
		logger.Logger.Fatalf("Failed to initialize Switch SDK: %v", err)
	}

	logger.Logger.Info("SDK initialized successfully")

	// 启动演示
	ctx = context.WithValue(ctx, "user_name", "张三")
	go runDemo(ctx)
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	logger.Logger.Info("Received shutdown signal")
	switchsdk.Shutdown()
	logger.Logger.Info("Demo stopped")
}

func runDemo(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			testSwitches(ctx)
		}
	}
}

// testSwitches 测试开关功能
func testSwitches(ctx context.Context) {
	// 测试功能开关
	switchName := "feature_enabled"
	if _switch.IsOpen(ctx, switchName) {
		logger.Logger.Infof("Env '%s' Switch '%s' is ON", env, switchName)
	} else {
		logger.Logger.Infof("Env '%s' Switch '%s' is OFF", env, switchName)
	}
}
