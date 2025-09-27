package funcs

import (
	"context"
	"fmt"
	"go-backend/database/ent"
	"go-backend/database/ent/logging"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

type LoggingFunc struct{}

// GetAllLoggings 获取所有日志记录
func (LoggingFunc) GetAllLoggings(ctx context.Context) ([]*models.LoggingResponse, error) {
	records, err := database.Client.Logging.Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get loggings: %w", err)
	}
	var loggings []*models.LoggingResponse
	for _, record := range records {
		loggingResponse := &models.LoggingResponse{
			ID:         utils.Uint64ToString(record.ID),
			Level:      record.Level.String(),
			Type:       record.Type.String(),
			Message:    record.Message,
			Method:     record.Method,
			Path:       record.Path,
			IP:         record.IP,
			Query:      record.Query,
			Code:       record.Code,
			UserAgent:  record.UserAgent,
			Data:       record.Data,
			Stack:      record.Stack,
			CreateTime: utils.FormatDateTime(record.CreateTime),
			UpdateTime: utils.FormatDateTime(record.UpdateTime),
		}
		loggings = append(loggings, loggingResponse)
	}
	return loggings, nil
}

// GetLoggingById 根据ID获取日志记录
func (LoggingFunc) GetLoggingById(ctx context.Context, id uint64) (*models.LoggingResponse, error) {
	log, err := database.Client.Logging.Query().
		Where(logging.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("logging with id %d not found", id)
		}
		return nil, err
	}
	loggingResponse := &models.LoggingResponse{
		ID:         utils.Uint64ToString(log.ID),
		Level:      log.Level.String(),
		Type:       log.Type.String(),
		Message:    log.Message,
		Method:     log.Method,
		Path:       log.Path,
		IP:         log.IP,
		Query:      log.Query,
		Code:       log.Code,
		UserAgent:  log.UserAgent,
		Data:       log.Data,
		Stack:      log.Stack,
		CreateTime: utils.FormatDateTime(log.CreateTime),
		UpdateTime: utils.FormatDateTime(log.UpdateTime),
	}
	return loggingResponse, nil
}

// CreateLogging 创建日志记录
func (LoggingFunc) CreateLogging(ctx context.Context, req *models.CreateLoggingRequest) (*ent.Logging, error) {
	builder := database.Client.Logging.Create().
		SetMessage(req.Message)

	// 设置默认值或用户提供的值
	if req.Level != "" {
		builder = builder.SetLevel(logging.Level(req.Level))
	} else {
		builder = builder.SetLevel(logging.LevelInfo) // 默认值
	}

	if req.Type != "" {
		builder = builder.SetType(logging.Type(req.Type))
	} else {
		builder = builder.SetType(logging.TypeManul) // 默认值
	}

	// 可选字段
	if req.Method != "" {
		builder = builder.SetMethod(req.Method)
	}
	if req.Path != "" {
		builder = builder.SetPath(req.Path)
	}
	if req.IP != "" {
		builder = builder.SetIP(req.IP)
	}
	if req.Query != "" {
		builder = builder.SetQuery(req.Query)
	}
	if req.Code != nil {
		builder = builder.SetCode(*req.Code)
	}
	if req.UserAgent != "" {
		builder = builder.SetUserAgent(req.UserAgent)
	}
	if req.Data != nil {
		builder = builder.SetData(req.Data)
	}
	if req.Stack != "" {
		builder = builder.SetStack(req.Stack)
	}

	return builder.Save(ctx)
}

