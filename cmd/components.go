package cmd

import (
	"fmt"
	basedatabase "go-backend/database"
	database "go-backend/database/ent"
	"go-backend/database/events"
	"go-backend/internal/routes"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	pkgdatabase "go-backend/pkg/database"
	"go-backend/pkg/email"
	"go-backend/pkg/jwt"
	"go-backend/pkg/logging"
	"go-backend/pkg/openai"
	"go-backend/pkg/s3"
	"go-backend/pkg/sms"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// InitResults 保存初始化结果
type InitResults struct {
	dbClient    *database.Client
	redisClient *redis.Client
	engine      *gin.Engine
	errors      map[string]error
}

// setupLogging 设置日志系统
func setupLogging(config *configs.AppConfig) {
	logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
	logging.SetPrefix(config.Logging.Prefix)

	// 设置各个包的logger
	events.SetLogger(logging.WithName("EventBus"))
	pkgdatabase.SetLogger(logging.WithName("Database"))
	caching.SetLogger(logging.WithName("Caching"))
	s3.SetLogger(logging.WithName("S3Client"))
	email.SetLogger(logging.WithName("EmailClient"))
	sms.SetLogger(logging.WithName("SMSClient"))
}

// initializeComponents 并行初始化各个组件
func initializeComponents(config *configs.AppConfig) (*InitResults, error) {
	results := &InitResults{
		errors: make(map[string]error),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并行初始化S3客户端
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s3.InitClient(&config.S3); err != nil {
			mu.Lock()
			results.errors["s3"] = err
			mu.Unlock()
			logging.Warn("Failed to initialize S3 client: %v", err)
		} else {
			logging.Info("S3 client initialized successfully")
		}
	}()

	// 并行初始化邮件客户端
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := email.InitializeClient(&config.Email); err != nil {
			mu.Lock()
			results.errors["email"] = err
			mu.Unlock()
			logging.Warn("Failed to initialize email client: %v", err)
		} else {
			logging.Info("Email client initialized successfully")
		}
	}()

	// 并行初始化短信客户端
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := sms.InitializeClient(&config.SMS); err != nil {
			mu.Lock()
			results.errors["sms"] = err
			mu.Unlock()
			logging.Warn("Failed to initialize SMS client: %v", err)
		} else {
			logging.Info("SMS client initialized successfully")
		}
	}()

	// 并行初始化JWT服务
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := jwt.InitializeService(&config.JWT); err != nil {
			mu.Lock()
			results.errors["jwt"] = err
			mu.Unlock()
			logging.Warn("Failed to initialize JWT service: %v", err)
		} else {
			logging.Info("JWT service initialized successfully")
		}
	}()

	// 并行初始化数据库连接
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbClient := pkgdatabase.InitInstance(&config.Database)
		mu.Lock()
		results.dbClient = dbClient
		mu.Unlock()
		logging.Info("Database client initialized successfully with driver: %s", config.Database.Driver)
	}()

	// 并行初始化Redis连接
	wg.Add(1)
	go func() {
		defer wg.Done()
		redisClient := caching.InitInstance(&config.Redis)
		mu.Lock()
		results.redisClient = redisClient
		mu.Unlock()
		logging.Info("Redis client initialized successfully with address: %s", config.Redis.Addr)
	}()

	// 并行创建和配置Gin引擎
	wg.Add(1)
	go func() {
		defer wg.Done()
		engine := createGinEngine(config)
		mu.Lock()
		results.engine = engine
		mu.Unlock()
		logging.Info("Gin engine created and configured successfully")
	}()

	// openaiClient
	wg.Add(1)
	go func() {
		defer wg.Done()
		openai.SetLogger(logging.WithName("OpenAiClient"))
		err := openai.InitializeClient(&config.OpenAI)
		if err != nil {
			mu.Lock()
			results.errors["openai"] = err
			mu.Unlock()
			logging.Warn("Failed to initialize OpenAI client: %v", err)
		} else {
			logging.Info("OpenAI client initialized successfully")
		}
	}()

	wg.Wait()

	var importantCmps []string = append(make([]string, 0), "s3", "email", "sms", "jwt", "openai")

	// // 检查关键组件的初始化错误
	for _, cmp := range importantCmps {
		if results.errors[cmp] != nil {
			return nil, fmt.Errorf(cmp+"服务初始化失败: %w", results.errors[cmp])
		}
	}

	return results, nil
}

// createGinEngine 创建和配置Gin引擎
func createGinEngine(config *configs.AppConfig) *gin.Engine {
	engine := gin.Default()

	// 配置CORS跨域中间件
	if config.Server.CORS.Enabled {
		setupCORS(engine, config)
	} else {
		logging.Info("CORS middleware is disabled")
	}

	// 配置静态文件服务
	if config.Server.Static.Enabled {
		setupStaticFiles(engine, config)
	} else {
		logging.Info("Static file service is disabled")
	}

	return engine
}

// setupCORS 配置CORS中间件
func setupCORS(engine *gin.Engine, config *configs.AppConfig) {
	var corsConfig cors.Config

	if config.Server.Debug {
		corsConfig = cors.Config{
			AllowAllOrigins:  true,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
			AllowHeaders:     []string{"*"},
			ExposeHeaders:    []string{"*"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}
		logging.Warn("Enable CORS - Allow all origins (debug mode)")
	} else {
		corsConfig = cors.Config{
			AllowAllOrigins:  config.Server.CORS.AllowAllOrigins,
			AllowOrigins:     config.Server.CORS.AllowOrigins,
			AllowMethods:     config.Server.CORS.AllowMethods,
			AllowHeaders:     config.Server.CORS.AllowHeaders,
			ExposeHeaders:    config.Server.CORS.ExposeHeaders,
			AllowCredentials: config.Server.CORS.AllowCredentials,
			MaxAge:           time.Duration(config.Server.CORS.MaxAge) * time.Second,
		}
		logging.Info("CORS enabled - Allow origins: %v", config.Server.CORS.AllowOrigins)
	}

	engine.Use(cors.New(corsConfig))
}

// setupStaticFiles 配置静态文件服务
func setupStaticFiles(engine *gin.Engine, config *configs.AppConfig) {
	staticRoot, err := configs.ResolveStaticPath(config.Server.Static.Root)
	if err != nil {
		logging.Warn("Failed to resolve static root path: %v, static file service disabled", err)
		// 如果解析失败，禁用静态文件服务
		config.Server.Static.Enabled = false
		return
	}

	engine.Static(config.Server.Static.Path, staticRoot)
	logging.Info("Static file service enabled at %s, root: %s", config.Server.Static.Path, staticRoot)
}

// initializeEventSystem 初始化事件系统
func initializeEventSystem() {
	basedatabase.InitEventSystem()
	logging.Info("Event system initialized")
}

// setupRoutes 设置路由
func setupRoutes(config *configs.AppConfig, engine *gin.Engine) *gin.Engine {
	router := routes.NewRouter()
	router.SetupRoutes(config, engine)
	logging.Info("Routes setup completed")
	return engine
}
