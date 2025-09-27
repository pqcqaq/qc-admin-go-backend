package funcs

import (
	"context"
	"fmt"
	"math"

	"go-backend/database/ent"
	"go-backend/database/ent/permission"
	"go-backend/database/ent/rolepermission"
	"go-backend/database/ent/scope"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// PermissionService 权限服务
type PermissionFuncs struct{}

// GetAllPermissions 获取所有权限
func (PermissionFuncs) GetAllPermissions(ctx context.Context) ([]*ent.Permission, error) {
	return database.Client.Permission.Query().
		All(ctx)
}

// GetPermissionByID 根据ID获取权限
func (PermissionFuncs) GetPermissionByID(ctx context.Context, id uint64) (*ent.Permission, error) {
	permission, err := database.Client.Permission.Query().
		Where(permission.ID(id)).
		WithScope().
		WithRolePermissions(func(rp *ent.RolePermissionQuery) {
			rp.WithRole()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, err
	}
	return permission, nil
}

// CreatePermission 创建权限
func (PermissionFuncs) CreatePermission(ctx context.Context, req *models.CreatePermissionRequest) (*ent.Permission, error) {
	builder := database.Client.Permission.Create().
		SetName(req.Name).
		SetAction(req.Action)

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.ScopeId != "" {
		scopeId := utils.StringToUint64(req.ScopeId)
		builder = builder.SetScopeID(scopeId)
	}

	permission, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return PermissionFuncs{}.GetPermissionByID(ctx, permission.ID)
}

// UpdatePermission 更新权限
func (PermissionFuncs) UpdatePermission(ctx context.Context, id uint64, req *models.UpdatePermissionRequest) (*ent.Permission, error) {
	builder := database.Client.Permission.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	if req.Action != "" {
		builder = builder.SetAction(req.Action)
	}

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.ScopeId != "" {
		scopeId := utils.StringToUint64(req.ScopeId)
		builder = builder.SetScopeID(scopeId)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, err
	}

	return PermissionFuncs{}.GetPermissionByID(ctx, id)
}

// DeletePermission 删除权限
func (PermissionFuncs) DeletePermission(ctx context.Context, id uint64) error {
	err := database.Client.Permission.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("permission not found")
		}
		return err
	}
	return nil
}

// GetPermissionsWithPagination 分页获取权限列表
func (PermissionFuncs) GetPermissionsWithPagination(ctx context.Context, req *models.GetPermissionsRequest) (*models.PermissionsListResponse, error) {
	query := database.Client.Permission.Query().
		WithScope().
		WithRolePermissions(func(rp *ent.RolePermissionQuery) {
			rp.WithRole()
		})

	// 添加搜索条件
	if req.Name != "" {
		query = query.Where(permission.NameContains(req.Name))
	}

	if req.Action != "" {
		query = query.Where(permission.ActionContains(req.Action))
	}

	if req.Description != "" {
		query = query.Where(permission.DescriptionContains(req.Description))
	}

	if req.ScopeId != "" {
		scopeId := utils.StringToUint64(req.ScopeId)
		query = query.Where(permission.HasScopeWith(scope.ID(scopeId)))
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
				query = query.Order(ent.Desc(permission.FieldName))
			} else {
				query = query.Order(ent.Asc(permission.FieldName))
			}
		case "action":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(permission.FieldAction))
			} else {
				query = query.Order(ent.Asc(permission.FieldAction))
			}
		case "createTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(permission.FieldCreateTime))
			} else {
				query = query.Order(ent.Asc(permission.FieldCreateTime))
			}
		case "updateTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(permission.FieldUpdateTime))
			} else {
				query = query.Order(ent.Asc(permission.FieldUpdateTime))
			}
		}
	} else {
		// 默认按创建时间倒序
		query = query.Order(ent.Desc(permission.FieldCreateTime))
	}

	// 执行查询
	permissions, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	permissionResponses := make([]*models.PermissionResponse, 0, len(permissions))
	for _, p := range permissions {
		permissionResponses = append(permissionResponses, PermissionFuncs{}.ConvertPermissionToResponse(p))
	}

	return &models.PermissionsListResponse{
		Data: permissionResponses,
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

// ConvertPermissionToResponse 将权限实体转换为响应格式
func (PermissionFuncs) ConvertPermissionToResponse(p *ent.Permission) *models.PermissionResponse {
	resp := &models.PermissionResponse{
		ID:          utils.Uint64ToString(p.ID),
		Name:        p.Name,
		Action:      p.Action,
		Description: p.Description,
		CreateTime:  utils.FormatDateTime(p.CreateTime),
		UpdateTime:  utils.FormatDateTime(p.UpdateTime),
	}

	// 转换权限域
	if p.Edges.Scope != nil {
		resp.Scope = ScopeFuncs{}.ConvertScopeToResponse(p.Edges.Scope)
	}

	return resp
}

// GetRolePermissionsWithPagination 分页获取角色权限关联列表
func (PermissionFuncs) GetRolePermissionsWithPagination(ctx context.Context, req *models.GetRolePermissionsRequest) (*models.RolePermissionsListResponse, error) {
	query := database.Client.RolePermission.Query().
		WithRole().
		WithPermission()

	// 添加搜索条件
	if req.RoleId != "" {
		roleId := utils.StringToUint64(req.RoleId)
		query = query.Where(rolepermission.RoleID(roleId))
	}

	if req.PermissionId != "" {
		permissionId := utils.StringToUint64(req.PermissionId)
		query = query.Where(rolepermission.PermissionID(permissionId))
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
		case "createTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(rolepermission.FieldCreateTime))
			} else {
				query = query.Order(ent.Asc(rolepermission.FieldCreateTime))
			}
		case "updateTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(rolepermission.FieldUpdateTime))
			} else {
				query = query.Order(ent.Asc(rolepermission.FieldUpdateTime))
			}
		}
	} else {
		// 默认按创建时间倒序
		query = query.Order(ent.Desc(rolepermission.FieldCreateTime))
	}

	// 执行查询
	rolePermissions, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	rolePermissionResponses := make([]*models.RolePermissionResponse, 0, len(rolePermissions))
	for _, rp := range rolePermissions {
		resp := &models.RolePermissionResponse{
			ID:           utils.Uint64ToString(rp.ID),
			RoleId:       utils.Uint64ToString(rp.RoleID),
			PermissionId: utils.Uint64ToString(rp.PermissionID),
			CreateTime:   utils.FormatDateTime(rp.CreateTime),
			UpdateTime:   utils.FormatDateTime(rp.UpdateTime),
		}

		if rp.Edges.Role != nil {
			resp.Role = RoleFuncs{}.ConvertRoleToResponse(rp.Edges.Role)
		}

		if rp.Edges.Permission != nil {
			resp.Permission = PermissionFuncs{}.ConvertPermissionToResponse(rp.Edges.Permission)
		}

		rolePermissionResponses = append(rolePermissionResponses, resp)
	}

	return &models.RolePermissionsListResponse{
		Data: rolePermissionResponses,
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
