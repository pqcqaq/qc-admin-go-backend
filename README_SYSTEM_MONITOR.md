# 系统监控功能

## 概述

系统监控功能用于实时监控服务器的系统状态，包括 CPU、内存、磁盘、网络等资源使用情况，并将数据存储到数据库中以供历史查询和分析。

## 后端实现

### 1. Schema 定义

文件位置：`database/schema/system_monitor.go`

定义了系统监控记录的数据结构，包含以下主要字段：

- **CPU 信息**：使用率、核心数
- **内存信息**：总量、已用、空闲、使用率
- **磁盘信息**：总量、已用、空闲、使用率
- **网络信息**：发送/接收字节数
- **系统信息**：操作系统、平台、版本、主机名、运行时间
- **Go 运行时信息**：Goroutine 数量、堆内存、GC 次数
- **系统负载**：1/5/15 分钟平均负载（仅 Unix 系统）

### 2. Funcs 实现

文件位置：`internal/funcs/system_monitor_func.go`

提供了完整的系统监控服务，包括：

#### 初始化函数

```go
// 在服务器启动后调用此函数初始化系统监控
func InitSystemMonitor(interval time.Duration, retentionDays int) error
```

**参数说明：**
- `interval`: 监控数据采集间隔（例如：30 * time.Second 表示每 30 秒采集一次）
- `retentionDays`: 数据保留天数（例如：7 表示保留最近 7 天的数据，更早的数据会被自动清理）

**功能：**
- 创建定时任务，按指定间隔采集系统状态
- 自动清理过期的历史数据
- 在后台协程中运行，不阻塞主程序

#### 停止监控

```go
func StopSystemMonitor()
```

停止系统监控定时任务。

#### 数据查询函数

```go
// 获取最新的系统状态
func (SystemMonitorFuncs) GetLatestSystemStatus(ctx context.Context) (*ent.SystemMonitor, error)

// 获取指定时间段的历史记录
func (SystemMonitorFuncs) GetSystemStatusHistory(ctx context.Context, limit int, hours int) ([]*ent.SystemMonitor, error)

// 根据时间范围获取系统状态
func (SystemMonitorFuncs) GetSystemStatusByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*ent.SystemMonitor, error)

// 获取系统状态统计摘要（包含平均值、最大值、最小值）
func (SystemMonitorFuncs) GetSystemStatsSummary(ctx context.Context, hours int) (map[string]interface{}, error)
```

### 3. 使用示例

在服务器启动文件（如 `cmd/api/main.go`）中添加初始化代码：

```go
import (
    "time"
    "go-backend/internal/funcs"
)

func main() {
    // ... 其他初始化代码 ...

    // 初始化系统监控
    // 每 30 秒采集一次数据，保留 7 天的历史记录
    if err := funcs.InitSystemMonitor(30*time.Second, 7); err != nil {
        log.Printf("Failed to initialize system monitor: %v", err)
    }

    // 确保程序退出时停止监控
    defer funcs.StopSystemMonitor()

    // ... 启动服务器 ...
}
```

### 4. 依赖安装

系统监控功能依赖 `gopsutil` 库，需要在 `go.mod` 中添加：

```bash
go get github.com/shirou/gopsutil/v3
```

主要使用的包：
- `github.com/shirou/gopsutil/v3/cpu` - CPU 信息
- `github.com/shirou/gopsutil/v3/mem` - 内存信息
- `github.com/shirou/gopsutil/v3/disk` - 磁盘信息
- `github.com/shirou/gopsutil/v3/net` - 网络信息
- `github.com/shirou/gopsutil/v3/host` - 主机信息
- `github.com/shirou/gopsutil/v3/load` - 系统负载

## 前端实现

### 页面位置

`src/views/system/monitor/index.vue`

### 功能特性

1. **实时状态卡片**
   - CPU 使用率和核心数
   - 内存使用率和容量
   - 磁盘使用率和容量
   - Goroutine 数量和 GC 次数
   - 带动态颜色的进度条

2. **系统信息展示**
   - 操作系统详细信息
   - 运行时信息
   - 网络统计
   - 系统运行时间

