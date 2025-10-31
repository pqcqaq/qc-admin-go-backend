package funcs

import (
	"context"
	"fmt"
	"math"

	"go-backend/database/ent"
	"go-backend/database/ent/area"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// AreaFuncs 地区服务函数
type AreaFuncs struct{}

// GetAllAreas 获取所有地区
func (AreaFuncs) GetAllAreas(ctx context.Context) ([]*models.AreaResponse, error) {
	areas, err := database.Client.Area.Query().
		WithParent().
		WithChildren().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	areaResponses := make([]*models.AreaResponse, 0, len(areas))
	for _, a := range areas {
		areaResponses = append(areaResponses, AreaFuncs{}.ConvertAreaToResponse(a))
	}

	return areaResponses, nil
}

// GetAreaByID 根据ID获取地区
func (AreaFuncs) GetAreaByID(ctx context.Context, id uint64) (*models.AreaResponse, error) {
	areaEntity, err := database.Client.Area.Query().
		Where(area.ID(id)).
		WithParent().
		WithChildren().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("area not found")
		}
		return nil, err
	}
	return AreaFuncs{}.ConvertAreaToResponse(areaEntity), nil
}

// CreateArea 创建地区
func (AreaFuncs) CreateArea(ctx context.Context, req *models.CreateAreaRequest) (*models.AreaResponse, error) {
	builder := database.Client.Area.Create().
		SetName(req.Name).
		SetLevel(area.Level(req.Level)).
		SetDepth(req.Depth).
		SetCode(req.Code).
		SetLatitude(req.Latitude).
		SetLongitude(req.Longitude)

	if req.Color != "" {
		builder = builder.SetColor(req.Color)
	}

	if req.ParentId != "" {
		parentId := utils.StringToUint64(req.ParentId)
		builder = builder.SetParentID(parentId)
	}

	areaEntity, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return AreaFuncs{}.GetAreaByID(ctx, areaEntity.ID)
}

// UpdateArea 更新地区
func (AreaFuncs) UpdateArea(ctx context.Context, id uint64, req *models.UpdateAreaRequest) (*models.AreaResponse, error) {
	builder := database.Client.Area.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	if req.Level != "" {
		builder = builder.SetLevel(area.Level(req.Level))
	}

	if req.Depth != nil {
		builder = builder.SetDepth(*req.Depth)
	}

	if req.Code != "" {
		builder = builder.SetCode(req.Code)
	}

	if req.Latitude != nil {
		builder = builder.SetLatitude(*req.Latitude)
	}

	if req.Longitude != nil {
		builder = builder.SetLongitude(*req.Longitude)
	}

	if req.Color != "" {
		builder = builder.SetColor(req.Color)
	}

	if req.ParentId != "" {
		parentId := utils.StringToUint64(req.ParentId)
		builder = builder.SetParentID(parentId)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("area not found")
		}
		return nil, err
	}

	return AreaFuncs{}.GetAreaByID(ctx, id)
}

// DeleteArea 删除地区(级联删除子地区)
func (AreaFuncs) DeleteArea(ctx context.Context, id uint64) error {
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return err
	}

	// 递归删除所有子地区
	_, err = tx.Area.Delete().Where(area.ParentIDEQ(id)).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除当前地区
	err = tx.Area.DeleteOneID(id).Exec(ctx)
	if err != nil {
		tx.Rollback()
		if ent.IsNotFound(err) {
			return fmt.Errorf("area not found")
		}
		return err
	}

	return tx.Commit()
}

