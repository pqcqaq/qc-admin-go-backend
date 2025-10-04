package main

import (
	"context"
	"fmt"
	"go-backend/cmd/socket/handlers"
	"go-backend/pkg/configs"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/websocket"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

// runServer 启动服务器的主要逻辑
func runServer(cmd *cobra.Command, args []string) error {
	// 解析配置文件路径
	resolvedConfigPath, err := configs.ResolveConfigPath(configFile)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	// 加载配置
	config, err := configs.LoadConfig(resolvedConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 应用命令行覆盖配置
	applyFlagOverrides(config)

	// 初始化日志系统
	setupLogging(config)

	logging.Info("Load config successfully: %s", resolvedConfigPath)
	logging.Info("Set log level to: %s", config.Logging.Level)

	// 并行初始化各个组件
	results, err := initializeComponents(config)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// 启动服务器
	return startServer(config, results.redisClient)
}

// startServer 启动HTTP服务器并处理优雅关闭
func startServer(config *configs.AppConfig, redisClient *redis.Client) error {
	ctx := context.Background()

	messaging.CreateGroup(ctx)

	// 创建WebSocket服务器实例
	wsServer := websocket.NewWsServer()
	sender := wsServer.CreateSender()
	handlers.SetSender(sender)

	// 在goroutine中启动服务器
	go func() {
		// 尝试从banner.txt读取并显示字符图
		if bannerContent, err := os.ReadFile("banner.txt"); err == nil {
			logging.Info(string(bannerContent))
		}
		logging.Info("WsServer is starting on %s", config.Socket.Port)
		if err := wsServer.Start(config.Socket.Port); err != nil {
			logging.Error("WebSocket Server failed to start: %v", err)
			os.Exit(1)
		}
	}()

	go func() {
		logging.Info("Started messaging consumer")
		consumer := messaging.NewMessageConsumer("qc-admin_socket")
		consumer.Consume(ctx)
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Info("Received shutdown signal, shutting down gracefully...")

	// 优雅关闭服务器
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	wsServer.Shutdown()

	// 关闭Redis连接
	if err := redisClient.Close(); err != nil {
		logging.Error("Redis connection close failed: %v", err)
	} else {
		logging.Info("Redis connection closed")
	}

	logging.Info("Server exited gracefully")
	return nil
}
