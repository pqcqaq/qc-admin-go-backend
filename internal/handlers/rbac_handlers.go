package handlers

import (
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
// @Summary      获取所有角色
// @Description  获取系统中所有角色的列表（不分页）
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]models.RoleResponse,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles/all [get]
func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	roles, err := funcs.GetAllRoles(middleware.GetRequestContext(c))
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
// @Summary      分页获取角色列表
// @Description  根据分页参数获取角色列表
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        order     query     string  false  "排序方式"  default(asc)
// @Param        order_by  query     string  false  "排序字段"  default(id)
// @Success      200  {object}  object{success=bool,data=[]models.RoleResponse,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles [get]
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

	result, err := funcs.GetRolesWithPagination(middleware.GetRequestContext(c), &req)
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
// @Summary      根据ID获取角色
// @Description  根据角色ID获取角色详细信息
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  object{success=bool,data=models.RoleResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id} [get]
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

	role, err := funcs.GetRoleByID(middleware.GetRequestContext(c), id)
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
// @Summary      创建角色
// @Description  创建新的角色
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        role  body      models.CreateRoleRequest  true  "角色信息"
// @Success      201   {object}  object{success=bool,data=models.RoleResponse,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /rbac/roles [post]
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

	role, err := funcs.CreateRole(middleware.GetRequestContext(c), &req)
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
// @Summary      更新角色
// @Description  根据ID更新角色信息
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "角色ID"
// @Param        role  body      models.UpdateRoleRequest  true  "角色信息"
// @Success      200   {object}  object{success=bool,data=models.RoleResponse,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id} [put]
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

	role, err := funcs.UpdateRole(middleware.GetRequestContext(c), id, &req)
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
// @Summary      删除角色
// @Description  根据ID删除角色
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteRole(middleware.GetRequestContext(c), id)
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
// @Summary      分配角色权限
// @Description  为角色分配权限
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id          path      int                                  true  "角色ID"
// @Param        permissions body      models.AssignRolePermissionsRequest  true  "权限ID列表"
// @Success      200         {object}  object{success=bool,message=string}
// @Failure      400         {object}  object{success=bool,message=string}
// @Failure      500         {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/permissions [post]
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

	err = funcs.AssignRolePermissions(middleware.GetRequestContext(c), id, &req)
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
// @Summary      撤销角色权限
// @Description  撤销角色的指定权限
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id           path      int  true  "角色ID"
// @Param        permissionId path      int  true  "权限ID"
// @Success      200          {object}  object{success=bool,message=string}
// @Failure      400          {object}  object{success=bool,message=string}
// @Failure      404          {object}  object{success=bool,message=string}
// @Failure      500          {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/permissions/{permissionId} [delete]
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

	err = funcs.RevokeRolePermission(middleware.GetRequestContext(c), roleID, permissionID)
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
// @Summary      获取所有权限
// @Description  获取系统中所有权限的列表（不分页）
// @Tags         rbac-permissions
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]models.PermissionResponse,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/permissions/all [get]
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := funcs.GetAllPermissions(middleware.GetRequestContext(c))
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
// @Summary      分页获取权限列表
// @Description  根据分页参数获取权限列表
// @Tags         rbac-permissions
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        order     query     string  false  "排序方式"  default(asc)
// @Param        order_by  query     string  false  "排序字段"  default(id)
// @Success      200  {object}  object{success=bool,data=[]models.PermissionResponse,pagination=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/permissions [get]
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

	result, err := funcs.GetPermissionsWithPagination(middleware.GetRequestContext(c), &req)
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
// @Summary      根据ID获取权限
// @Description  根据权限ID获取权限详细信息
// @Tags         rbac-permissions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "权限ID"
// @Success      200  {object}  object{success=bool,data=models.PermissionResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/permissions/{id} [get]
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

	permission, err := funcs.GetPermissionByID(middleware.GetRequestContext(c), id)
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
// @Summary      创建权限
// @Description  创建新的权限
// @Tags         rbac-permissions
// @Accept       json
// @Produce      json
// @Param        permission  body      models.CreatePermissionRequest  true  "权限信息"
// @Success      201         {object}  object{success=bool,data=models.PermissionResponse,message=string}
// @Failure      400         {object}  object{success=bool,message=string}
// @Failure      500         {object}  object{success=bool,message=string}
// @Router       /rbac/permissions [post]
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

	permission, err := funcs.CreatePermission(middleware.GetRequestContext(c), &req)
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
// @Summary      更新权限
// @Description  根据ID更新权限信息
// @Tags         rbac-permissions
// @Accept       json
// @Produce      json
// @Param        id          path      int                             true  "权限ID"
// @Param        permission  body      models.UpdatePermissionRequest  true  "权限信息"
// @Success      200         {object}  object{success=bool,data=models.PermissionResponse,message=string}
// @Failure      400         {object}  object{success=bool,message=string}
// @Failure      404         {object}  object{success=bool,message=string}
// @Failure      500         {object}  object{success=bool,message=string}
// @Router       /rbac/permissions/{id} [put]
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

	permission, err := funcs.UpdatePermission(middleware.GetRequestContext(c), id, &req)
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
// @Summary      删除权限
// @Description  根据ID删除权限
// @Tags         rbac-permissions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "权限ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("权限ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeletePermission(middleware.GetRequestContext(c), id)
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

