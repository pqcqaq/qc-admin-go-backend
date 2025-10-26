package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupAreaRoutes 设置地区相关路由
func (r *Router) setupAreaRoutes(rg *gin.RouterGroup) {

	areaHandler := handlers.NewAreaHandler()

	areas := rg.Group("/areas")
	{
		// 基本CRUD操作
		areas.GET("", areaHandler.GetAreas)                    // 获取所有地区
		areas.GET("/page", areaHandler.GetAreasWithPagination) // 分页获取地区列表
		areas.GET("/:id", areaHandler.GetArea)                 // 根据ID获取地区
		areas.POST("", areaHandler.CreateArea)                 // 创建地区
		areas.PUT("/:id", areaHandler.UpdateArea)              // 更新地区
		areas.DELETE("/:id", areaHandler.DeleteArea)           // 删除地区

		// 特殊查询接口
		areas.GET("/tree", areaHandler.GetAreaTree)            // 获取地区树形结构
		areas.GET("/children", areaHandler.GetAreasByParentID) // 根据父级ID获取下一级地区
		areas.GET("/level", areaHandler.GetAreasByLevel)       // 根据级别获取地区
		areas.GET("/depth", areaHandler.GetAreasByDepth)       // 根据深度获取地区
	}
}
