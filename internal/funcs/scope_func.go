package funcs

import (
	"context"
	"fmt"
	"math"

	"go-backend/database/ent"
	"go-backend/database/ent/scope"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// ScopeService 权限域服务

// GetAllScopes 获取所有权限域
func GetAllScopes(ctx context.Context) ([]*ent.Scope, error) {
	return database.Client.Scope.Query().
		WithParent().
		WithChildren().
		WithPermissions().
		All(ctx)
}

// GetScopeByID 根据ID获取权限域
func GetScopeByID(ctx context.Context, id uint64) (*ent.Scope, error) {
	scope, err := database.Client.Scope.Query().
		Where(scope.ID(id)).
		WithParent().
		WithChildren().
		WithPermissions().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("scope not found")
		}
		return nil, err
	}
	return scope, nil
}

// CreateScope 创建权限域
func CreateScope(ctx context.Context, req *models.CreateScopeRequest) (*ent.Scope, error) {
	builder := database.Client.Scope.Create().
		SetName(req.Name).
		SetType(scope.Type(req.Type)).
		SetOrder(req.Order).
		SetHidden(req.Hidden).
		SetDisabled(req.Disabled)

	if req.Icon != "" {
		builder = builder.SetIcon(req.Icon)
	}

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.Action != "" {
		builder = builder.SetAction(req.Action)
	}

	if req.Path != "" {
		builder = builder.SetPath(req.Path)
	}

	if req.Component != "" {
		builder = builder.SetComponent(req.Component)
	}

	if req.Redirect != "" {
		builder = builder.SetRedirect(req.Redirect)
	}

	if req.ParentId != "" {
		parentId := utils.StringToUint64(req.ParentId)
		builder = builder.SetParentID(parentId)
	}

	scope, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return GetScopeByID(ctx, scope.ID)
}

// UpdateScope 更新权限域
func UpdateScope(ctx context.Context, id uint64, req *models.UpdateScopeRequest) (*ent.Scope, error) {
	builder := database.Client.Scope.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	if req.Type != "" {
		builder = builder.SetType(scope.Type(req.Type))
	}

	if req.Icon != "" {
		builder = builder.SetIcon(req.Icon)
	}

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.Action != "" {
		builder = builder.SetAction(req.Action)
	}

	if req.Path != "" {
		builder = builder.SetPath(req.Path)
	}

	if req.Component != "" {
		builder = builder.SetComponent(req.Component)
	}

	if req.Redirect != "" {
		builder = builder.SetRedirect(req.Redirect)
	}

	if req.Order != nil {
		builder = builder.SetOrder(*req.Order)
	}

	if req.Hidden != nil {
		builder = builder.SetHidden(*req.Hidden)
	}

	if req.Disabled != nil {
		builder = builder.SetDisabled(*req.Disabled)
	}

	if req.ParentId != "" {
		parentId := utils.StringToUint64(req.ParentId)
		builder = builder.SetParentID(parentId)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("scope not found")
		}
		return nil, err
	}

	return GetScopeByID(ctx, id)
}

// DeleteScope 删除权限域
func DeleteScope(ctx context.Context, id uint64) error {
	err := database.Client.Scope.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("scope not found")
		}
		return err
	}
	return nil
}

