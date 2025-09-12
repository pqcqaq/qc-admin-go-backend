package funcs

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-backend/database/ent"
	"go-backend/database/ent/credential"
	"go-backend/database/ent/permission"
	"go-backend/database/ent/role"
	"go-backend/database/ent/user"
	"go-backend/database/ent/userrole"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// UserService 用户服务

// GetAllUsers 获取所有用户
func GetAllUsers(ctx context.Context) ([]*models.UserResponse, error) {
	list, err := database.Client.User.Query().All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	// 格式转换
	var res []*models.UserResponse
	for _, item := range list {
		res = append(res, &models.UserResponse{
			ID:     utils.Uint64ToString(item.ID),
			Name:   item.Name,
			Age:    &item.Age,
			Sex:    utils.ToString(item.Sex),
			Status: utils.ToString(item.Status),
		})
	}
	return res, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(ctx context.Context, id uint64) (*ent.User, error) {
	user, err := database.Client.User.Query().Where(user.ID(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}
	return user, nil
}

// CreateUser 创建用户
func CreateUser(ctx context.Context, req *models.CreateUserRequest) (*ent.User, error) {
	builder := database.Client.User.Create().
		SetName(req.Name).
		SetAge(*req.Age).
		SetSex(user.Sex(req.Sex)).
		SetStatus(user.Status(req.Status))

	if req.Age != nil {
		builder = builder.SetAge(*req.Age)
	}

	return builder.Save(ctx)
}

// UpdateUser 更新用户
func UpdateUser(ctx context.Context, id uint64, req *models.UpdateUserRequest) (*ent.User, error) {
	// 首先检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user with id %d not found", id)
	}

	builder := database.Client.User.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}
	if req.Age != nil {
		builder = builder.SetAge(*req.Age)
	}
	// Sex
	if req.Sex != "" {
		builder = builder.SetSex(user.Sex(req.Sex))
	}
	// status
	if req.Status != "" {
		builder = builder.SetStatus(user.Status(req.Status))
	}

	return builder.Save(ctx)
}

// DeleteUser 删除用户
func DeleteUser(ctx context.Context, id uint64) error {
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return err
	}
	// 首先检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user with id %d not found", id)
	}

	// 需要先删除userRole关联
	_, err = tx.UserRole.Delete().Where(userrole.UserID(id)).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 然后删除user_credential关联
	_, err = tx.Credential.Delete().Where(credential.HasUserWith(user.ID(id))).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.User.DeleteOneID(id).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务，检查提交错误
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

// GetUsersWithPagination 分页获取用户列表
func GetUsersWithPagination(ctx context.Context, req *models.GetUsersRequest) (*models.UsersListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.OrderBy == "" {
		req.OrderBy = "create_time"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	// 构建查询条件
	query := database.Client.User.Query()

	// 添加搜索条件
	if req.Name != "" {
		query = query.Where(user.NameContains(req.Name))
	}

	// sex
	if req.Sex != "" {
		query = query.Where(user.SexEQ(user.Sex(req.Sex)))
	}

	// status
	if req.Status != "" {
		query = query.Where(user.StatusEQ(user.Status(req.Status)))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 计算分页信息
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	offset := (req.Page - 1) * req.PageSize

	// 添加排序和分页
	switch strings.ToLower(req.OrderBy) {
	case "id":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(user.FieldID))
		} else {
			query = query.Order(ent.Asc(user.FieldID))
		}
	case "name":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(user.FieldName))
		} else {
			query = query.Order(ent.Asc(user.FieldName))
		}
	case "age":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(user.FieldAge))
		} else {
			query = query.Order(ent.Asc(user.FieldAge))
		}
	case "create_time", "created_at":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(user.FieldCreateTime))
		} else {
			query = query.Order(ent.Asc(user.FieldCreateTime))
		}
	case "update_time", "updated_at":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(user.FieldUpdateTime))
		} else {
			query = query.Order(ent.Asc(user.FieldUpdateTime))
		}
	default:
		// 默认按创建时间降序排列
		query = query.Order(ent.Desc(user.FieldCreateTime))
	}

	// 执行分页查询
	users, err := query.
		Offset(offset).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	// 转换为响应格式
	userResponses := make([]*models.UserResponse, len(users))
	for i, u := range users {
		var age *int
		if u.Age != 0 { // age为0表示未设置
			age = &u.Age
		}

		userResponses[i] = &models.UserResponse{
			ID:         utils.Uint64ToString(u.ID),
			Name:       u.Name,
			Age:        age,
			Sex:        string(u.Sex),
			Status:     string(u.Status),
			CreateTime: utils.FormatDateTime(u.CreateTime),
			UpdateTime: utils.FormatDateTime(u.UpdateTime),
		}
	}

	// 构建分页信息
	pagination := models.Pagination{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      int64(total),
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &models.UsersListResponse{
		Data:       userResponses,
		Pagination: pagination,
	}, nil
}

