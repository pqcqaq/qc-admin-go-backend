package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func (r *Router) setupAPIAuthRoutes(rg *gin.RouterGroup) {
	apiAuthHandler := handlers.NewAPIAuthHandler()

	// API认证路由组
	apiAuthGroup := rg.Group("/apiauth")
	{
		// 获取所有API认证记录
		apiAuthGroup.GET("", apiAuthHandler.GetAPIAuths)

		// 分页获取API认证记录
		apiAuthGroup.GET("/page", apiAuthHandler.GetAPIAuthsWithPagination)

		// 导出API认证记录为Excel
		apiAuthGroup.GET("/export", apiAuthHandler.ExportAPIAuthsToExcel)

		// 根据ID获取API认证记录
		apiAuthGroup.GET("/:id", apiAuthHandler.GetAPIAuth)

		// 创建API认证记录
		apiAuthGroup.POST("", apiAuthHandler.CreateAPIAuth)

		// 更新API认证记录
		apiAuthGroup.PUT("/:id", apiAuthHandler.UpdateAPIAuth)

		// 删除API认证记录
		apiAuthGroup.DELETE("/:id", apiAuthHandler.DeleteAPIAuth)
	}
}
