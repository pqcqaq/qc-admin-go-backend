package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupScanRoutes 设置扫描相关路由
func (r *Router) setupScanRoutes(rg *gin.RouterGroup) {

	scanHandler := handlers.NewScanHandler()

	scans := rg.Group("/scans")
	{
		scans.GET("", scanHandler.GetScans)
		scans.GET("/page", scanHandler.GetScansWithPagination)
		scans.GET("/:id", scanHandler.GetScan)
		scans.POST("", scanHandler.CreateScan)
		scans.PUT("/:id", scanHandler.UpdateScan)
		scans.DELETE("/:id", scanHandler.DeleteScan)
	}
}
