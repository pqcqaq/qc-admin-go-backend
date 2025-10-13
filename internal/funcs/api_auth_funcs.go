package funcs

import (
	"context"
	"fmt"
	"go-backend/database/ent"
	"go-backend/database/ent/apiauth"
	"go-backend/database/ent/permission"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

type ApiAuthFuncs struct{}

// GetAllAPIAuths 获取所有API认证记录
func (ApiAuthFuncs) GetAllAPIAuths(ctx context.Context) ([]*models.APIAuthResponse, error) {
	records, err := database.Client.APIAuth.Query().
		WithPermissions().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get api auths: %w", err)
	}
	var apiAuths []*models.APIAuthResponse
	for _, record := range records {
		apiAuthResponse := &models.APIAuthResponse{
			ID:          utils.Uint64ToString(record.ID),
			CreateTime:  utils.FormatDateTime(record.CreateTime),
			UpdateTime:  utils.FormatDateTime(record.UpdateTime),
			Type:        record.Type.String(),
			Name:        record.Name,
			Description: record.Description,
			Method:      record.Method,
			Path:        record.Path,
			IsPublic:    record.IsPublic,
			IsActive:    record.IsActive,
			Metadata:    record.Metadata,
		}

		apiAuths = append(apiAuths, apiAuthResponse)
	}
	return apiAuths, nil
}

// GetAPIAuthById 根据ID获取API认证记录
func (ApiAuthFuncs) GetAPIAuthById(ctx context.Context, id uint64) (*models.APIAuthResponse, error) {
	apiAuth, err := database.Client.APIAuth.Query().
		Where(apiauth.ID(id)).
		WithPermissions().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("api auth with id %d not found", id)
		}
		return nil, err
	}
	apiAuthResponse := &models.APIAuthResponse{
		ID:          utils.Uint64ToString(apiAuth.ID),
		CreateTime:  utils.FormatDateTime(apiAuth.CreateTime),
		UpdateTime:  utils.FormatDateTime(apiAuth.UpdateTime),
		Name:        apiAuth.Name,
		Type:        apiAuth.Type.String(),
		Description: apiAuth.Description,
		Method:      apiAuth.Method,
		Path:        apiAuth.Path,
		IsPublic:    apiAuth.IsPublic,
		IsActive:    apiAuth.IsActive,
		Metadata:    apiAuth.Metadata,
	}
	return apiAuthResponse, nil
}

// CreateAPIAuth 创建API认证记录
func (ApiAuthFuncs) CreateAPIAuth(ctx context.Context, req *models.CreateAPIAuthRequest) (*ent.APIAuth, error) {
	pIdsList := ApiAuthFuncs{}.permissionsListToIDs(req.Permissions)

	builder := database.Client.APIAuth.Create().
		SetName(req.Name).
		SetMethod(req.Method).
		SetPath(req.Path).
		SetIsPublic(*req.IsPublic).
		SetIsActive(*req.IsActive).
		SetType(apiauth.Type(req.Type)).
		AddPermissionIDs(pIdsList...)

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.Metadata != nil {
		builder = builder.SetMetadata(req.Metadata)
	}

	rec, err := builder.Save(ctx)
	// 更新缓存
	if rec.Type == apiauth.TypeWebsocket {
		updateCache(rec)
	}
	return rec, err
}