// AssignUserRole 为用户分配角色
func AssignUserRole(ctx context.Context, req *models.AssignUserRoleRequest) (*ent.UserRole, error) {
	// 检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(utils.StringToUint64(req.UserID))).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// 检查角色是否存在
	exists, err = database.Client.Role.Query().Where(role.ID(utils.StringToUint64(req.RoleID))).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("role not found")
	}

	// 检查是否已经分配了该角色
	exists, err = database.Client.UserRole.Query().
		Where(
			userrole.UserID(utils.StringToUint64(req.UserID)),
			userrole.RoleID(utils.StringToUint64(req.RoleID)),
		).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user role already exists")
	}

	// 创建用户角色关联
	userRole, err := database.Client.UserRole.Create().
		SetUserID(utils.StringToUint64(req.UserID)).
		SetRoleID(utils.StringToUint64(req.RoleID)).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 加载关联数据
	userRole, err = database.Client.UserRole.Query().
		Where(userrole.ID(userRole.ID)).
		WithUser().
		WithRole().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return userRole, nil
}

// RevokeUserRole 撤销用户角色
func RevokeUserRole(ctx context.Context, userID, roleID uint64) error {
	// 查找用户角色关联
	userRole, err := database.Client.UserRole.Query().
		Where(
			userrole.UserID(userID),
			userrole.RoleID(roleID),
		).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("user role not found")
		}
		return err
	}

	// 删除关联
	err = database.Client.UserRole.DeleteOne(userRole).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetUserRoles 获取用户的所有角色
