package funcs

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-backend/database/ent"
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
func GetAllUsers(ctx context.Context) ([]*ent.User, error) {
	return database.Client.User.Query().All(ctx)
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
		SetEmail(req.Email)

	if req.Age != nil {
		builder = builder.SetAge(*req.Age)
	}

	if req.Phone != "" {
		builder = builder.SetPhone(req.Phone)
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
	if req.Email != "" {
		builder = builder.SetEmail(req.Email)
	}
	if req.Age != nil {
		builder = builder.SetAge(*req.Age)
	}
	if req.Phone != "" {
		builder = builder.SetPhone(req.Phone)
	}

	return builder.Save(ctx)
}

// DeleteUser 删除用户
func DeleteUser(ctx context.Context, id uint64) error {
	// 首先检查用户是否存在
	exists, err := database.Client.User.Query().Where(user.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user with id %d not found", id)
	}

	return database.Client.User.DeleteOneID(id).Exec(ctx)
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
	if req.Email != "" {
		query = query.Where(user.EmailContains(req.Email))
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
	case "email":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(user.FieldEmail))
		} else {
			query = query.Order(ent.Asc(user.FieldEmail))
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
			ID:    utils.Uint64ToString(u.ID),
			Name:  u.Name,
			Email: u.Email,
			Age:   age,
			Phone: u.Phone,
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
	exists, err := database.Client.User.Query().Where(user.ID(req.UserID)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// 检查角色是否存在
	exists, err = database.Client.Role.Query().Where(role.ID(req.RoleID)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("role not found")
	}

	// 检查是否已经分配了该角色
	exists, err = database.Client.UserRole.Query().
		Where(
			userrole.UserID(req.UserID),
			userrole.RoleID(req.RoleID),
		).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user role already exists")
	}

	// 创建用户角色关联
	userRole, err := database.Client.UserRole.Create().
		SetUserID(req.UserID).
		SetRoleID(req.RoleID).
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
		UserID:     userRole.UserID,
		RoleID:     userRole.RoleID,
		CreateTime: utils.JSONTime(userRole.CreateTime),
		UpdateTime: utils.JSONTime(userRole.UpdateTime),
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

// ConvertUserToResponse 将User实体转换为响应格式
func ConvertUserToResponse(user *ent.User) *models.UserResponse {
	return &models.UserResponse{
		ID:         utils.Uint64ToString(user.ID),
		Name:       user.Name, // 假设Name字段对应Username
		Email:      user.Email,
		CreateTime: utils.TimeToDateTimeString(&user.CreateTime),
		UpdateTime: utils.TimeToDateTimeString(&user.UpdateTime),
	}
}
