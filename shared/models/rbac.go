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
	UserID uint64 `json:"userId" binding:"required"`
	RoleID uint64 `json:"roleId" binding:"required"`
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
	UserID uint64 `json:"userId" binding:"required"`
	RoleID uint64 `json:"roleId" binding:"required"`
}

// UserRoleResponse 用户角色关联响应结构
type UserRoleResponse struct {
	ID         string        `json:"id"`
	UserID     uint64        `json:"userId"`
	RoleID     uint64        `json:"roleId"`
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
