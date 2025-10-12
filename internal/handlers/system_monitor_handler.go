package handlers

import (
	"net/http"
	"time"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// SystemMonitorHandler 系统监控处理器
type SystemMonitorHandler struct {
}

// NewSystemMonitorHandler 创建新的系统监控处理器
func NewSystemMonitorHandler() *SystemMonitorHandler {
	return &SystemMonitorHandler{}
}

// GetLatest 获取最新的系统监控状态
// @Summary      获取最新系统监控状态
// @Description  获取系统最新的监控数据，包括CPU、内存、磁盘、网络等信息
// @Tags         system-monitor
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=models.SystemMonitorResponse}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /system/monitor/latest [get]
func (h *SystemMonitorHandler) GetLatest(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)

	monitor, err := funcs.SystemMonitorFunc{}.GetLatestSystemMonitor(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取最新系统监控状态失败", err.Error()))
		return
	}

	if monitor == nil {
		middleware.ThrowError(c, middleware.NotFoundError("未找到系统监控数据", ""))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitor,
	})
}

// GetHistory 获取系统监控历史记录
// @Summary      获取系统监控历史记录
// @Description  根据查询参数获取最近的系统监控历史数据
// @Tags         system-monitor
// @Accept       json
// @Produce      json
// @Param        limit  query     int  false  "返回记录数量(1-1000)"  default(100)
// @Param        hours  query     int  false  "查询最近多少小时(1-168)" default(1)
// @Success      200  {object}  object{success=bool,data=[]models.SystemMonitorResponse,count=int}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /system/monitor/history [get]
func (h *SystemMonitorHandler) GetHistory(c *gin.Context) {
	var req models.SystemMonitorHistoryRequest

	// 设置默认值
	defaultLimit := 100
	defaultHours := 1
	req.Limit = &defaultLimit
	req.Hours = &defaultHours

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	monitors, err := funcs.SystemMonitorFunc{}.GetSystemMonitorHistory(ctx, *req.Limit, *req.Hours)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取系统监控历史记录失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitors,
		"count":   len(monitors),
	})
}

// GetByRange 根据时间范围获取系统监控数据
// @Summary      根据时间范围获取系统监控数据
// @Description  根据指定的开始和结束时间获取系统监控数据
// @Tags         system-monitor
// @Accept       json
// @Produce      json
// @Param        start  query     string  true  "开始时间(ISO 8601格式)"
// @Param        end    query     string  true  "结束时间(ISO 8601格式)"
// @Success      200  {object}  object{success=bool,data=[]models.SystemMonitorResponse,count=int}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /system/monitor/range [get]
func (h *SystemMonitorHandler) GetByRange(c *gin.Context) {
	var req models.SystemMonitorRangeRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 验证时间格式
	startTime, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		middleware.ThrowError(c, middleware.ValidationError("开始时间格式错误，请使用ISO 8601格式", err.Error()))
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		middleware.ThrowError(c, middleware.ValidationError("结束时间格式错误，请使用ISO 8601格式", err.Error()))
		return
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		middleware.ThrowError(c, middleware.ValidationError("结束时间不能早于开始时间", ""))
		return
	}

	ctx := middleware.GetRequestContext(c)
	monitors, err := funcs.SystemMonitorFunc{}.GetSystemMonitorByRange(ctx, startTime, endTime)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取系统监控数据失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitors,
		"count":   len(monitors),
	})
}

// GetSummary 获取系统监控统计摘要
// @Summary      获取系统监控统计摘要
// @Description  获取指定时间范围内的系统监控统计信息，包括平均值、最大值、最小值等
// @Tags         system-monitor
// @Accept       json
// @Produce      json
// @Param        hours  query     int  false  "查询最近多少小时(1-720)" default(24)
// @Success      200  {object}  object{success=bool,data=models.SystemMonitorSummaryResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /system/monitor/summary [get]
func (h *SystemMonitorHandler) GetSummary(c *gin.Context) {
	var req models.SystemMonitorSummaryRequest

	// 设置默认值
	defaultHours := 24
	req.Hours = &defaultHours

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	summary, err := funcs.SystemMonitorFunc{}.GetSystemMonitorSummary(ctx, *req.Hours)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取系统监控统计摘要失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// Delete 删除系统监控记录
// @Summary      删除系统监控记录
// @Description  根据ID删除指定的系统监控记录
// @Tags         system-monitor
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "监控记录ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /system/monitor/{id} [delete]
func (h *SystemMonitorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.ThrowError(c, middleware.ValidationError("监控记录ID不能为空", ""))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err := funcs.SystemMonitorFunc{}.DeleteSystemMonitor(ctx, id)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("删除系统监控记录失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统监控记录删除成功",
	})
}

// DeleteByRange 根据时间范围删除系统监控记录
// @Summary      根据时间范围删除系统监控记录
// @Description  删除指定时间范围内的所有系统监控记录
// @Tags         system-monitor
// @Accept       json
// @Produce      json
// @Param        start  query     string  true  "开始时间(ISO 8601格式)"
// @Param        end    query     string  true  "结束时间(ISO 8601格式)"
// @Success      200  {object}  object{success=bool,data=models.DeleteSystemMonitorRangeResponse,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /system/monitor/range [delete]
func (h *SystemMonitorHandler) DeleteByRange(c *gin.Context) {
	var req models.SystemMonitorRangeRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 验证时间格式
	startTime, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		middleware.ThrowError(c, middleware.ValidationError("开始时间格式错误，请使用ISO 8601格式", err.Error()))
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		middleware.ThrowError(c, middleware.ValidationError("结束时间格式错误，请使用ISO 8601格式", err.Error()))
		return
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		middleware.ThrowError(c, middleware.ValidationError("结束时间不能早于开始时间", ""))
		return
	}

	ctx := middleware.GetRequestContext(c)
	deleted, err := funcs.SystemMonitorFunc{}.DeleteSystemMonitorByRange(ctx, startTime, endTime)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("删除系统监控记录失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.DeleteSystemMonitorRangeResponse{
			Deleted: deleted,
		},
		"message": "系统监控记录批量删除成功",
	})
}
