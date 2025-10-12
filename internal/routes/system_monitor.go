package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupSystemMonitorRoutes 设置系统监控相关路由
func (r *Router) setupSystemMonitorRoutes(rg *gin.RouterGroup) {
	systemMonitorHandler := handlers.NewSystemMonitorHandler()

	monitor := rg.Group("/system/monitor")
	{
		// 查询接口
		monitor.GET("/latest", systemMonitorHandler.GetLatest)   // 获取最新状态
		monitor.GET("/history", systemMonitorHandler.GetHistory) // 获取历史记录
		monitor.GET("/range", systemMonitorHandler.GetByRange)   // 按时间范围查询
		monitor.GET("/summary", systemMonitorHandler.GetSummary) // 获取统计摘要

		// 删除接口
		monitor.DELETE("/:id", systemMonitorHandler.Delete)          // 删除单条记录
		monitor.DELETE("/range", systemMonitorHandler.DeleteByRange) // 按时间范围删除
	}
}
