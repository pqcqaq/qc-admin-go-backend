package models

// CreateRoleRequest 创建角色请求结构
type CreateRoleRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description,omitempty"`
	InheritsFrom []string `json:"inheritsFrom,omitempty"` // 继承的父角色ID列表
}

// UpdateRoleRequest 更新角色请求结构
type UpdateRoleRequest struct {
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	InheritsFrom []string `json:"inheritsFrom,omitempty"` // 继承的父角色ID列表
}

// RoleResponse 角色响应结构
type RoleResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description,omitempty"`
	InheritsFrom []*RoleResponse       `json:"inheritsFrom,omitempty"` // 继承的父角色
	InheritedBy  []*RoleResponse       `json:"inheritedBy,omitempty"`  // 被哪些角色继承
	Permissions  []*PermissionResponse `json:"permissions,omitempty"`  // 角色拥有的权限
	CreateTime   string                `json:"createTime"`
	UpdateTime   string                `json:"updateTime"`
}

// GetRolesRequest 获取角色列表请求结构
type GetRolesRequest struct {
	PaginationRequest
	Name        string `form:"name" json:"name"`               // 按角色名称模糊搜索
	Description string `form:"description" json:"description"` // 按描述模糊搜索
}

// RolesListResponse 角色列表响应结构
type RolesListResponse struct {
	Data       []*RoleResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// AssignRolePermissionsRequest 分配角色权限请求结构
type AssignRolePermissionsRequest struct {
	PermissionIds []string `json:"permissionIds" binding:"required"` // 权限ID列表
}

// AssignUserRoleRequest 分配用户角色请求结构
type AssignUserRoleRequest struct {
	UserID string `json:"userId" binding:"required"`
	RoleID string `json:"roleId" binding:"required"`
}

// CreatePermissionRequest 创建权限请求结构
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description,omitempty"`
	ScopeId     string `json:"scopeId,omitempty"` // 权限域ID
}

// UpdatePermissionRequest 更新权限请求结构
type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty"`
	Action      string `json:"action,omitempty"`
	Description string `json:"description,omitempty"`
	ScopeId     string `json:"scopeId,omitempty"` // 权限域ID
}

// PermissionResponse 权限响应结构
type PermissionResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Action      string         `json:"action"`
	Description string         `json:"description,omitempty"`
	Scope       *ScopeResponse `json:"scope,omitempty"` // 所属权限域
	CreateTime  string         `json:"createTime"`
	UpdateTime  string         `json:"updateTime"`
}

// GetPermissionsRequest 获取权限列表请求结构
type GetPermissionsRequest struct {
	PaginationRequest
	Name        string `form:"name" json:"name"`               // 按权限名称模糊搜索
	Action      string `form:"action" json:"action"`           // 按操作类型搜索
	Description string `form:"description" json:"description"` // 按描述模糊搜索
	ScopeId     string `form:"scopeId" json:"scopeId"`         // 按权限域ID搜索
}

// PermissionsListResponse 权限列表响应结构
type PermissionsListResponse struct {
	Data       []*PermissionResponse `json:"data"`
	Pagination Pagination            `json:"pagination"`
}

// CreateUserRoleRequest 创建用户角色关联请求结构
type CreateUserRoleRequest struct {
	UserID string `json:"userId" binding:"required"`
	RoleID string `json:"roleId" binding:"required"`
}

// UserRoleResponse 用户角色关联响应结构
type UserRoleResponse struct {
	ID         string        `json:"id"`
	UserID     string        `json:"userId"`
	RoleID     string        `json:"roleId"`
	User       *UserResponse `json:"user,omitempty"`
	Role       *RoleResponse `json:"role,omitempty"`
	CreateTime string        `json:"createTime"`
	UpdateTime string        `json:"updateTime"`
}

// GetUserRolesRequest 获取用户角色列表请求结构
type GetUserRolesRequest struct {
	PaginationRequest
	UserId string `form:"userId" json:"userId"` // 按用户ID搜索
	RoleId string `form:"roleId" json:"roleId"` // 按角色ID搜索
}

// UserRolesListResponse 用户角色列表响应结构
type UserRolesListResponse struct {
	Data       []*UserRoleResponse `json:"data"`
	Pagination Pagination          `json:"pagination"`
}

