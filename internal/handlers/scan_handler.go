package handlers

import (
	"context"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// ScanHandler 扫描处理器
type ScanHandler struct {
}

// NewScanHandler 创建新的扫描处理器
func NewScanHandler() *ScanHandler {
	return &ScanHandler{}
}

// GetScans 获取所有扫描记录
func (h *ScanHandler) GetScans(c *gin.Context) {
	scans, err := funcs.GetAllScans(context.Background())
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取扫描记录列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    scans,
		"count":   len(scans),
	})
}

// GetScansWithPagination 分页获取扫描记录列表
func (h *ScanHandler) GetScansWithPagination(c *gin.Context) {
	var req models.PageScansRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "desc"
	req.OrderBy = "create_time"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 调用服务层方法
	result, err := funcs.GetScanWithPagination(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取扫描记录列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetScan 根据ID获取扫描记录
func (h *ScanHandler) GetScan(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	scan, err := funcs.GetScanById(context.Background(), id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "scan not found" ||
			err.Error() == "scan with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("扫描记录未找到", map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询扫描记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    scan,
	})
}

// CreateScan 创建扫描记录
func (h *ScanHandler) CreateScan(c *gin.Context) {
	var req models.CreateScanRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Content == "" {
		middleware.ThrowError(c, middleware.BadRequestError("扫描内容不能为空", nil))
		return
	}

	scan, err := funcs.CreateScan(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建扫描记录失败", err.Error()))
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    scan,
		"message": "扫描记录创建成功",
	})
}

// UpdateScan 更新扫描记录
func (h *ScanHandler) UpdateScan(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Content == "" {
		middleware.ThrowError(c, middleware.BadRequestError("扫描内容不能为空", nil))
		return
	}

	scan, err := funcs.UpdateScan(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "scan not found" ||
			err.Error() == "scan with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("扫描记录未找到", map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新扫描记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    scan,
		"message": "扫描记录更新成功",
	})
}

// DeleteScan 删除扫描记录
func (h *ScanHandler) DeleteScan(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteScan(context.Background(), id)
	if err != nil {
		if err.Error() == "scan not found" ||
			err.Error() == "scan with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("扫描记录未找到", map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除扫描记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "扫描记录删除成功",
	})
}