func GetUserRoles(ctx context.Context, userID uint64) ([]*ent.Role, error) {
	// 检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(userID)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// 获取用户的所有角色
	roles, err := database.Client.Role.Query().
		Where(role.HasUserRolesWith(userrole.UserID(userID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRoleUsers 获取拥有指定角色的所有用户
func GetRoleUsers(ctx context.Context, roleID uint64) ([]*ent.User, error) {
	// 检查角色是否存在
	exists, err := database.Client.Role.Query().Where(role.ID(roleID)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("role not found")
	}

	// 获取拥有该角色的所有用户
	users, err := database.Client.User.Query().
		Where(user.HasUserRolesWith(userrole.RoleID(roleID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserPermissions 获取用户的所有权限（通过角色继承）
func GetUserPermissions(ctx context.Context, userID uint64) ([]*ent.Permission, error) {
	// 检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(userID)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// 获取用户的所有角色
	userRoles, err := database.Client.Role.Query().
		Where(role.HasUserRolesWith(userrole.UserID(userID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []*ent.Permission{}, nil
	}

	// 获取所有角色的权限（包括继承的权限）
	permissionMap := make(map[uint64]*ent.Permission)

	for _, userRole := range userRoles {
		// 获取角色及其继承链的所有权限
		rolePermissions, err := GetRoleInheritedPermissions(ctx, userRole.ID)
		if err != nil {
			return nil, err
		}

		// 添加到权限映射中（去重）
		for _, perm := range rolePermissions {
			permissionMap[perm.ID] = perm
		}
	}

	// 转换为切片
	permissions := make([]*ent.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CheckUserPermission 检查用户是否拥有指定权限
func CheckUserPermission(ctx context.Context, userID, permissionID uint64) (bool, error) {
	// 检查权限是否存在
	exists, err := database.Client.Permission.Query().Where(permission.ID(permissionID)).Exist(ctx)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	// 获取用户的所有权限
	userPermissions, err := GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// 检查是否包含指定权限
	for _, perm := range userPermissions {
		if perm.ID == permissionID {
			return true, nil
		}
	}

	return false, nil
}

// ConvertUserRoleToResponse 将UserRole实体转换为响应格式
func ConvertUserRoleToResponse(userRole *ent.UserRole) *models.UserRoleResponse {
	response := &models.UserRoleResponse{
		ID:         strconv.FormatUint(userRole.ID, 10),
		UserID:     utils.ToString(userRole.UserID),
		RoleID:     utils.ToString(userRole.RoleID),
		CreateTime: utils.FormatDateTime(userRole.CreateTime),
		UpdateTime: utils.FormatDateTime(userRole.UpdateTime),
	}

	// 如果加载了用户关联数据
	if userRole.Edges.User != nil {
		response.User = ConvertUserToResponse(userRole.Edges.User)
	}

	// 如果加载了角色关联数据
	if userRole.Edges.Role != nil {
		response.Role = ConvertRoleToResponse(userRole.Edges.Role)
	}

	return response
}

// GetUserRolesWithPagination 分页获取用户角色关联列表
func GetUserRolesWithPagination(ctx context.Context, req *models.GetUserRolesRequest) (*models.UserRolesListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.OrderBy == "" {
		req.OrderBy = "create_time"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	// 构建查询条件
	query := database.Client.UserRole.Query().
		WithUser().
		WithRole()

	// 添加搜索条件
	if req.UserId != "" {
		userId := utils.StringToUint64(req.UserId)
		query = query.Where(userrole.UserID(userId))
	}

	if req.RoleId != "" {
		roleId := utils.StringToUint64(req.RoleId)
		query = query.Where(userrole.RoleID(roleId))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count user roles: %w", err)
	}

	// 计算分页信息
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	offset := (req.Page - 1) * req.PageSize

	// 添加排序和分页
	switch strings.ToLower(req.OrderBy) {
	case "id":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(userrole.FieldID))
		} else {
			query = query.Order(ent.Asc(userrole.FieldID))
		}
	case "user_id":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(userrole.FieldUserID))
		} else {
			query = query.Order(ent.Asc(userrole.FieldUserID))
		}
	case "role_id":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(userrole.FieldRoleID))
		} else {
			query = query.Order(ent.Asc(userrole.FieldRoleID))
		}
	case "create_time", "created_at":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(userrole.FieldCreateTime))
		} else {
			query = query.Order(ent.Asc(userrole.FieldCreateTime))
		}
	case "update_time", "updated_at":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(userrole.FieldUpdateTime))
		} else {
			query = query.Order(ent.Asc(userrole.FieldUpdateTime))
		}
	default:
		// 默认按创建时间降序排列
		query = query.Order(ent.Desc(userrole.FieldCreateTime))
	}

	// 执行分页查询
	userRoles, err := query.
		Offset(offset).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 转换为响应格式
	userRoleResponses := make([]*models.UserRoleResponse, len(userRoles))
	for i, ur := range userRoles {
		userRoleResponses[i] = ConvertUserRoleToResponse(ur)
	}

	// 构建分页信息
	pagination := models.Pagination{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      int64(total),
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &models.UserRolesListResponse{
		Data:       userRoleResponses,
		Pagination: pagination,
	}, nil
}

// ConvertUserToResponse 将User实体转换为响应格式
func ConvertUserToResponse(user *ent.User) *models.UserResponse {
	return &models.UserResponse{
		ID:         utils.Uint64ToString(user.ID),
		Name:       user.Name, // 假设Name字段对应Username
		Sex:        string(user.Sex),
		Status:     string(user.Status),
		CreateTime: utils.FormatDateTime(user.CreateTime),
		UpdateTime: utils.FormatDateTime(user.UpdateTime),
	}
}

// GetUserMenuTree 获取用户的菜单树
func GetUserMenuTree(ctx context.Context, userID uint64) ([]*models.ScopeResponse, error) {
	// 1. 获取用户的所有角色
	userRoles, err := database.Client.UserRole.Query().
		Where(userrole.UserID(userID)).
		WithRole().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 2. 收集所有角色ID
	roleIDs := make([]uint64, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	// 如果用户没有角色，返回空树
	if len(roleIDs) == 0 {
		return []*models.ScopeResponse{}, nil
	}

	// 3. 获取这些角色的所有权限（包括继承的权限）
	var allPermissions []*ent.Permission
	for _, roleID := range roleIDs {
		rolePermissions, err := GetRoleInheritedPermissions(ctx, roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to query role permissions for role %d: %w", roleID, err)
		}
		allPermissions = append(allPermissions, rolePermissions...)
	}

	// 4. 去重权限并提取scope ID
	permissionMap := make(map[uint64]*ent.Permission)
	scopeIDs := make(map[uint64]bool)

	for _, perm := range allPermissions {
		permissionMap[perm.ID] = perm
	}

	// 获取权限对应的scope信息，现在是Permission->Scope的关系
	for permID := range permissionMap {
		perm, err := database.Client.Permission.Query().
			Where(permission.ID(permID)).
			WithScope().
			Only(ctx)
		if err != nil {
			continue // 忽略查询错误的权限
		}
		if perm.Edges.Scope != nil {
			scopeIDs[perm.Edges.Scope.ID] = true
		}
	}

	// 5. 获取所有相关的scope，包括父级scope
	allScopes, err := database.Client.Scope.Query().
		Order(ent.Asc("order")).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query scopes: %w", err)
	}

	// 6. 过滤出用户有权限的scope和它们的父级
	accessibleScopes := make(map[uint64]*ent.Scope)

	// 首先添加用户直接有权限的scope
	for _, scope := range allScopes {
		if scopeIDs[scope.ID] {
			accessibleScopes[scope.ID] = scope
		}
	}

	// 然后添加这些scope的所有父级（确保菜单路径完整）
	for scopeID := range scopeIDs {
		scope := findScopeByID(allScopes, scopeID)
		if scope != nil {
			addParentScopes(scope, allScopes, accessibleScopes)
		}
	}

	// 7. 过滤掉隐藏和禁用的scope，但保留按钮类型的权限
	filteredScopes := make(map[uint64]*ent.Scope)
	for id, scope := range accessibleScopes {
		// 只显示启用且非隐藏的菜单和页面，或者是按钮类型
		if (!scope.Hidden && !scope.Disabled) || scope.Type == "button" {
			filteredScopes[id] = scope
		}
	}

	// 8. 构建树形结构
	scopeMap := make(map[uint64]*models.ScopeResponse)
	var rootScopes []*models.ScopeResponse

	// 创建所有节点
	for _, scope := range filteredScopes {
		scopeResp := ConvertScopeToResponseForTree(scope)
		scopeMap[scope.ID] = scopeResp

		// 如果没有父节点或父节点不在可访问列表中，则为根节点
		if scope.ParentID == 0 || filteredScopes[scope.ParentID] == nil {
			// 只有非按钮类型才能作为根节点显示
			if scope.Type != "button" {
				rootScopes = append(rootScopes, scopeResp)
			}
		}
	}

	// 构建父子关系
	for _, scope := range filteredScopes {
		if scope.ParentID != 0 && filteredScopes[scope.ParentID] != nil {
			parent := scopeMap[scope.ParentID]
			child := scopeMap[scope.ID]
			if parent != nil && child != nil {
				// 只有非按钮类型才添加到树中
				// if scope.Type != "button" {
				if parent.Children == nil {
					parent.Children = make([]*models.ScopeResponse, 0)
				}
				parent.Children = append(parent.Children, child)
				// }
			}
		}
	}

	return rootScopes, nil
}

// findScopeByID 根据ID查找scope
func findScopeByID(scopes []*ent.Scope, id uint64) *ent.Scope {
	for _, scope := range scopes {
		if scope.ID == id {
			return scope
		}
	}
	return nil
}

// addParentScopes 递归添加父级scope
func addParentScopes(scope *ent.Scope, allScopes []*ent.Scope, accessibleScopes map[uint64]*ent.Scope) {
	if scope.ParentID != 0 {
		parent := findScopeByID(allScopes, scope.ParentID)
		if parent != nil {
			accessibleScopes[parent.ID] = parent
			addParentScopes(parent, allScopes, accessibleScopes)
		}
	}
}
