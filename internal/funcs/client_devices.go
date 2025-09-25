package funcs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go-backend/database/ent"
	"go-backend/database/ent/clientdevice"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// generateClientCode 生成客户端设备code
func generateClientCode() (string, error) {
	bytes := make([]byte, 32) // 32字节 = 64字符的十六进制
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetAllClientDevices 获取所有客户端设备
func GetAllClientDevices(ctx context.Context) ([]*models.ClientDeviceResponse, error) {
	records, err := database.Client.ClientDevice.Query().
		WithRoles().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client devices: %w", err)
	}

	var devices []*models.ClientDeviceResponse
	for _, record := range records {
		deviceResponse := &models.ClientDeviceResponse{
			ID:                 utils.Uint64ToString(record.ID),
			CreateTime:         utils.FormatDateTime(record.CreateTime),
			UpdateTime:         utils.FormatDateTime(record.UpdateTime),
			Name:               record.Name,
			Code:               record.Code,
			Description:        record.Description,
			Enabled:            record.Enabled,
			AccessTokenExpiry:  record.AccessTokenExpiry,
			RefreshTokenExpiry: record.RefreshTokenExpiry,
			Anonymous:          record.Anonymous,
		}

		// 处理关联的角色
		if record.Edges.Roles != nil {
			for _, role := range record.Edges.Roles {
				deviceResponse.Roles = append(deviceResponse.Roles, models.RoleInfo{
					ID:          utils.Uint64ToString(role.ID),
					Name:        role.Name,
					Description: role.Description,
				})
			}
		}

		devices = append(devices, deviceResponse)
	}
	return devices, nil
}

func GetClientDeviceByIdInner(ctx context.Context, id uint64) (*ent.ClientDevice, error) {
	device, err := database.Client.ClientDevice.Query().
		Where(clientdevice.ID(id)).
		WithRoles().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("client device with id %d not found", id)
		}
		return nil, err
	}
	return device, nil
}

// GetClientDeviceById 根据ID获取客户端设备
func GetClientDeviceById(ctx context.Context, id uint64) (*models.ClientDeviceResponse, error) {
	device, err := GetClientDeviceByIdInner(ctx, id)

	if err != nil {
		return nil, err
	}

	deviceResponse := &models.ClientDeviceResponse{
		ID:                 utils.Uint64ToString(device.ID),
		CreateTime:         utils.FormatDateTime(device.CreateTime),
		UpdateTime:         utils.FormatDateTime(device.UpdateTime),
		Name:               device.Name,
		Code:               device.Code,
		Description:        device.Description,
		Enabled:            device.Enabled,
		AccessTokenExpiry:  device.AccessTokenExpiry,
		RefreshTokenExpiry: device.RefreshTokenExpiry,
		Anonymous:          device.Anonymous,
	}

	// 处理关联的角色
	if device.Edges.Roles != nil {
		for _, role := range device.Edges.Roles {
			deviceResponse.Roles = append(deviceResponse.Roles, models.RoleInfo{
				ID:          utils.Uint64ToString(role.ID),
				Name:        role.Name,
				Description: role.Description,
			})
		}
	}

	return deviceResponse, nil
}

// GetClientDeviceByCode 根据code获取客户端设备
func GetClientDeviceByCodeInner(ctx context.Context, code string) (*ent.ClientDevice, error) {
	device, err := database.Client.ClientDevice.Query().
		Where(clientdevice.Code(code)).
		WithRoles().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("client device with code %s not found", code)
		}
		return nil, err
	}

	return device, nil
}

// GetClientDeviceByCode 根据code获取客户端设备
func GetClientDeviceByCode(ctx context.Context, code string) (*models.ClientDeviceByCodeResponse, error) {
	device, err := GetClientDeviceByCodeInner(ctx, code)

	if err != nil {
		return nil, err
	}

	deviceResponse := &models.ClientDeviceByCodeResponse{
		ID:                 utils.Uint64ToString(device.ID),
		Name:               device.Name,
		Code:               device.Code,
		Enabled:            device.Enabled,
		AccessTokenExpiry:  device.AccessTokenExpiry,
		RefreshTokenExpiry: device.RefreshTokenExpiry,
		Anonymous:          device.Anonymous,
	}

	// 处理关联的角色
	if device.Edges.Roles != nil {
		for _, role := range device.Edges.Roles {
			deviceResponse.Roles = append(deviceResponse.Roles, models.RoleInfo{
				ID:          utils.Uint64ToString(role.ID),
				Name:        role.Name,
				Description: role.Description,
			})
		}
	}

	return deviceResponse, nil
}

