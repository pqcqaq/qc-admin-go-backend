package handlers

import (
	"context"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// UserHandler 使用新错误处理中间件的用户处理器示例
type UserHandler struct {
}

// NewUserHandler 创建新的用户处理器V2
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetUsers 获取所有用户
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := funcs.GetAllUsers(context.Background())
	if err != nil {
		// 使用自定义错误处理
		middleware.ThrowError(c, middleware.DatabaseError("获取用户列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    users,
		"count":   len(users),
	})
}

// GetUsersWithPagination 分页获取用户列表
func (h *UserHandler) GetUsersWithPagination(c *gin.Context) {
	var req models.GetUsersRequest

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
	result, err := funcs.GetUsersWithPagination(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取用户列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetUser 根据ID获取用户
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	user, err := funcs.GetUserByID(context.Background(), id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "user not found" {
			middleware.ThrowError(c, middleware.UserNotFoundError(map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询用户失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    user,
	})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("用户名不能为空", nil))
		return
	}

	if req.Email == "" {
		middleware.ThrowError(c, middleware.BadRequestError("邮箱不能为空", nil))
		return
	}

	user, err := funcs.CreateUser(context.Background(), &req)
	if err != nil {
		// 根据错误内容判断错误类型
		if err.Error() == "user already exists" {
			middleware.ThrowError(c, middleware.UserExistsError(map[string]interface{}{
				"email": req.Email,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("创建用户失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    user,
		"message": "用户创建成功",
	})
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	user, err := funcs.UpdateUser(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			middleware.ThrowError(c, middleware.UserNotFoundError(map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新用户失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    user,
		"message": "用户更新成功",
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteUser(context.Background(), id)
	if err != nil {
		if err.Error() == "user not found" {
			middleware.ThrowError(c, middleware.UserNotFoundError(map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除用户失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "用户删除成功",
	})
}
