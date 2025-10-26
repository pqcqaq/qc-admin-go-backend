# 从数据库Schema到完整API实现的开发指南

## 概述

本文档描述了如何从一个数据库Schema定义快速生成完整的后端API接口和前端API封装的系统化方法。这套方法论可以确保代码的一致性、完整性和可维护性。

## 核心开发流程

```
数据库Schema (Ent) 
    ↓
分析实体关系和字段
    ↓
设计API接口规划
    ↓
实现后端代码 (Models → Funcs → Handlers → Routes)
    ↓
实现前端API封装 (TypeScript)
    ↓
测试验证
```

## 第一步：分析数据库Schema

### 1.1 从Schema中提取关键信息

以 `area.go` 为例，需要关注：

```go
// Area 地区实体(存储全国行政区划数据)
type Area struct {
    ent.Schema
}

func (Area) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").MaxLen(32).NotEmpty(),
        field.Enum("level").Values("country", "province", "city", "district", "street"),
        field.Int("depth").Min(0).Max(4),
        field.String("code").MaxLen(12).NotEmpty(),
        field.Float("latitude"),
        field.Float("longitude"),
        field.Uint64("parent_id").Optional(),
        field.String("color").MaxLen(20).Optional(),
    }
}

func (Area) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("parent", Area.Type).Field("parent_id").Unique().From("children"),
    }
}
```

**关键提取信息：**

1. **实体名称**: Area (地区)
2. **必填字段**: name, level, depth, code, latitude, longitude
3. **可选字段**: parent_id, color
4. **字段类型**: string, enum, int, float, uint64
5. **字段约束**: MaxLen, Min, Max, NotEmpty, Optional
6. **枚举值**: country, province, city, district, street (level字段)
7. **自关联关系**: parent/children (树形结构)
8. **业务逻辑**: 
   - 层级结构 (0-4层深度)
   - 地理位置 (经纬度)
   - 行政区划编码

### 1.2 确定Mixin特性

```go
func (Area) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixins.BaseMixin{},          // ID, CreateTime, UpdateTime
        mixins.SoftDeleteMixin{},    // DeleteTime, 软删除
    }
}
```

**继承的通用字段：**
- `ID` (uint64): 主键
- `CreateTime` (time.Time): 创建时间
- `UpdateTime` (time.Time): 更新时间
- `DeleteTime` (*time.Time): 删除时间（软删除）

### 1.3 分析索引和查询需求

```go
func (Area) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("code", "delete_time").Unique(),  // 编码唯一
        index.Fields("name"),                          // 按名称查询
        index.Fields("level"),                         // 按层级查询
        index.Fields("depth"),                         // 按深度查询
        index.Fields("parent_id"),                     // 按父级查询
    }
}
```

**推导查询接口需求：**
- 按名称模糊搜索
- 按层级类型筛选
- 按深度筛选
- 按父级ID获取子项（树形查询的核心）
- 编码唯一性校验

## 第二步：设计API接口规划

### 2.1 标准CRUD接口（必备）

基于RESTful规范，每个实体都需要：

| 功能 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 列表 | GET | `/areas` | 获取所有记录（谨慎使用） |
| 分页 | GET | `/areas/page` | 分页获取记录（推荐） |
| 详情 | GET | `/areas/:id` | 获取单条记录 |
| 创建 | POST | `/areas` | 创建新记录 |
| 更新 | PUT | `/areas/:id` | 更新记录 |
| 删除 | DELETE | `/areas/:id` | 删除记录 |

### 2.2 业务特定接口（根据Schema分析）

基于Area的特性，增加：

| 功能 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 树形结构 | GET | `/areas/tree` | 获取完整树形结构 |
| 子级查询 | GET | `/areas/children?parentId=xxx` | 获取指定父级的子项 |
| 层级查询 | GET | `/areas/level?level=province` | 按层级类型查询 |
| 深度查询 | GET | `/areas/depth?depth=1` | 按深度查询 |

**设计依据：**
- **树形结构**: 因为有 parent/children 自关联
- **子级查询**: 常见的级联选择需求（省→市→区）
- **层级/深度查询**: 有 level 和 depth 枚举字段

### 2.3 查询参数设计原则

**分页查询支持的参数：**
```
page: 页码
pageSize: 每页数量
order: 排序方向 (asc/desc)
orderBy: 排序字段 (name/code/depth/createTime/updateTime)
name: 名称模糊搜索
level: 层级精确匹配
depth: 深度精确匹配
code: 编码模糊搜索
parentId: 父级ID精确匹配
```

## 第三步：实现Models层

### 3.1 创建请求/响应模型

**文件位置**: `shared/models/area.go`

#### 3.1.1 Create请求模型

```go
type CreateAreaRequest struct {
    // 必填字段 - binding:"required"
    Name  string  `json:"name" binding:"required"`
    Level string  `json:"level" binding:"required,oneof=country province city district street"`
    Depth int     `json:"depth" binding:"required,min=0,max=4"`
    Code  string  `json:"code" binding:"required"`
    
    // 有默认值的字段 - 不标记required
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    
    // 可选字段 - omitempty
    ParentId string `json:"parentId,omitempty"`
    Color    string `json:"color,omitempty"`
}
```

