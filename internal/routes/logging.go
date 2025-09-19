package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupLoggingRoutes 设置日志相关路由
func (r *Router) setupLoggingRoutes(rg *gin.RouterGroup) {

	loggingHandler := handlers.NewLoggingHandler()

	loggings := rg.Group("/loggings")
	{
		loggings.GET("", loggingHandler.GetLoggings)
		loggings.GET("/page", loggingHandler.GetLoggingsWithPagination)
		loggings.GET("/export", loggingHandler.ExportLoggingsToExcel)
		loggings.GET("/:id", loggingHandler.GetLogging)
		loggings.POST("", loggingHandler.CreateLogging)
		loggings.PUT("/:id", loggingHandler.UpdateLogging)
		loggings.DELETE("/:id", loggingHandler.DeleteLogging)
	}
}
