# 系统监控 API 文档

## 概述

系统监控功能提供了对服务器实时状态和历史数据的监控能力。该功能会自动收集系统的 CPU、内存、磁盘、网络等关键指标，并提供了完整的 RESTful API 来查询和管理这些数据。

## 初始化

在服务器启动时，需要初始化系统监控功能：

```go
import "go-backend/internal/funcs"

// 初始化系统监控
// 参数1: 收集间隔时间（例如：30 * time.Second 表示每30秒收集一次）
// 参数2: 数据保留天数（例如：7 表示保留最近7天的数据，超过的会自动清理）
err := funcs.InitSystemMonitor(30 * time.Second, 7)
if err != nil {
    log.Fatal("Failed to initialize system monitor:", err)
}

// 在服务器关闭时停止监控
defer funcs.StopSystemMonitor()
```

## API 端点

所有 API 端点的基础路径为：`/api/v1/system/monitor`

### 1. 获取最新系统状态

获取系统最新的监控数据。

**请求：**
```
GET /system/monitor/latest
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "cpuUsagePercent": 45.5,
    "cpuCores": 8,
    "memoryTotal": 17179869184,
    "memoryUsed": 8589934592,
    "memoryFree": 8589934592,
    "memoryUsagePercent": 50.0,
    "diskTotal": 512110190592,
    "diskUsed": 256055095296,
    "diskFree": 256055095296,
    "diskUsagePercent": 50.0,
    "networkBytesSent": 1024000000,
    "networkBytesRecv": 2048000000,
    "os": "linux",
    "platform": "ubuntu",
    "platformVersion": "22.04",
    "hostname": "server01",
    "goroutinesCount": 150,
    "heapAlloc": 10485760,
    "heapSys": 20971520,
    "gcCount": 25,
    "loadAvg1": 1.5,
    "loadAvg5": 1.2,
    "loadAvg15": 1.0,
    "uptime": 864000,
    "recordedAt": "2025-10-12T10:30:00Z",
    "createdAt": "2025-10-12T10:30:00Z",
    "updatedAt": "2025-10-12T10:30:00Z"
  }
}
```

### 2. 获取历史记录

获取指定时间范围内的系统监控历史数据。

**请求：**
```
GET /system/monitor/history?limit=100&hours=1
```

**查询参数：**
- `limit` (可选): 返回的记录数量，范围 1-1000，默认 100
- `hours` (可选): 查询最近多少小时的数据，范围 1-168（7天），默认 1

**响应示例：**
```json
{
  "success": true,
  "data": [
    {
      "id": "123",
      "cpuUsagePercent": 45.5,
      "memoryUsagePercent": 50.0,
      // ... 其他字段
    },
    {
      "id": "122",
      "cpuUsagePercent": 43.2,
      "memoryUsagePercent": 48.5,
      // ... 其他字段
    }
  ],
  "count": 120
}
```

### 3. 按时间范围查询

根据指定的开始和结束时间获取系统监控数据。

**请求：**
```
GET /system/monitor/range?start=2025-10-12T00:00:00Z&end=2025-10-12T23:59:59Z
```

**查询参数：**
- `start` (必需): 开始时间，ISO 8601 格式
- `end` (必需): 结束时间，ISO 8601 格式

**响应示例：**
```json
{
  "success": true,
  "data": [
    // ... 监控记录数组
  ],
  "count": 2880
}
```

### 4. 获取统计摘要

获取指定时间范围内的系统监控统计信息，包括 CPU、内存、磁盘的平均值、最大值、最小值。

**请求：**
```
GET /system/monitor/summary?hours=24
```

**查询参数：**
- `hours` (可选): 查询最近多少小时的数据，范围 1-720（30天），默认 24

**响应示例：**
```json
{
  "success": true,
  "data": {
    "count": 2880,
    "cpu": {
      "avg": 45.5,
      "max": 85.2,
      "min": 15.3
    },
    "memory": {
      "avg": 50.0,
      "max": 78.5,
      "min": 32.1
    },
    "disk": {
      "avg": 50.0,
      "max": 52.3,
      "min": 49.8
    },
    "period": {
      "start": "2025-10-11T10:30:00Z",
      "end": "2025-10-12T10:30:00Z",
      "hours": 24.0
    }
  }
}
```

### 5. 删除监控记录

删除指定 ID 的系统监控记录。

**请求：**
```
DELETE /system/monitor/{id}
```

**路径参数：**
- `id`: 监控记录的 ID

**响应示例：**
```json
{
  "success": true,
  "message": "系统监控记录删除成功"
}
```

### 6. 按时间范围删除

删除指定时间范围内的所有系统监控记录。

**请求：**
```
DELETE /system/monitor/range?start=2025-10-01T00:00:00Z&end=2025-10-10T23:59:59Z
```

**查询参数：**
- `start` (必需): 开始时间，ISO 8601 格式
- `end` (必需): 结束时间，ISO 8601 格式

