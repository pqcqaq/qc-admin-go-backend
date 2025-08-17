package funcs

import (
	"context"
	"fmt"
	"math"

	"go-backend/database/ent"
	"go-backend/database/ent/role"
	"go-backend/database/ent/rolepermission"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// RoleService 角色服务

// GetAllRoles 获取所有角色
func GetAllRoles(ctx context.Context) ([]*ent.Role, error) {
	return database.Client.Role.Query().
		WithInheritsFrom().
		WithInheritedBy().
		WithRolePermissions(func(rp *ent.RolePermissionQuery) {
			rp.WithPermission()
		}).
		All(ctx)
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(ctx context.Context, id uint64) (*ent.Role, error) {
	role, err := database.Client.Role.Query().
		Where(role.ID(id)).
		WithInheritsFrom().
		WithInheritedBy().
		WithRolePermissions(func(rp *ent.RolePermissionQuery) {
			rp.WithPermission()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}
	return role, nil
}

// CreateRole 创建角色
func CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*ent.Role, error) {
	builder := database.Client.Role.Create().
		SetName(req.Name)

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	// 如果有父角色，设置继承关系
	if len(req.InheritsFrom) > 0 {
		for _, parentIdStr := range req.InheritsFrom {
			parentId := utils.StringToUint64(parentIdStr)
			builder.AddInheritsFromIDs(parentId)
		}
		// err = database.Client.Role.UpdateOneID(role.ID).
		// 	AddInheritsFromIDs(parentIds...).
		// 	Exec(ctx)
		// if err != nil {
		// 	// 如果设置继承关系失败，删除已创建的角色
		// 	database.Client.Role.DeleteOneID(role.ID).ExecX(ctx)
		// 	return nil, fmt.Errorf("failed to set role inheritance: %v", err)
		// }
	}

	// 创建角色
	role, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// 重新查询角色以获取完整信息
	return GetRoleByID(ctx, role.ID)
}

// UpdateRole 更新角色
func UpdateRole(ctx context.Context, id uint64, req *models.UpdateRoleRequest) (*ent.Role, error) {
	builder := database.Client.Role.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	// 如果有父角色更新
	if len(req.InheritsFrom) > 0 {
		parentIds := make([]uint64, 0, len(req.InheritsFrom))
		for _, parentIdStr := range req.InheritsFrom {
			parentId := utils.StringToUint64(parentIdStr)
			parentIds = append(parentIds, parentId)
		}

		// 先清除现有的继承关系，再设置新的
		builder = builder.ClearInheritsFrom().AddInheritsFromIDs(parentIds...)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	return GetRoleByID(ctx, id)
}

// DeleteRole 删除角色
func DeleteRole(ctx context.Context, id uint64) error {
	err := database.Client.Role.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role not found")
		}
		return err
	}
	return nil
}

// GetRolesWithPagination 分页获取角色列表
func GetRolesWithPagination(ctx context.Context, req *models.GetRolesRequest) (*models.RolesListResponse, error) {
	query := database.Client.Role.Query().
		WithInheritsFrom().
		WithInheritedBy().
		WithRolePermissions(func(rp *ent.RolePermissionQuery) {
			rp.WithPermission()
		})

	// 添加搜索条件
	if req.Name != "" {
		query = query.Where(role.NameContains(req.Name))
	}

	if req.Description != "" {
		query = query.Where(role.DescriptionContains(req.Description))
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
				query = query.Order(ent.Desc(role.FieldName))
			} else {
				query = query.Order(ent.Asc(role.FieldName))
			}
		case "createTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(role.FieldCreateTime))
			} else {
				query = query.Order(ent.Asc(role.FieldCreateTime))
			}
		case "updateTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(role.FieldUpdateTime))
			} else {
				query = query.Order(ent.Asc(role.FieldUpdateTime))
			}
		}
	} else {
		// 默认按创建时间倒序
		query = query.Order(ent.Desc(role.FieldCreateTime))
	}

	// 执行查询
	roles, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	roleResponses := make([]*models.RoleResponse, 0, len(roles))
	for _, r := range roles {
		roleResponses = append(roleResponses, ConvertRoleToResponse(r))
	}

	return &models.RolesListResponse{
		Data: roleResponses,
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

// AssignRolePermissions 分配角色权限
func AssignRolePermissions(ctx context.Context, roleID uint64, req *models.AssignRolePermissionsRequest) error {
	// 先删除现有的角色权限关联
	_, err := database.Client.RolePermission.Delete().
		Where(rolepermission.RoleID(roleID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear existing permissions: %v", err)
	}

	// 创建新的角色权限关联
	if len(req.PermissionIds) > 0 {
		bulk := make([]*ent.RolePermissionCreate, 0, len(req.PermissionIds))
		for _, permissionIdStr := range req.PermissionIds {
			permissionId := utils.StringToUint64(permissionIdStr)
			bulk = append(bulk, database.Client.RolePermission.Create().
				SetRoleID(roleID).
				SetPermissionID(permissionId))
		}

		_, err = database.Client.RolePermission.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to assign permissions: %v", err)
		}
	}

	return nil
}

// GetRolePermissions 获取角色的权限列表
func GetRolePermissions(ctx context.Context, roleId uint64) ([]*ent.Permission, error) {
	role, err := database.Client.Role.Query().
		Where(role.ID(roleId)).
		WithRolePermissions(func(rp *ent.RolePermissionQuery) {
			rp.WithPermission()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	permissions := make([]*ent.Permission, 0, len(role.Edges.RolePermissions))
	for _, rp := range role.Edges.RolePermissions {
		if rp.Edges.Permission != nil {
			permissions = append(permissions, rp.Edges.Permission)
		}
	}

	return permissions, nil
}

// ConvertRoleToResponse 将角色实体转换为响应格式
func ConvertRoleToResponse(r *ent.Role) *models.RoleResponse {
	resp := &models.RoleResponse{
		ID:          utils.Uint64ToString(r.ID),
		Name:        r.Name,
		Description: r.Description,
		CreateTime:  utils.FormatDateTime(r.CreateTime),
		UpdateTime:  utils.FormatDateTime(r.UpdateTime),
	}

	// 转换父角色
	if r.Edges.InheritsFrom != nil {
		resp.InheritsFrom = make([]*models.RoleResponse, 0, len(r.Edges.InheritsFrom))
		for _, parent := range r.Edges.InheritsFrom {
			resp.InheritsFrom = append(resp.InheritsFrom, &models.RoleResponse{
				ID:          utils.Uint64ToString(parent.ID),
				Name:        parent.Name,
				Description: parent.Description,
				CreateTime:  utils.FormatDateTime(parent.CreateTime),
				UpdateTime:  utils.FormatDateTime(parent.UpdateTime),
			})
		}
	}

	// 转换子角色
	if r.Edges.InheritedBy != nil {
		resp.InheritedBy = make([]*models.RoleResponse, 0, len(r.Edges.InheritedBy))
		for _, child := range r.Edges.InheritedBy {
			resp.InheritedBy = append(resp.InheritedBy, &models.RoleResponse{
				ID:          utils.Uint64ToString(child.ID),
				Name:        child.Name,
				Description: child.Description,
				CreateTime:  utils.FormatDateTime(child.CreateTime),
				UpdateTime:  utils.FormatDateTime(child.UpdateTime),
			})
		}
	}

	// 转换权限
	if r.Edges.RolePermissions != nil {
		resp.Permissions = make([]*models.PermissionResponse, 0, len(r.Edges.RolePermissions))
		for _, rp := range r.Edges.RolePermissions {
			if rp.Edges.Permission != nil {
				resp.Permissions = append(resp.Permissions, ConvertPermissionToResponse(rp.Edges.Permission))
			}
		}
	}

	return resp
}

// GetRoleInheritedPermissions 获取角色及其继承链的所有权限
func GetRoleInheritedPermissions(ctx context.Context, roleID uint64) ([]*ent.Permission, error) {
	// 获取角色的所有继承链
	visited := make(map[uint64]bool)
	var getAllRolePermissions func(uint64) ([]*ent.Permission, error)

	getAllRolePermissions = func(currentRoleID uint64) ([]*ent.Permission, error) {
		// 防止循环继承
		if visited[currentRoleID] {
			return []*ent.Permission{}, nil
		}
		visited[currentRoleID] = true

		// 获取当前角色
		currentRole, err := database.Client.Role.Query().
			Where(role.ID(currentRoleID)).
			WithInheritsFrom().
			WithRolePermissions(func(rp *ent.RolePermissionQuery) {
				rp.WithPermission()
			}).
			Only(ctx)
		if err != nil {
			return nil, err
		}

		var allPermissions []*ent.Permission

		// 获取当前角色的直接权限
		if currentRole.Edges.RolePermissions != nil {
			for _, rp := range currentRole.Edges.RolePermissions {
				if rp.Edges.Permission != nil {
					allPermissions = append(allPermissions, rp.Edges.Permission)
				}
			}
		}

		// 递归获取继承角色的权限
		if currentRole.Edges.InheritsFrom != nil {
			for _, parentRole := range currentRole.Edges.InheritsFrom {
				parentPermissions, err := getAllRolePermissions(parentRole.ID)
				if err != nil {
					return nil, err
				}
				allPermissions = append(allPermissions, parentPermissions...)
			}
		}

		return allPermissions, nil
	}

	permissions, err := getAllRolePermissions(roleID)
	if err != nil {
		return nil, err
	}

	// 去重
	permissionMap := make(map[uint64]*ent.Permission)
	for _, perm := range permissions {
		permissionMap[perm.ID] = perm
	}

	// 转换为切片
	uniquePermissions := make([]*ent.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		uniquePermissions = append(uniquePermissions, perm)
	}

	return uniquePermissions, nil
}

// RevokeRolePermission 撤销角色权限
func RevokeRolePermission(ctx context.Context, roleID, permissionID uint64) error {
	// 查找角色权限关联
	rolePermission, err := database.Client.RolePermission.Query().
		Where(
			rolepermission.RoleID(roleID),
			rolepermission.PermissionID(permissionID),
		).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role permission not found")
		}
		return err
	}

	// 删除关联
	err = database.Client.RolePermission.DeleteOne(rolePermission).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