3. **历史数据图表**
   - CPU 使用率趋势图
   - 内存使用率趋势图
   - 磁盘使用率趋势图
   - Goroutine 数量趋势图
   - 支持切换时间范围（1小时/6小时/24小时/7天）

4. **自动刷新**
   - 每 30 秒自动刷新数据
   - 支持手动刷新

### 技术栈

- Vue 3 + TypeScript
- Element Plus（UI 组件）
- ECharts（图表库）
- Day.js（时间处理）

### API 接口说明

前端需要以下 API 接口（需要自行实现）：

```typescript
// 获取最新系统状态
GET /api/system/monitor/latest
Response: SystemStatus

// 获取历史数据
GET /api/system/monitor/history?limit=100&hours=1
Response: SystemStatus[]

// 根据时间范围获取数据
GET /api/system/monitor/range?start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z
Response: SystemStatus[]

// 获取统计摘要
GET /api/system/monitor/summary?hours=24
Response: {
  count: number;
  cpu: { avg: number; max: number; min: number };
  memory: { avg: number; max: number; min: number };
  disk: { avg: number; max: number; min: number };
  period: { start: string; end: string; hours: number };
}
```

### 数据类型定义

```typescript
interface SystemStatus {
  id: string;
  cpuUsagePercent: number;
  cpuCores: number;
  memoryTotal: number;
  memoryUsed: number;
  memoryFree: number;
  memoryUsagePercent: number;
  diskTotal: number;
  diskUsed: number;
  diskFree: number;
  diskUsagePercent: number;
  networkBytesSent: number;
  networkBytesRecv: number;
  os: string;
  platform: string;
  platformVersion: string;
  hostname: string;
  goroutinesCount: number;
  heapAlloc: number;
  heapSys: number;
  gcCount: number;
  loadAvg1?: number;
  loadAvg5?: number;
  loadAvg15?: number;
  uptime: number;
  recordedAt: string;
  createdAt: string;
}
```

## 数据库迁移

生成 Ent 代码并创建数据库表：

```bash
# 生成 Ent 代码
go generate ./database/ent

# 运行数据库迁移
# 根据你的项目配置执行相应的迁移命令
```

## 性能优化建议

1. **采集间隔**：建议设置为 30 秒到 1 分钟，避免过于频繁导致性能开销
2. **数据保留期**：根据实际需求设置，建议 7-30 天
3. **索引优化**：已在 Schema 中为 `recorded_at` 和 `created_at` 字段创建索引
4. **分区表**：如果数据量特别大，可以考虑使用数据库分区功能

## 监控告警

可以基于收集的数据实现告警功能：

```go
// 示例：检查 CPU 使用率是否超过阈值
func checkCPUAlert(ctx context.Context, threshold float64) error {
    status, err := SystemMonitorFuncs{}.GetLatestSystemStatus(ctx)
    if err != nil {
        return err
    }
    
    if status.CPUUsagePercent > threshold {
        // 发送告警通知
        log.Printf("Alert: CPU usage is %.2f%%, exceeds threshold %.2f%%", 
            status.CPUUsagePercent, threshold)
    }
    
    return nil
}
```

## 常见问题

### 1. 在 Windows 上某些指标无法获取

部分系统指标（如 `load average`）仅在 Unix 系统上可用，在 Windows 上会返回 `nil`。代码已经处理了这种情况。

### 2. 权限问题

某些系统信息可能需要管理员权限才能读取，请确保程序有足够的权限。

### 3. 磁盘路径

默认监控根目录（`/`），Windows 系统可能需要修改为 `C:\`：

```go
// 在 collectSystemMetrics 函数中修改
diskInfo, err := disk.Usage("C:\\")  // Windows
```

## 扩展建议

1. **多磁盘监控**：监控所有挂载的磁盘分区
2. **进程监控**：监控特定进程的资源使用
3. **数据库连接池监控**：监控数据库连接状态
4. **自定义指标**：添加业务相关的自定义监控指标
5. **导出功能**：支持导出监控数据为 CSV 或其他格式
6. **对比分析**：支持多个时间段的数据对比

## 许可证

与主项目相同
