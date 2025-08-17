package routes

import (
	"go-backend/internal/handlers"
	"go-backend/internal/middleware"

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
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// 注册错误处理中间件
	engine.Use(middleware.ErrorHandlerMiddleware()) // 处理panic恢复
	engine.Use(middleware.ErrorHandler())           // 处理gin.Error

	// Swagger文档路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 健康检查端点
	healthHandler := handlers.NewHealthHandler()
	engine.GET("/health", healthHandler.Health)

	// API v1 路由组
	api := engine.Group("/api/v1")
	{
		r.setupUserRoutes(api)
		r.setupAttachmentRoutes(api)
		r.setupScanRoutes(api)
		r.setupDemoRoutes(api)
		r.setupRBACRoutes(api)
	}
}
