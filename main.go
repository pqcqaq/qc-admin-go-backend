package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go-backend/internal/routes"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	"go-backend/pkg/database"
	"go-backend/pkg/logging"
	"go-backend/pkg/s3"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 定义命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径")
	flag.StringVar(&configPath, "c", "config.yaml", "配置文件路径（简写）")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "使用方法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s                           # 使用默认配置文件 config.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c config.dev.yaml        # 使用开发环境配置\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config config.prod.yaml # 使用生产环境配置\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c /path/to/config.yaml   # 使用绝对路径配置\n", os.Args[0])
	}

	// 解析命令行参数
	flag.Parse()

	// 处理配置文件路径（支持相对路径和绝对路径）
	resolvedConfigPath, err := resolveConfigPath(configPath)
	if err != nil {
		log.Fatalf("Failed to resolve config path: %v", err)
	}

	// 加载配置
	config, err := configs.LoadConfig(resolvedConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志系统
	logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
	logging.SetPrefix(config.Logging.Prefix)

	// 设置数据库包的logger
	database.SetLogger(logging.GetInstance())

	// 设置缓存包的logger
	caching.SetLogger(logging.GetInstance())

	// 设置S3包的logger
	s3.SetLogger(logging.GetInstance())

	logging.Info("Config loaded successfully from: %s", resolvedConfigPath)
	logging.Info("Log level set to: %s", config.Logging.Level)

	// 设置Gin模式
	gin.SetMode(config.Server.Mode)
	logging.Info("Gin mode set to: %s", config.Server.Mode)

	// 创建数据库连接
	dbClient := database.InitInstance(&config.Database)
	redisClient := caching.InitInstance(&config.Redis)

	// 初始化S3客户端
	if err := s3.InitClient(&config.S3); err != nil {
		logging.Warn("Failed to initialize S3 client: %v", err)
	} else {
		logging.Info("S3 client initialized successfully")
	}

	logging.Info("Database connection established")

	// 创建Gin引擎
	engine := gin.Default()

	// 配置CORS跨域中间件
	if config.Server.CORS.Enabled {
		var corsConfig cors.Config

		// 如果是debug模式，允许所有来源
		if config.Server.Debug {
			corsConfig = cors.Config{
				AllowAllOrigins:  true,
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
				AllowHeaders:     []string{"*"},
				ExposeHeaders:    []string{"*"},
				AllowCredentials: true,
				MaxAge:           12 * time.Hour,
			}
			logging.Warn("CORS enabled with allow all origins (debug mode)")
		} else {
			// 生产环境使用配置文件中的设置
			corsConfig = cors.Config{
				AllowAllOrigins:  config.Server.CORS.AllowAllOrigins,
				AllowOrigins:     config.Server.CORS.AllowOrigins,
				AllowMethods:     config.Server.CORS.AllowMethods,
				AllowHeaders:     config.Server.CORS.AllowHeaders,
				ExposeHeaders:    config.Server.CORS.ExposeHeaders,
				AllowCredentials: config.Server.CORS.AllowCredentials,
				MaxAge:           time.Duration(config.Server.CORS.MaxAge) * time.Second,
			}
			logging.Info("CORS enabled with configured origins: %v", config.Server.CORS.AllowOrigins)
		}

		engine.Use(cors.New(corsConfig))
	} else {
		logging.Info("CORS disabled")
	}

	// 配置静态文件服务
	if config.Server.Static.Enabled {
		// 解析静态文件根目录的绝对路径
		staticRoot, err := resolveStaticPath(config.Server.Static.Root)
		if err != nil {
			logging.Warn("Failed to resolve static root path: %v, static file service disabled", err)
		} else {
			engine.Static(config.Server.Static.Path, staticRoot)
			logging.Info("Static file service enabled - Path: %s, Root: %s", config.Server.Static.Path, staticRoot)
		}
	} else {
		logging.Info("Static file service disabled")
	}

	// 设置路由
	router := routes.NewRouter()
	router.SetupRoutes(engine)
	logging.Info("Routes configured")

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    config.Server.Port,
		Handler: engine,
	}

	// 在goroutine中启动服务器
	go func() {
		logging.Info("Server starting on %s", config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Fatal("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// kill (no param) 默认发送 syscall.SIGTERM
	// kill -2 发送 syscall.SIGINT Ctrl+C
	// kill -9 发送 syscall.SIGKILL 但不能被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Info("Shutting down server...")

	// 上下文用于通知服务器它有5秒的时间完成当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := srv.Shutdown(ctx); err != nil {
		logging.Fatal("Server forced to shutdown: %v", err)
	}

	// 关闭数据库连接
	if err := dbClient.Close(); err != nil {
		logging.Error("Failed to close database connection: %v", err)
	} else {
		logging.Info("Database connection closed")
	}

	// 关闭Redis连接
	if err := redisClient.Close(); err != nil {
		logging.Error("Failed to close Redis connection: %v", err)
	} else {
		logging.Info("Redis connection closed")
	}

	logging.Info("Server exiting")
}

// resolveConfigPath 解析配置文件路径，支持相对路径和绝对路径
func resolveConfigPath(configPath string) (string, error) {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(configPath) {
		// 检查文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found: %s", configPath)
		}
		return configPath, nil
	}

	// 相对路径：相对于当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	resolvedPath := filepath.Join(workDir, configPath)

	// 检查文件是否存在（可选，因为Viper会处理文件不存在的情况）
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		log.Printf("Warning: Config file not found at %s, will use defaults", resolvedPath)
	}

	return resolvedPath, nil
}

// resolveStaticPath 解析静态文件目录路径，支持相对路径和绝对路径
func resolveStaticPath(staticPath string) (string, error) {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(staticPath) {
		// 检查目录是否存在
		if _, err := os.Stat(staticPath); os.IsNotExist(err) {
			return "", fmt.Errorf("static directory not found: %s", staticPath)
		}
		return staticPath, nil
	}

	// 相对路径：相对于当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	resolvedPath := filepath.Join(workDir, staticPath)

	// 检查目录是否存在，如果不存在则创建
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		logging.Info("Static directory not found, creating: %s", resolvedPath)
		if err := os.MkdirAll(resolvedPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create static directory: %w", err)
		}
	}

	return resolvedPath, nil
}
