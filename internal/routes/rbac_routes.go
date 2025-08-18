package routes

import (
	"go-backend/internal/handlers"
	"go-backend/internal/middleware"

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
	rbacGroup.Use(middleware.JWTAuthMiddleware())
	{
		// 角色路由
		roleGroup := rbacGroup.Group("/roles")
		{
			roleGroup.GET("", roleHandler.GetRoles)                                              // 获取角色列表(分页)
			roleGroup.GET("/all", roleHandler.GetAllRoles)                                       // 获取所有角色(不分页)
			roleGroup.POST("", roleHandler.CreateRole)                                           // 创建角色
			roleGroup.GET("/:id", roleHandler.GetRole)                                           // 获取单个角色
			roleGroup.PUT("/:id", roleHandler.UpdateRole)                                        // 更新角色
			roleGroup.DELETE("/:id", roleHandler.DeleteRole)                                     // 删除角色
			roleGroup.POST("/:id/permissions", roleHandler.AssignRolePermissions)                // 分配角色权限
			roleGroup.DELETE("/:id/permissions/:permissionId", roleHandler.RevokeRolePermission) // 撤销角色权限
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
			userRoleGroup.POST("", userRoleHandler.AssignRole)                                                       // 分配用户角色
			userRoleGroup.DELETE("/users/:userID/roles/:roleID", userRoleHandler.RevokeRole)                         // 撤销用户角色
			userRoleGroup.GET("/users/:userID/roles", userRoleHandler.GetUserRoles)                                  // 获取用户的所有角色
			userRoleGroup.GET("/roles/:roleID/users", userRoleHandler.GetRoleUsers)                                  // 获取拥有指定角色的用户
			userRoleGroup.GET("/users/:userID/permissions", userRoleHandler.GetUserPermissions)                      // 获取用户的所有权限
			userRoleGroup.GET("/users/:userID/permissions/:permissionID/check", userRoleHandler.CheckUserPermission) // 检查用户权限
		}
	}
}
