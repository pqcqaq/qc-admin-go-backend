package handlers

import (
	"net/http"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// AreaHandler 地区处理器
type AreaHandler struct{}

// NewAreaHandler 创建新的地区处理器
func NewAreaHandler() *AreaHandler {
	return &AreaHandler{}
}

// GetAreas 获取所有地区
// @Summary      获取所有地区
// @Description  获取系统中所有地区的列表
// @Tags         areas
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas [get]
func (h *AreaHandler) GetAreas(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	areas, err := funcs.AreaFuncs{}.GetAllAreas(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    areas,
		"count":   len(areas),
	})
}

// GetAreasWithPagination 分页获取地区列表
// @Summary      分页获取地区列表
// @Description  根据分页参数获取地区列表
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"         default(1)
// @Param        page_size query     int     false  "每页数量"      default(10)
// @Param        order     query     string  false  "排序方式"      default(asc)
// @Param        order_by  query     string  false  "排序字段"      default(depth)
// @Param        name      query     string  false  "地区名称"
// @Param        level     query     string  false  "层级类型"
// @Param        depth     query     int     false  "深度"
// @Param        code      query     string  false  "地区编码"
// @Param        parentId  query     string  false  "父级ID"
// @Success      200  {object}  object{success=bool,data=[]object,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/page [get]
func (h *AreaHandler) GetAreasWithPagination(c *gin.Context) {
	var req models.GetAreasRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "asc"
	req.OrderBy = "depth"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 调用服务层方法
	ctx := middleware.GetRequestContext(c)
	result, err := funcs.AreaFuncs{}.GetAreasWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetArea 根据ID获取地区
// @Summary      根据ID获取地区
// @Description  根据地区ID获取地区详细信息
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "地区ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/{id} [get]
func (h *AreaHandler) GetArea(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("地区ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("地区ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	area, err := funcs.AreaFuncs{}.GetAreaByID(ctx, id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "area not found" {
			middleware.ThrowError(c, middleware.NotFoundError("地区未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询地区失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    area,
	})
}

// CreateArea 创建地区
// @Summary      创建地区
// @Description  创建新的地区
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        area  body      models.CreateAreaRequest  true  "地区信息"
// @Success      201   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /areas [post]
func (h *AreaHandler) CreateArea(c *gin.Context) {
	var req models.CreateAreaRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("地区名称不能为空", nil))
		return
	}

	if req.Code == "" {
		middleware.ThrowError(c, middleware.BadRequestError("地区编码不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	area, err := funcs.AreaFuncs{}.CreateArea(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建地区失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    area,
		"message": "地区创建成功",
	})
}

// UpdateArea 更新地区
// @Summary      更新地区
// @Description  根据ID更新地区信息
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "地区ID"
// @Param        area  body      models.UpdateAreaRequest  true  "地区信息"
// @Success      200   {object}  object{success=bool,data=object}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /areas/{id} [put]
func (h *AreaHandler) UpdateArea(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("地区ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	area, err := funcs.AreaFuncs{}.UpdateArea(ctx, id, &req)
	if err != nil {
		if err.Error() == "area not found" {
			middleware.ThrowError(c, middleware.NotFoundError("地区未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新地区失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    area,
		"message": "地区更新成功",
	})
}

// DeleteArea 删除地区
// @Summary      删除地区
// @Description  根据ID删除地区(会级联删除子地区)
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "地区ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/{id} [delete]
func (h *AreaHandler) DeleteArea(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("地区ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.AreaFuncs{}.DeleteArea(ctx, id)
	if err != nil {
		if err.Error() == "area not found" {
			middleware.ThrowError(c, middleware.NotFoundError("地区未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除地区失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "地区删除成功",
	})
}

// GetAreasByParentID 根据父级ID获取下一级地区
// @Summary      根据父级ID获取下一级地区
// @Description  获取指定父级ID的直接子地区列表,parentId为0时获取顶级地区
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        parentId  query     string  true  "父级ID(0表示获取顶级地区)"
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/children [get]
func (h *AreaHandler) GetAreasByParentID(c *gin.Context) {
	parentIdStr := c.Query("parentId")

	var err error

	ctx := middleware.GetRequestContext(c)
	areas, err := funcs.AreaFuncs{}.GetAreasByParentID(ctx, parentIdStr)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取子地区列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    areas,
		"count":   len(areas),
	})
}

// GetAreasByLevel 根据级别获取地区
// @Summary      根据级别获取地区
// @Description  获取指定层级类型的所有地区
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        level  query     string  true  "层级类型(country/province/city/district/street)"
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/level [get]
func (h *AreaHandler) GetAreasByLevel(c *gin.Context) {
	level := c.Query("level")

	if level == "" {
		middleware.ThrowError(c, middleware.BadRequestError("层级类型不能为空", nil))
		return
	}

	// 验证level值
	validLevels := map[string]bool{
		"country":  true,
		"province": true,
		"city":     true,
		"district": true,
		"street":   true,
	}

	if !validLevels[level] {
		middleware.ThrowError(c, middleware.BadRequestError("无效的层级类型", map[string]any{
			"provided_level": level,
			"valid_levels":   []string{"country", "province", "city", "district", "street"},
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	areas, err := funcs.AreaFuncs{}.GetAreasByLevel(ctx, level)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    areas,
		"count":   len(areas),
	})
}

// GetAreasByDepth 根据深度获取地区
// @Summary      根据深度获取地区
// @Description  获取指定深度的所有地区(0=国家、1=省、2=市、3=区、4=街道)
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        depth  query     int  true  "深度(0-4)"
// @Success      200  {object}  object{success=bool,data=[]object,count=int}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/depth [get]
func (h *AreaHandler) GetAreasByDepth(c *gin.Context) {
	depthStr := c.Query("depth")

	if depthStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("深度不能为空", nil))
		return
	}

	depth, err := strconv.Atoi(depthStr)
	if err != nil || depth < 0 || depth > 4 {
		middleware.ThrowError(c, middleware.BadRequestError("深度值无效,必须在0-4之间", map[string]any{
			"provided_depth": depthStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	areas, err := funcs.AreaFuncs{}.GetAreasByDepth(ctx, depth)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    areas,
		"count":   len(areas),
	})
}

// GetAreaTree 获取地区树形结构
// @Summary      获取地区树形结构
// @Description  获取完整的地区树形结构
// @Tags         areas
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]object}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/tree [get]
func (h *AreaHandler) GetAreaTree(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	result, err := funcs.AreaFuncs{}.GetAreaTree(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取地区树形结构失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Data,
	})
}
