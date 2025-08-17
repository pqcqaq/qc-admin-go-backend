package handlers

import (
	"context"
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
func (h *ScopeHandler) GetScopes(c *gin.Context) {
	scopes, err := funcs.GetAllScopes(context.Background())
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限域列表失败", err.Error()))
		return
	}

	// 转换为响应格式
	scopeResponses := make([]*models.ScopeResponse, 0, len(scopes))
	for _, scope := range scopes {
		scopeResponses = append(scopeResponses, funcs.ConvertScopeToResponse(scope))
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    scopeResponses,
		"count":   len(scopeResponses),
	})
}

// GetScopesWithPagination 分页获取权限域列表
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

	result, err := funcs.GetScopesWithPagination(context.Background(), &req)
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

	scope, err := funcs.GetScopeByID(context.Background(), id)
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
		"data":    funcs.ConvertScopeToResponse(scope),
	})
}

// CreateScope 创建权限域
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

	scope, err := funcs.CreateScope(context.Background(), &req)
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
		"data":    funcs.ConvertScopeToResponse(scope),
		"message": "权限域创建成功",
	})
}

// UpdateScope 更新权限域
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

	scope, err := funcs.UpdateScope(context.Background(), id, &req)
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
		"data":    funcs.ConvertScopeToResponse(scope),
		"message": "权限域更新成功",
	})
}

// DeleteScope 删除权限域
func (h *ScopeHandler) DeleteScope(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限域ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteScope(context.Background(), id)
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
func (h *ScopeHandler) GetScopeTree(c *gin.Context) {
	result, err := funcs.GetScopeTree(context.Background())
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限域树失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result.Data,
	})
}