// GetAreasWithPagination 分页获取地区列表
func (AreaFuncs) GetAreasWithPagination(ctx context.Context, req *models.GetAreasRequest) (*models.AreasListResponse, error) {
	query := database.Client.Area.Query().
		WithParent().
		WithChildren()

	// 添加搜索条件
	if req.Name != "" {
		query = query.Where(area.NameContains(req.Name))
	}

	if req.Level != "" {
		query = query.Where(area.LevelEQ(area.Level(req.Level)))
	}

	if req.Depth != nil {
		query = query.Where(area.DepthEQ(*req.Depth))
	}

	if req.Code != "" {
		query = query.Where(area.CodeContains(req.Code))
	}

	if req.ParentId != "" {
		parentId := utils.StringToUint64(req.ParentId)
		query = query.Where(area.ParentIDEQ(parentId))
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
				query = query.Order(ent.Desc(area.FieldName))
			} else {
				query = query.Order(ent.Asc(area.FieldName))
			}
		case "code":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(area.FieldCode))
			} else {
				query = query.Order(ent.Asc(area.FieldCode))
			}
		case "spell":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(area.FieldSpell))
			} else {
				query = query.Order(ent.Asc(area.FieldSpell))
			}
		case "depth":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(area.FieldDepth))
			} else {
				query = query.Order(ent.Asc(area.FieldDepth))
			}
		case "createTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(area.FieldCreateTime))
			} else {
				query = query.Order(ent.Asc(area.FieldCreateTime))
			}
		case "updateTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(area.FieldUpdateTime))
			} else {
				query = query.Order(ent.Asc(area.FieldUpdateTime))
			}
		}
	} else {
		// 默认按深度和名称排序
		query = query.Order(ent.Asc(area.FieldDepth), ent.Asc(area.FieldName))
	}

	// 执行查询
	areas, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	areaResponses := make([]*models.AreaResponse, 0, len(areas))
	for _, a := range areas {
		areaResponses = append(areaResponses, AreaFuncs{}.ConvertAreaToResponse(a))
	}

	return &models.AreasListResponse{
		Data: areaResponses,
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

// GetAreasByParentID 根据父级ID获取下一级地区
func (AreaFuncs) GetAreasByParentID(ctx context.Context, parentId string) ([]*models.AreaResponse, error) {
	var query *ent.AreaQuery

	if parentId == "" {
		// 如果parentId为空,获取顶级地区(没有父级)
		query = database.Client.Area.Query().
			Where(area.ParentIDIsNil())
	} else {
		parentIdUint64 := utils.StringToUint64(parentId)
		// 获取指定父级的子地区
		query = database.Client.Area.Query().
			Where(area.ParentIDEQ(parentIdUint64))
	}

	areas, err := query.Order(ent.Asc(area.FieldName)).All(ctx)
	if err != nil {
		return nil, err
	}

	areaResponses := make([]*models.AreaResponse, 0, len(areas))
	for _, a := range areas {
		// 只返回简单信息,不需要子级和父级
		areaResponses = append(areaResponses, &models.AreaResponse{
			ID:        utils.Uint64ToString(a.ID),
			Name:      a.Name,
			Spell:     a.Spell,
			Level:     string(a.Level),
			Depth:     a.Depth,
			Code:      a.Code,
			Latitude:  a.Latitude,
			Longitude: a.Longitude,
			Color:     a.Color,
		})
	}

	return areaResponses, nil
}

// GetAreasByLevel 根据级别获取地区
func (AreaFuncs) GetAreasByLevel(ctx context.Context, level string) ([]*models.AreaResponse, error) {
	areas, err := database.Client.Area.Query().
		Where(area.LevelEQ(area.Level(level))).
		WithParent().
		WithChildren().
		Order(ent.Asc(area.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	areaResponses := make([]*models.AreaResponse, 0, len(areas))
	for _, a := range areas {
		areaResponses = append(areaResponses, AreaFuncs{}.ConvertAreaToResponse(a))
	}

	return areaResponses, nil
}

// GetAreasByDepth 根据深度获取地区
func (AreaFuncs) GetAreasByDepth(ctx context.Context, depth int) ([]*models.AreaResponse, error) {
	areas, err := database.Client.Area.Query().
		Where(area.DepthEQ(depth)).
		WithParent().
		WithChildren().
		Order(ent.Asc(area.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	areaResponses := make([]*models.AreaResponse, 0, len(areas))
	for _, a := range areas {
		areaResponses = append(areaResponses, AreaFuncs{}.ConvertAreaToResponse(a))
	}

	return areaResponses, nil
}

// GetAreaTree 获取地区树形结构
func (AreaFuncs) GetAreaTree(ctx context.Context) (*models.AreaTreeResponse, error) {
	// 获取所有地区
	allAreas, err := database.Client.Area.Query().
		WithParent().
		Order(ent.Asc(area.FieldDepth), ent.Asc(area.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	areaMap := make(map[uint64]*models.AreaResponse)
	var rootAreas []*models.AreaResponse

	// 先创建所有节点
	for _, a := range allAreas {
		areaResp := AreaFuncs{}.ConvertAreaToResponseForTree(a)
		areaMap[a.ID] = areaResp

		// 如果没有父节点,则为根节点
		if a.ParentID == 0 {
			rootAreas = append(rootAreas, areaResp)
		}
	}

	// 构建父子关系
	for _, a := range allAreas {
		if a.ParentID != 0 {
			parent := areaMap[a.ParentID]
			child := areaMap[a.ID]
			if parent != nil && child != nil {
				if parent.Children == nil {
					parent.Children = make([]*models.AreaResponse, 0)
				}
				parent.Children = append(parent.Children, child)
				// 设置子节点的父节点信息
				child.Parent = &models.AreaResponse{
					ID:    parent.ID,
					Name:  parent.Name,
					Level: parent.Level,
					Depth: parent.Depth,
					Code:  parent.Code,
				}
			}
		}
	}

	return &models.AreaTreeResponse{
		Data: rootAreas,
	}, nil
}

// ConvertAreaToResponse 将地区实体转换为响应格式
func (AreaFuncs) ConvertAreaToResponse(a *ent.Area) *models.AreaResponse {
	resp := &models.AreaResponse{
		ID:         utils.Uint64ToString(a.ID),
		Name:       a.Name,
		Spell:      a.Spell,
		Level:      string(a.Level),
		Depth:      a.Depth,
		Code:       a.Code,
		Latitude:   a.Latitude,
		Longitude:  a.Longitude,
		Color:      a.Color,
		CreateTime: utils.FormatDateTime(a.CreateTime),
		UpdateTime: utils.FormatDateTime(a.UpdateTime),
	}

	if a.ParentID != 0 {
		resp.ParentId = utils.Uint64ToString(a.ParentID)
	}

	// 转换父级地区(简单信息)
	if a.Edges.Parent != nil {
		resp.Parent = &models.AreaResponse{
			ID:    utils.Uint64ToString(a.Edges.Parent.ID),
			Name:  a.Edges.Parent.Name,
			Spell: a.Edges.Parent.Spell,
			Level: string(a.Edges.Parent.Level),
			Depth: a.Edges.Parent.Depth,
			Code:  a.Edges.Parent.Code,
		}
	}

	// 转换子级地区(简单信息)
	if len(a.Edges.Children) > 0 {
		resp.Children = make([]*models.AreaResponse, 0, len(a.Edges.Children))
		for _, child := range a.Edges.Children {
			resp.Children = append(resp.Children, &models.AreaResponse{
				ID:    utils.Uint64ToString(child.ID),
				Name:  child.Name,
				Spell: child.Spell,
				Level: string(child.Level),
				Depth: child.Depth,
				Code:  child.Code,
			})
		}
	}

	return resp
}

// ConvertAreaToResponseForTree 将地区实体转换为响应格式(专用于树形结构)
func (AreaFuncs) ConvertAreaToResponseForTree(a *ent.Area) *models.AreaResponse {
	resp := &models.AreaResponse{
		ID:         utils.Uint64ToString(a.ID),
		Name:       a.Name,
		Spell:      a.Spell,
		Level:      string(a.Level),
		Depth:      a.Depth,
		Code:       a.Code,
		Latitude:   a.Latitude,
		Longitude:  a.Longitude,
		Color:      a.Color,
		CreateTime: utils.FormatDateTime(a.CreateTime),
		UpdateTime: utils.FormatDateTime(a.UpdateTime),
	}

	if a.ParentID != 0 {
		resp.ParentId = utils.Uint64ToString(a.ParentID)
	}

	// 转换父级地区(简单信息)
	if a.Edges.Parent != nil {
		resp.Parent = &models.AreaResponse{
			ID:    utils.Uint64ToString(a.Edges.Parent.ID),
			Name:  a.Edges.Parent.Name,
			Spell: a.Edges.Parent.Spell,
			Level: string(a.Edges.Parent.Level),
			Depth: a.Edges.Parent.Depth,
			Code:  a.Edges.Parent.Code,
		}
	}

	return resp
}