**关键点：**
1. 从Schema的Required/Optional判断`binding`标签
2. 枚举字段使用`oneof`验证
3. 数值范围使用`min/max`验证
4. 可选字段使用`omitempty`
5. ID字段统一使用string类型（JSON友好）

#### 3.1.2 Update请求模型

```go
type UpdateAreaRequest struct {
    // 全部字段都是可选的
    Name      string   `json:"name,omitempty"`
    Level     string   `json:"level,omitempty"`
    Depth     *int     `json:"depth,omitempty"`      // 指针类型：区分0值和未设置
    Code      string   `json:"code,omitempty"`
    Latitude  *float64 `json:"latitude,omitempty"`   // 指针类型：区分0值和未设置
    Longitude *float64 `json:"longitude,omitempty"`
    ParentId  string   `json:"parentId,omitempty"`
    Color     string   `json:"color,omitempty"`
}
```

**关键点：**
1. Update模型所有字段都可选
2. 数值类型（int, float）使用指针，区分"未设置"和"设为0"
3. 字符串类型可以用空字符串表示未设置

#### 3.1.3 Response模型

```go
type AreaResponse struct {
    // 基础字段
    ID         string  `json:"id"`
    Name       string  `json:"name"`
    Level      string  `json:"level"`
    Depth      int     `json:"depth"`
    Code       string  `json:"code"`
    Latitude   float64 `json:"latitude"`
    Longitude  float64 `json:"longitude"`
    ParentId   string  `json:"parentId,omitempty"`
    Color      string  `json:"color,omitempty"`
    
    // 关联数据
    Parent   *AreaResponse   `json:"parent,omitempty"`
    Children []*AreaResponse `json:"children,omitempty"`
    
    // 时间字段（Mixin继承）
    CreateTime string `json:"createTime"`
    UpdateTime string `json:"updateTime"`
}
```

**关键点：**
1. 所有字段都包含（展示用）
2. 关联数据用指针/切片表示（可能为nil）
3. 时间格式化为string（前端友好）
4. 使用驼峰命名（JSON标准）

#### 3.1.4 查询请求模型

```go
type GetAreasRequest struct {
    PaginationRequest  // 嵌入分页参数
    
    // 查询条件 - 对应索引字段
    Name     string `form:"name" json:"name"`
    Level    string `form:"level" json:"level"`
    Depth    *int   `form:"depth" json:"depth"`
    Code     string `form:"code" json:"code"`
    ParentId string `form:"parentId" json:"parentId"`
}
```

**关键点：**
1. 嵌入`PaginationRequest`（page, pageSize, order, orderBy）
2. 查询字段对应索引字段
3. 使用`form`和`json`双标签（支持GET/POST）

#### 3.1.5 响应列表模型

```go
type AreasListResponse struct {
    Data       []*AreaResponse `json:"data"`
    Pagination Pagination      `json:"pagination"`
}

type AreaTreeResponse struct {
    Data []*AreaResponse `json:"data"`
}
```

## 第四步：实现Funcs层（业务逻辑）

### 4.1 结构和命名规范

**文件位置**: `internal/funcs/area_func.go`

```go
type AreaFuncs struct{}  // 空结构体，用于组织方法
```

**命名规范：**
- 文件名：`{实体名}_func.go`
- 结构体：`{实体名}Funcs`
- 方法：使用值接收器 `func (AreaFuncs) MethodName()`

### 4.2 基础CRUD实现模板

#### 4.2.1 GetAll - 获取所有记录

```go
func (AreaFuncs) GetAllAreas(ctx context.Context) ([]*ent.Area, error) {
    return database.Client.Area.Query().
        WithParent().      // 加载关联的父级
        WithChildren().    // 加载关联的子级
        All(ctx)
}
```

**模板要点：**
- 返回Ent实体类型（不转换）
- 使用`With{Edge}()`预加载关联数据
- 简单直接，Handler层负责转换

#### 4.2.2 GetByID - 根据ID获取

```go
func (AreaFuncs) GetAreaByID(ctx context.Context, id uint64) (*ent.Area, error) {
    area, err := database.Client.Area.Query().
        Where(area.ID(id)).
        WithParent().
        WithChildren().
        Only(ctx)  // Only()期望唯一结果
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, fmt.Errorf("area not found")
        }
        return nil, err
    }
    return area, nil
}
```

**模板要点：**
1. 使用`Where()`过滤
2. 使用`Only()`获取唯一结果
3. 处理`NotFound`错误，返回友好消息
4. 预加载必要的关联数据

#### 4.2.3 Create - 创建记录

