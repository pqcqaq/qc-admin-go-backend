package handlers

import (
	"context"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色处理器
type RoleHandler struct {
}

// NewRoleHandler 创建新的角色处理器
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{}
}

// PermissionHandler 权限处理器
type PermissionHandler struct {
}

// NewPermissionHandler 创建新的权限处理器
func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{}
}

// === 角色相关方法 ===

// GetAllRoles 获取所有角色（不分页）
func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	roles, err := funcs.GetAllRoles(context.Background())
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取角色列表失败", err.Error()))
		return
	}

	// 转换为响应格式
	roleResponses := make([]*models.RoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, funcs.ConvertRoleToResponse(role))
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    roleResponses,
		"count":   len(roleResponses),
	})
}

// GetRoles 分页获取角色列表
func (h *RoleHandler) GetRoles(c *gin.Context) {
	var req models.GetRolesRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "asc"
	req.OrderBy = "id"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	result, err := funcs.GetRolesWithPagination(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取角色列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetRole 根据ID获取角色
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	role, err := funcs.GetRoleByID(context.Background(), id)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询角色失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    funcs.ConvertRoleToResponse(role),
	})
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("角色名称不能为空", nil))
		return
	}

	role, err := funcs.CreateRole(context.Background(), &req)
	if err != nil {
		if err.Error() == "role already exists" {
			middleware.ThrowError(c, middleware.BadRequestError("角色已存在", map[string]any{
				"name": req.Name,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("创建角色失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    funcs.ConvertRoleToResponse(role),
		"message": "角色创建成功",
	})
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	role, err := funcs.UpdateRole(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新角色失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    funcs.ConvertRoleToResponse(role),
		"message": "角色更新成功",
	})
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteRole(context.Background(), id)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除角色失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "角色删除成功",
	})
}

// AssignRolePermissions 分配角色权限
func (h *RoleHandler) AssignRolePermissions(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	err = funcs.AssignRolePermissions(context.Background(), id, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("分配角色权限失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "角色权限分配成功",
	})
}

// RevokeRolePermission 撤销角色权限
func (h *RoleHandler) RevokeRolePermission(c *gin.Context) {
	roleIDStr := c.Param("id")
	permissionIDStr := c.Param("permissionId")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_role_id": roleIDStr,
		}))
		return
	}

	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限ID格式无效", map[string]any{
			"provided_permission_id": permissionIDStr,
		}))
		return
	}

	err = funcs.RevokeRolePermission(context.Background(), roleID, permissionID)
	if err != nil {
		if err.Error() == "role permission not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色权限关联不存在", map[string]any{
				"role_id":       roleID,
				"permission_id": permissionID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("撤销角色权限失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "角色权限撤销成功",
	})
}

// === 权限相关方法 ===

// GetAllPermissions 获取所有权限（不分页）
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := funcs.GetAllPermissions(context.Background())
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限列表失败", err.Error()))
		return
	}

	// 转换为响应格式
	permissionResponses := make([]*models.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		permissionResponses = append(permissionResponses, funcs.ConvertPermissionToResponse(permission))
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    permissionResponses,
		"count":   len(permissionResponses),
	})
}

// GetPermissions 分页获取权限列表
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	var req models.GetPermissionsRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "asc"
	req.OrderBy = "id"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	result, err := funcs.GetPermissionsWithPagination(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取权限列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetPermission 根据ID获取权限
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("权限ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	permission, err := funcs.GetPermissionByID(context.Background(), id)
	if err != nil {
		if err.Error() == "permission not found" {
			middleware.ThrowError(c, middleware.NotFoundError("权限不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询权限失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    funcs.ConvertPermissionToResponse(permission),
	})
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req models.CreatePermissionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("权限名称不能为空", nil))
		return
	}

	if req.Action == "" {
		middleware.ThrowError(c, middleware.BadRequestError("权限操作不能为空", nil))
		return
	}

	permission, err := funcs.CreatePermission(context.Background(), &req)
	if err != nil {
		if err.Error() == "permission already exists" {
			middleware.ThrowError(c, middleware.BadRequestError("权限已存在", map[string]any{
				"name":   req.Name,
				"action": req.Action,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("创建权限失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    funcs.ConvertPermissionToResponse(permission),
		"message": "权限创建成功",
	})
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	permission, err := funcs.UpdatePermission(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "permission not found" {
			middleware.ThrowError(c, middleware.NotFoundError("权限不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新权限失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    funcs.ConvertPermissionToResponse(permission),
		"message": "权限更新成功",
	})
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeletePermission(context.Background(), id)
	if err != nil {
		if err.Error() == "permission not found" {
			middleware.ThrowError(c, middleware.NotFoundError("权限不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除权限失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "权限删除成功",
	})
}