// UpdateAPIAuth 更新API认证记录
func (ApiAuthFuncs) UpdateAPIAuth(ctx context.Context, id uint64, req *models.UpdateAPIAuthRequest) (*ent.APIAuth, error) {
	// ent 的事务开启方法
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	// defer 里根据 err 判断提交还是回滚
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // 继续抛 panic
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 首先检查API认证记录是否存在
	exists, err := tx.APIAuth.Query().Where(apiauth.ID(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("api auth with id %d not found", id)
	}

	oldIdList, err := tx.APIAuth.Query().Where(apiauth.ID(id)).QueryPermissions().IDs(ctx)
	if err != nil {
		return nil, err
	}
	newIdsList := ApiAuthFuncs{}.permissionsListToIDs(req.Permissions)

	toRemove, toAdd := utils.DiffUint64Slices(oldIdList, newIdsList)

	builder := tx.APIAuth.UpdateOneID(id).
		SetName(req.Name).
		SetMethod(req.Method).
		SetPath(req.Path).
		SetIsPublic(*req.IsPublic).
		SetIsActive(*req.IsActive).
		SetType(apiauth.Type(req.Type)).
		RemovePermissionIDs(toRemove...).
		AddPermissionIDs(toAdd...)

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.Metadata != nil {
		builder = builder.SetMetadata(req.Metadata)
	}

	rec, err := builder.Save(ctx)

	// 更新缓存
	updateCache(rec)

	return rec, err
}

func (ApiAuthFuncs) permissionsListToIDs(permissionsList []*models.PermissionsList) []uint64 {
	var ids []uint64 = make([]uint64, 0)
	for _, perm := range permissionsList {
		id := utils.StringToUint64(perm.ID)
		ids = append(ids, id)
	}
	return ids
}

// DeleteAPIAuth 删除API认证记录
func (ApiAuthFuncs) DeleteAPIAuth(ctx context.Context, id uint64) error {
	// 首先检查API认证记录是否存在
	exists, err := database.Client.APIAuth.Query().Where(apiauth.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("api auth with id %d not found", id)
	}

	deleteByKey(id)

	return database.Client.APIAuth.DeleteOneID(id).Exec(ctx)
}

func (ApiAuthFuncs) GetAPIAuthWithPagination(ctx context.Context, req *models.PageAPIAuthRequest) (*models.PageAPIAuthResponse, error) {
	// 构建查询
	query := database.Client.APIAuth.Query().
		WithPermissions()

	startAt, err := utils.ParseTime(req.BeginTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}
	endAt, err := utils.ParseTime(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	// 是否公开
	if req.IsPublic != nil {
		query = query.Where(apiauth.IsPublicEQ(*req.IsPublic))
	}

	// 是否启用
	if req.IsActive != nil {
		query = query.Where(apiauth.IsActiveEQ(*req.IsActive))
	}

	// 时间
	if !(startAt.IsZero() || endAt.IsZero()) {
		query = query.Where(
			apiauth.CreateTimeGTE(startAt),
			apiauth.CreateTimeLTE(endAt),
		)
	}

	// 排序
	switch req.OrderBy {
	case "create_time":
		if req.Order == "asc" {
			query = query.Order(ent.Asc("create_time"))
		} else {
			query = query.Order(ent.Desc("create_time"))
		}
	case "update_time":
		if req.Order == "asc" {
			query = query.Order(ent.Asc("update_time"))
		} else {
			query = query.Order(ent.Desc("update_time"))
		}
	case "name":
		if req.Order == "asc" {
			query = query.Order(ent.Asc("name"))
		} else {
			query = query.Order(ent.Desc("name"))
		}
	default:
		query = query.Order(ent.Desc("create_time"))
	}

	// type过滤
	if req.Type != "" {
		query = query.Where(apiauth.TypeEQ(apiauth.Type(req.Type)))
	}

	// 模糊搜索
	if req.Name != "" {
		query = query.Where(apiauth.NameContains(req.Name))
	}

	if req.Method != "" {
		query = query.Where(apiauth.MethodEQ(req.Method))
	}

	if req.Path != "" {
		query = query.Where(apiauth.PathContains(req.Path))
	}

	// 分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count api auths: %w", err)
	}

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	query.WithPermissions(func(pq *ent.PermissionQuery) {
		pq.Select(permission.FieldAction, permission.FieldName, permission.FieldID)
	})
	query.Order(ent.Desc("create_time"))

	apiAuths, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get api auths: %w", err)
	}

	var apiAuthResponses []*models.APIAuthResponse
	for _, apiAuth := range apiAuths {

		var permissions []*models.PermissionResponse = make([]*models.PermissionResponse, 0)

		for _, perm := range apiAuth.Edges.Permissions {
			permissions = append(permissions, &models.PermissionResponse{
				ID:     utils.Uint64ToString(perm.ID),
				Name:   perm.Name,
				Action: perm.Action,
			})
		}

		apiAuthResponse := &models.APIAuthResponse{
			ID:          utils.Uint64ToString(apiAuth.ID),
			CreateTime:  utils.FormatDateTime(apiAuth.CreateTime),
			UpdateTime:  utils.FormatDateTime(apiAuth.UpdateTime),
			Name:        apiAuth.Name,
			Type:        apiAuth.Type.String(),
			Description: apiAuth.Description,
			Method:      apiAuth.Method,
			Path:        apiAuth.Path,
			IsPublic:    apiAuth.IsPublic,
			IsActive:    apiAuth.IsActive,
			Metadata:    apiAuth.Metadata,
			Permissions: permissions,
		}

		apiAuthResponses = append(apiAuthResponses, apiAuthResponse)
	}

	if apiAuthResponses == nil {
		apiAuthResponses = make([]*models.APIAuthResponse, 0)
	}

	return &models.PageAPIAuthResponse{
		Data: apiAuthResponses,
		Pagination: models.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      int64(total),
			TotalPages: (total + req.PageSize - 1) / req.PageSize, // 向上取整
			HasNext:    req.Page < (total+req.PageSize-1)/req.PageSize,
			HasPrev:    req.Page > 1,
		},
	}, nil
}