```go
func (AreaFuncs) CreateArea(ctx context.Context, req *models.CreateAreaRequest) (*ent.Area, error) {
    // 构建创建器
    builder := database.Client.Area.Create().
        SetName(req.Name).
        SetLevel(area.Level(req.Level)).  // 枚举类型转换
        SetDepth(req.Depth).
        SetCode(req.Code).
        SetLatitude(req.Latitude).
        SetLongitude(req.Longitude)

    // 可选字段单独处理
    if req.Color != "" {
        builder = builder.SetColor(req.Color)
    }

    if req.ParentId != "" {
        parentId := utils.StringToUint64(req.ParentId)  // ID转换
        builder = builder.SetParentID(parentId)
    }

    // 保存并重新查询（包含关联数据）
    area, err := builder.Save(ctx)
    if err != nil {
        return nil, err
    }

    return AreaFuncs{}.GetAreaByID(ctx, area.ID)
}
```

**模板要点：**
1. 必填字段直接设置
2. 可选字段检查后设置
3. 枚举类型需要类型转换：`area.Level(req.Level)`
4. ID字段需要string→uint64转换
5. 创建后重新查询，获取完整数据（包括关联）

#### 4.2.4 Update - 更新记录

```go
func (AreaFuncs) UpdateArea(ctx context.Context, id uint64, req *models.UpdateAreaRequest) (*ent.Area, error) {
    builder := database.Client.Area.UpdateOneID(id)

    // 只更新提供的字段
    if req.Name != "" {
        builder = builder.SetName(req.Name)
    }

    if req.Level != "" {
        builder = builder.SetLevel(area.Level(req.Level))
    }

    // 指针类型字段
    if req.Depth != nil {
        builder = builder.SetDepth(*req.Depth)
    }

    if req.Latitude != nil {
        builder = builder.SetLatitude(*req.Latitude)
    }

    if req.Longitude != nil {
        builder = builder.SetLongitude(*req.Longitude)
    }

    if req.Code != "" {
        builder = builder.SetCode(req.Code)
    }

    if req.Color != "" {
        builder = builder.SetColor(req.Color)
    }

    if req.ParentId != "" {
        parentId := utils.StringToUint64(req.ParentId)
        builder = builder.SetParentID(parentId)
    }

    // 执行更新
    err := builder.Exec(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, fmt.Errorf("area not found")
        }
        return nil, err
    }

    return AreaFuncs{}.GetAreaByID(ctx, id)
}
```

**模板要点：**
1. 使用`UpdateOneID(id)`
2. 每个字段都要判断是否提供（部分更新）
3. 字符串字段检查非空
4. 指针字段检查非nil
5. 最后`Exec()`执行更新
6. 更新后重新查询返回完整数据

#### 4.2.5 Delete - 删除记录

```go
func (AreaFuncs) DeleteArea(ctx context.Context, id uint64) error {
    // 使用事务（涉及级联删除）
    tx, err := database.Client.Tx(ctx)
    if err != nil {
        return err
    }

    // 删除所有子记录（级联删除）
    _, err = tx.Area.Delete().Where(area.ParentIDEQ(id)).Exec(ctx)
    if err != nil {
        tx.Rollback()
        return err
    }

    // 删除当前记录
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
```

**模板要点：**
1. 有级联关系时使用事务
2. 先删除关联数据（子记录）
3. 再删除主记录
4. 错误时回滚事务
5. 成功时提交事务

### 4.3 分页查询实现模板

```go
func (AreaFuncs) GetAreasWithPagination(ctx context.Context, req *models.GetAreasRequest) (*models.AreasListResponse, error) {
    // 1. 构建基础查询
    query := database.Client.Area.Query().
        WithParent().
        WithChildren()

    // 2. 添加过滤条件
    if req.Name != "" {
        query = query.Where(area.NameContains(req.Name))  // 模糊搜索
    }

    if req.Level != "" {
        query = query.Where(area.LevelEQ(area.Level(req.Level)))  // 精确匹配
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

    // 3. 获取总数（在分页前）
    total, err := query.Count(ctx)
    if err != nil {
        return nil, err
    }

    // 4. 计算分页参数
    offset := (req.Page - 1) * req.PageSize
    totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

    // 5. 添加排序
    if req.OrderBy != "" {
        switch req.OrderBy {
        case "name":
            if req.Order == "desc" {
                query = query.Order(ent.Desc(area.FieldName))
            } else {
                query = query.Order(ent.Asc(area.FieldName))
            }
        case "code":
            // ... 其他字段
        case "depth":
            // ...
        case "createTime":
            // ...
        case "updateTime":
            // ...
        }
    } else {
        // 默认排序（业务相关）
        query = query.Order(ent.Asc(area.FieldDepth), ent.Asc(area.FieldName))
    }

    // 6. 执行分页查询
    areas, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
    if err != nil {
        return nil, err
    }

    // 7. 转换为响应格式
    areaResponses := make([]*models.AreaResponse, 0, len(areas))
    for _, a := range areas {
        areaResponses = append(areaResponses, AreaFuncs{}.ConvertAreaToResponse(a))
    }

    // 8. 返回结果
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
```

**分页查询模板关键步骤：**
1. 构建基础查询 + 预加载关联
2. 应用所有过滤条件
3. **先计算总数**（在Offset/Limit前）
4. 计算分页参数（offset, totalPages）
5. 应用排序规则
6. 应用分页（Offset + Limit）
7. 执行查询
8. 转换为响应格式
9. 构造分页元信息

### 4.4 业务特定查询实现

#### 4.4.1 树形查询（根据ParentID）

