package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupClientDeviceRoutes 设置客户端设备相关路由
func (r *Router) setupClientDeviceRoutes(rg *gin.RouterGroup) {

	clientDeviceHandler := handlers.NewClientDeviceHandler()

	clientDevices := rg.Group("/client-devices")
	{
		// 基础CRUD操作
		clientDevices.GET("", clientDeviceHandler.GetClientDevices)
		clientDevices.GET("/page", clientDeviceHandler.GetClientDevicesWithPagination)
		clientDevices.GET("/export", clientDeviceHandler.ExportClientDevicesToExcel)
		clientDevices.GET("/:id", clientDeviceHandler.GetClientDevice)
		clientDevices.POST("", clientDeviceHandler.CreateClientDevice)
		clientDevices.PUT("/:id", clientDeviceHandler.UpdateClientDevice)
		clientDevices.DELETE("/:id", clientDeviceHandler.DeleteClientDevice)

		// 根据code获取设备信息
		clientDevices.GET("/code/:code", clientDeviceHandler.GetClientDeviceByCode)

		// 检查客户端访问权限
		clientDevices.POST("/check-access", clientDeviceHandler.CheckClientAccess)
	}
}
