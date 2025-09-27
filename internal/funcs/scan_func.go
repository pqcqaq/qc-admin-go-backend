package funcs

import (
	"context"
	"fmt"
	"go-backend/database/ent"
	"go-backend/database/ent/scan"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

type ScanFuncs struct{}

// GetAllScans 获取所有扫描记录
func (ScanFuncs) GetAllScans(ctx context.Context) ([]*models.ScanResponse, error) {
	records, err := database.Client.Scan.Query().
		WithAttachment().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %w", err)
	}
	var scans []*models.ScanResponse
	for _, record := range records {
		scanResponse := &models.ScanResponse{
			ID:         utils.Uint64ToString(record.ID),
			Content:    record.Content,
			Success:    record.Success,
			CreateTime: utils.FormatDateTime(record.CreateTime),
		}

		// 由于使用了 WithAttachment()，关联数据已经在 Edges 中了
		if record.Edges.Attachment != nil {
			scanResponse.ImageId = utils.Uint64ToString(record.Edges.Attachment.ID)
			scanResponse.ImageUrl = record.Edges.Attachment.URL
		}

		scans = append(scans, scanResponse)
	}
	return scans, nil
}

// GetScanById 根据ID获取扫描记录
func (ScanFuncs) GetScanById(ctx context.Context, id uint64) (*models.ScanResponse, error) {
	scan, err := database.Client.Scan.Query().
		Where(scan.ID(id)).
		WithAttachment().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("scan with id %d not found", id)
		}
		return nil, err
	}
	scanResponse := &models.ScanResponse{
		ID:         utils.Uint64ToString(scan.ID),
		Content:    scan.Content,
		Success:    scan.Success,
		CreateTime: utils.FormatDateTime(scan.CreateTime),
	}
	if scan.Edges.Attachment != nil {
		scanResponse.ImageId = utils.Uint64ToString(scan.Edges.Attachment.ID)
		scanResponse.ImageUrl = scan.Edges.Attachment.URL
	}
	return scanResponse, nil
}

// CreateScan 创建扫描记录
func (ScanFuncs) CreateScan(ctx context.Context, req *models.CreateScanRequest) (*ent.Scan, error) {
	builder := database.Client.Scan.Create().
		SetContent(req.Content).
		SetLength(len(req.Content)).
		SetSuccess(*req.Success)

	if req.ImageId != "" {
		builder = builder.SetAttachmentID(utils.StringToUint64(req.ImageId))
	}

	return builder.Save(ctx)
}

// UpdateScan 更新扫描记录
func (ScanFuncs) UpdateScan(ctx context.Context, id uint64, req *models.UpdateScanRequest) (*ent.Scan, error) {
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

	// 首先检查扫描记录是否存在
	exists, err := tx.Scan.Query().Where(scan.ID(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("scan with id %d not found", id)
	}

	builder := tx.Scan.UpdateOneID(id).
		SetContent(req.Content).
		SetSuccess(*req.Success)

	if req.ImageId != "" {
		builder = builder.SetAttachmentID(utils.StringToUint64(req.ImageId))
	}

	return builder.Save(ctx)
}

// DeleteScan 删除扫描记录
func (ScanFuncs) DeleteScan(ctx context.Context, id uint64) error {
	// 首先检查扫描记录是否存在
	exists, err := database.Client.Scan.Query().Where(scan.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("scan with id %d not found", id)
	}

	return database.Client.Scan.DeleteOneID(id).Exec(ctx)
}

func (ScanFuncs) GetScanWithPagination(ctx context.Context, req *models.PageScansRequest) (*models.PageScansResponse, error) {
	// 构建查询
	query := database.Client.Scan.Query().
		WithAttachment()

	startAt, err := utils.ParseTime(req.BeginTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}
	endAt, err := utils.ParseTime(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	// 是否成功
	if req.Success != nil {
		query = query.Where(scan.SuccessEQ(*req.Success))
	}

	// 时间
	if !(startAt.IsZero() || endAt.IsZero()) {
		query = query.Where(
			scan.CreateTimeGTE(startAt),
			scan.CreateTimeLTE(endAt),
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
	default:
		query = query.Order(ent.Desc("create_time"))
	}

	// 模糊
	if req.Content != "" {
		query = query.Where(scan.ContentContains(req.Content))
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
		return nil, fmt.Errorf("failed to count scans: %w", err)
	}

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	scans, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %w", err)
	}

	var scanResponses []*models.ScanResponse
	for _, scan := range scans {
		scanResponse := &models.ScanResponse{
			ID:         utils.Uint64ToString(scan.ID),
			Content:    scan.Content,
			Success:    scan.Success,
			CreateTime: utils.FormatDateTime(scan.CreateTime),
		}

		// 正确处理可能为空的 attachment 关联
		if scan.Edges.Attachment != nil {
			scanResponse.ImageId = utils.Uint64ToString(scan.Edges.Attachment.ID)
			scanResponse.ImageUrl = scan.Edges.Attachment.URL
		}

		scanResponses = append(scanResponses, scanResponse)
	}

	if scanResponses == nil {
		scanResponses = make([]*models.ScanResponse, 0)
	}

	return &models.PageScansResponse{
		Data: scanResponses,
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