```go
func (AreaFuncs) GetAreasByParentID(ctx context.Context, parentId uint64) ([]*models.AreaResponse, error) {
    var query *ent.AreaQuery

    if parentId == 0 {
        // 获取根节点
        query = database.Client.Area.Query().
            Where(area.ParentIDIsNil()).
            WithChildren()
    } else {
        // 获取指定父级的子节点
        query = database.Client.Area.Query().
            Where(area.ParentIDEQ(parentId)).
            WithChildren()
    }

    areas, err := query.Order(ent.Asc(area.FieldName)).All(ctx)
    if err != nil {
        return nil, err
    }

    areaResponses := make([]*models.AreaResponse, 0, len(areas))
    for _, a := range areas {
        areaResponses = append(areaResponses, AreaFuncs{}.ConvertAreaToResponse(a))
    }

    return areaResponses, nil
}
```

#### 4.4.2 按枚举字段查询

```go
func (AreaFuncs) GetAreasByLevel(ctx context.Context, level string) ([]*models.AreaResponse, error) {
    areas, err := database.Client.Area.Query().
        Where(area.LevelEQ(area.Level(level))).  // 枚举转换
        WithParent().
        WithChildren().
        Order(ent.Asc(area.FieldName)).
        All(ctx)
    if err != nil {
        return nil, err
    }

    // 转换逻辑...
}
```

#### 4.4.3 完整树形结构构建

```go
func (AreaFuncs) GetAreaTree(ctx context.Context) (*models.AreaTreeResponse, error) {
    // 1. 获取所有节点（按层级排序）
    allAreas, err := database.Client.Area.Query().
        WithParent().
        Order(ent.Asc(area.FieldDepth), ent.Asc(area.FieldName)).
        All(ctx)
    if err != nil {
        return nil, err
    }

    // 2. 构建节点映射 + 识别根节点
    areaMap := make(map[uint64]*models.AreaResponse)
    var rootAreas []*models.AreaResponse

    for _, a := range allAreas {
        areaResp := AreaFuncs{}.ConvertAreaToResponseForTree(a)
        areaMap[a.ID] = areaResp

        if a.ParentID == 0 {
            rootAreas = append(rootAreas, areaResp)
        }
    }

    // 3. 建立父子关系
    for _, a := range allAreas {
        if a.ParentID != 0 {
            parent := areaMap[a.ParentID]
            child := areaMap[a.ID]
            if parent != nil && child != nil {
                if parent.Children == nil {
                    parent.Children = make([]*models.AreaResponse, 0)
                }
                parent.Children = append(parent.Children, child)
                
                // 设置子节点的父引用（简化版）
                child.Parent = &models.AreaResponse{
                    ID:    parent.ID,
                    Name:  parent.Name,
                    Level: parent.Level,
                    // ... 关键字段
                }
            }
        }
    }

    return &models.AreaTreeResponse{
        Data: rootAreas,
    }, nil
}
```

**树形结构构建步骤：**
1. 按层级顺序查询所有节点
2. 创建ID→节点的映射表
3. 识别根节点（parentId == 0）
4. 遍历建立父子关系
5. 返回根节点数组

### 4.5 实体转换函数

#### 4.5.1 完整转换（包含关联）

```go
func (AreaFuncs) ConvertAreaToResponse(a *ent.Area) *models.AreaResponse {
    resp := &models.AreaResponse{
        ID:         utils.Uint64ToString(a.ID),
        Name:       a.Name,
        Level:      string(a.Level),  // 枚举转字符串
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

    // 转换父级（简化信息）
    if a.Edges.Parent != nil {
        resp.Parent = &models.AreaResponse{
            ID:    utils.Uint64ToString(a.Edges.Parent.ID),
            Name:  a.Edges.Parent.Name,
            Level: string(a.Edges.Parent.Level),
            Depth: a.Edges.Parent.Depth,
            Code:  a.Edges.Parent.Code,
        }
    }

    // 转换子级（简化信息）
    if len(a.Edges.Children) > 0 {
        resp.Children = make([]*models.AreaResponse, 0, len(a.Edges.Children))
        for _, child := range a.Edges.Children {
            resp.Children = append(resp.Children, &models.AreaResponse{
                ID:    utils.Uint64ToString(child.ID),
                Name:  child.Name,
                Level: string(child.Level),
                Depth: child.Depth,
                Code:  child.Code,
            })
        }
    }

    return resp
}
```

**转换函数要点：**
1. uint64 → string（ID转换）
2. enum → string（枚举转换）
3. time.Time → string（时间格式化）
4. 关联对象递归转换（避免循环引用）
5. 关联对象只包含关键字段（减少数据量）

#### 4.5.2 树形专用转换（不包含Children）

```go
func (AreaFuncs) ConvertAreaToResponseForTree(a *ent.Area) *models.AreaResponse {
    resp := &models.AreaResponse{
        // ... 同上，但不转换Children
    }
    
    // 只转换父级信息
    if a.Edges.Parent != nil {
        resp.Parent = &models.AreaResponse{ /* ... */ }
    }
    
    // 不转换Children，由树构建逻辑手动添加
    
    return resp
}
```

