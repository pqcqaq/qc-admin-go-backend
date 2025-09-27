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

// APIAuthHandler API认证处理器
type APIAuthHandler struct {
}

// NewAPIAuthHandler 创建新的API认证处理器
func NewAPIAuthHandler() *APIAuthHandler {
	return &APIAuthHandler{}
}

// GetAPIAuths 获取所有API认证记录
// @Summary      获取所有API认证记录
// @Description  获取系统中所有API认证记录的列表
// @Tags         apiauth
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /apiauth [get]
func (h *APIAuthHandler) GetAPIAuths(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	apiAuths, err := funcs.ApiAuthFuncs{}.GetAllAPIAuths(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取API认证记录列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    apiAuths,
		"count":   len(apiAuths),
	})
}

// GetAPIAuthsWithPagination 分页获取API认证记录列表
// @Summary      分页获取API认证记录列表
// @Description  根据分页参数获取API认证记录列表
// @Tags         apiauth
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        order     query     string  false  "排序方式"  default(desc)
// @Param        order_by  query     string  false  "排序字段"  default(create_time)
// @Param        name      query     string  false  "按名称模糊搜索"
// @Param        method    query     string  false  "按HTTP方法过滤"
// @Param        path      query     string  false  "按路径模糊搜索"
// @Param        isPublic  query     bool    false  "按是否公开过滤"
// @Param        isActive  query     bool    false  "按是否启用过滤"
// @Param        beginTime query     string  false  "开始时间"
// @Param        endTime   query     string  false  "结束时间"
// @Success      200  {object}  object{success=bool,data=[]object,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /apiauth/page [get]
func (h *APIAuthHandler) GetAPIAuthsWithPagination(c *gin.Context) {
	var req models.PageAPIAuthRequest

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
	result, err := funcs.ApiAuthFuncs{}.GetAPIAuthWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取API认证记录列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetAPIAuth 根据ID获取API认证记录
// @Summary      根据ID获取API认证记录
// @Description  根据API认证记录ID获取API认证记录详细信息
// @Tags         apiauth
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "API认证记录ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /apiauth/{id} [get]
func (h *APIAuthHandler) GetAPIAuth(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("API认证记录ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("API认证记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	apiAuth, err := funcs.ApiAuthFuncs{}.GetAPIAuthById(ctx, id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "api auth not found" ||
			err.Error() == "api auth with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("API认证记录未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询API认证记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    apiAuth,
	})
}

// CreateAPIAuth 创建API认证记录
// @Summary      创建API认证记录
// @Description  创建新的API认证记录
// @Tags         apiauth
// @Accept       json
// @Produce      json
// @Param        apiauth  body      models.CreateAPIAuthRequest  true  "API认证记录信息"
// @Success      201   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /apiauth [post]
func (h *APIAuthHandler) CreateAPIAuth(c *gin.Context) {
	var req models.CreateAPIAuthRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("API名称不能为空", nil))
		return
	}
	if req.Method == "" {
		middleware.ThrowError(c, middleware.BadRequestError("HTTP方法不能为空", nil))
		return
	}
	if req.Path == "" {
		middleware.ThrowError(c, middleware.BadRequestError("API路径不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	apiAuth, err := funcs.ApiAuthFuncs{}.CreateAPIAuth(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建API认证记录失败", err.Error()))
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    apiAuth,
		"message": "API认证记录创建成功",
	})
}

// UpdateAPIAuth 更新API认证记录
// @Summary      更新API认证记录
// @Description  根据ID更新API认证记录信息
// @Tags         apiauth
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "API认证记录ID"
// @Param        apiauth  body      models.UpdateAPIAuthRequest  true  "API认证记录信息"
// @Success      200   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /apiauth/{id} [put]
func (h *APIAuthHandler) UpdateAPIAuth(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("API认证记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateAPIAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("API名称不能为空", nil))
		return
	}
	if req.Method == "" {
		middleware.ThrowError(c, middleware.BadRequestError("HTTP方法不能为空", nil))
		return
	}
	if req.Path == "" {
		middleware.ThrowError(c, middleware.BadRequestError("API路径不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	apiAuth, err := funcs.ApiAuthFuncs{}.UpdateAPIAuth(ctx, id, &req)
	if err != nil {
		if err.Error() == "api auth not found" ||
			err.Error() == "api auth with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("API认证记录未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新API认证记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    apiAuth,
		"message": "API认证记录更新成功",
	})
}

// DeleteAPIAuth 删除API认证记录
// @Summary      删除API认证记录
// @Description  根据ID删除API认证记录
// @Tags         apiauth
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "API认证记录ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /apiauth/{id} [delete]
func (h *APIAuthHandler) DeleteAPIAuth(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("API认证记录ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.ApiAuthFuncs{}.DeleteAPIAuth(ctx, id)
	if err != nil {
		if err.Error() == "api auth not found" ||
			err.Error() == "api auth with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("API认证记录未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除API认证记录失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "API认证记录删除成功",
	})
}

// ExportAPIAuthsToExcel 导出API认证记录为Excel
// @Summary      导出API认证记录为Excel
// @Description  将API认证记录导出为Excel文件
// @Tags         apiauth
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10000)
// @Param        order     query     string  false  "排序方式"  default(desc)
// @Param        order_by  query     string  false  "排序字段"  default(create_time)
// @Param        name      query     string  false  "按名称模糊搜索"
// @Param        method    query     string  false  "按HTTP方法过滤"
// @Param        path      query     string  false  "按路径模糊搜索"
// @Param        isPublic  query     bool    false  "按是否公开过滤"
// @Param        isActive  query     bool    false  "按是否启用过滤"
// @Param        beginTime query     string  false  "开始时间"
// @Param        endTime   query     string  false  "结束时间"
// @Success      200  {file}   file    "Excel文件"
// @Failure      400  {object} object{success=bool,message=string}
// @Failure      500  {object} object{success=bool,message=string}
// @Router       /apiauth/export [get]
func (h *APIAuthHandler) ExportAPIAuthsToExcel(c *gin.Context) {
	var req models.PageAPIAuthRequest

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
	result, err := funcs.ApiAuthFuncs{}.GetAPIAuthWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取API认证记录失败", err.Error()))
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
			Header:    "API名称",
			Width:     25,
			FieldName: "Name",
		},
		{
			Header:    "API描述",
			Width:     40,
			FieldName: "Description",
		},
		{
			Header:    "HTTP方法",
			Width:     15,
			FieldName: "Method",
		},
		{
			Header:    "API路径",
			Width:     40,
			FieldName: "Path",
		},
		{
			Header:    "是否公开",
			Width:     15,
			FieldName: "IsPublic",
			Formatter: excel.BoolFormatter("是", "否"),
		},
		{
			Header:    "是否启用",
			Width:     15,
			FieldName: "IsActive",
			Formatter: excel.BoolFormatter("启用", "禁用"),
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
	processor := excel.NewExcelProcessor("API认证记录", columns)

	// 生成Excel文件
	file, err := processor.GenerateExcelStream(result.Data)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("生成Excel文件失败", err.Error()))
		return
	}

	// 生成文件名
	filename := excel.GenerateFilename("API认证记录")

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
