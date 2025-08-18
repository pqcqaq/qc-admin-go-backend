// @title           Go Backend API
// @version         1.0
// @description     这是一个基于Go和Gin框架的后端API服务
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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
	"sync"
	"syscall"
	"time"

	basedatabase "go-backend/database"
	"go-backend/internal/routes"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	pkgdatabase "go-backend/pkg/database"
	"go-backend/pkg/email"
	"go-backend/pkg/jwt"
	"go-backend/pkg/logging"
	"go-backend/pkg/s3"
	"go-backend/pkg/sms"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	// 导入ent runtime以注册schema hooks
	database "go-backend/database/ent"
	_ "go-backend/database/ent/runtime"
	"go-backend/database/events"
)

func main() {
	// 定义命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径")
	flag.StringVar(&configPath, "c", "config.yaml", "配置文件路径（简写）")
	var autoMigrate string
	flag.StringVar(&autoMigrate, "migrate", "none", "是否自动迁移数据库模式")
	flag.StringVar(&autoMigrate, "m", "none", "是否自动迁移数据库模式（简写）")

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
		fmt.Fprintf(os.Stderr, "  %s -m          (skip|auto)   # 启动时自动迁移数据库模式\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --migrate   (skip|auto)   # 启动时自动迁移数据库模式（长选项）\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(0)
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

	// 通过命令行传入的配置项也合并到配置中
	{
		// 是否自动迁移，这个选项优先读取main.go中的命令行参数，否则就配置文件
		switch autoMigrate {
		case "skip":
			config.Database.SkipMigrateCheck = true
			config.Database.AutoMigrate = false
		case "auto":
			config.Database.SkipMigrateCheck = false
			config.Database.AutoMigrate = true
		default:
			// 默认不自动迁移，有不适配直接报错
			config.Database.SkipMigrateCheck = false
			config.Database.AutoMigrate = false
		}
	}

	// 初始化日志系统
	logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
	logging.SetPrefix(config.Logging.Prefix)

	// 设置数据库包的logger
	pkgdatabase.SetLogger(logging.WithName("Database"))

	// 设置缓存包的logger
	caching.SetLogger(logging.WithName("Caching"))

	// 设置S3包的logger
	s3.SetLogger(logging.WithName("S3Client"))

	// 设置邮件包的logger
	email.SetLogger(logging.WithName("EmailClient"))

	// 设置短信包的logger
	sms.SetLogger(logging.WithName("SMSClient"))

	logging.Info("Config loaded successfully from: %s", resolvedConfigPath)
	logging.Info("Log level set to: %s", config.Logging.Level)

	// 设置Gin模式
	gin.SetMode(config.Server.Mode)
	logging.Info("Gin mode set to: %s", config.Server.Mode)

	// 使用sync.WaitGroup并行初始化各个组件
	type InitResults struct {
		dbClient    *database.Client
		redisClient *redis.Client
		s3Error     error
		emailError  error
		smsError    error
		jwtError    error
		engine      *gin.Engine
	}

	results := &InitResults{}

	// 创建WaitGroup用于等待所有并行初始化完成
	var initWaitGroup sync.WaitGroup
	var initMutex sync.Mutex

	// 并行初始化S3客户端
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()
		if err := s3.InitClient(&config.S3); err != nil {
			initMutex.Lock()
			results.s3Error = err
			initMutex.Unlock()
			logging.Warn("Failed to initialize S3 client: %v", err)
		} else {
			logging.Info("S3 client initialized successfully")
		}
	}()

	// 并行初始化邮件客户端
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()
		if err := email.InitializeClient(&config.Email); err != nil {
			initMutex.Lock()
			results.emailError = err
			initMutex.Unlock()
			logging.Warn("Failed to initialize email client: %v", err)
		} else {
			logging.Info("Email client initialized successfully")
		}
	}()

	// 并行初始化短信客户端
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()
		if err := sms.InitializeClient(&config.SMS); err != nil {
			initMutex.Lock()
			results.smsError = err
			initMutex.Unlock()
			logging.Warn("Failed to initialize SMS client: %v", err)
		} else {
			logging.Info("SMS client initialized successfully")
		}
	}()

	// 并行初始化JWT服务
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()
		if err := jwt.InitializeService(&config.JWT); err != nil {
			initMutex.Lock()
			results.jwtError = err
			initMutex.Unlock()
			logging.Warn("Failed to initialize JWT service: %v", err)
		} else {
			logging.Info("JWT service initialized successfully")
		}
	}()

	// 并行初始化数据库连接
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()
		dbClient := pkgdatabase.InitInstance(&config.Database)
		initMutex.Lock()
		results.dbClient = dbClient
		initMutex.Unlock()
		logging.Info("Database connection established")
	}()

	// 并行初始化Redis连接
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()
		redisClient := caching.InitInstance(&config.Redis)
		initMutex.Lock()
		results.redisClient = redisClient
		initMutex.Unlock()
		logging.Info("Redis connection established")
	}()

	// 并行创建和配置Gin引擎
	initWaitGroup.Add(1)
	go func() {
		defer initWaitGroup.Done()

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

		initMutex.Lock()
		results.engine = engine
		initMutex.Unlock()
		logging.Info("Gin engine configured")
	}()

	// 等待所有并行初始化完成
	initWaitGroup.Wait()

	// 从结果中获取初始化的组件
	dbClient := results.dbClient
	redisClient := results.redisClient
	engine := results.engine

	// 初始化事件系统（必须在数据库初始化之后）
	events.SetLogger(logging.WithName("EventBus"))
	basedatabase.InitEventSystem()
	logging.Info("Event system initialized successfully")

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