**响应示例：**
```json
{
  "success": true,
  "data": {
    "deleted": 28800
  },
  "message": "系统监控记录批量删除成功"
}
```

## 数据模型

### SystemMonitorResponse

| 字段 | 类型 | 描述 |
|------|------|------|
| id | string | 记录ID |
| cpuUsagePercent | number | CPU使用率(%) |
| cpuCores | number | CPU核心数 |
| memoryTotal | number | 总内存(字节) |
| memoryUsed | number | 已使用内存(字节) |
| memoryFree | number | 空闲内存(字节) |
| memoryUsagePercent | number | 内存使用率(%) |
| diskTotal | number | 总磁盘空间(字节) |
| diskUsed | number | 已使用磁盘空间(字节) |
| diskFree | number | 空闲磁盘空间(字节) |
| diskUsagePercent | number | 磁盘使用率(%) |
| networkBytesSent | number | 网络发送字节数 |
| networkBytesRecv | number | 网络接收字节数 |
| os | string | 操作系统 |
| platform | string | 平台 |
| platformVersion | string | 平台版本 |
| hostname | string | 主机名 |
| goroutinesCount | number | Goroutine数量 |
| heapAlloc | number | 堆内存分配(字节) |
| heapSys | number | 堆系统内存(字节) |
| gcCount | number | GC次数 |
| loadAvg1 | number? | 1分钟平均负载(可选) |
| loadAvg5 | number? | 5分钟平均负载(可选) |
| loadAvg15 | number? | 15分钟平均负载(可选) |
| uptime | number | 系统运行时间(秒) |
| recordedAt | string | 记录时间 |
| createdAt | string | 创建时间 |
| updatedAt | string | 更新时间 |

## 错误处理

所有 API 在发生错误时会返回如下格式：

```json
{
  "success": false,
  "message": "错误信息描述",
  "error": "详细错误信息"
}
```

常见的 HTTP 状态码：
- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

## 前端集成

前端可以使用 `qc-admin-api-common` 包中的 API 函数：

```typescript
import {
  getLatestSystemMonitor,
  getSystemMonitorHistory,
  getSystemMonitorByRange,
  getSystemMonitorSummary,
  deleteSystemMonitor,
  deleteSystemMonitorByRange
} from '@/api-common';

// 获取最新状态
const latest = await getLatestSystemMonitor();

// 获取最近1小时的历史数据，最多100条
const history = await getSystemMonitorHistory({ limit: 100, hours: 1 });

// 获取指定时间范围的数据
const rangeData = await getSystemMonitorByRange({
  start: '2025-10-12T00:00:00Z',
  end: '2025-10-12T23:59:59Z'
});

// 获取最近24小时的统计摘要
const summary = await getSystemMonitorSummary({ hours: 24 });
```

## 性能建议

1. **数据保留策略**: 根据实际需求设置合理的数据保留天数，避免数据库过大
2. **查询限制**: 使用 `limit` 参数限制返回的数据量，避免一次性加载过多数据
3. **时间范围**: 查询历史数据时，尽量使用较小的时间范围
4. **定期清理**: 利用自动清理功能，或手动删除过期数据
5. **采集间隔**: 根据实际需求设置合理的采集间隔，不宜过于频繁

## 监控指标说明

### CPU 使用率
- 值范围：0-100%
- 说明：所有 CPU 核心的平均使用率

### 内存使用率
- 值范围：0-100%
- 说明：已使用内存占总内存的百分比

### 磁盘使用率
- 值范围：0-100%
- 说明：已使用磁盘空间占总空间的百分比（通常监控根分区 `/`）

### 系统负载 (Load Average)
- loadAvg1: 过去1分钟的平均负载
- loadAvg5: 过去5分钟的平均负载
- loadAvg15: 过去15分钟的平均负载
- 说明：仅在 Unix 系统上可用，Windows 系统此字段为 null

### Go 运行时指标
- goroutinesCount: 当前 Goroutine 数量
- heapAlloc: 堆内存分配量
- heapSys: 从系统获取的堆内存
- gcCount: GC 执行次数

## 故障排查

### 问题：无法获取最新数据
1. 检查系统监控是否已初始化
2. 查看日志确认数据收集是否正常
3. 检查数据库连接是否正常

### 问题：Load Average 值为 null
- 这是正常的，Windows 系统不支持 Load Average 指标

### 问题：数据库增长过快
- 调整数据保留天数
- 增加采集间隔时间
- 手动删除历史数据

## 相关文件

- Schema: `database/schema/system_monitor.go`
- Functions: `internal/funcs/system_monitor_func.go`
- Models: `shared/models/system_monitor.go`
- Handlers: `internal/handlers/system_monitor_handler.go`
- Routes: `internal/routes/system_monitor.go`
- Frontend API: `qc-admin-api-common/src/system_monitor.ts`
- Frontend Components: `qc-admin/src/views/system/monitor/`
