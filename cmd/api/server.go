package main

import (
	"context"
	"fmt"
	database "go-backend/database/ent"
	"go-backend/pkg/configs"
	"go-backend/pkg/logging"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

	// 设置Gin模式
	gin.SetMode(config.Server.Mode)
	logging.Info("Set Gin mode to: %s", config.Server.Mode)

	// 并行初始化各个组件
	results, err := initializeComponents(config)
	if err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// 初始化事件系统
	initializeEventSystem()

	// 设置路由
	engine := setupRoutes(config, results.engine)

	// 启动服务器
	return startServer(config, engine, results.dbClient, results.redisClient)
}

// startServer 启动HTTP服务器并处理优雅关闭
func startServer(config *configs.AppConfig, engine *gin.Engine, dbClient *database.Client, redisClient *redis.Client) error {
	srv := &http.Server{
		Addr:    config.Server.Port,
		Handler: engine,
	}

	// 在goroutine中启动服务器
	go func() {
		logging.Info(
			// easy_study banner
			` 

 ██████   ██████        █████  ██████  ███    ███ ██ ███    ██ 
██    ██ ██            ██   ██ ██   ██ ████  ████ ██ ████   ██ 
██    ██ ██      █████ ███████ ██   ██ ██ ████ ██ ██ ██ ██  ██ 
██ ▄▄ ██ ██            ██   ██ ██   ██ ██  ██  ██ ██ ██  ██ ██ 
 ██████   ██████       ██   ██ ██████  ██      ██ ██ ██   ████ 
    ▀▀                                                         
---------------------------QC-ADMIN---------------------------
			`,
		)
		logging.Info("Server is starting on %s", config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Fatal("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Info("Received shutdown signal, shutting down gracefully...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := srv.Shutdown(ctx); err != nil {
		logging.Fatal("Server shutdown failed: %v", err)
	}

	// 关闭数据库连接
	if err := dbClient.Close(); err != nil {
		logging.Error("Database connection close failed: %v", err)
	} else {
		logging.Info("Database connection closed")
	}

	// 关闭Redis连接
	if err := redisClient.Close(); err != nil {
		logging.Error("Redis connection close failed: %v", err)
	} else {
		logging.Info("Redis connection closed")
	}

	logging.Info("Server exited gracefully")
	return nil
}
