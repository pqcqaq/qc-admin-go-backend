package handlers

import (
	"context"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// UserRoleHandler 用户角色关联处理器
type UserRoleHandler struct {
}

// NewUserRoleHandler 创建新的用户角色关联处理器
func NewUserRoleHandler() *UserRoleHandler {
	return &UserRoleHandler{}
}

// AssignRole 为用户分配角色
func (h *UserRoleHandler) AssignRole(c *gin.Context) {
	var req models.AssignUserRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.UserID == 0 {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID不能为空", nil))
		return
	}

	if req.RoleID == 0 {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID不能为空", nil))
		return
	}

	userRole, err := funcs.AssignUserRole(context.Background(), &req)
	if err != nil {
		if err.Error() == "user role already exists" {
			middleware.ThrowError(c, middleware.BadRequestError("用户已拥有此角色", map[string]any{
				"user_id": req.UserID,
				"role_id": req.RoleID,
			}))
		} else if err.Error() == "user not found" {
			middleware.ThrowError(c, middleware.NotFoundError("用户不存在", map[string]any{
				"user_id": req.UserID,
			}))
		} else if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"role_id": req.RoleID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("分配角色失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    funcs.ConvertUserRoleToResponse(userRole),
		"message": "角色分配成功",
	})
}

// RevokeRole 撤销用户角色
func (h *UserRoleHandler) RevokeRole(c *gin.Context) {
	userIDStr := c.Param("userID")
	roleIDStr := c.Param("roleID")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]any{
			"provided_user_id": userIDStr,
		}))
		return
	}

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_role_id": roleIDStr,
		}))
		return
	}

	err = funcs.RevokeUserRole(context.Background(), userID, roleID)
	if err != nil {
		if err.Error() == "user role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("用户角色关联不存在", map[string]any{
				"user_id": userID,
				"role_id": roleID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("撤销角色失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "角色撤销成功",
	})
}

// GetUserRoles 获取用户的所有角色
func (h *UserRoleHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("userID")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]any{
			"provided_user_id": userIDStr,
		}))
		return
	}

	roles, err := funcs.GetUserRoles(context.Background(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			middleware.ThrowError(c, middleware.NotFoundError("用户不存在", map[string]any{
				"user_id": userID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询用户角色失败", err.Error()))
		}
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

// GetRoleUsers 获取拥有指定角色的所有用户
func (h *UserRoleHandler) GetRoleUsers(c *gin.Context) {
	roleIDStr := c.Param("roleID")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_role_id": roleIDStr,
		}))
		return
	}

	users, err := funcs.GetRoleUsers(context.Background(), roleID)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"role_id": roleID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询角色用户失败", err.Error()))
		}
		return
	}

	// 转换为响应格式
	userResponses := make([]*models.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, funcs.ConvertUserToResponse(user))
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    userResponses,
		"count":   len(userResponses),
	})
}

// GetUserPermissions 获取用户的所有权限（通过角色继承）
func (h *UserRoleHandler) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("userID")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]any{
			"provided_user_id": userIDStr,
		}))
		return
	}

	permissions, err := funcs.GetUserPermissions(context.Background(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			middleware.ThrowError(c, middleware.NotFoundError("用户不存在", map[string]any{
				"user_id": userID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询用户权限失败", err.Error()))
		}
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

// CheckUserPermission 检查用户是否拥有指定权限
func (h *UserRoleHandler) CheckUserPermission(c *gin.Context) {
	userIDStr := c.Param("userID")
	permissionIDStr := c.Param("permissionID")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]any{
			"provided_user_id": userIDStr,
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

	hasPermission, err := funcs.CheckUserPermission(context.Background(), userID, permissionID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("检查用户权限失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":        true,
		"has_permission": hasPermission,
		"data": map[string]any{
			"user_id":        userID,
			"permission_id":  permissionID,
			"has_permission": hasPermission,
		},
	})
}
