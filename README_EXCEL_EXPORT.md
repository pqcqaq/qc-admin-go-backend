# Excel导出功能说明

## 概述

新增了Excel导出功能，支持将扫描记录导出为Excel文件并下载。该功能使用流式处理以提升性能，支持大量数据的导出。

## 主要组件

### 1. Excel处理器 (`pkg/excel/excel.go`)

提供了一个通用的Excel处理器，支持：
- 自定义列配置
- 字段格式化器
- 流式Excel生成
- 自动样式设置

#### 主要类型

```go
type ColumnConfig struct {
    Header    string                      // 列标题
    Width     float64                     // 列宽度
    Formatter func(any) string    // 格式化函数
    FieldName string                      // 对应的字段名
}

type ExcelProcessor struct {
    SheetName string
    Columns   []ColumnConfig
}
```

#### 内置格式化器

- `TimeFormatter(layout string)` - 时间格式化
- `BoolFormatter(trueText, falseText string)` - 布尔值格式化
- `NumberFormatter(precision int)` - 数字格式化

### 2. 新增API接口

#### 导出扫描记录为Excel

- **URL**: `GET /api/scans/export`
- **参数**: 与分页查询相同的参数
  - `content` - 按内容模糊搜索
  - `success` - 按扫描结果过滤
  - `begin_time` - 开始时间
  - `end_time` - 结束时间
  - `order` - 排序方式 (asc/desc)
  - `order_by` - 排序字段 (create_time/update_time)

- **响应**: 直接返回Excel文件流，浏览器会自动下载

## 使用示例

### 1. 基本导出

```bash
GET /api/scans/export
```

### 2. 按条件导出

```bash
GET /api/scans/export?success=true&begin_time=2024-01-01&end_time=2024-12-31
```

### 3. 按内容搜索导出

```bash
GET /api/scans/export?content=测试&order=asc&order_by=create_time
```

## Excel输出格式

导出的Excel文件包含以下列：

| 列名 | 字段 | 说明 |
|------|------|------|
| ID | ID | 扫描记录ID |
| 扫描内容 | Content | 扫描的内容 |
| 是否成功 | Success | 成功/失败 |
| 创建时间 | CreateTime | YYYY-MM-DD HH:mm:ss |
| 图片ID | ImageId | 关联的图片ID |
| 图片URL | ImageUrl | 图片访问地址 |

## 性能特性

1. **流式处理**: 直接将Excel数据写入HTTP响应流，不会将整个文件加载到内存
2. **分页限制**: 默认最大导出10000条记录，可以通过修改`PageSize`调整
3. **内存优化**: 使用excelize库的流式API，减少内存占用

## 扩展使用

如果需要为其他数据类型添加Excel导出，可以参考以下步骤：

### 1. 定义列配置

```go
columns := []excel.ColumnConfig{
    {
        Header:    "用户名",
        Width:     20,
        FieldName: "Username",
    },
    {
        Header:    "注册时间",
        Width:     25,
        FieldName: "CreatedAt",
        Formatter: excel.TimeFormatter("2006-01-02 15:04:05"),
    },
    {
        Header:    "是否活跃",
        Width:     15,
        FieldName: "IsActive",
        Formatter: excel.BoolFormatter("活跃", "不活跃"),
    },
}
```

### 2. 创建处理器并生成Excel

```go
processor := excel.NewExcelProcessor("用户列表", columns)
file, err := processor.GenerateExcelStream(userData)
if err != nil {
    // 处理错误
}

// 设置响应头并写入流
c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
c.Header("Content-Disposition", "attachment; filename="+filename)
file.Write(c.Writer)
```

## 依赖库

- `github.com/xuri/excelize/v2` - Excel文件处理库

## 测试

运行测试：

```bash
go test ./pkg/excel/... -v
```

测试覆盖：
- Excel文件生成功能
- 文件名生成
- 列号转换功能
- 格式化器功能
