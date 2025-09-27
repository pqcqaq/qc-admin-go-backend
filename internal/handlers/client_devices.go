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

// ClientDeviceHandler 客户端设备处理器
type ClientDeviceHandler struct {
}

// NewClientDeviceHandler 创建新的客户端设备处理器
func NewClientDeviceHandler() *ClientDeviceHandler {
	return &ClientDeviceHandler{}
}

// GetClientDevices 获取所有客户端设备
// @Summary      获取所有客户端设备
// @Description  获取系统中所有客户端设备的列表
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /client-devices [get]
func (h *ClientDeviceHandler) GetClientDevices(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	devices, err := funcs.ClientDeviceFuncs{}.GetAllClientDevices(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取客户端设备列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    devices,
		"count":   len(devices),
	})
}

// GetClientDevicesWithPagination 分页获取客户端设备列表
// @Summary      分页获取客户端设备列表
// @Description  根据分页参数获取客户端设备列表
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        order     query     string  false  "排序方式"  default(desc)
// @Param        order_by  query     string  false  "排序字段"  default(create_time)
// @Param        name      query     string  false  "按名称模糊搜索"
// @Param        code      query     string  false  "按code精确搜索"
// @Param        enabled   query     bool    false  "按启用状态过滤"
// @Param        anonymous query     bool    false  "按匿名登录过滤"
// @Success      200  {object}  object{success=bool,data=[]object,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /client-devices/page [get]
func (h *ClientDeviceHandler) GetClientDevicesWithPagination(c *gin.Context) {
	var req models.PageClientDevicesRequest

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
	result, err := funcs.ClientDeviceFuncs{}.GetClientDevicesWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取客户端设备列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetClientDevice 根据ID获取客户端设备
// @Summary      根据ID获取客户端设备
// @Description  根据客户端设备ID获取设备详细信息
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "客户端设备ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /client-devices/{id} [get]
func (h *ClientDeviceHandler) GetClientDevice(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("客户端设备ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("客户端设备ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	device, err := funcs.ClientDeviceFuncs{}.GetClientDeviceById(ctx, id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "client device not found" ||
			err.Error() == "client device with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("客户端设备未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询客户端设备失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    device,
	})
}

// GetClientDeviceByCode 根据code获取客户端设备
// @Summary      根据code获取客户端设备
// @Description  根据客户端设备code获取设备信息
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        code   path      string  true  "客户端设备code"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /client-devices/code/{code} [get]
func (h *ClientDeviceHandler) GetClientDeviceByCode(c *gin.Context) {
	code := c.Param("code")

	// 验证code参数
	if code == "" {
		middleware.ThrowError(c, middleware.BadRequestError("客户端设备code不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	device, err := funcs.ClientDeviceFuncs{}.GetClientDeviceByCode(ctx, code)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "client device not found" ||
			err.Error() == "client device with code "+code+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("客户端设备未找到", map[string]any{
				"code": code,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询客户端设备失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    device,
	})
}

// CreateClientDevice 创建客户端设备
// @Summary      创建客户端设备
// @Description  创建新的客户端设备
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        device  body      models.CreateClientDeviceRequest  true  "客户端设备信息"
// @Success      201   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /client-devices [post]
func (h *ClientDeviceHandler) CreateClientDevice(c *gin.Context) {
	var req models.CreateClientDeviceRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("设备名称不能为空", nil))
		return
	}

	if req.AccessTokenExpiry < 1000 {
		middleware.ThrowError(c, middleware.BadRequestError("AccessToken超时时间不能小于1000毫秒", nil))
		return
	}

	if req.RefreshTokenExpiry < 1000 {
		middleware.ThrowError(c, middleware.BadRequestError("RefreshToken超时时间不能小于1000毫秒", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	device, err := funcs.ClientDeviceFuncs{}.CreateClientDevice(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建客户端设备失败", err.Error()))
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    device,
		"message": "客户端设备创建成功",
	})
}

// UpdateClientDevice 更新客户端设备
// @Summary      更新客户端设备
// @Description  根据ID更新客户端设备信息
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        id      path      int                             true  "客户端设备ID"
// @Param        device  body      models.UpdateClientDeviceRequest  true  "客户端设备信息"
// @Success      200   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /client-devices/{id} [put]
func (h *ClientDeviceHandler) UpdateClientDevice(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("客户端设备ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateClientDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("设备名称不能为空", nil))
		return
	}

	if req.AccessTokenExpiry != nil && *req.AccessTokenExpiry < 1000 {
		middleware.ThrowError(c, middleware.BadRequestError("AccessToken超时时间不能小于1000毫秒", nil))
		return
	}

	if req.RefreshTokenExpiry != nil && *req.RefreshTokenExpiry < 1000 {
		middleware.ThrowError(c, middleware.BadRequestError("RefreshToken超时时间不能小于1000毫秒", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	device, err := funcs.ClientDeviceFuncs{}.UpdateClientDevice(ctx, id, &req)
	if err != nil {
		if err.Error() == "client device not found" ||
			err.Error() == "client device with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("客户端设备未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新客户端设备失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    device,
		"message": "客户端设备更新成功",
	})
}

// DeleteClientDevice 删除客户端设备
// @Summary      删除客户端设备
// @Description  根据ID删除客户端设备
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "客户端设备ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /client-devices/{id} [delete]
func (h *ClientDeviceHandler) DeleteClientDevice(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("客户端设备ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.ClientDeviceFuncs{}.DeleteClientDevice(ctx, id)
	if err != nil {
		if err.Error() == "client device not found" ||
			err.Error() == "client device with id "+strconv.FormatUint(id, 10)+" not found" {
			middleware.ThrowError(c, middleware.NotFoundError("客户端设备未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除客户端设备失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "客户端设备删除成功",
	})
}

// CheckClientAccess 检查客户端访问权限
// @Summary      检查客户端访问权限
// @Description  检查用户是否能使用指定客户端登录
// @Tags         client-devices
// @Accept       json
// @Produce      json
// @Param        request  body      models.CheckClientAccessRequest  true  "检查请求"
// @Success      200  {object}  object{success=bool,data=models.CheckClientAccessResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /client-devices/check-access [post]
func (h *ClientDeviceHandler) CheckClientAccess(c *gin.Context) {
	var req models.CheckClientAccessRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Code == "" {
		middleware.ThrowError(c, middleware.BadRequestError("客户端code不能为空", nil))
		return
	}

	if req.UserId == "" {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	result, err := funcs.ClientDeviceFuncs{}.CheckClientAccess(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("检查客户端访问权限失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result,
	})
}

// ExportClientDevicesToExcel 导出客户端设备为Excel
// @Summary      导出客户端设备为Excel
// @Description  将客户端设备导出为Excel文件
// @Tags         client-devices
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10000)
// @Param        order     query     string  false  "排序方式"  default(desc)
// @Param        order_by  query     string  false  "排序字段"  default(create_time)
// @Param        name      query     string  false  "按名称模糊搜索"
// @Param        code      query     string  false  "按code精确搜索"
// @Param        enabled   query     bool    false  "按启用状态过滤"
// @Param        anonymous query     bool    false  "按匿名登录过滤"
// @Success      200  {file}   file    "Excel文件"
// @Failure      400  {object} object{success=bool,message=string}
// @Failure      500  {object} object{success=bool,message=string}
// @Router       /client-devices/export [get]
func (h *ClientDeviceHandler) ExportClientDevicesToExcel(c *gin.Context) {
	var req models.PageClientDevicesRequest

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
	result, err := funcs.ClientDeviceFuncs{}.GetClientDevicesWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取客户端设备失败", err.Error()))
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
			Header:    "设备名称",
			Width:     25,
			FieldName: "Name",
		},
		{
			Header:    "设备Code",
			Width:     70,
			FieldName: "Code",
		},
		{
			Header:    "是否启用",
			Width:     15,
			FieldName: "Enabled",
			Formatter: excel.BoolFormatter("启用", "禁用"),
		},
		{
			Header:    "AccessToken超时(ms)",
			Width:     25,
			FieldName: "AccessTokenExpiry",
		},
		{
			Header:    "RefreshToken超时(ms)",
			Width:     25,
			FieldName: "RefreshTokenExpiry",
		},
		{
			Header:    "匿名登录",
			Width:     15,
			FieldName: "Anonymous",
			Formatter: excel.BoolFormatter("允许", "不允许"),
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
	processor := excel.NewExcelProcessor("客户端设备", columns)

	// 生成Excel文件
	file, err := processor.GenerateExcelStream(result.Data)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("生成Excel文件失败", err.Error()))
		return
	}

	// 生成文件名
	filename := excel.GenerateFilename("客户端设备")

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