## 第五步：实现Handlers层（HTTP处理）

### 5.1 Handler结构

**文件位置**: `internal/handlers/area_handler.go`

```go
type AreaHandler struct{}

func NewAreaHandler() *AreaHandler {
    return &AreaHandler{}
}
```

### 5.2 Handler方法模板

#### 5.2.1 GetAll Handler

```go
func (h *AreaHandler) GetAreas(c *gin.Context) {
    // 1. 获取请求上下文
    ctx := middleware.GetRequestContext(c)
    
    // 2. 调用业务逻辑
    areas, err := funcs.AreaFuncs{}.GetAllAreas(ctx)
    if err != nil {
        middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
        return
    }

    // 3. 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    areas,
        "count":   len(areas),
    })
}
```

**Handler模板结构：**
1. 获取请求上下文（ctx）
2. 提取并验证参数
3. 调用业务逻辑（funcs层）
4. 处理错误
5. 返回JSON响应

#### 5.2.2 GetByID Handler

```go
func (h *AreaHandler) GetArea(c *gin.Context) {
    // 1. 提取路径参数
    idStr := c.Param("id")

    // 2. 验证参数
    if idStr == "" {
        middleware.ThrowError(c, middleware.BadRequestError("地区ID不能为空", nil))
        return
    }

    // 3. 类型转换
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        middleware.ThrowError(c, middleware.BadRequestError("地区ID格式无效", map[string]any{
            "provided_id": idStr,
        }))
        return
    }

    // 4. 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    area, err := funcs.AreaFuncs{}.GetAreaByID(ctx, id)
    if err != nil {
        // 5. 错误分类处理
        if err.Error() == "area not found" {
            middleware.ThrowError(c, middleware.NotFoundError("地区未找到", map[string]any{
                "id": id,
            }))
        } else {
            middleware.ThrowError(c, middleware.DatabaseError("查询地区失败", err.Error()))
        }
        return
    }

    // 6. 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    area,
    })
}
```

**参数提取方法：**
- 路径参数：`c.Param("id")`
- 查询参数：`c.Query("name")` 或 `c.ShouldBindQuery(&req)`
- Body参数：`c.ShouldBindJSON(&req)`

#### 5.2.3 Create Handler

```go
func (h *AreaHandler) CreateArea(c *gin.Context) {
    var req models.CreateAreaRequest

    // 1. 绑定并验证请求体
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
        return
    }

    // 2. 额外的业务验证（可选）
    if req.Name == "" {
        middleware.ThrowError(c, middleware.BadRequestError("地区名称不能为空", nil))
        return
    }

    // 3. 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    area, err := funcs.AreaFuncs{}.CreateArea(ctx, &req)
    if err != nil {
        middleware.ThrowError(c, middleware.DatabaseError("创建地区失败", err.Error()))
        return
    }

    // 4. 返回201 Created
    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data":    area,
        "message": "地区创建成功",
    })
}
```

#### 5.2.4 Update Handler

```go
func (h *AreaHandler) UpdateArea(c *gin.Context) {
    // 1. 提取ID
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        middleware.ThrowError(c, middleware.BadRequestError("地区ID格式无效", map[string]any{
            "provided_id": idStr,
        }))
        return
    }

    // 2. 绑定请求体
    var req models.UpdateAreaRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
        return
    }

    // 3. 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    area, err := funcs.AreaFuncs{}.UpdateArea(ctx, id, &req)
    if err != nil {
        if err.Error() == "area not found" {
            middleware.ThrowError(c, middleware.NotFoundError("地区未找到", map[string]any{
                "id": id,
            }))
        } else {
            middleware.ThrowError(c, middleware.DatabaseError("更新地区失败", err.Error()))
        }
        return
    }

    // 4. 返回200 OK
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    area,
        "message": "地区更新成功",
    })
}
```

#### 5.2.5 Delete Handler

```go
func (h *AreaHandler) DeleteArea(c *gin.Context) {
    // 提取ID
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        middleware.ThrowError(c, middleware.BadRequestError("地区ID格式无效", map[string]any{
            "provided_id": idStr,
        }))
        return
    }

    // 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    err = funcs.AreaFuncs{}.DeleteArea(ctx, id)
    if err != nil {
        if err.Error() == "area not found" {
            middleware.ThrowError(c, middleware.NotFoundError("地区未找到", map[string]any{
                "id": id,
            }))
        } else {
            middleware.ThrowError(c, middleware.DatabaseError("删除地区失败", err.Error()))
        }
        return
    }

    // 返回成功（无数据）
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "地区删除成功",
    })
}
```

#### 5.2.6 分页查询 Handler

```go
func (h *AreaHandler) GetAreasWithPagination(c *gin.Context) {
    var req models.GetAreasRequest

    // 1. 设置默认值
    req.Page = 1
    req.PageSize = 10
    req.Order = "asc"
    req.OrderBy = "depth"

    // 2. 绑定查询参数
    if err := c.ShouldBindQuery(&req); err != nil {
        middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
        return
    }

    // 3. 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    result, err := funcs.AreaFuncs{}.GetAreasWithPagination(ctx, &req)
    if err != nil {
        middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
        return
    }

    // 4. 返回结果
    c.JSON(http.StatusOK, gin.H{
        "success":    true,
        "data":       result.Data,
        "pagination": result.Pagination,
    })
}
```

