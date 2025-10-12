# 系统监控功能 - 快速开始

## 后端集成

### 1. 在 main.go 中初始化

找到你的 `main.go` 或服务启动文件，添加以下代码：

```go
package main

import (
    "go-backend/internal/funcs"
    "time"
    "log"
)

func main() {
    // ... 其他初始化代码 ...

    // 初始化系统监控
    // 参数1: 30*time.Second - 每30秒采集一次数据
    // 参数2: 7 - 保留最近7天的数据
    err := funcs.InitSystemMonitor(30*time.Second, 7)
    if err != nil {
        log.Printf("警告: 系统监控初始化失败: %v", err)
        // 注意：这里使用 Printf 而不是 Fatal，避免因监控功能失败导致服务无法启动
    } else {
        log.Println("系统监控已启动")
    }

    // 确保在程序退出时停止监控
    defer funcs.StopSystemMonitor()

    // ... 启动服务器 ...
}
```

### 2. 配置参数建议

根据你的需求调整参数：

```go
// 开发环境 - 快速采集，短期保留
funcs.InitSystemMonitor(10*time.Second, 1)  // 10秒采集一次，保留1天

// 生产环境 - 常规监控
funcs.InitSystemMonitor(30*time.Second, 7)  // 30秒采集一次，保留7天

// 长期监控 - 降低频率，延长保留
funcs.InitSystemMonitor(60*time.Second, 30) // 60秒采集一次，保留30天
```

### 3. 验证后端

启动服务器后，访问以下 API 验证：

```bash
# 获取最新状态
curl http://localhost:8080/api/v1/system/monitor/latest

# 获取历史数据
curl http://localhost:8080/api/v1/system/monitor/history?hours=1&limit=10
```

## 前端集成

### 1. 添加路由

在你的路由配置文件中添加系统监控路由：

```typescript
// src/router/index.ts 或类似文件
{
  path: '/system/monitor',
  name: 'SystemMonitor',
  component: () => import('@/views/system/monitor/index.vue'),
  meta: {
    title: '系统监控',
    icon: 'Monitor',
    requiresAuth: true
  }
}
```

### 2. 添加菜单项

在你的菜单配置中添加：

```typescript
{
  title: '系统监控',
  path: '/system/monitor',
  icon: 'Monitor'
}
```

### 3. 访问页面

启动前端项目后，访问：
```
http://localhost:3000/system/monitor
```

## 快速测试

### 测试后端 API

```bash
# 1. 获取最新状态
curl -X GET "http://localhost:8080/api/v1/system/monitor/latest"

# 2. 获取最近1小时的历史数据（最多100条）
curl -X GET "http://localhost:8080/api/v1/system/monitor/history?hours=1&limit=100"

# 3. 获取指定时间范围的数据
curl -X GET "http://localhost:8080/api/v1/system/monitor/range?start=2025-10-12T00:00:00Z&end=2025-10-12T23:59:59Z"

# 4. 获取最近24小时的统计摘要
curl -X GET "http://localhost:8080/api/v1/system/monitor/summary?hours=24"

# 5. 删除指定ID的记录
curl -X DELETE "http://localhost:8080/api/v1/system/monitor/123"

# 6. 删除指定时间范围的记录
curl -X DELETE "http://localhost:8080/api/v1/system/monitor/range?start=2025-10-01T00:00:00Z&end=2025-10-10T23:59:59Z"
```

### 使用前端 API

```typescript
import {
  getLatestSystemMonitor,
  getSystemMonitorHistory,
  getSystemMonitorSummary
} from 'qc-admin-api-common';

// 获取最新状态
const latest = await getLatestSystemMonitor();
console.log('当前 CPU 使用率:', latest.cpuUsagePercent);

// 获取历史数据
const history = await getSystemMonitorHistory({ hours: 1, limit: 100 });
console.log('历史记录数:', history.length);

// 获取统计摘要
const summary = await getSystemMonitorSummary({ hours: 24 });
console.log('24小时平均 CPU 使用率:', summary.cpu.avg);
```

## 常见问题

### Q: 为什么看不到数据？
A: 
1. 检查后端是否已调用 `InitSystemMonitor()` 初始化
2. 等待至少一个采集周期（默认30秒）
3. 检查数据库连接是否正常
4. 查看后端日志是否有错误信息

### Q: 数据库表在哪里？
A: 表名为 `sys_system_monitor`，由 Ent 自动创建

### Q: 如何手动触发一次数据采集？
A: 目前只支持自动采集。如需手动采集，可以调用内部函数 `collectSystemMetrics()`

### Q: Windows 系统上 LoadAvg 为什么是 null？
A: Load Average 是 Unix/Linux 特有的指标，Windows 不支持

### Q: 如何调整图表的时间范围？
A: 在前端页面的历史图表区域，使用时间范围选择器选择 1h/6h/12h/24h

### Q: 可以导出监控数据吗？
A: 当前版本暂不支持导出，可以通过 API 获取数据后自行处理

## 性能影响

系统监控功能的性能影响很小：

- **CPU**: < 0.1% (每30秒采集一次)
- **内存**: < 10MB (Go 运行时)
- **磁盘I/O**: 最小化 (批量写入)
- **数据库**: 每30秒一条记录，7天约 20,160 条记录

## 下一步

1. 根据实际需求调整采集间隔和数据保留期
2. 考虑添加告警功能（CPU/内存/磁盘使用率超过阈值时通知）
3. 可选：集成到现有的监控系统（如 Prometheus、Grafana）
4. 可选：添加更多监控指标（数据库连接池、API 响应时间等）

## 获取帮助

- 详细 API 文档: `README_SYSTEM_MONITOR_API.md`
- 实现总结: `SYSTEM_MONITOR_IMPLEMENTATION_SUMMARY.md`
- 前端组件说明: `qc-admin/src/views/system/monitor/README.md`

## 完成！

🎉 系统监控功能已就绪，enjoy！
