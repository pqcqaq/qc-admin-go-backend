package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupRBACRoutes 设置RBAC相关路由
func (r *Router) setupRBACRoutes(rg *gin.RouterGroup) {
	// 创建处理器实例
	roleHandler := handlers.NewRoleHandler()
	permissionHandler := handlers.NewPermissionHandler()
	scopeHandler := handlers.NewScopeHandler()
	userRoleHandler := handlers.NewUserRoleHandler()

	// RBAC API组
	rbacGroup := rg.Group("/rbac")
	// 全部需要auth中间件保护
	// 角色路由
	roleGroup := rbacGroup.Group("/roles")
	{
		roleGroup.GET("", roleHandler.GetRoles)                                              // 获取角色列表(分页)
		roleGroup.GET("/all", roleHandler.GetAllRoles)                                       // 获取所有角色(不分页)
		roleGroup.GET("/tree", roleHandler.GetRoleTree)                                      // 获取角色树结构
		roleGroup.POST("", roleHandler.CreateRole)                                           // 创建角色
		roleGroup.GET("/:id", roleHandler.GetRole)                                           // 获取单个角色
		roleGroup.GET("/:id/permissions/detailed", roleHandler.GetRoleWithPermissions)       // 获取角色详细权限信息
		roleGroup.PUT("/:id", roleHandler.UpdateRole)                                        // 更新角色
		roleGroup.DELETE("/:id", roleHandler.DeleteRole)                                     // 删除角色
		roleGroup.POST("/:id/permissions", roleHandler.AssignRolePermissions)                // 分配角色权限
		roleGroup.DELETE("/:id/permissions/:permissionId", roleHandler.RevokeRolePermission) // 撤销角色权限
		roleGroup.GET("/:id/assignable-permissions", roleHandler.GetAssignablePermissions)   // 获取可分配的权限

		// 角色继承管理
		roleGroup.POST("/:id/children", roleHandler.CreateChildRole)             // 创建子角色
		roleGroup.DELETE("/:id/parents/:parentId", roleHandler.RemoveParentRole) // 移除父角色继承关系
		roleGroup.POST("/:id/parents/:parentId", roleHandler.AddParentRole)      // 添加父角色继承关系

		// 角色用户管理
		roleGroup.GET("/:id/users", roleHandler.GetRoleUsersWithPagination)             // 获取角色下的用户（分页）
		roleGroup.POST("/:id/users/batch-assign", roleHandler.BatchAssignUsersToRole)   // 批量分配用户到角色
		roleGroup.POST("/:id/users/batch-remove", roleHandler.BatchRemoveUsersFromRole) // 批量从角色移除用户
	}

	// 权限路由
	permissionGroup := rbacGroup.Group("/permissions")
	{
		permissionGroup.GET("", permissionHandler.GetPermissions)          // 获取权限列表(分页)
		permissionGroup.GET("/all", permissionHandler.GetAllPermissions)   // 获取所有权限(不分页)
		permissionGroup.POST("", permissionHandler.CreatePermission)       // 创建权限
		permissionGroup.GET("/:id", permissionHandler.GetPermission)       // 获取单个权限
		permissionGroup.PUT("/:id", permissionHandler.UpdatePermission)    // 更新权限
		permissionGroup.DELETE("/:id", permissionHandler.DeletePermission) // 删除权限
	}

	// 权限域路由
	scopeGroup := rbacGroup.Group("/scopes")
	{
		scopeGroup.GET("", scopeHandler.GetScopesWithPagination) // 获取权限域列表(分页)
		scopeGroup.GET("/all", scopeHandler.GetScopes)           // 获取所有权限域(不分页)
		scopeGroup.GET("/tree", scopeHandler.GetScopeTree)       // 获取权限域树形结构
		scopeGroup.POST("", scopeHandler.CreateScope)            // 创建权限域
		scopeGroup.GET("/:id", scopeHandler.GetScope)            // 获取单个权限域
		scopeGroup.PUT("/:id", scopeHandler.UpdateScope)         // 更新权限域
		scopeGroup.DELETE("/:id", scopeHandler.DeleteScope)      // 删除权限域
	}

	// 用户角色路由
	userRoleGroup := rbacGroup.Group("/user-roles")
	{
		userRoleGroup.GET("", userRoleHandler.GetUserRolesWithPagination)                                        // 分页获取用户角色关联列表
		userRoleGroup.POST("", userRoleHandler.AssignRole)                                                       // 分配用户角色
		userRoleGroup.DELETE("/users/:userID/roles/:roleID", userRoleHandler.RevokeRole)                         // 撤销用户角色
		userRoleGroup.GET("/users/:userID/roles", userRoleHandler.GetUserRoles)                                  // 获取用户的所有角色
		userRoleGroup.GET("/roles/:roleID/users", userRoleHandler.GetRoleUsers)                                  // 获取拥有指定角色的用户
		userRoleGroup.GET("/users/:userID/permissions", userRoleHandler.GetUserPermissions)                      // 获取用户的所有权限
		userRoleGroup.GET("/users/:userID/permissions/:permissionID/check", userRoleHandler.CheckUserPermission) // 检查用户权限
	}
}