// CreateClientDevice 创建客户端设备
func CreateClientDevice(ctx context.Context, req *models.CreateClientDeviceRequest) (*ent.ClientDevice, error) {
	// 生成唯一的code
	code, err := generateClientCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client code: %w", err)
	}

	// 检查code是否已存在（虽然概率很小，但还是要检查）
	exists, err := database.Client.ClientDevice.Query().
		Where(clientdevice.Code(code)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check code existence: %w", err)
	}
	if exists {
		// 如果存在，重新生成
		return CreateClientDevice(ctx, req)
	}

	// 开始事务
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 设置默认值
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	anonymous := true
	if req.Anonymous != nil {
		anonymous = *req.Anonymous
	}

	desc := ""
	if utils.IsNotEmpty(req.Description) {
		desc = req.Description
	}

	// 创建客户端设备
	builder := tx.ClientDevice.Create().
		SetName(req.Name).
		SetCode(code).
		SetEnabled(enabled).
		SetDescription(desc).
		SetAccessTokenExpiry(req.AccessTokenExpiry).
		SetRefreshTokenExpiry(req.RefreshTokenExpiry).
		SetAnonymous(anonymous)

	// 如果指定了角色ID，添加角色关联
	if len(req.RoleIds) > 0 {
		var roleIds []uint64
		for _, roleIdStr := range req.RoleIds {
			roleId := utils.StringToUint64(roleIdStr)
			roleIds = append(roleIds, roleId)
		}
		builder = builder.AddRoleIDs(roleIds...)
	}

	return builder.Save(ctx)
}

// UpdateClientDevice 更新客户端设备
func UpdateClientDevice(ctx context.Context, id uint64, req *models.UpdateClientDeviceRequest) (*ent.ClientDevice, error) {
	// 开始事务
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 检查设备是否存在
	exists, err := tx.ClientDevice.Query().Where(clientdevice.ID(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("client device with id %d not found", id)
	}

	// 更新设备基本信息
	// 更新设备基本信息
	builder := tx.ClientDevice.UpdateOneID(id)

	// 只有非空字符串才设置 Name
	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	// 只有非 nil 指针才设置 Enabled
	if req.Enabled != nil {
		builder = builder.SetEnabled(*req.Enabled)
	}

	// 只有非空字符串才设置 Description
	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	// 只有非零值才设置 AccessTokenExpiry
	if req.AccessTokenExpiry != nil {
		builder = builder.SetAccessTokenExpiry(*req.AccessTokenExpiry)
	}

	// 只有非零值才设置 RefreshTokenExpiry
	if req.RefreshTokenExpiry != nil {
		builder = builder.SetRefreshTokenExpiry(*req.RefreshTokenExpiry)
	}

	// 只有非 nil 指针才设置 Anonymous
	if req.Anonymous != nil {
		builder = builder.SetAnonymous(*req.Anonymous)
	}

	oldRoles, err := tx.ClientDevice.Query().Where(clientdevice.IDEQ(id)).QueryRoles().All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			oldRoles = make([]*ent.Role, 0)
		} else {
			return nil, fmt.Errorf("failed to query roles for clientdevice: %d", id)
		}
	}
	var oldRoleIds []uint64 = make([]uint64, 0)

	for _, role := range oldRoles {
		oldRoleIds = append(oldRoleIds, role.ID)
	}

	roleIds := utils.StringToUint64Slice(req.RoleIds)

	toRemove, toAdd := utils.DiffUint64Slices(oldRoleIds, roleIds)

	builder.RemoveRoleIDs(toRemove...).AddRoleIDs(toAdd...)

	return builder.Save(ctx)
}

// DeleteClientDevice 删除客户端设备
func DeleteClientDevice(ctx context.Context, id uint64) error {
	// 检查设备是否存在
	exists, err := database.Client.ClientDevice.Query().Where(clientdevice.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("client device with id %d not found", id)
	}

	return database.Client.ClientDevice.DeleteOneID(id).Exec(ctx)
}

