package handlers

import (
	"net/http"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/pkg/excel"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// LoggingHandler 日志处理器
type LoggingHandler struct {
}

// NewLoggingHandler 创建新的日志处理器
func NewLoggingHandler() *LoggingHandler {
	return &LoggingHandler{}
}

// GetLoggings 获取所有日志记录
// @Summary      获取所有日志记录
// @Description  获取系统中所有日志记录的列表
// @Tags         loggings
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /loggings [get]
func (h *LoggingHandler) GetLoggings(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	loggings, err := funcs.LoggingFunc{}.GetAllLoggings(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取日志记录列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    loggings,
		"count":   len(loggings),
	})
}

// GetLoggingsWithPagination 分页获取日志记录列表
// @Summary      分页获取日志记录列表
// @Description  根据分页参数获取日志记录列表，支持多种过滤条件
// @Tags         loggings
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "页码"          default(1)
// @Param        page_size  query     int     false  "每页数量"       default(10)
// @Param        order      query     string  false  "排序方式"       default(desc)
// @Param        order_by   query     string  false  "排序字段"       default(create_time)
// @Param        level      query     string  false  "日志级别"       Enums(debug,info,error,warn,fatal)
// @Param        type       query     string  false  "日志类型"       Enums(Error,Panic,manul)
// @Param        message    query     string  false  "消息内容"
// @Param        method     query     string  false  "HTTP方法"
// @Param        path       query     string  false  "请求路径"
// @Param        ip         query     string  false  "IP地址"
// @Param        code       query     int     false  "状态码"
// @Param        beginTime  query     string  false  "开始时间"
// @Param        endTime    query     string  false  "结束时间"
// @Success      200  {object}  object{success=bool,data=[]object,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /loggings/page [get]
func (h *LoggingHandler) GetLoggingsWithPagination(c *gin.Context) {
	var req models.PageLoggingRequest

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
	ctx := middleware.GetRequestContext(c)
	result, err := funcs.LoggingFunc{}.GetLoggingWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取日志记录列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetLogging 根据ID获取日志记录
// @Summary      根据ID获取日志记录
// @Description  根据日志记录ID获取日志记录详细信息
// @Tags         loggings
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "日志记录ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /loggings/{id} [get]
func (h *LoggingHandler) GetLogging(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("日志记录ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("日志记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	logging, err := funcs.LoggingFunc{}.GetLoggingById(ctx, id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "logging not found" ||
			err.Error() == "logging with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("日志记录未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询日志记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    logging,
	})
}

// CreateLogging 创建日志记录
// @Summary      创建日志记录
// @Description  创建新的日志记录
// @Tags         loggings
// @Accept       json
// @Produce      json
// @Param        logging  body      models.CreateLoggingRequest  true  "日志记录信息"
// @Success      201      {object}  object{success=bool,data=object}
// @Failure      400      {object}  object{success=bool,message=string}
// @Failure      500      {object}  object{success=bool,message=string}
// @Router       /loggings [post]
func (h *LoggingHandler) CreateLogging(c *gin.Context) {
	var req models.CreateLoggingRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Message == "" {
		middleware.ThrowError(c, middleware.BadRequestError("日志消息不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	logging, err := funcs.LoggingFunc{}.CreateLogging(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建日志记录失败", err.Error()))
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    logging,
		"message": "日志记录创建成功",
	})
}

