package funcs

import (
	"context"
	"fmt"

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
func GetUserByID(ctx context.Context, id int64) (*ent.User, error) {
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
func UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*ent.User, error) {
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
func DeleteUser(ctx context.Context, id int64) error {
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
