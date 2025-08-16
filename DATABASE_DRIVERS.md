# 数据库驱动条件编译指南

本项目使用条件编译来控制数据库驱动的加载，这样可以减少最终二进制文件的大小，并且只包含实际需要的数据库驱动。

## 支持的数据库类型

| 数据库 | 编译标签 | 驱动包 | 说明 |
|--------|----------|---------|------|
| SQLite | `sqlite` | `github.com/mattn/go-sqlite3` | 轻量级文件数据库 |
| MySQL | `mysql` | `github.com/go-sql-driver/mysql` | 流行的关系型数据库 |
| PostgreSQL | `postgres` | `github.com/lib/pq` | 高级关系型数据库 |
| ClickHouse | `clickhouse` | `github.com/ClickHouse/clickhouse-go/v2` | 列式分析数据库 |
| SQL Server | `sqlserver` | `github.com/denisenkom/go-mssqldb` | Microsoft SQL Server |
| Oracle | `oracle` | `github.com/godror/godror` | Oracle 数据库 |

## 编译选项

### 1. 包含所有驱动（默认）
```bash
go build -tags "all" -o server .
# 或者使用 Makefile
make build
```

### 2. 仅包含特定驱动
```bash
# 仅包含 SQLite
go build -tags "sqlite" -o server-sqlite .
make build DB_TAGS=sqlite

# 仅包含 MySQL
go build -tags "mysql" -o server-mysql .
make build DB_TAGS=mysql

# 仅包含 PostgreSQL
go build -tags "postgres" -o server-postgres .
make build DB_TAGS=postgres
```

### 3. 包含多个驱动
```bash
# 包含 MySQL 和 PostgreSQL
go build -tags "mysql postgres" -o server-multi .
make build DB_TAGS="mysql postgres"

# 包含 SQLite 和 ClickHouse
go build -tags "sqlite clickhouse" -o server-analytics .
make build DB_TAGS="sqlite clickhouse"
```

### 4. 使用预定义的构建目标
```bash
# 构建所有单驱动版本
make build-all-variants

# 构建特定版本
make build-sqlite    # 仅 SQLite
make build-mysql     # 仅 MySQL
make build-postgres  # 仅 PostgreSQL
```

## 二进制文件大小对比

不同编译选项生成的二进制文件大小对比（大致估计）：

- 仅 SQLite: ~15MB
- 仅 MySQL: ~18MB
- 仅 PostgreSQL: ~20MB
- 仅 ClickHouse: ~25MB
- 包含所有驱动: ~35MB

## 配置文件示例

参考 `config.example.yaml` 文件，了解如何配置不同类型的数据库。

### SQLite 配置
```yaml
database:
  driver: "sqlite3"
  dsn: "file:ent.db?cache=shared&_fk=1"
```

### MySQL 配置
```yaml
database:
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local"
```

### PostgreSQL 配置
```yaml
database:
  driver: "postgres"
  dsn: "host=localhost user=username password=password dbname=database_name port=5432 sslmode=disable TimeZone=Asia/Shanghai"
```

### ClickHouse 配置
```yaml
database:
  driver: "clickhouse"
  dsn: "tcp://localhost:9000?database=default&username=default&password="
```

## 运行时检查

应用程序启动时会自动检查配置的数据库驱动是否已编译包含：

```go
// 检查驱动是否支持
supportedDrivers := database.GetSupportedDrivers()
fmt.Println("支持的数据库驱动:", supportedDrivers)

// 列出已加载的驱动
fmt.Println(database.ListLoadedDrivers())
```

## 注意事项

1. **默认驱动**: 如果没有指定任何编译标签，系统会自动包含 SQLite 作为默认驱动。

2. **驱动验证**: 应用程序启动时会验证配置文件中指定的驱动是否已编译包含，如果不匹配会返回错误。

3. **CGO 依赖**: 
   - SQLite 驱动需要 CGO 支持
   - Oracle 驱动需要安装 Oracle Instant Client

4. **交叉编译**: 使用 CGO 的驱动（如 SQLite）在交叉编译时可能需要额外配置。

## 开发环境配置

在开发环境中，推荐使用包含所有驱动的版本以便测试：

```bash
# 开发模式运行
make run

# 或者指定特定驱动进行测试
make run DB_TAGS=mysql
```

## 生产环境部署

在生产环境中，建议只包含实际需要的数据库驱动以减少部署包大小：

```bash
# 生产环境只使用 PostgreSQL
make build DB_TAGS=postgres

# 生产环境使用 MySQL
make build DB_TAGS=mysql
```

## 故障排除

### 1. 驱动未找到错误
如果遇到类似错误：
```
unsupported database driver: mysql. Supported drivers: [sqlite3]
```

这说明二进制文件没有包含 MySQL 驱动，需要重新编译：
```bash
make build DB_TAGS=mysql
```

### 2. 查看已编译的驱动
程序启动时会输出已加载的驱动信息，或者可以调用：
```go
fmt.Println(database.ListLoadedDrivers())
```
