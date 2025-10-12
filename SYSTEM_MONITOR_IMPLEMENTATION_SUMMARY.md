# 系统监控功能实现总结

## 已完成的工作

### 1. 后端实现

#### Schema (数据库模型)
- **文件**: `database/schema/system_monitor.go`
- **功能**: 定义了系统监控数据的数据库结构
- **字段**: 包含 CPU、内存、磁盘、网络、系统信息、Go 运行时信息等 25+ 个字段
- **索引**: 对 `recorded_at` 和 `create_time` 字段建立了降序索引，优化查询性能

#### Models (数据模型)
- **文件**: `shared/models/system_monitor.go`
- **功能**: 定义了 API 请求和响应的数据结构
- **包含**:
  - `SystemMonitorResponse`: 监控数据响应结构
  - `SystemMonitorHistoryRequest`: 历史记录查询请求
  - `SystemMonitorRangeRequest`: 时间范围查询请求
  - `SystemMonitorSummaryRequest`: 统计摘要请求
  - `SystemMonitorSummaryResponse`: 统计摘要响应
  - `SystemMonitorMetricsSummary`: 指标统计摘要
  - `SystemMonitorPeriodSummary`: 时间周期摘要
  - `DeleteSystemMonitorRangeResponse`: 批量删除响应

#### Functions (业务逻辑)
- **文件**: `internal/funcs/system_monitor_func.go`
- **核心功能**:
  1. **初始化和管理**:
     - `InitSystemMonitor()`: 初始化监控，设置采集间隔和数据保留策略
     - `StopSystemMonitor()`: 停止监控
     - `collectSystemMetrics()`: 收集系统指标
     - `cleanupOldRecords()`: 自动清理过期数据

  2. **查询功能**:
     - `GetLatestSystemMonitor()`: 获取最新监控状态
     - `GetSystemMonitorHistory()`: 获取历史记录
     - `GetSystemMonitorByRange()`: 按时间范围查询
     - `GetSystemMonitorSummary()`: 获取统计摘要

  3. **删除功能**:
     - `DeleteSystemMonitor()`: 删除单条记录
     - `DeleteSystemMonitorByRange()`: 按时间范围批量删除

  4. **辅助功能**:
     - `convertToResponse()`: 将数据库实体转换为 API 响应
     - `convertToResponseList()`: 批量转换

#### Handlers (HTTP 处理器)
- **文件**: `internal/handlers/system_monitor_handler.go`
- **功能**: 处理所有系统监控相关的 HTTP 请求
- **端点处理器**:
  - `GetLatest()`: 处理获取最新状态的请求
  - `GetHistory()`: 处理获取历史记录的请求
  - `GetByRange()`: 处理按时间范围查询的请求
  - `GetSummary()`: 处理获取统计摘要的请求
  - `Delete()`: 处理删除单条记录的请求
  - `DeleteByRange()`: 处理批量删除的请求
- **特性**: 包含完整的参数验证、错误处理和 Swagger 文档注释

#### Routes (路由配置)
- **文件**: `internal/routes/system_monitor.go`
- **功能**: 定义系统监控的 API 路由
- **路由**:
  - `GET /system/monitor/latest`: 获取最新状态
  - `GET /system/monitor/history`: 获取历史记录
  - `GET /system/monitor/range`: 按时间范围查询
  - `GET /system/monitor/summary`: 获取统计摘要
  - `DELETE /system/monitor/:id`: 删除单条记录
  - `DELETE /system/monitor/range`: 按时间范围删除

- **集成**: 已在 `internal/routes/routes.go` 中注册

### 2. 前端实现

#### API 接口定义
- **文件**: `qc-admin-api-common/src/system_monitor.ts`
- **功能**: 定义前端调用后端 API 的函数
- **导出函数**:
  - `getLatestSystemMonitor()`
  - `getSystemMonitorHistory()`
  - `getSystemMonitorByRange()`
  - `getSystemMonitorSummary()`
  - `deleteSystemMonitor()`
  - `deleteSystemMonitorByRange()`
- **已导出**: 在 `qc-admin-api-common/src/index.ts` 中已导出

#### 组件实现

##### 主页面
- **文件**: `qc-admin/src/views/system/monitor/index.vue`
- **功能**: 系统监控主页面，整合所有子组件
- **特性**: 
  - 自动刷新（30秒间隔）
  - 响应式布局
  - 生命周期管理

##### 状态卡片组件
- **文件**: `qc-admin/src/views/system/monitor/components/StatusCards.vue`
- **功能**: 展示 CPU、内存、磁盘使用率的卡片
- **特性**: 
  - 颜色编码（正常/警告/危险）
  - 图标展示
  - 响应式网格布局

##### 系统信息组件
- **文件**: `qc-admin/src/views/system/monitor/components/SystemInfo.vue`
- **功能**: 展示操作系统、主机名、平台等基本信息
- **特性**: 
  - 表格形式展示
  - 手动刷新按钮
  - 加载状态

##### 运行时信息组件
- **文件**: `qc-admin/src/views/system/monitor/components/RuntimeInfo.vue`
- **功能**: 展示 Go 运行时信息（Goroutines、堆内存、GC次数等）
- **特性**: 
  - 内存单位格式化（MB/GB）
  - 系统运行时间格式化
  - 响应式表格