#### 5.2.7 特殊查询 Handler（查询参数）

```go
func (h *AreaHandler) GetAreasByParentID(c *gin.Context) {
    // 1. 提取查询参数
    parentIdStr := c.Query("parentId")

    var parentId uint64
    var err error

    // 2. 参数转换（允许空值）
    if parentIdStr != "" {
        parentId, err = strconv.ParseUint(parentIdStr, 10, 64)
        if err != nil {
            middleware.ThrowError(c, middleware.BadRequestError("父级ID格式无效", map[string]any{
                "provided_parentId": parentIdStr,
            }))
            return
        }
    }

    // 3. 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    areas, err := funcs.AreaFuncs{}.GetAreasByParentID(ctx, parentId)
    if err != nil {
        middleware.ThrowError(c, middleware.DatabaseError("获取子地区列表失败", err.Error()))
        return
    }

    // 4. 返回结果
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    areas,
        "count":   len(areas),
    })
}
```

#### 5.2.8 枚举参数验证 Handler

```go
func (h *AreaHandler) GetAreasByLevel(c *gin.Context) {
    level := c.Query("level")

    // 1. 必填验证
    if level == "" {
        middleware.ThrowError(c, middleware.BadRequestError("层级类型不能为空", nil))
        return
    }

    // 2. 枚举值验证
    validLevels := map[string]bool{
        "country":  true,
        "province": true,
        "city":     true,
        "district": true,
        "street":   true,
    }

    if !validLevels[level] {
        middleware.ThrowError(c, middleware.BadRequestError("无效的层级类型", map[string]any{
            "provided_level": level,
            "valid_levels":   []string{"country", "province", "city", "district", "street"},
        }))
        return
    }

    // 3. 调用业务逻辑
    ctx := middleware.GetRequestContext(c)
    areas, err := funcs.AreaFuncs{}.GetAreasByLevel(ctx, level)
    if err != nil {
        middleware.ThrowError(c, middleware.DatabaseError("获取地区列表失败", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    areas,
        "count":   len(areas),
    })
}
```

### 5.3 Swagger文档注解模板

```go
// GetArea 根据ID获取地区
// @Summary      根据ID获取地区
// @Description  根据地区ID获取地区详细信息
// @Tags         areas
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "地区ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /areas/{id} [get]
func (h *AreaHandler) GetArea(c *gin.Context) { /* ... */ }
```

## 第六步：实现Routes层（路由配置）

### 6.1 路由文件模板

**文件位置**: `internal/routes/area.go`

```go
package routes

import (
    "go-backend/internal/handlers"
    "github.com/gin-gonic/gin"
)

// setupAreaRoutes 设置地区相关路由
func (r *Router) setupAreaRoutes(rg *gin.RouterGroup) {
    // 创建Handler实例
    areaHandler := handlers.NewAreaHandler()

    // 创建路由组
    areas := rg.Group("/areas")
    {
        // 基本CRUD操作
        areas.GET("", areaHandler.GetAreas)                    // 列表
        areas.GET("/page", areaHandler.GetAreasWithPagination) // 分页
        areas.GET("/:id", areaHandler.GetArea)                 // 详情
        areas.POST("", areaHandler.CreateArea)                 // 创建
        areas.PUT("/:id", areaHandler.UpdateArea)              // 更新
        areas.DELETE("/:id", areaHandler.DeleteArea)           // 删除
        
        // 特殊查询接口（放在/:id之后，避免冲突）
        areas.GET("/tree", areaHandler.GetAreaTree)            // 树形结构
        areas.GET("/children", areaHandler.GetAreasByParentID) // 子级查询
        areas.GET("/level", areaHandler.GetAreasByLevel)       // 层级查询
        areas.GET("/depth", areaHandler.GetAreasByDepth)       // 深度查询
    }
}
```

**路由顺序规则：**
1. 精确路径在前（如 `/tree`, `/children`）
2. 参数路径在后（如 `/:id`）
3. 否则会被参数路由拦截

### 6.2 注册到主路由

**文件位置**: `internal/routes/routes.go`

```go
func (r *Router) setupRoutes(engine *gin.Engine, config *ServerConfig) {
    // ...
    
    api := prefixGroup.Group("/v1")
    {
        r.setupTestRoutes(api)
        r.setupAuthRoutes(api)
        r.setupUserRoutes(api)
        r.setupAttachmentRoutes(api)
        r.setupScanRoutes(api)
        r.setupAreaRoutes(api)  // 新增
        // ...
    }
}
```

## 第七步：实现前端API封装（TypeScript）

### 7.1 API文件模板

**文件位置**: `qc-admin-api-common/src/area.ts`

#### 7.1.1 导入依赖

```typescript
import type { Pagination } from ".";
import { http } from "./http";
```

#### 7.1.2 类型定义

