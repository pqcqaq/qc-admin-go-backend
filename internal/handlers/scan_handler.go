package handlers

import (
	"context"
	"net/http"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/pkg/excel"
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
// @Summary      获取所有扫描记录
// @Description  获取系统中所有扫描记录的列表
// @Tags         scans
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /scans [get]
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
// @Summary      分页获取扫描记录列表
// @Description  根据分页参数获取扫描记录列表
// @Tags         scans
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        order     query     string  false  "排序方式"  default(desc)
// @Param        order_by  query     string  false  "排序字段"  default(create_time)
// @Success      200  {object}  object{success=bool,data=[]object,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /scans/page [get]
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
// @Summary      根据ID获取扫描记录
// @Description  根据扫描记录ID获取扫描记录详细信息
// @Tags         scans
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "扫描记录ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /scans/{id} [get]
func (h *ScanHandler) GetScan(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	scan, err := funcs.GetScanById(context.Background(), id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "scan not found" ||
			err.Error() == "scan with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("扫描记录未找到", map[string]any{
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
// @Summary      创建扫描记录
// @Description  创建新的扫描记录
// @Tags         scans
// @Accept       json
// @Produce      json
// @Param        scan  body      models.CreateScanRequest  true  "扫描记录信息"
// @Success      201   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /scans [post]
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
// @Summary      更新扫描记录
// @Description  根据ID更新扫描记录信息
// @Tags         scans
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "扫描记录ID"
// @Param        scan  body      models.UpdateScanRequest  true  "扫描记录信息"
// @Success      200   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /scans/{id} [put]
func (h *ScanHandler) UpdateScan(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID格式无效", map[string]any{
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
			middleware.ThrowError(c, middleware.NotFoundError("扫描记录未找到", map[string]any{
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
		middleware.ThrowError(c, middleware.BadRequestError("扫描记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteScan(context.Background(), id)
	if err != nil {
		if err.Error() == "scan not found" ||
			err.Error() == "scan with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("扫描记录未找到", map[string]any{
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

// ExportScansToExcel 导出扫描记录为Excel
// @Summary      导出扫描记录为Excel
// @Description  将扫描记录导出为Excel文件
// @Tags         scans
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10000)
// @Param        order     query     string  false  "排序方式"  default(desc)
// @Param        order_by  query     string  false  "排序字段"  default(create_time)
// @Success      200  {file}   file    "Excel文件"
// @Failure      400  {object} object{success=bool,message=string}
// @Failure      500  {object} object{success=bool,message=string}
// @Router       /scans/export [get]
func (h *ScanHandler) ExportScansToExcel(c *gin.Context) {
	var req models.PageScansRequest

	// 设置默认值，但不限制数量（用于导出）
	req.Page = 1
	req.PageSize = 10000 // 设置一个较大的值来获取所有数据
	req.Order = "desc"
	req.OrderBy = "create_time"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 获取数据
	result, err := funcs.GetScanWithPagination(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取扫描记录失败", err.Error()))
		return
	}

	// 配置Excel列
	columns := []excel.ColumnConfig{
		{
			Header:    "ID",
			Width:     15,
			FieldName: "ID",
		},
		{
			Header:    "扫描内容",
			Width:     40,
			FieldName: "Content",
		},
		{
			Header:    "是否成功",
			Width:     15,
			FieldName: "Success",
			Formatter: excel.BoolFormatter("成功", "失败"),
		},
		{
			Header:    "创建时间",
			Width:     25,
			FieldName: "CreateTime",
			Formatter: excel.TimeFormatter("2006-01-02 15:04:05"),
		},
		{
			Header:    "图片ID",
			Width:     15,
			FieldName: "ImageId",
		},
		{
			Header:    "图片URL",
			Width:     50,
			FieldName: "ImageUrl",
		},
	}

	// 创建Excel处理器
	processor := excel.NewExcelProcessor("扫描记录", columns)

	// 生成Excel文件
	file, err := processor.GenerateExcelStream(result.Data)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("生成Excel文件失败", err.Error()))
		return
	}

	// 生成文件名
	filename := excel.GenerateFilename("扫描记录")

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Cache-Control", "no-cache")

	// 将Excel文件写入响应流
	if err := file.Write(c.Writer); err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("写入Excel文件失败", err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
