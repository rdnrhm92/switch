package main

import (
	"context"
	"errors"
	"fmt"

	"gitee.com/fatzeng/switch-admin/internal/admin_driver"
	"gitee.com/fatzeng/switch-admin/internal/api"
	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-admin/internal/notifier"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"gitee.com/fatzeng/switch-admin/internal/ws"
	"gitee.com/fatzeng/switch-components/drivers"
	"gitee.com/fatzeng/switch-components/logging"
	"gitee.com/fatzeng/switch-components/recovery"
	"gitee.com/fatzeng/switch-sdk-core/driver"
	"gitee.com/fatzeng/switch-sdk-core/logger"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 应用启动后加载初始化配置
	if err := initializeApp(ctx); err != nil {
		panic(err)
	}

	config.GlobalContext = ctx

	r := api.SetupRouter(config.GlobalConfig)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler: r,
	}

	// http服务启动
	go startServer(ctx, srv)

	// 优雅关闭
	waitForShutdownSignal()

	// 执行优雅关闭
	shutdownServer(srv)

	// 关闭所有驱动连接
	if err := driver.GetManager().Close(); err != nil {
		logger.Logger.Errorf("failed to close drivers: %v", err)
	}

	logger.Logger.Info("Server exiting gracefully")
}

// initializeApp 封装了所有的应用初始化逻辑。
func initializeApp(ctx context.Context) error {
	if err := logInit(); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	if err := admin_driver.InitializeDriver(); err != nil {
		return fmt.Errorf("failed to initialize admin_driver: %w", err)
	}

	notifier.Init()
	if err := notifier.PreloadDrivers(ctx); err != nil {
		logger.Logger.Warnf("failed to preload drivers, notifications may be delayed on first request: %v", err)
	}
	// mysql表以及数据初始化
	repository.SeedData(repository.GetDB())

	// 全局IP池维护
	if err := drivers.StartIPPoolManager(ctx, config.GlobalConfig.Retry.IPConnectivity); err != nil {
		return err
	}

	// 全局缓存管理器启动
	if err := ws.StartSwitchCache(ctx, config.GlobalConfig.Cache); err != nil {
		return err
	}

	// ws长连接启动
	if err := ws.StartWebSocketServer(ctx); err != nil {
		return err
	}

	return nil
}

// startServer 安全的启动一个http服务
func startServer(ctx context.Context, srv *http.Server) {
	recovery.SafeGo(ctx, func(ctx context.Context) error {
		logger.Logger.Infof("Server is running on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Errorf("HTTP server ListenAndServe error: %v", err)
			return err
		}
		return nil
	}, "api-server")
}

// waitForShutdownSignal 优雅关机
func waitForShutdownSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("Shutdown signal received, starting graceful shutdown...")
}

// shutdownServer 关闭http服务设置5秒超时防止无限制挂机
func shutdownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown: ", err)
	}

	if err := drivers.StopIPPoolManager(); err != nil {
		logger.Logger.Warnf("failed to close drivers: %v", err)
	}

	ws.StopSwitchCache()

	if err := ws.StopWebSocketServer(); err != nil {
		logger.Logger.Warnf("failed to close websocket server: %v", err)
	}

	logger.Logger.Info("Server exiting gracefully")
}

func logInit() error {
	logConfig := config.GlobalConfig.LogConfig
	log, err := logging.New(logConfig)
	if err != nil {
		return err
	}
	logger.Logger = log
	return nil
}