```typescript
// 主实体类型
export type Area = {
  /** 地区ID */
  id: string;
  /** 地区名称 */
  name: string;
  /** 层级类型 */
  level: "country" | "province" | "city" | "district" | "street";
  /** 深度 */
  depth: number;
  /** 地区编码 */
  code: string;
  /** 纬度 */
  latitude: number;
  /** 经度 */
  longitude: number;
  /** 父级ID */
  parentId?: string;
  /** 颜色 */
  color?: string;
  /** 父级地区 */
  parent?: Area;
  /** 子级地区 */
  children?: Array<Area>;
  /** 创建时间 */
  createTime: string;
  /** 更新时间 */
  updateTime: string;
};

// 创建请求类型
export type CreateAreaRequest = {
  /** 地区名称 */
  name: string;
  /** 层级类型 */
  level: "country" | "province" | "city" | "district" | "street";
  /** 深度 */
  depth: number;
  /** 地区编码 */
  code: string;
  /** 纬度 */
  latitude?: number;
  /** 经度 */
  longitude?: number;
  /** 父级ID */
  parentId?: string;
  /** 颜色 */
  color?: string;
};

// 更新请求类型
export type UpdateAreaRequest = {
  /** 地区名称 */
  name?: string;
  /** 层级类型 */
  level?: "country" | "province" | "city" | "district" | "street";
  /** 深度 */
  depth?: number;
  /** 地区编码 */
  code?: string;
  /** 纬度 */
  latitude?: number;
  /** 经度 */
  longitude?: number;
  /** 父级ID */
  parentId?: string;
  /** 颜色 */
  color?: string;
};

// 响应类型
export type AreaResult = {
  success: boolean;
  data: Area;
  message?: string;
};

export type AreaListResult = {
  success: boolean;
  data: Array<Area>;
  pagination?: Pagination;
  count?: number;
};

// 查询参数类型
export type GetAreasParams = {
  /** 页码 */
  page?: number;
  /** 每页数量 */
  pageSize?: number;
  /** 排序方式 */
  order?: "asc" | "desc";
  /** 排序字段 */
  orderBy?: string;
  /** 地区名称 */
  name?: string;
  /** 层级类型 */
  level?: "country" | "province" | "city" | "district" | "street";
  /** 深度 */
  depth?: number;
  /** 地区编码 */
  code?: string;
  /** 父级ID */
  parentId?: string;
};
```

**TypeScript类型定义原则：**
1. 枚举字段使用联合类型（`"country" | "province" | ...`）
2. 可选字段使用`?:`标记
3. 数组使用`Array<T>`或`T[]`
4. 添加JSDoc注释（`/** 注释 */`）
5. 保持与后端模型一致

#### 7.1.3 API函数

```typescript
/** 获取所有地区列表 */
export const getAreaList = () => {
  return http.get<AreaListResult, null>("/api/v1/areas");
};

/** 获取地区分页列表 */
export const getAreaListWithPagination = (params?: GetAreasParams) => {
  return http.get<AreaListResult, GetAreasParams>("/api/v1/areas/page", {
    params
  });
};

/** 获取单个地区 */
export const getArea = (id: string) => {
  return http.get<AreaResult, null>(`/api/v1/areas/${id}`);
};

/** 创建地区 */
export const createArea = (data: CreateAreaRequest) => {
  return http.post<AreaResult, CreateAreaRequest>("/api/v1/areas", {
    data
  });
};

/** 更新地区 */
export const updateArea = (id: string, data: UpdateAreaRequest) => {
  return http.put<AreaResult, UpdateAreaRequest>(`/api/v1/areas/${id}`, {
    data
  });
};

/** 删除地区 */
export const deleteArea = (id: string) => {
  return http.delete<{ success: boolean; message: string }, null>(
    `/api/v1/areas/${id}`
  );
};

/** 获取地区树形结构 */
export const getAreaTree = () => {
  return http.get<AreaListResult, null>("/api/v1/areas/tree");
};

/** 根据父级ID获取下一级地区 */
export const getAreasByParentId = (parentId: string) => {
  return http.get<AreaListResult, { parentId: string }>(
    "/api/v1/areas/children",
    {
      params: { parentId }
    }
  );
};

/** 根据层级类型获取地区 */
export const getAreasByLevel = (
  level: "country" | "province" | "city" | "district" | "street"
) => {
  return http.get<AreaListResult, { level: string }>("/api/v1/areas/level", {
    params: { level }
  });
};

/** 根据深度获取地区 */
export const getAreasByDepth = (depth: number) => {
  return http.get<AreaListResult, { depth: number }>("/api/v1/areas/depth", {
    params: { depth }
  });
};
```

**API函数模板：**
```typescript
export const functionName = (params) => {
  return http.method<ResponseType, RequestType>(url, config);
};
```

- **GET**: 使用`params`传递查询参数
- **POST/PUT**: 使用`data`传递请求体
- **DELETE**: 通常无请求体
- **路径参数**: 使用模板字符串拼接（`` `/api/v1/areas/${id}` ``）

## 关键设计原则总结

### 1. 命名规范