// GetClientDevicesWithPagination 分页获取客户端设备
func GetClientDevicesWithPagination(ctx context.Context, req *models.PageClientDevicesRequest) (*models.PageClientDevicesResponse, error) {
	// 构建查询
	query := database.Client.ClientDevice.Query().
		WithRoles()

	// 时间过滤
	startAt, err := utils.ParseTime(req.BeginTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}
	endAt, err := utils.ParseTime(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	if !(startAt.IsZero() || endAt.IsZero()) {
		query = query.Where(
			clientdevice.CreateTimeGTE(startAt),
			clientdevice.CreateTimeLTE(endAt),
		)
	}

	// 启用状态过滤
	if req.Enabled != nil {
		query = query.Where(clientdevice.EnabledEQ(*req.Enabled))
	}

	// 匿名登录过滤
	if req.Anonymous != nil {
		query = query.Where(clientdevice.AnonymousEQ(*req.Anonymous))
	}

	// 名称模糊搜索
	if req.Name != "" {
		query = query.Where(clientdevice.NameContains(req.Name))
	}

	// code精确搜索
	if req.Code != "" {
		query = query.Where(clientdevice.CodeEQ(req.Code))
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

	// 分页参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count client devices: %w", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	devices, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client devices: %w", err)
	}

	var deviceResponses []*models.ClientDeviceResponse
	for _, device := range devices {
		deviceResponse := &models.ClientDeviceResponse{
			ID:                 utils.Uint64ToString(device.ID),
			CreateTime:         utils.FormatDateTime(device.CreateTime),
			UpdateTime:         utils.FormatDateTime(device.UpdateTime),
			Name:               device.Name,
			Code:               device.Code,
			Enabled:            device.Enabled,
			Description:        device.Description,
			AccessTokenExpiry:  device.AccessTokenExpiry,
			RefreshTokenExpiry: device.RefreshTokenExpiry,
			Anonymous:          device.Anonymous,
		}

		// 处理关联的角色
		if device.Edges.Roles != nil {
			for _, role := range device.Edges.Roles {
				deviceResponse.Roles = append(deviceResponse.Roles, models.RoleInfo{
					ID:          utils.Uint64ToString(role.ID),
					Name:        role.Name,
					Description: role.Description,
				})
			}
		}

		deviceResponses = append(deviceResponses, deviceResponse)
	}

	if deviceResponses == nil {
		deviceResponses = make([]*models.ClientDeviceResponse, 0)
	}

	return &models.PageClientDevicesResponse{
		Data: deviceResponses,
		Pagination: models.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      int64(total),
			TotalPages: (total + req.PageSize - 1) / req.PageSize,
			HasNext:    req.Page < (total+req.PageSize-1)/req.PageSize,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

// CheckClientAccess 检查用户是否能使用指定客户端登录
func CheckClientAccess(ctx context.Context, req *models.CheckClientAccessRequest) (*models.CheckClientAccessResponse, error) {
	// 根据code获取客户端设备
	device, err := database.Client.ClientDevice.Query().
		Where(clientdevice.Code(req.Code)).
		WithRoles().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &models.CheckClientAccessResponse{
				Allowed: false,
				Reason:  "客户端设备不存在",
			}, nil
		}
		return nil, fmt.Errorf("failed to get client device: %w", err)
	}

	// 检查设备是否启用
	if !device.Enabled {
		return &models.CheckClientAccessResponse{
			Allowed: false,
			Reason:  "客户端设备已禁用",
		}, nil
	}

	// 如果设备允许匿名登录，直接允许
	if device.Anonymous {
		return &models.CheckClientAccessResponse{
			Allowed: true,
			Reason:  "",
		}, nil
	}

	// 如果不允许匿名登录，检查用户角色是否匹配
	if len(req.Roles) == 0 {
		return &models.CheckClientAccessResponse{
			Allowed: false,
			Reason:  "用户没有任何角色",
		}, nil
	}

	// 获取设备关联的角色ID
	var deviceRoleIds []string
	if device.Edges.Roles != nil {
		for _, role := range device.Edges.Roles {
			deviceRoleIds = append(deviceRoleIds, utils.Uint64ToString(role.ID))
		}
	}

	// 如果设备没有关联任何角色，不允许登录
	if len(deviceRoleIds) == 0 {
		return &models.CheckClientAccessResponse{
			Allowed: false,
			Reason:  "客户端设备未配置允许的角色",
		}, nil
	}

	// 检查用户角色是否与设备角色有交集
	for _, userRole := range req.Roles {
		for _, deviceRole := range deviceRoleIds {
			if userRole == deviceRole {
				return &models.CheckClientAccessResponse{
					Allowed: true,
					Reason:  "",
				}, nil
			}
		}
	}

	return &models.CheckClientAccessResponse{
		Allowed: false,
		Reason:  "用户角色不匹配客户端设备要求",
	}, nil
}
