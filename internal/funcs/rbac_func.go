package funcs

import (
	"context"
	"fmt"
	"math"

	"go-backend/database/ent"
	"go-backend/database/ent/permission"
	entRole "go-backend/database/ent/role"
	"go-backend/database/ent/rolepermission"
	"go-backend/database/ent/user"
	"go-backend/database/ent/userrole"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"

	"entgo.io/ent/dialect/sql"
)

// HasCircularInheritance 检查角色继承是否存在循环引用
// roleID: 当前角色ID
// parentID: 要设置的父角色ID
func HasCircularInheritance(ctx context.Context, client *ent.Client, roleID, parentID uint64) error {
	// 如果要设置自己为父角色，直接返回错误
	if roleID == parentID {
		return fmt.Errorf("角色不能继承自己")
	}

	// 使用深度优先搜索检测循环
	visited := make(map[uint64]bool)
	return dfsCheckCircular(ctx, client, parentID, roleID, visited, 0)
}

// dfsCheckCircular 使用深度优先搜索检测循环继承
// currentID: 当前遍历的角色ID
// targetID: 目标角色ID（我们要检查是否能到达的角色）
// visited: 已访问的角色ID集合
// depth: 当前继承深度
func dfsCheckCircular(ctx context.Context, client *ent.Client, currentID, targetID uint64, visited map[uint64]bool, depth int) error {
	// 检查继承深度是否超过限制（防止过深的继承链）
	const maxInheritanceDepth = 10
	if depth > maxInheritanceDepth {
		return fmt.Errorf("角色继承深度超过限制(%d层)", maxInheritanceDepth)
	}

	// 如果当前角色就是目标角色，说明存在循环
	if currentID == targetID {
		return fmt.Errorf("检测到角色继承循环")
	}

	// 如果已经访问过这个角色，说明存在循环
	if visited[currentID] {
		return fmt.Errorf("检测到角色继承循环")
	}

	// 标记当前角色为已访问
	visited[currentID] = true

	// 查询当前角色的所有父角色（inherits_from关系）
	parentRoles, err := client.Role.Query().
		Where(entRole.HasInheritedByWith(entRole.ID(currentID))).
		All(ctx)

	if err != nil {
		return fmt.Errorf("查询父角色失败: %v", err)
	}

	// 递归检查每个父角色
	for _, parentRole := range parentRoles {
		if err := dfsCheckCircular(ctx, client, parentRole.ID, targetID, visited, depth+1); err != nil {
			return err
		}
	}

	// 移除访问标记（回溯）
	delete(visited, currentID)
	return nil
}