| 层级 | 文件名 | 结构体/函数 |
|------|--------|------------|
| Schema | `area.go` | `type Area struct` |
| Models | `area.go` | `CreateAreaRequest`, `AreaResponse` |
| Funcs | `area_func.go` | `type AreaFuncs struct`, `GetAreaByID()` |
| Handlers | `area_handler.go` | `type AreaHandler struct`, `GetArea()` |
| Routes | `area.go` | `setupAreaRoutes()` |
| API | `area.ts` | `getArea()`, `createArea()` |

### 2. 字段类型映射

| Ent Schema | Go Model | TypeScript | 说明 |
|------------|----------|------------|------|
| `field.String` | `string` | `string` | - |
| `field.Int` | `int` | `number` | - |
| `field.Float` | `float64` | `number` | - |
| `field.Uint64` (ID) | `string` | `string` | JSON友好 |
| `field.Enum` | `string` | `"a" \| "b"` | 联合类型 |
| `field.Time` | `string` | `string` | 格式化后 |
| `field.Bool` | `bool` | `boolean` | - |
| `.Optional()` | `*T` / `omitempty` | `T?` | 可选字段 |

### 3. RESTful路径设计

```
GET    /areas           - 获取所有
GET    /areas/page      - 分页获取
GET    /areas/:id       - 获取单个
POST   /areas           - 创建
PUT    /areas/:id       - 更新
DELETE /areas/:id       - 删除

GET    /areas/tree      - 树形结构（业务特定）
GET    /areas/children  - 子级查询（业务特定）
GET    /areas/level     - 层级查询（业务特定）
GET    /areas/depth     - 深度查询（业务特定）
```

### 4. 错误处理层次

```
Handler层: 参数验证、格式转换
    ↓
Funcs层: 业务逻辑、数据验证
    ↓
Ent层: 数据库约束、唯一性检查
```

### 5. 数据流向

```
请求 → Handler(参数提取) → Funcs(业务逻辑) → Ent(数据库) 
                                              ↓
响应 ← Handler(JSON) ← Funcs(转换) ← Ent(实体)
```

## 快速检查清单

开发完成后，检查以下项目：

### Models层
- [ ] CreateRequest包含所有必填字段
- [ ] UpdateRequest所有字段可选
- [ ] Response包含所有字段+关联
- [ ] GetRequest包含分页+查询条件
- [ ] 枚举字段有验证标签

### Funcs层
- [ ] GetAll() 预加载关联
- [ ] GetByID() 处理NotFound
- [ ] Create() 可选字段单独处理
- [ ] Update() 部分更新逻辑
- [ ] Delete() 级联删除（如需要）
- [ ] GetWithPagination() 完整实现
- [ ] 业务特定查询函数
- [ ] Convert函数（实体→响应）

### Handlers层
- [ ] 所有方法有Swagger注释
- [ ] 参数验证完整
- [ ] 错误分类处理
- [ ] 统一返回格式
- [ ] HTTP状态码正确

### Routes层
- [ ] 路由顺序正确（精确在前）
- [ ] 已注册到主路由
- [ ] 路由组命名清晰

### TypeScript层
- [ ] 类型定义完整
- [ ] 枚举使用联合类型
- [ ] API函数类型安全
- [ ] JSDoc注释完整
- [ ] 导出所有类型和函数

## 常见陷阱和注意事项

### 1. ID类型转换
- Go: `uint64` ↔ `string`
- JSON: 始终使用`string`
- 使用工具函数: `utils.StringToUint64()`, `utils.Uint64ToString()`

### 2. 枚举类型处理
- Schema: `field.Enum("level").Values("a", "b")`
- Go: `area.Level(string)` 类型转换
- TypeScript: `"a" | "b"` 联合类型
- 验证: Handler层验证枚举值

### 3. 可选字段处理
- Create: 检查非空字符串 / 非nil指针
- Update: 所有字段都检查
- Response: 使用`omitempty`

### 4. 时间字段
- 存储: `time.Time`
- 传输: `string` (格式化)
- 使用: `utils.FormatDateTime()`

### 5. 分页查询
- 先Count再查询
- 默认值设置（page=1, pageSize=10）
- 计算totalPages
- 返回分页元信息

### 6. 树形结构
- 避免循环引用
- 关联数据简化（只包含关键字段）
- Children手动构建（不依赖WithChildren）

### 7. 软删除
- 查询时自动过滤（Ent处理）
- 唯一索引要包含delete_time
- 恢复功能需要特殊处理

### 8. 事务处理
- 级联操作使用事务
- 错误时回滚
- 成功时提交

## 代码模板总结

使用本指南时，按以下步骤操作：

1. **分析Schema** - 提取字段、关系、约束
2. **规划接口** - 标准CRUD + 业务特定
3. **实现Models** - Request + Response
4. **实现Funcs** - 业务逻辑 + 转换
5. **实现Handlers** - HTTP处理 + 验证
6. **配置Routes** - 路由组织 + 注册
7. **实现API** - TypeScript封装
8. **测试验证** - 功能测试 + 集成测试

每一层都有清晰的职责分离，代码结构一致，便于维护和扩展。

你在编写TODO时需要每一步建立todo然后一步步执行