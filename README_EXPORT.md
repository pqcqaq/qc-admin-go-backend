# Database Export 功能使用说明

## 功能概述

这个功能可以通过反射自动发现数据库中的所有实体，并将每个表的数据导出为JSON格式的文件。每个实体会生成一个对应的JSON文件，文件名为实体名的小写形式。

## 主要特性

- 自动发现所有实体：通过反射扫描 `*Client` 字段
- 支持自定义导出配置：输出目录、格式化选项、包含/排除列表等
- 错误处理和结果统计：详细的导出结果和错误信息
- 支持超时控制：可以设置导出操作的超时时间
- 灵活的过滤选项：可以选择性导出特定表或排除某些表

## 基本使用

### 1. 使用默认配置导出所有表

```go
package main

import (
    "go-backend/pkg/database"
    "go-backend/pkg/configs"
)

func main() {
    // 初始化数据库连接
    config := &configs.DatabaseConfig{
        Driver: "postgres",
        DSN:    "postgres://user:password@localhost/dbname?sslmode=disable",
        // ... 其他配置
    }
    
    client := database.MustNewClient(config)
    defer client.Close()
    
    // 导出所有表到默认目录 "./exports"
    result, err := database.ExportAllTablesWithDefaultConfig(client)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("导出完成: %d个表, 成功: %d, 失败: %d\n", 
        result.TotalEntities, result.SuccessCount, result.FailedCount)
}
```

### 2. 使用全局客户端导出

```go
package main

import (
    "go-backend/pkg/database"
    "go-backend/pkg/configs"
)

func main() {
    // 初始化全局数据库实例
    config := &configs.DatabaseConfig{
        // ... 配置
    }
    database.InitInstance(config)
    
    // 使用全局客户端导出所有表
    result, err := database.ExportAllTablesGlobal("./my_exports")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("导出结果: %+v\n", result)
}
```

### 3. 导出指定的表

```go
// 只导出 User 和 Role 表
entityNames := []string{"User", "Role"}
result, err := database.ExportSpecificTables(client, entityNames, "./specific_exports")
if err != nil {
    panic(err)
}
```

### 4. 排除特定表

```go
// 导出所有表，但排除日志和验证码表
excludeList := []string{"Logging", "VerifyCode"}
result, err := database.ExportExcludeTables(client, excludeList, "./filtered_exports")
if err != nil {
    panic(err)
}
```

### 5. 自定义配置导出

```go
config := &database.ExportConfig{
    OutputDir:    "./custom_exports",
    PrettyFormat: false, // 紧凑JSON格式
    Context:      context.WithTimeout(context.Background(), time.Minute*5),
    ExcludeEntities: []string{"Logging", "VerifyCode"},
    // IncludeEntities: []string{"User", "Role"}, // 或者使用包含列表
}

result, err := database.ExportAllTables(client, config)
if err != nil {
    panic(err)
}
```

## 配置选项

### ExportConfig 结构

```go
type ExportConfig struct {
    // OutputDir 输出目录，默认为 "./exports"
    OutputDir string
    
    // PrettyFormat 是否格式化JSON输出，默认为true
    PrettyFormat bool
    
    // Context 上下文，用于控制超时等
    Context context.Context
    
    // ExcludeEntities 排除的实体名称列表
    ExcludeEntities []string
    
    // IncludeEntities 仅包含的实体名称列表（如果设置，则只导出这些实体）
    IncludeEntities []string
}
```

## 导出结果

### ExportResult 结构

```go
type ExportResult struct {
    TotalEntities   int                   // 总实体数量
    SuccessCount    int                   // 成功导出的数量
    FailedCount     int                   // 失败的数量
    OutputDirectory string                // 输出目录
    Results         []EntityExportResult  // 每个实体的详细结果
}

type EntityExportResult struct {
    EntityName  string // 实体名称
    FilePath    string // 导出文件路径
    RecordCount int    // 记录数量
    Success     bool   // 是否成功
    Error       string // 错误信息（如果有）
}
```

## 命令行工具

项目提供了一个命令行工具 `cmd/export/main.go`，可以直接从命令行导出数据：

```bash
# 基本用法
go run cmd/export/main.go

# 指定配置文件和输出目录
go run cmd/export/main.go -config config.yaml -output ./my_exports

# 只导出指定的表
go run cmd/export/main.go -include "User,Role,Permission"

# 排除特定的表
go run cmd/export/main.go -exclude "Logging,VerifyCode"

# 紧凑格式输出
go run cmd/export/main.go -pretty=false

# 设置超时时间
go run cmd/export/main.go -timeout 5m

# 显示详细结果
go run cmd/export/main.go -result
```

## 输出文件格式

导出的文件按照以下规则命名和组织：

```text
exports/
├── user.json           # User 实体的数据
├── role.json           # Role 实体的数据
├── permission.json     # Permission 实体的数据
├── attachment.json     # Attachment 实体的数据
├── ...
└── export_result.json  # 导出结果统计（命令行工具生成）
```

每个JSON文件包含对应表的所有记录的数组：

```json
[
  {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "created_at": "2023-01-01T00:00:00Z",
    ...
  },
  {
    "id": 2,
    "username": "user",
    "email": "user@example.com",
    "created_at": "2023-01-02T00:00:00Z",
    ...
  }
]
```

## 注意事项

1. **大数据量**：对于包含大量数据的表，导出可能需要较长时间，建议设置合适的超时时间
2. **磁盘空间**：确保输出目录有足够的磁盘空间存储导出的JSON文件
3. **内存使用**：导出大表时会将所有数据加载到内存中，注意内存使用情况
4. **权限**：确保程序对输出目录有写入权限
5. **数据敏感性**：导出的文件包含原始数据，注意数据安全和隐私保护

## 错误处理

函数会返回详细的错误信息，包括：

- 数据库连接错误
- 查询执行错误
- JSON序列化错误
- 文件写入错误

每个实体的导出结果都会单独记录，即使某些表导出失败，其他表的导出仍会继续进行。