// UpdateLogging 更新日志记录
// @Summary      更新日志记录
// @Description  根据ID更新日志记录信息
// @Tags         loggings
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "日志记录ID"
// @Param        logging  body      models.UpdateLoggingRequest  true  "日志记录信息"
// @Success      200      {object}  object{success=bool,data=object}
// @Failure      400      {object}  object{success=bool,message=string}
// @Failure      404      {object}  object{success=bool,message=string}
// @Failure      500      {object}  object{success=bool,message=string}
// @Router       /loggings/{id} [put]
func (h *LoggingHandler) UpdateLogging(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("日志记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateLoggingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Message == "" {
		middleware.ThrowError(c, middleware.BadRequestError("日志消息不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	logging, err := funcs.LoggingFunc{}.UpdateLogging(ctx, id, &req)
	if err != nil {
		if err.Error() == "logging not found" ||
			err.Error() == "logging with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("日志记录未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新日志记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    logging,
		"message": "日志记录更新成功",
	})
}

// DeleteLogging 删除日志记录
// @Summary      删除日志记录
// @Description  根据ID删除日志记录
// @Tags         loggings
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "日志记录ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /loggings/{id} [delete]
func (h *LoggingHandler) DeleteLogging(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("日志记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.LoggingFunc{}.DeleteLogging(ctx, id)
	if err != nil {
		if err.Error() == "logging not found" ||
			err.Error() == "logging with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("日志记录未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除日志记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "日志记录删除成功",
	})
}

// ExportLoggingsToExcel 导出日志记录为Excel
// @Summary      导出日志记录为Excel
// @Description  将日志记录导出为Excel文件
// @Tags         loggings
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        page       query     int     false  "页码"          default(1)
// @Param        page_size  query     int     false  "每页数量"       default(10000)
// @Param        order      query     string  false  "排序方式"       default(desc)
// @Param        order_by   query     string  false  "排序字段"       default(create_time)
// @Param        level      query     string  false  "日志级别"       Enums(debug,info,error,warn,fatal)
// @Param        type       query     string  false  "日志类型"       Enums(Error,Panic,manul)
// @Param        message    query     string  false  "消息内容"
// @Param        method     query     string  false  "HTTP方法"
// @Param        path       query     string  false  "请求路径"
// @Param        ip         query     string  false  "IP地址"
// @Param        code       query     int     false  "状态码"
// @Param        beginTime  query     string  false  "开始时间"
// @Param        endTime    query     string  false  "结束时间"
// @Success      200  {file}   file    "Excel文件"
// @Failure      400  {object} object{success=bool,message=string}
// @Failure      500  {object} object{success=bool,message=string}
// @Router       /loggings/export [get]
func (h *LoggingHandler) ExportLoggingsToExcel(c *gin.Context) {
	var req models.PageLoggingRequest

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
	ctx := middleware.GetRequestContext(c)
	result, err := funcs.LoggingFunc{}.GetLoggingWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取日志记录失败", err.Error()))
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
			Header:    "日志级别",
			Width:     15,
			FieldName: "Level",
		},
		{
			Header:    "日志类型",
			Width:     15,
			FieldName: "Type",
		},
		{
			Header:    "消息内容",
			Width:     50,
			FieldName: "Message",
		},
		{
			Header:    "HTTP方法",
			Width:     15,
			FieldName: "Method",
		},
		{
			Header:    "请求路径",
			Width:     40,
			FieldName: "Path",
		},
		{
			Header:    "IP地址",
			Width:     20,
			FieldName: "IP",
		},
		{
			Header:    "查询参数",
			Width:     30,
			FieldName: "Query",
		},
		{
			Header:    "状态码",
			Width:     15,
			FieldName: "Code",
		},
		{
			Header:    "用户代理",
			Width:     50,
			FieldName: "UserAgent",
		},
		{
			Header:    "创建时间",
			Width:     25,
			FieldName: "CreateTime",
			Formatter: excel.TimeFormatter("2006-01-02 15:04:05"),
		},
		{
			Header:    "更新时间",
			Width:     25,
			FieldName: "UpdateTime",
			Formatter: excel.TimeFormatter("2006-01-02 15:04:05"),
		},
	}

	// 创建Excel处理器
	processor := excel.NewExcelProcessor("日志记录", columns)

	// 生成Excel文件
	file, err := processor.GenerateExcelStream(result.Data)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("生成Excel文件失败", err.Error()))
		return
	}

	// 生成文件名
	filename := excel.GenerateFilename("日志记录")

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