// RolePermissionResponse 角色权限关联响应结构
type RolePermissionResponse struct {
	ID           string              `json:"id"`
	RoleId       string              `json:"roleId"`
	PermissionId string              `json:"permissionId"`
	Role         *RoleResponse       `json:"role,omitempty"`
	Permission   *PermissionResponse `json:"permission,omitempty"`
	CreateTime   string              `json:"createTime"`
	UpdateTime   string              `json:"updateTime"`
}

// GetRolePermissionsRequest 获取角色权限列表请求结构
type GetRolePermissionsRequest struct {
	PaginationRequest
	RoleId       string `form:"roleId" json:"roleId"`             // 按角色ID搜索
	PermissionId string `form:"permissionId" json:"permissionId"` // 按权限ID搜索
}

// RolePermissionsListResponse 角色权限列表响应结构
type RolePermissionsListResponse struct {
	Data       []*RolePermissionResponse `json:"data"`
	Pagination Pagination                `json:"pagination"`
}

// === 新增的RBAC管理页面接口模型 ===

// RoleTreeResponse 角色树响应结构
type RoleTreeResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	UserCount    int                 `json:"userCount"`              // 拥有该角色的用户数量
	InheritsFrom []*RoleResponse     `json:"inheritsFrom,omitempty"` // 继承的父角色
	Children     []*RoleTreeResponse `json:"children,omitempty"`     // 子角色
	CreateTime   string              `json:"createTime"`
	UpdateTime   string              `json:"updateTime"`
}

// RoleDetailedPermissionsResponse 角色详细权限响应结构
type RoleDetailedPermissionsResponse struct {
	Role                 *RoleResponse           `json:"role"`
	DirectPermissions    []*PermissionWithSource `json:"directPermissions"`    // 直接分配的权限
	InheritedPermissions []*PermissionWithSource `json:"inheritedPermissions"` // 继承的权限
	AllPermissions       []*PermissionResponse   `json:"allPermissions"`       // 所有权限
}

// PermissionWithSource 权限来源信息
type PermissionWithSource struct {
	Permission *PermissionResponse `json:"permission"`
	Source     string              `json:"source"`     // 来源类型: "直接分配" | "角色继承"
	SourceRole *RoleResponse       `json:"sourceRole"` // 来源角色
}

// InheritedPermissionInfo 继承权限信息
type InheritedPermissionInfo struct {
	Permission *PermissionResponse `json:"permission"`
	FromRole   *RoleResponse       `json:"fromRole"` // 来源角色
}

// CreateChildRoleRequest 创建子角色请求结构
type CreateChildRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// GetRoleUsersRequest 获取角色用户请求结构
type GetRoleUsersRequest struct {
	PaginationRequest
	Keyword string `form:"keyword" json:"keyword"` // 搜索关键字（用户名）
}

// RoleUsersResponse 角色用户响应结构
type RoleUsersResponse struct {
	Users      []*UserResponse `json:"users"`
	Pagination Pagination      `json:"pagination"`
}

// RoleUserResponse 角色用户响应结构
type RoleUserResponse struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Nickname   string          `json:"nickname,omitempty"`
	OtherRoles []*RoleResponse `json:"otherRoles,omitempty"` // 除当前角色外的其他角色
	CreateTime string          `json:"createTime"`
	UpdateTime string          `json:"updateTime"`
}

// RoleUsersListResponse 角色用户列表响应结构
type RoleUsersListResponse struct {
	Data       []*RoleUserResponse `json:"data"`
	Pagination Pagination          `json:"pagination"`
}

// BatchAssignUsersToRoleRequest 批量分配用户到角色请求结构
type BatchAssignUsersToRoleRequest struct {
	UserIds []string `json:"userIds" binding:"required"` // 用户ID列表
}

// BatchRemoveUsersFromRoleRequest 批量从角色移除用户请求结构
type BatchRemoveUsersFromRoleRequest struct {
	UserIds []string `json:"userIds" binding:"required"` // 用户ID列表
}

// UserWithRolesResponse 带角色信息的用户响应结构
type UserWithRolesResponse struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Nickname   string          `json:"nickname,omitempty"`
	Age        int             `json:"age,omitempty"`
	Sex        string          `json:"sex,omitempty"`
	Status     string          `json:"status,omitempty"`
	Roles      []*RoleResponse `json:"roles,omitempty"` // 用户的所有角色
	CreateTime string          `json:"createTime"`
	UpdateTime string          `json:"updateTime"`
}

// UsersWithRolesListResponse 带角色信息的用户列表响应结构
type UsersWithRolesListResponse struct {
	Data       []*UserWithRolesResponse `json:"data"`
	Pagination Pagination               `json:"pagination"`
}