// === 新增的RBAC管理页面接口 ===

// GetRoleTree 获取角色树形结构
// @Summary      获取角色树形结构
// @Description  获取角色的层级树形结构，包含用户统计信息
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]models.RoleTreeResponse}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles/tree [get]
func (h *RoleHandler) GetRoleTree(c *gin.Context) {
	roleTree, err := funcs.GetRoleTree(middleware.GetRequestContext(c))
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取角色树失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    roleTree,
	})
}

// GetRoleWithPermissions 获取角色详细权限信息
// @Summary      获取角色详细权限信息
// @Description  获取角色的详细权限信息，区分直接分配和继承的权限
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  object{success=bool,data=models.RoleDetailedPermissionsResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/permissions/detailed [get]
func (h *RoleHandler) GetRoleWithPermissions(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	detailedPermissions, err := funcs.GetRoleWithPermissions(middleware.GetRequestContext(c), id)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("获取角色权限详情失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    detailedPermissions,
	})
}

// CreateChildRole 创建子角色
// @Summary      创建子角色
// @Description  在指定父角色下创建子角色，自动建立继承关系
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        parentId  path      int                           true  "父角色ID"
// @Param        role      body      models.CreateChildRoleRequest true  "子角色信息"
// @Success      201       {object}  object{success=bool,data=models.RoleResponse,message=string}
// @Failure      400       {object}  object{success=bool,message=string}
// @Failure      404       {object}  object{success=bool,message=string}
// @Failure      500       {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{parentId}/children [post]
func (h *RoleHandler) CreateChildRole(c *gin.Context) {
	parentIDStr := c.Param("id")

	parentID, err := strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("父角色ID格式无效", map[string]any{
			"provided_parent_id": parentIDStr,
		}))
		return
	}

	var req models.CreateChildRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("角色名称不能为空", nil))
		return
	}

	role, err := funcs.CreateChildRole(middleware.GetRequestContext(c), parentID, &req)
	if err != nil {
		if err.Error() == "parent role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("父角色不存在", map[string]any{
				"parent_id": parentID,
			}))
		} else if err.Error() == "role already exists" {
			middleware.ThrowError(c, middleware.BadRequestError("角色已存在", map[string]any{
				"name": req.Name,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("创建子角色失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    funcs.ConvertRoleToResponse(role),
		"message": "子角色创建成功",
	})
}

// RemoveParentRole 解除父角色依赖
// @Summary      解除父角色依赖
// @Description  解除指定角色对某个父角色的继承关系
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "角色ID"
// @Param        parentId path      int  true  "父角色ID"
// @Success      200      {object}  object{success=bool,message=string}
// @Failure      400      {object}  object{success=bool,message=string}
// @Failure      404      {object}  object{success=bool,message=string}
// @Failure      500      {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/parents/{parentId} [delete]
func (h *RoleHandler) RemoveParentRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	parentIDStr := c.Param("parentId")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_role_id": roleIDStr,
		}))
		return
	}

	parentID, err := strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("父角色ID格式无效", map[string]any{
			"provided_parent_id": parentIDStr,
		}))
		return
	}

	err = funcs.RemoveParentRole(middleware.GetRequestContext(c), roleID, parentID)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"role_id": roleID,
			}))
		} else if err.Error() == "parent role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("父角色不存在", map[string]any{
				"parent_id": parentID,
			}))
		} else if err.Error() == "inheritance relationship not found" {
			middleware.ThrowError(c, middleware.NotFoundError("继承关系不存在", map[string]any{
				"role_id":   roleID,
				"parent_id": parentID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("解除继承关系失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "解除继承关系成功",
	})
}

// AddParentRole 添加父角色依赖
// @Summary      添加父角色依赖
// @Description  为指定角色添加父角色继承关系
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "角色ID"
// @Param        parentId path      int  true  "父角色ID"
// @Success      200      {object}  object{success=bool,message=string}
// @Failure      400      {object}  object{success=bool,message=string}
// @Failure      404      {object}  object{success=bool,message=string}
// @Failure      500      {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/parents/{parentId} [post]
func (h *RoleHandler) AddParentRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	parentIDStr := c.Param("parentId")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_role_id": roleIDStr,
		}))
		return
	}

	parentID, err := strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("父角色ID格式无效", map[string]any{
			"provided_parent_id": parentIDStr,
		}))
		return
	}

	err = funcs.AddParentRole(middleware.GetRequestContext(c), roleID, parentID)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"role_id": roleID,
			}))
		} else if err.Error() == "parent role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("父角色不存在", map[string]any{
				"parent_id": parentID,
			}))
		} else if err.Error() == "circular inheritance detected" {
			middleware.ThrowError(c, middleware.BadRequestError("检测到循环继承", map[string]any{
				"role_id":   roleID,
				"parent_id": parentID,
			}))
		} else if err.Error() == "inheritance relationship already exists" {
			middleware.ThrowError(c, middleware.BadRequestError("继承关系已存在", map[string]any{
				"role_id":   roleID,
				"parent_id": parentID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("添加继承关系失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "添加继承关系成功",
	})
}

// GetAssignablePermissions 获取可分配的权限
// @Summary      获取可分配的权限
// @Description  获取角色可以分配的权限（排除已有的直接权限和继承权限）
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  object{success=bool,data=[]models.PermissionResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/permissions/assignable [get]
func (h *RoleHandler) GetAssignablePermissions(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	permissions, err := funcs.GetAssignablePermissions(middleware.GetRequestContext(c), id)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("获取可分配权限失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    permissions,
	})
}

// GetRoleUsers 获取拥有指定角色的用户（支持分页）
// @Summary      获取拥有指定角色的用户
// @Description  获取拥有指定角色的用户列表，支持分页和搜索
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id        path      int     true   "角色ID"
// @Param        page      query     int     false  "页码"     default(1)
// @Param        page_size query     int     false  "每页数量"  default(10)
// @Param        keyword   query     string  false  "搜索关键字"
// @Success      200       {object}  object{success=bool,data=[]models.RoleUserResponse,pagination=object}
// @Failure      400       {object}  object{success=bool,message=string}
// @Failure      404       {object}  object{success=bool,message=string}
// @Failure      500       {object}  object{success=bool,message=string}
// @Router       /rbac/user-roles/roles/{id}/users [get]
func (h *RoleHandler) GetRoleUsersWithPagination(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.GetRoleUsersRequest

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

	result, err := funcs.GetRoleUsersWithPagination(middleware.GetRequestContext(c), id, &req)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("获取角色用户列表失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Users,
		"pagination": result.Pagination,
	})
}

// BatchAssignUsersToRole 批量分配用户到角色
// @Summary      批量分配用户到角色
// @Description  批量将用户分配到指定角色
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id    path      int                                 true  "角色ID"
// @Param        users body      models.BatchAssignUsersToRoleRequest true  "用户ID列表"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/users/batch [post]
func (h *RoleHandler) BatchAssignUsersToRole(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.BatchAssignUsersToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if len(req.UserIds) == 0 {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID列表不能为空", nil))
		return
	}

	err = funcs.BatchAssignUsersToRole(middleware.GetRequestContext(c), id, &req)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("批量分配用户失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "批量分配用户成功",
	})
}

// BatchRemoveUsersFromRole 批量从角色移除用户
// @Summary      批量从角色移除用户
// @Description  批量从指定角色移除用户
// @Tags         rbac-roles
// @Accept       json
// @Produce      json
// @Param        id    path      int                                    true  "角色ID"
// @Param        users body      models.BatchRemoveUsersFromRoleRequest true  "用户ID列表"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /rbac/roles/{id}/users/batch [delete]
func (h *RoleHandler) BatchRemoveUsersFromRole(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("角色ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.BatchRemoveUsersFromRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if len(req.UserIds) == 0 {
		middleware.ThrowError(c, middleware.BadRequestError("用户ID列表不能为空", nil))
		return
	}

	err = funcs.BatchRemoveUsersFromRole(middleware.GetRequestContext(c), id, &req)
	if err != nil {
		if err.Error() == "role not found" {
			middleware.ThrowError(c, middleware.NotFoundError("角色不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("批量移除用户失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "批量移除用户成功",
	})
}