// UpdateLogging 更新日志记录
func (LoggingFunc) UpdateLogging(ctx context.Context, id uint64, req *models.UpdateLoggingRequest) (*ent.Logging, error) {
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

	// 首先检查日志记录是否存在
	exists, err := tx.Logging.Query().Where(logging.ID(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("logging with id %d not found", id)
	}

	builder := tx.Logging.UpdateOneID(id).
		SetMessage(req.Message)

	// 设置级别和类型
	if req.Level != "" {
		builder = builder.SetLevel(logging.Level(req.Level))
	}
	if req.Type != "" {
		builder = builder.SetType(logging.Type(req.Type))
	}

	// 可选字段
	if req.Method != "" {
		builder = builder.SetMethod(req.Method)
	} else {
		builder = builder.ClearMethod()
	}
	if req.Path != "" {
		builder = builder.SetPath(req.Path)
	} else {
		builder = builder.ClearPath()
	}
	if req.IP != "" {
		builder = builder.SetIP(req.IP)
	} else {
		builder = builder.ClearIP()
	}
	if req.Query != "" {
		builder = builder.SetQuery(req.Query)
	} else {
		builder = builder.ClearQuery()
	}
	if req.Code != nil {
		builder = builder.SetCode(*req.Code)
	} else {
		builder = builder.ClearCode()
	}
	if req.UserAgent != "" {
		builder = builder.SetUserAgent(req.UserAgent)
	} else {
		builder = builder.ClearUserAgent()
	}
	if req.Data != nil {
		builder = builder.SetData(req.Data)
	} else {
		builder = builder.ClearData()
	}
	if req.Stack != "" {
		builder = builder.SetStack(req.Stack)
	} else {
		builder = builder.ClearStack()
	}

	return builder.Save(ctx)
}

// DeleteLogging 删除日志记录
func (LoggingFunc) DeleteLogging(ctx context.Context, id uint64) error {
	// 首先检查日志记录是否存在
	exists, err := database.Client.Logging.Query().Where(logging.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("logging with id %d not found", id)
	}

	return database.Client.Logging.DeleteOneID(id).Exec(ctx)
}

// GetLoggingWithPagination 分页获取日志记录
func (LoggingFunc) GetLoggingWithPagination(ctx context.Context, req *models.PageLoggingRequest) (*models.PageLoggingResponse, error) {
	// 构建查询
	query := database.Client.Logging.Query()

	// 时间过滤
	if req.BeginTime != "" || req.EndTime != "" {
		startAt, err := utils.ParseTime(req.BeginTime)
		if err != nil && req.BeginTime != "" {
			return nil, fmt.Errorf("invalid start time format: %w", err)
		}
		endAt, err := utils.ParseTime(req.EndTime)
		if err != nil && req.EndTime != "" {
			return nil, fmt.Errorf("invalid end time format: %w", err)
		}

		if !startAt.IsZero() && !endAt.IsZero() {
			query = query.Where(
				logging.CreateTimeGTE(startAt),
				logging.CreateTimeLTE(endAt),
			)
		} else if !startAt.IsZero() {
			query = query.Where(logging.CreateTimeGTE(startAt))
		} else if !endAt.IsZero() {
			query = query.Where(logging.CreateTimeLTE(endAt))
		}
	}

	// 日志级别过滤
	if req.Level != "" {
		query = query.Where(logging.LevelEQ(logging.Level(req.Level)))
	}

	// 日志类型过滤
	if req.Type != "" {
		query = query.Where(logging.TypeEQ(logging.Type(req.Type)))
	}

	// 消息内容模糊搜索
	if req.Message != "" {
		query = query.Where(logging.MessageContains(req.Message))
	}

	// HTTP方法过滤
	if req.Method != "" {
		query = query.Where(logging.MethodEQ(req.Method))
	}

	// 路径模糊搜索
	if req.Path != "" {
		query = query.Where(logging.PathContains(req.Path))
	}

	// IP过滤
	if req.IP != "" {
		query = query.Where(logging.IPEQ(req.IP))
	}

	// 状态码过滤
	if req.Code != nil {
		query = query.Where(logging.CodeEQ(*req.Code))
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
	case "level":
		if req.Order == "asc" {
			query = query.Order(ent.Asc("level"))
		} else {
			query = query.Order(ent.Desc("level"))
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
		return nil, fmt.Errorf("failed to count loggings: %w", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	loggings, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get loggings: %w", err)
	}

	var loggingResponses []*models.LoggingResponse
	for _, log := range loggings {
		loggingResponse := &models.LoggingResponse{
			ID:         utils.Uint64ToString(log.ID),
			Level:      log.Level.String(),
			Type:       log.Type.String(),
			Message:    log.Message,
			Method:     log.Method,
			Path:       log.Path,
			IP:         log.IP,
			Query:      log.Query,
			Code:       log.Code,
			UserAgent:  log.UserAgent,
			Data:       log.Data,
			Stack:      log.Stack,
			CreateTime: utils.FormatDateTime(log.CreateTime),
			UpdateTime: utils.FormatDateTime(log.UpdateTime),
		}
		loggingResponses = append(loggingResponses, loggingResponse)
	}

	if loggingResponses == nil {
		loggingResponses = make([]*models.LoggingResponse, 0)
	}

	return &models.PageLoggingResponse{
		Data: loggingResponses,
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
