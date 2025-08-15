package funcs

import (
	"context"
	"fmt"
	"math"
	"strings"

	"go-backend/database/ent"
	"go-backend/database/ent/user"
	"go-backend/pkg/database"
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
			ID:    int(u.ID),
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
