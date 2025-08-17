package models

import "go-backend/pkg/utils"

// CreateScopeRequest 创建权限域请求结构
type CreateScopeRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=menu page button"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	Action      string `json:"action,omitempty"`
	Path        string `json:"path,omitempty"`
	Component   string `json:"component,omitempty"`
	Redirect    string `json:"redirect,omitempty"`
	Order       int    `json:"order"`
	Hidden      bool   `json:"hidden"`
	Disabled    bool   `json:"disabled"`
	ParentId    string `json:"parentId,omitempty"` // 父级ID
}

// UpdateScopeRequest 更新权限域请求结构
type UpdateScopeRequest struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	Action      string `json:"action,omitempty"`
	Path        string `json:"path,omitempty"`
	Component   string `json:"component,omitempty"`
	Redirect    string `json:"redirect,omitempty"`
	Order       *int   `json:"order,omitempty"`
	Hidden      *bool  `json:"hidden,omitempty"`
	Disabled    *bool  `json:"disabled,omitempty"`
	ParentId    string `json:"parentId,omitempty"` // 父级ID
}

// ScopeResponse 权限域响应结构
type ScopeResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Type        string                `json:"type"`
	Icon        string                `json:"icon,omitempty"`
	Description string                `json:"description,omitempty"`
	Action      string                `json:"action,omitempty"`
	Path        string                `json:"path,omitempty"`
	Component   string                `json:"component,omitempty"`
	Redirect    string                `json:"redirect,omitempty"`
	Order       int                   `json:"order"`
	Hidden      bool                  `json:"hidden"`
	Disabled    bool                  `json:"disabled"`
	ParentId    string                `json:"parentId,omitempty"`
	Parent      *ScopeResponse        `json:"parent,omitempty"`      // 父级权限域
	Children    []*ScopeResponse      `json:"children,omitempty"`    // 子级权限域
	Permissions []*PermissionResponse `json:"permissions,omitempty"` // 关联的权限
	CreateTime  utils.JSONTime        `json:"createTime"`
	UpdateTime  utils.JSONTime        `json:"updateTime"`
}

// GetScopesRequest 获取权限域列表请求结构
type GetScopesRequest struct {
	PaginationRequest
	Name        string `form:"name" json:"name"`               // 按名称模糊搜索
	Type        string `form:"type" json:"type"`               // 按类型搜索
	Description string `form:"description" json:"description"` // 按描述模糊搜索
	ParentId    string `form:"parentId" json:"parentId"`       // 按父级ID搜索
	Hidden      *bool  `form:"hidden" json:"hidden"`           // 是否隐藏
	Disabled    *bool  `form:"disabled" json:"disabled"`       // 是否禁用
}

// ScopesListResponse 权限域列表响应结构
type ScopesListResponse struct {
	Data       []*ScopeResponse `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// ScopeTreeResponse 权限域树形结构响应
type ScopeTreeResponse struct {
	Data []*ScopeResponse `json:"data"` // 树形结构的权限域
}
