package routes

import (
	"go-backend/internal/handlers"
	"go-backend/internal/middleware"
	"go-backend/pkg/configs"
	"go-backend/pkg/logging"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-backend/docs" // 导入swagger文档
)

// Router 路由配置结构
type Router struct {
}

// NewRouter 创建新的路由配置
func NewRouter() *Router {
	return &Router{}
}

// SetupRoutes 设置所有路由
func (r *Router) SetupRoutes(config *configs.AppConfig, engine *gin.Engine) {
	// 注册错误处理中间件
	engine.Use(middleware.ErrorHandlerMiddleware()) // 处理panic恢复
	engine.Use(middleware.ErrorHandler())           // 处理gin.Error
	// 注册API认证中间件
	engine.Use(middleware.APIAuthMiddleware(engine))
	engine.Use(middleware.JWTAuthMiddleware())

	middleware.RegisterConfigMiddlewares(engine)

	// Swagger文档路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 健康检查端点
	healthHandler := handlers.NewHealthHandler()
	engine.GET("/health", healthHandler.Health)

	logging.WithName("Router").Info("Setting up routes with prefix: %s", config.Server.Prefix)
	prefixGroup := engine.Group(config.Server.Prefix)

	// API v1 路由组
	api := prefixGroup.Group("/v1")
	{
		r.setupTestRoutes(api)
		r.setupAuthRoutes(api)
		r.setupUserRoutes(api)
		r.setupAttachmentRoutes(api)
		r.setupScanRoutes(api)
		r.setupDemoRoutes(api)
		r.setupRBACRoutes(api)
		r.setupLoginRecordRoutes(api)
		r.setupAPIAuthRoutes(api)
		r.setupLoggingRoutes(api)
		r.setupClientDeviceRoutes(api)
		r.setupSystemMonitorRoutes(api)
	}
}