// getAllAncestorRoles 获取角色的所有祖先角色（用于权限计算）
func GetAllAncestorRoles(ctx context.Context, client *ent.Client, roleID uint64) ([]*ent.Role, error) {
	var ancestors []*ent.Role
	visited := make(map[uint64]bool)

	err := collectAncestors(ctx, client, roleID, visited, &ancestors, 0)
	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

// collectAncestors 递归收集祖先角色
func collectAncestors(ctx context.Context, client *ent.Client, roleID uint64, visited map[uint64]bool, ancestors *[]*ent.Role, depth int) error {
	// 防止无限递归
	const maxDepth = 10
	if depth > maxDepth {
		return fmt.Errorf("角色继承深度超过限制")
	}

	if visited[roleID] {
		return nil // 已经访问过，避免重复
	}
	visited[roleID] = true

	// 查询当前角色的父角色（inherits_from关系）
	parentRoles, err := client.Role.Query().
		Where(entRole.HasInheritedByWith(entRole.ID(roleID))).
		All(ctx)

	if err != nil {
		return err
	}

	for _, parentRole := range parentRoles {
		*ancestors = append(*ancestors, parentRole)
		// 递归获取父角色的祖先
		if err := collectAncestors(ctx, client, parentRole.ID, visited, ancestors, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// GetRoleTree 获取角色树结构
func GetRoleTree(ctx context.Context) ([]*models.RoleTreeResponse, error) {
	// 获取所有角色
	roles, err := database.Client.Role.Query().
		WithInheritsFrom().
		WithInheritedBy().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %v", err)
	}

	// 创建角色映射
	roleMap := make(map[uint64]*models.RoleTreeResponse)
	for _, r := range roles {
		roleMap[r.ID] = &models.RoleTreeResponse{
			ID:          utils.Uint64ToString(r.ID),
			Name:        r.Name,
			Description: r.Description,
			Children:    []*models.RoleTreeResponse{},
		}
	}

	// 构建树结构
	var rootRoles []*models.RoleTreeResponse
	for _, r := range roles {
		node := roleMap[r.ID]

		// 如果有父角色，将自己添加到父角色的children中
		if len(r.Edges.InheritsFrom) > 0 {
			for _, parent := range r.Edges.InheritsFrom {
				if parentNode, exists := roleMap[parent.ID]; exists {
					parentNode.Children = append(parentNode.Children, node)
				}
			}
		} else {
			// 没有父角色，作为根节点
			rootRoles = append(rootRoles, node)
		}
	}

	return rootRoles, nil
}

// GetRoleWithPermissions 获取角色详细权限信息
func GetRoleWithPermissions(ctx context.Context, roleID uint64) (*models.RoleDetailedPermissionsResponse, error) {
	// 获取角色基本信息
	role, err := database.Client.Role.Query().
		Where(entRole.ID(roleID)).
		WithInheritsFrom().
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

	// 获取直接权限
	var directPermissions []*models.PermissionWithSource
	if role.Edges.RolePermissions != nil {
		for _, rp := range role.Edges.RolePermissions {
			if rp.Edges.Permission != nil {
				directPermissions = append(directPermissions, &models.PermissionWithSource{
					Permission: ConvertPermissionToResponse(rp.Edges.Permission),
					Source:     "direct",
					SourceRole: &models.RoleResponse{
						ID:          utils.Uint64ToString(role.ID),
						Name:        role.Name,
						Description: role.Description,
					},
				})
			}
		}
	}

	// 获取继承权限
	var inheritedPermissions []*models.PermissionWithSource
	if role.Edges.InheritsFrom != nil {
		for _, parentRole := range role.Edges.InheritsFrom {
			parentRoleWithPerms, err := database.Client.Role.Query().
				Where(entRole.ID(parentRole.ID)).
				WithRolePermissions(func(rp *ent.RolePermissionQuery) {
					rp.WithPermission()
				}).
				Only(ctx)
			if err != nil {
				continue
			}

			if parentRoleWithPerms.Edges.RolePermissions != nil {
				for _, rp := range parentRoleWithPerms.Edges.RolePermissions {
					if rp.Edges.Permission != nil {
						inheritedPermissions = append(inheritedPermissions, &models.PermissionWithSource{
							Permission: ConvertPermissionToResponse(rp.Edges.Permission),
							Source:     "inherit",
							SourceRole: &models.RoleResponse{
								ID:          utils.Uint64ToString(parentRole.ID),
								Name:        parentRole.Name,
								Description: parentRole.Description,
							},
						})
					}
				}
			}
		}
	}

	return &models.RoleDetailedPermissionsResponse{
		Role:                 ConvertRoleToResponse(role),
		DirectPermissions:    directPermissions,
		InheritedPermissions: inheritedPermissions,
	}, nil
}

// CreateChildRole 创建子角色
func CreateChildRole(ctx context.Context, parentID uint64, req *models.CreateChildRoleRequest) (*ent.Role, error) {
	// 检查父角色是否存在
	_, err := database.Client.Role.Query().
		Where(entRole.ID(parentID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("parent role not found")
		}
		return nil, err
	}

	// 创建子角色
	childRole, err := database.Client.Role.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		AddInheritsFromIDs(parentID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create child role: %v", err)
	}

	// 返回完整的角色信息
	return GetRoleByID(ctx, childRole.ID)
}

// RemoveParentRole 移除父角色继承关系
func RemoveParentRole(ctx context.Context, roleID, parentID uint64) error {
	// 检查角色是否存在
	exists, err := database.Client.Role.Query().
		Where(entRole.ID(roleID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role not found")
	}

	// 检查父角色是否存在
	parentExists, err := database.Client.Role.Query().
		Where(entRole.ID(parentID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !parentExists {
		return fmt.Errorf("parent role not found")
	}

	// 移除继承关系
	err = database.Client.Role.UpdateOneID(roleID).
		RemoveInheritsFromIDs(parentID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove parent role: %v", err)
	}

	return nil
}

// AddParentRole 添加父角色继承关系
func AddParentRole(ctx context.Context, roleID, parentID uint64) error {
	// 检查循环继承
	if err := HasCircularInheritance(ctx, database.Client, roleID, parentID); err != nil {
		return err
	}

	// 检查角色是否存在
	exists, err := database.Client.Role.Query().
		Where(entRole.ID(roleID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role not found")
	}

	// 检查父角色是否存在
	parentExists, err := database.Client.Role.Query().
		Where(entRole.ID(parentID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !parentExists {
		return fmt.Errorf("parent role not found")
	}

	// 添加继承关系
	err = database.Client.Role.UpdateOneID(roleID).
		AddInheritsFromIDs(parentID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add parent role: %v", err)
	}

	return nil
}

// GetAssignablePermissions 获取可分配的权限列表（排除已分配的权限）
func GetAssignablePermissions(ctx context.Context, roleID uint64) ([]*models.PermissionResponse, error) {
	// 获取所有权限
	allPermissions, err := database.Client.Permission.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %v", err)
	}

	// 获取角色已有的权限
	rolePermissions, err := GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 创建已分配权限的映射
	assignedPermissions := make(map[uint64]bool)
	for _, perm := range rolePermissions {
		assignedPermissions[perm.ID] = true
	}

	// 过滤出可分配的权限
	var assignablePermissions []*models.PermissionResponse
	for _, perm := range allPermissions {
		if !assignedPermissions[perm.ID] {
			assignablePermissions = append(assignablePermissions, ConvertPermissionToResponse(perm))
		}
	}

	return assignablePermissions, nil
}

// GetRoleUsersWithPagination 获取角色下的用户（分页）
func GetRoleUsersWithPagination(ctx context.Context, roleID uint64, req *models.GetRoleUsersRequest) (*models.RoleUsersResponse, error) {
	// 检查角色是否存在
	exists, err := database.Client.Role.Query().
		Where(entRole.ID(roleID)).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("role not found")
	}

	// 第一步：查询用户角色关联表，获取所有属于该角色的用户ID
	userRoles, err := database.Client.UserRole.Query().
		Where(userrole.RoleID(roleID)).
		WithRole(func(rq *ent.RoleQuery) {
			rq.Select(entRole.FieldID, entRole.FieldName)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 第二步：提取用户ID
	userIDs := make([]uint64, len(userRoles))
	for i, ur := range userRoles {
		userIDs[i] = ur.UserID
	}

	if len(userIDs) == 0 {
		// 如果没有用户，直接返回空结果
		return &models.RoleUsersResponse{
			Users: []*models.UserFromRoleResponse{},
			Pagination: models.Pagination{
				Page:       req.Page,
				PageSize:   req.PageSize,
				Total:      0,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			},
		}, nil
	}

	// 第三步：构建用户查询，使用IN过滤
	query := database.Client.User.Query().
		Where(user.IDIn(userIDs...))

	// 添加搜索条件
	if req.Keyword != "" {
		query = query.Where(user.NameContains(req.Keyword))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// 计算分页
	offset := (req.Page - 1) * req.PageSize
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	// 第四步：分页查询用户
	users, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 第五步：创建用户角色映射
	userRoleMap := make(map[uint64][]*ent.UserRole)
	for _, ur := range userRoles {
		userRoleMap[ur.UserID] = append(userRoleMap[ur.UserID], ur)
	}

	// 转换为响应格式
	userResponses := make([]*models.UserFromRoleResponse, len(users))
	for i, u := range users {
		var age *int
		if u.Age != 0 { // age为0表示未设置
			age = &u.Age
		}

		// 从映射中提取用户角色名称
		var roles []string = make([]string, 0)
		if userRoleList, exists := userRoleMap[u.ID]; exists {
			for _, userRole := range userRoleList {
				if userRole.Edges.Role != nil {
					roles = append(roles, userRole.Edges.Role.Name)
				}
			}
		}

		var roleList []models.RoleResponse = make([]models.RoleResponse, 0)

		foundRoles, err := GetUserRoles(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot found user-roles")
		}

		for _, oneRole := range foundRoles {
			if oneRole.ID != roleID {
				roleList = append(roleList, models.RoleResponse{
					ID:          utils.Uint64ToString(oneRole.ID),
					Name:        oneRole.Name,
					Description: oneRole.Description,
					CreateTime:  utils.FormatDateTime(oneRole.CreateTime),
					UpdateTime:  utils.FormatDateTime(oneRole.UpdateTime),
				})
			}
		}

		userResponses[i] = &models.UserFromRoleResponse{
			ID:         utils.Uint64ToString(u.ID),
			Name:       u.Name,
			Age:        age,
			Sex:        string(u.Sex),
			Status:     string(u.Status),
			CreateTime: utils.FormatDateTime(u.CreateTime),
			UpdateTime: utils.FormatDateTime(u.UpdateTime),
			Roles:      roles,
			OtherRoles: roleList,
		}
	}

	return &models.RoleUsersResponse{
		Users: userResponses,
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

// BatchAssignUsersToRole 批量分配用户到角色
func BatchAssignUsersToRole(ctx context.Context, roleID uint64, req *models.BatchAssignUsersToRoleRequest) error {
	// 检查角色是否存在
	exists, err := database.Client.Role.Query().
		Where(entRole.ID(roleID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role not found")
	}

	// 转换用户ID
	userIDs := make([]uint64, 0, len(req.UserIds))
	for _, userIDStr := range req.UserIds {
		userID := utils.StringToUint64(userIDStr)
		userIDs = append(userIDs, userID)
	}

	// 检查用户是否存在
	existingUsers, err := database.Client.User.Query().
		Where(user.IDIn(userIDs...)).
		All(ctx)
	if err != nil {
		return err
	}
	if len(existingUsers) != len(userIDs) {
		return fmt.Errorf("some users not found")
	}

	// 批量创建用户角色关联
	bulk := make([]*ent.UserRoleCreate, 0, len(userIDs))
	for _, userID := range userIDs {
		// 检查是否已经存在关联
		exists, err := database.Client.UserRole.Query().
			Where(
				userrole.UserID(userID),
				userrole.RoleID(roleID),
			).Exist(ctx)
		if err != nil {
			return err
		}
		if !exists {
			bulk = append(bulk, database.Client.UserRole.Create().
				SetUserID(userID).
				SetRoleID(roleID))
		}
	}

	if len(bulk) > 0 {
		_, err = database.Client.UserRole.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to assign users to role: %v", err)
		}
	}

	return nil
}

// BatchRemoveUsersFromRole 批量移除用户从角色
func BatchRemoveUsersFromRole(ctx context.Context, roleID uint64, req *models.BatchRemoveUsersFromRoleRequest) error {
	// 检查角色是否存在
	exists, err := database.Client.Role.Query().
		Where(entRole.ID(roleID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role not found")
	}

	// 转换用户ID
	userIDs := make([]uint64, 0, len(req.UserIds))
	for _, userIDStr := range req.UserIds {
		userID := utils.StringToUint64(userIDStr)
		userIDs = append(userIDs, userID)
	}

	// 删除用户角色关联
	_, err = database.Client.UserRole.Delete().
		Where(
			userrole.RoleID(roleID),
			userrole.UserIDIn(userIDs...),
		).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove users from role: %v", err)
	}

	return nil
}

// collectRolePermissions 递归收集角色权限（处理角色继承）
func collectRolePermissions(ctx context.Context, roleID uint64, visited map[uint64]bool, permissionMap map[uint64]*ent.Permission) error {
	// 防止循环继承
	if visited[roleID] {
		return nil
	}
	visited[roleID] = true

	// 获取当前角色的直接权限
	directPermissions, err := database.Client.Permission.Query().
		Where(permission.HasRolePermissionsWith(
			rolepermission.RoleID(roleID),
		)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get direct permissions for role %d: %w", roleID, err)
	}

	// 添加直接权限到映射中
	for _, perm := range directPermissions {
		permissionMap[perm.ID] = perm
	}

	// 获取当前角色继承的角色
	inheritedRoles, err := database.Client.Role.Query().
		Where(entRole.HasInheritedByWith(entRole.ID(roleID))).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inherited roles for role %d: %w", roleID, err)
	}

	// 递归处理继承的角色
	for _, inheritedRole := range inheritedRoles {
		err = collectRolePermissions(ctx, inheritedRole.ID, visited, permissionMap)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasAnyPermissions 检查用户是否拥有指定权限列表中的任何一个权限（支持角色继承）
func HasAnyPermissions(ctx context.Context, userID uint64, permissions []string) (bool, error) {
	if len(permissions) == 0 {
		return false, nil
	}

	// 获取用户所有权限
	userPermissions, err := GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// 检查是否有任何匹配的权限
	for _, userPerm := range userPermissions {
		for _, requiredPerm := range permissions {
			if userPerm.Name == requiredPerm || userPerm.Action == requiredPerm {
				return true, nil
			}
		}
	}

	return false, nil
}

// HasAnyPermissionsOptimized 更高效的版本：直接通过数据库查询检查权限
func HasAnyPermissionsOptimized(ctx context.Context, userID uint64, permissions []string) (bool, error) {
	if len(permissions) == 0 {
		return true, nil
	}

	// 检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(userID)).Exist(ctx)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("user not found")
	}

	// 获取用户的所有角色ID（包括继承的角色）
	roleIDs, err := getAllUserRoleIDs(ctx, userID)
	if err != nil {
		return false, err
	}

	if len(roleIDs) == 0 {
		return false, nil
	}

	// 检查这些角色是否有任何所需的权限
	count, err := database.Client.Permission.Query().
		Where(
			permission.Or(
				permission.NameIn(permissions...),
				permission.ActionIn(permissions...),
			),
		).
		Where(permission.HasRolePermissionsWith(
			rolepermission.RoleIDIn(roleIDs...),
		)).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// getAllUserRoleIDs 获取用户所有角色ID（包括继承的角色）
func getAllUserRoleIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	// 获取用户的直接角色
	directRoles, err := database.Client.Role.Query().
		Where(entRole.HasUserRolesWith(func(s *sql.Selector) {
			s.Where(sql.EQ("user_id", userID)).Where(sql.IsNull("delete_time"))
		})).All(ctx)
	if err != nil {
		return nil, err
	}

	roleIDSet := make(map[uint64]bool)
	visited := make(map[uint64]bool)

	// 收集所有角色ID（包括继承的）
	for _, role := range directRoles {
		err = collectInheritedRoleIDs(ctx, role.ID, visited, roleIDSet)
		if err != nil {
			return nil, err
		}
	}

	// 转换为切片
	roleIDs := make([]uint64, 0, len(roleIDSet))
	for roleID := range roleIDSet {
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, nil
}

// collectInheritedRoleIDs 递归收集角色ID
func collectInheritedRoleIDs(ctx context.Context, roleID uint64, visited map[uint64]bool, roleIDSet map[uint64]bool) error {
	if visited[roleID] {
		return nil
	}
	visited[roleID] = true
	roleIDSet[roleID] = true

	// 获取继承的角色
	inheritedRoles, err := database.Client.Role.Query().
		Where(entRole.HasInheritedByWith(entRole.ID(roleID))).
		All(ctx)
	if err != nil {
		return err
	}

	// 递归处理
	for _, inheritedRole := range inheritedRoles {
		err = collectInheritedRoleIDs(ctx, inheritedRole.ID, visited, roleIDSet)
		if err != nil {
			return err
		}
	}

	return nil
}