// GetScopesWithPagination 分页获取权限域列表
func GetScopesWithPagination(ctx context.Context, req *models.GetScopesRequest) (*models.ScopesListResponse, error) {
	query := database.Client.Scope.Query().
		WithParent().
		WithChildren().
		WithPermissions()

	// 添加搜索条件
	if req.Name != "" {
		query = query.Where(scope.NameContains(req.Name))
	}

	if req.Type != "" {
		query = query.Where(scope.TypeEQ(scope.Type(req.Type)))
	}

	if req.Description != "" {
		query = query.Where(scope.DescriptionContains(req.Description))
	}

	if req.ParentId != "" {
		parentId := utils.StringToUint64(req.ParentId)
		query = query.Where(scope.ParentIDEQ(parentId))
	}

	if req.Hidden != nil {
		query = query.Where(scope.HiddenEQ(*req.Hidden))
	}

	if req.Disabled != nil {
		query = query.Where(scope.DisabledEQ(*req.Disabled))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// 计算分页
	offset := (req.Page - 1) * req.PageSize
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	// 设置排序
	if req.OrderBy != "" {
		switch req.OrderBy {
		case "name":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(scope.FieldName))
			} else {
				query = query.Order(ent.Asc(scope.FieldName))
			}
		case "order":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(scope.FieldOrder))
			} else {
				query = query.Order(ent.Asc(scope.FieldOrder))
			}
		case "createTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(scope.FieldCreateTime))
			} else {
				query = query.Order(ent.Asc(scope.FieldCreateTime))
			}
		case "updateTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(scope.FieldUpdateTime))
			} else {
				query = query.Order(ent.Asc(scope.FieldUpdateTime))
			}
		}
	} else {
		// 默认按order字段和创建时间排序
		query = query.Order(ent.Asc(scope.FieldOrder), ent.Desc(scope.FieldCreateTime))
	}

	// 执行查询
	scopes, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	scopeResponses := make([]*models.ScopeResponse, 0, len(scopes))
	for _, s := range scopes {
		scopeResponses = append(scopeResponses, ConvertScopeToResponse(s))
	}

	return &models.ScopesListResponse{
		Data: scopeResponses,
		Pagination: models.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      int64(total),
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

// GetScopeTree 获取权限域树形结构
func GetScopeTree(ctx context.Context) (*models.ScopeTreeResponse, error) {
	// 获取所有权限域
	allScopes, err := database.Client.Scope.Query().
		WithParent().
		WithPermissions().
		Order(ent.Asc(scope.FieldOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	scopeMap := make(map[uint64]*models.ScopeResponse)
	var rootScopes []*models.ScopeResponse

	// 先创建所有节点（不包含Children，避免重复）
	for _, s := range allScopes {
		scopeResp := ConvertScopeToResponseForTree(s)
		scopeMap[s.ID] = scopeResp

		// 如果没有父节点，则为根节点
		if s.ParentID == 0 {
			rootScopes = append(rootScopes, scopeResp)
		}
	}

	// 构建父子关系
	for _, s := range allScopes {
		if s.ParentID != 0 {
			parent := scopeMap[s.ParentID]
			child := scopeMap[s.ID]
			if parent != nil && child != nil {
				if parent.Children == nil {
					parent.Children = make([]*models.ScopeResponse, 0)
				}
				parent.Children = append(parent.Children, child)
				// 设置子节点的父节点信息
				child.Parent = &models.ScopeResponse{
					ID:   parent.ID,
					Name: parent.Name,
					Type: parent.Type,
				}
			}
		}
	}

	return &models.ScopeTreeResponse{
		Data: rootScopes,
	}, nil
}

// ConvertScopeToResponse 将权限域实体转换为响应格式
func ConvertScopeToResponse(s *ent.Scope) *models.ScopeResponse {
	resp := &models.ScopeResponse{
		ID:          utils.Uint64ToString(s.ID),
		Name:        s.Name,
		Type:        string(s.Type),
		Icon:        s.Icon,
		Description: s.Description,
		Action:      s.Action,
		Path:        s.Path,
		Component:   s.Component,
		Redirect:    s.Redirect,
		Order:       s.Order,
		Hidden:      s.Hidden,
		Disabled:    s.Disabled,
		CreateTime:  utils.FormatDateTime(s.CreateTime),
		UpdateTime:  utils.FormatDateTime(s.UpdateTime),
	}

	if s.ParentID != 0 {
		resp.ParentId = utils.Uint64ToString(s.ParentID)
	}

	// 转换父级权限域（简单信息）
	if s.Edges.Parent != nil {
		resp.Parent = &models.ScopeResponse{
			ID:   utils.Uint64ToString(s.Edges.Parent.ID),
			Name: s.Edges.Parent.Name,
			Type: string(s.Edges.Parent.Type),
		}
	}

	// 转换子级权限域（简单信息）
	if len(s.Edges.Children) > 0 {
		resp.Children = make([]*models.ScopeResponse, 0, len(s.Edges.Children))
		for _, child := range s.Edges.Children {
			resp.Children = append(resp.Children, &models.ScopeResponse{
				ID:       utils.Uint64ToString(child.ID),
				Name:     child.Name,
				Type:     string(child.Type),
				Order:    child.Order,
				Hidden:   child.Hidden,
				Disabled: child.Disabled,
			})
		}
	}

	// 转换权限（简单信息）
	if len(s.Edges.Permissions) > 0 {
		resp.Permissions = make([]*models.PermissionResponse, 0, len(s.Edges.Permissions))
		for _, permission := range s.Edges.Permissions {
			resp.Permissions = append(resp.Permissions, &models.PermissionResponse{
				ID:     utils.Uint64ToString(permission.ID),
				Name:   permission.Name,
				Action: permission.Action,
			})
		}
	}

	return resp
}

// ConvertScopeToResponseForTree 将权限域实体转换为响应格式（专用于树形结构，不包含Children避免重复）
func ConvertScopeToResponseForTree(s *ent.Scope) *models.ScopeResponse {
	resp := &models.ScopeResponse{
		ID:          utils.Uint64ToString(s.ID),
		Name:        s.Name,
		Type:        string(s.Type),
		Icon:        s.Icon,
		Description: s.Description,
		Action:      s.Action,
		Path:        s.Path,
		Component:   s.Component,
		Redirect:    s.Redirect,
		Order:       s.Order,
		Hidden:      s.Hidden,
		Disabled:    s.Disabled,
		CreateTime:  utils.FormatDateTime(s.CreateTime),
		UpdateTime:  utils.FormatDateTime(s.UpdateTime),
	}

	if s.ParentID != 0 {
		resp.ParentId = utils.Uint64ToString(s.ParentID)
	}

	// 转换父级权限域（简单信息）
	if s.Edges.Parent != nil {
		resp.Parent = &models.ScopeResponse{
			ID:   utils.Uint64ToString(s.Edges.Parent.ID),
			Name: s.Edges.Parent.Name,
			Type: string(s.Edges.Parent.Type),
		}
	}

	// 只转换权限信息，不包含Children（树形结构中手动构建）
	if len(s.Edges.Permissions) > 0 {
		resp.Permissions = make([]*models.PermissionResponse, 0, len(s.Edges.Permissions))
		for _, permission := range s.Edges.Permissions {
			resp.Permissions = append(resp.Permissions, &models.PermissionResponse{
				ID:     utils.Uint64ToString(permission.ID),
				Name:   permission.Name,
				Action: permission.Action,
			})
		}
	}

	return resp
}
