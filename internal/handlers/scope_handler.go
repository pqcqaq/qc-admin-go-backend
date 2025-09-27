package handlers

import (
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// ScopeHandler 权限域处理器
type ScopeHandler struct {
}

// NewScopeHandler 创建新的权限域处理器
func NewScopeHandler() *ScopeHandler {
	return &ScopeHandler{}
}

// GetScopes 获取所有权限域
// @Summary      获取所有权限域
// @Description  获取系统中所有权限域的列表（不分页）
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]models.ScopeResponse,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/scopes/all [get]
func (h *ScopeHandler) GetScopes(c *gin.Context) {
	scopes, err := funcs.ScopeFuncs{}.GetAllScopes(middleware.GetRequestContext(c))
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限域列表失败", err.Error()))
		return
	}

	// 转换为响应格式
	scopeResponses := make([]*models.ScopeResponse, 0, len(scopes))
	for _, scope := range scopes {
		scopeResponses = append(scopeResponses, funcs.ScopeFuncs{}.ConvertScopeToResponse(scope))
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    scopeResponses,
		"count":   len(scopeResponses),
	})
}

// GetScopesWithPagination 分页获取权限域列表
// @Summary      分页获取权限域列表
// @Description  根据分页参数获取权限域列表
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        order     query     string  false  "排序方式"  default(asc)
// @Param        order_by  query     string  false  "排序字段"  default(order)
// @Success      200  {object}  object{success=bool,data=[]models.ScopeResponse,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/scopes [get]
func (h *ScopeHandler) GetScopesWithPagination(c *gin.Context) {
	var req models.GetScopesRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "asc"
	req.OrderBy = "order"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	result, err := funcs.ScopeFuncs{}.GetScopesWithPagination(middleware.GetRequestContext(c), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限域列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetScope 根据ID获取权限域
// @Summary      根据ID获取权限域
// @Description  根据权限域ID获取权限域详细信息
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "权限域ID"
// @Success      200  {object}  object{success=bool,data=models.ScopeResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/scopes/{id} [get]
func (h *ScopeHandler) GetScope(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("权限域ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限域ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	scope, err := funcs.ScopeFuncs{}.GetScopeByID(middleware.GetRequestContext(c), id)
	if err != nil {
		if err.Error() == "scope not found" {
			middleware.ThrowError(c, middleware.NotFoundError("权限域不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询权限域失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    funcs.ScopeFuncs{}.ConvertScopeToResponse(scope),
	})
}

// CreateScope 创建权限域
// @Summary      创建权限域
// @Description  创建新的权限域
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Param        scope  body      models.CreateScopeRequest  true  "权限域信息"
// @Success      201    {object}  object{success=bool,data=models.ScopeResponse,message=string}
// @Failure      400    {object}  object{success=bool,message=string}
// @Failure      500    {object}  object{success=bool,message=string}
// @Router       /rbac/scopes [post]
func (h *ScopeHandler) CreateScope(c *gin.Context) {
	var req models.CreateScopeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("权限域名称不能为空", nil))
		return
	}

	if req.Type == "" {
		middleware.ThrowError(c, middleware.BadRequestError("权限域类型不能为空", nil))
		return
	}

	scope, err := funcs.ScopeFuncs{}.CreateScope(middleware.GetRequestContext(c), &req)
	if err != nil {
		if err.Error() == "scope already exists" {
			middleware.ThrowError(c, middleware.BadRequestError("权限域已存在", map[string]any{
				"name": req.Name,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("创建权限域失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    funcs.ScopeFuncs{}.ConvertScopeToResponse(scope),
		"message": "权限域创建成功",
	})
}

// UpdateScope 更新权限域
// @Summary      更新权限域
// @Description  根据ID更新权限域信息
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Param        id     path      int                        true  "权限域ID"
// @Param        scope  body      models.UpdateScopeRequest  true  "权限域信息"
// @Success      200    {object}  object{success=bool,data=models.ScopeResponse,message=string}
// @Failure      400    {object}  object{success=bool,message=string}
// @Failure      404    {object}  object{success=bool,message=string}
// @Failure      500    {object}  object{success=bool,message=string}
// @Router       /rbac/scopes/{id} [put]
func (h *ScopeHandler) UpdateScope(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限域ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	scope, err := funcs.ScopeFuncs{}.UpdateScope(middleware.GetRequestContext(c), id, &req)
	if err != nil {
		if err.Error() == "scope not found" {
			middleware.ThrowError(c, middleware.NotFoundError("权限域不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新权限域失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    funcs.ScopeFuncs{}.ConvertScopeToResponse(scope),
		"message": "权限域更新成功",
	})
}

// DeleteScope 删除权限域
// @Summary      删除权限域
// @Description  根据ID删除权限域
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "权限域ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/scopes/{id} [delete]
func (h *ScopeHandler) DeleteScope(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限域ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.ScopeFuncs{}.DeleteScope(middleware.GetRequestContext(c), id)
	if err != nil {
		if err.Error() == "scope not found" {
			middleware.ThrowError(c, middleware.NotFoundError("权限域不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除权限域失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "权限域删除成功",
	})
}

// GetScopeTree 获取权限域树形结构
// @Summary      获取权限域树形结构
// @Description  获取权限域的树形结构数据
// @Tags         rbac-scopes
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/scopes/tree [get]
func (h *ScopeHandler) GetScopeTree(c *gin.Context) {
	result, err := funcs.ScopeFuncs{}.GetScopeTree(middleware.GetRequestContext(c))
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限域树失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result.Data,
	})
}