##### 历史图表组件
- **文件**: `qc-admin/src/views/system/monitor/components/HistoryCharts.vue`
- **功能**: 展示 CPU、内存、磁盘、网络的历史趋势图表
- **特性**: 
  - 使用 ECharts 图表库
  - 时间范围选择（1h/6h/12h/24h）
  - 多个图表展示不同指标
  - 网络流量图表（字节格式化）
  - 响应式图表大小

##### 组合式函数
- **文件**: `qc-admin/src/views/system/monitor/composables/useSystemMonitor.ts`
- **功能**: 封装系统监控的状态管理和数据获取逻辑
- **导出**: 
  - 响应式状态（loading、currentStatus、historyData、timeRange）
  - 数据获取函数（fetchCurrentStatus、fetchHistoryData）

##### 组件索引
- **文件**: `qc-admin/src/views/system/monitor/components/index.ts`
- **功能**: 统一导出所有组件，简化导入

### 3. 文档

#### API 文档
- **文件**: `README_SYSTEM_MONITOR_API.md`
- **内容**:
  - API 端点详细说明
  - 请求/响应示例
  - 数据模型定义
  - 错误处理说明
  - 前端集成指南
  - 性能建议
  - 故障排查

#### 使用指南
- **文件**: `qc-admin/src/views/system/monitor/README.md`
- **内容**:
  - 功能概述
  - 组件说明
  - 使用方法
  - API 集成
  - 自定义配置

#### 组件拆分说明
- **文件**: `SYSTEM_MONITOR_COMPONENTS_SUMMARY.md`
- **内容**:
  - 组件架构
  - 文件结构
  - 各组件职责
  - 使用建议

## 技术栈

### 后端
- Go 1.x
- Ent (ORM)
- Gin (Web 框架)
- gopsutil (系统信息采集)

### 前端
- Vue 3
- TypeScript
- Element Plus (UI 组件库)
- ECharts (图表库)
- Composition API

## 核心特性

1. **自动采集**: 后台定时任务自动收集系统指标
2. **数据管理**: 自动清理过期数据，防止数据库膨胀
3. **实时监控**: 前端自动刷新，实时展示系统状态
4. **历史分析**: 提供历史数据查询和图表可视化
5. **统计摘要**: 计算指定时间范围内的平均值、最大值、最小值
6. **响应式设计**: 适配不同屏幕尺寸
7. **错误处理**: 完善的错误处理和用户提示
8. **性能优化**: 索引优化、数据分页、按需加载

## 使用方法

### 后端初始化

在服务器启动代码中添加：

```go
import (
    "go-backend/internal/funcs"
    "time"
)

// 初始化系统监控
// 每30秒采集一次，保留7天数据
err := funcs.InitSystemMonitor(30*time.Second, 7)
if err != nil {
    log.Fatal(err)
}

// 确保在程序退出时停止监控
defer funcs.StopSystemMonitor()
```

### 前端访问

1. 在路由中添加系统监控页面路由
2. 访问 `/system/monitor` 路径
3. 页面会自动加载当前状态和历史数据
4. 支持选择不同的时间范围查看历史趋势

## 数据流

```
1. 后台定时任务 -> 收集系统指标 -> 存入数据库
2. 前端请求 -> Handler -> Functions -> Database -> 返回数据
3. 前端接收数据 -> 更新状态 -> 渲染组件 -> 展示图表
```

## 下一步改进建议

1. **告警功能**: 添加阈值配置和告警通知
2. **数据导出**: 支持导出监控数据为 Excel 或 CSV
3. **更多指标**: 添加进程级监控、数据库连接池等
4. **对比分析**: 支持不同时间段的数据对比
5. **性能优化**: 使用时序数据库（如 InfluxDB）存储监控数据
6. **分布式监控**: 支持多服务器监控和集群视图
7. **自定义仪表盘**: 允许用户自定义监控面板

## 相关文件清单

### 后端
- `database/schema/system_monitor.go`
- `shared/models/system_monitor.go`
- `internal/funcs/system_monitor_func.go`
- `internal/handlers/system_monitor_handler.go`
- `internal/routes/system_monitor.go`
- `internal/routes/routes.go` (已更新)

### 前端
- `qc-admin-api-common/src/system_monitor.ts`
- `qc-admin-api-common/src/index.ts` (已更新)
- `qc-admin/src/views/system/monitor/index.vue`
- `qc-admin/src/views/system/monitor/components/StatusCards.vue`
- `qc-admin/src/views/system/monitor/components/SystemInfo.vue`
- `qc-admin/src/views/system/monitor/components/RuntimeInfo.vue`
- `qc-admin/src/views/system/monitor/components/HistoryCharts.vue`
- `qc-admin/src/views/system/monitor/components/index.ts`
- `qc-admin/src/views/system/monitor/composables/useSystemMonitor.ts`

### 文档
- `README_SYSTEM_MONITOR_API.md`
- `qc-admin/src/views/system/monitor/README.md`
- `SYSTEM_MONITOR_COMPONENTS_SUMMARY.md`

## 总结

系统监控功能已经完整实现，包括：
- ✅ 数据库 Schema 定义
- ✅ 后端业务逻辑（采集、存储、查询、删除）
- ✅ RESTful API 接口
- ✅ 前端 API 调用封装
- ✅ 前端 UI 组件（状态卡片、系统信息、运行时信息、历史图表）
- ✅ 自动刷新和实时监控
- ✅ 历史数据可视化
- ✅ 响应式设计
- ✅ 完整的文档

所有功能已准备就绪，可以直接使用。只需在服务器启动时调用 `InitSystemMonitor()` 初始化即可。
