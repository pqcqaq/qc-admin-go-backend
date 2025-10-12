package funcs

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/systemmonitor"
	"go-backend/pkg/database"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
	"go-backend/shared/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMonitorFuncs 系统监控服务
type SystemMonitorFunc struct{}

var (
	// 全局变量存储上次的网络统计
	lastNetStats *net.IOCountersStat
	// 定时器
	monitorTicker *time.Ticker
	// 停止信号
	stopChan chan struct{}
)

// convertToResponse 将 ent.SystemMonitor 转换为 models.SystemMonitorResponse
func convertToResponse(record *ent.SystemMonitor) *models.SystemMonitorResponse {
	if record == nil {
		return nil
	}

	// 处理可选的 LoadAvg 字段 - 如果值为0则设置为nil
	var loadAvg1, loadAvg5, loadAvg15 *float64
	if record.LoadAvg1 != 0 {
		loadAvg1 = &record.LoadAvg1
	}
	if record.LoadAvg5 != 0 {
		loadAvg5 = &record.LoadAvg5
	}
	if record.LoadAvg15 != 0 {
		loadAvg15 = &record.LoadAvg15
	}

	return &models.SystemMonitorResponse{
		ID:                 strconv.FormatUint(record.ID, 10),
		CPUUsagePercent:    record.CPUUsagePercent,
		CPUCores:           record.CPUCores,
		MemoryTotal:        record.MemoryTotal,
		MemoryUsed:         record.MemoryUsed,
		MemoryFree:         record.MemoryFree,
		MemoryUsagePercent: record.MemoryUsagePercent,
		DiskTotal:          record.DiskTotal,
		DiskUsed:           record.DiskUsed,
		DiskFree:           record.DiskFree,
		DiskUsagePercent:   record.DiskUsagePercent,
		NetworkBytesSent:   record.NetworkBytesSent,
		NetworkBytesRecv:   record.NetworkBytesRecv,
		OS:                 record.Os,
		Platform:           record.Platform,
		PlatformVersion:    record.PlatformVersion,
		Hostname:           record.Hostname,
		GoroutinesCount:    record.GoroutinesCount,
		HeapAlloc:          record.HeapAlloc,
		HeapSys:            record.HeapSys,
		GCCount:            record.GcCount,
		LoadAvg1:           loadAvg1,
		LoadAvg5:           loadAvg5,
		LoadAvg15:          loadAvg15,
		Uptime:             record.Uptime,
		RecordedAt:         record.RecordedAt.Format(time.RFC3339),
		CreatedAt:          record.CreateTime.Format(time.RFC3339),
		UpdatedAt:          record.UpdateTime.Format(time.RFC3339),
	}
}

// convertToResponseList 批量转换
func convertToResponseList(records []*ent.SystemMonitor) []*models.SystemMonitorResponse {
	result := make([]*models.SystemMonitorResponse, 0, len(records))
	for _, record := range records {
		result = append(result, convertToResponse(record))
	}
	return result
}

// InitSystemMonitor 初始化系统监控
// interval: 监控间隔时间 (例如: 30 * time.Second)
// retentionDays: 数据保留天数，超过此天数的数据将被自动清理
func InitSystemMonitor(interval time.Duration, retentionDays int) error {
	// 如果已经初始化，先停止
	if monitorTicker != nil {
		StopSystemMonitor()
	}

	// 初始化网络统计
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		lastNetStats = &netStats[0]
	}

	// 创建定时器
	monitorTicker = time.NewTicker(interval)
	stopChan = make(chan struct{})

	// 启动监控协程
	go func() {
		// 立即执行一次
		if _, err := collectSystemMetrics(); err != nil {
			logging.Error("Failed to collect system metrics: %v\n", err)
		}

		for {
			select {
			case <-monitorTicker.C:
				record, err := collectSystemMetrics()
				if err != nil {
					logging.Error("Failed to collect system metrics: %v\n", err)
				}
				// 清理过期数据
				if err := cleanupOldRecords(retentionDays); err != nil {
					logging.Error("Failed to cleanup old records: %v\n", err)
				}

				// 通过ws发送到前端
				data, err := utils.StructToMap(convertToResponse(record))
				if err != nil {
					logging.Error("Failed to convert system monitor record to map: %v\n", err)
					continue
				}
				messaging.Publish(context.Background(), messaging.MessageStruct{
					Type: messaging.ServerToUserSocket,
					Payload: messaging.SocketMessagePayload{
						UserId: nil,
						Topic:  "system/monitor/update",
						Data:   data,
					},
				})
			case <-stopChan:
				return
			}
		}
	}()

	logging.Info("System monitor initialized with interval: %v, retention: %d days\n", interval, retentionDays)
	return nil
}

// StopSystemMonitor 停止系统监控
func StopSystemMonitor() {
	if monitorTicker != nil {
		monitorTicker.Stop()
		close(stopChan)
		monitorTicker = nil
		stopChan = nil
		logging.Info("System monitor stopped")
	}
}

// collectSystemMetrics 收集系统指标
func collectSystemMetrics() (*ent.SystemMonitor, error) {
	ctx := context.Background()

	// 收集 CPU 信息
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	cpuCores, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU cores: %w", err)
	}

	// 收集内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// 收集磁盘信息
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	// 收集网络信息
	netStats, err := net.IOCounters(false)
	var bytesSent, bytesRecv uint64
	if err == nil && len(netStats) > 0 {
		bytesSent = netStats[0].BytesSent
		bytesRecv = netStats[0].BytesRecv
		lastNetStats = &netStats[0]
	}

	// 收集系统信息
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	// 收集负载信息 (仅Unix系统)
	loadInfo, _ := load.Avg()
	var load1, load5, load15 *float64
	if loadInfo != nil {
		l1, l5, l15 := loadInfo.Load1, loadInfo.Load5, loadInfo.Load15
		load1, load5, load15 = &l1, &l5, &l15
	}

	// 收集 Go 运行时信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 创建监控记录
	record, err := database.Client.SystemMonitor.Create().
		SetCPUUsagePercent(cpuUsage).
		SetCPUCores(cpuCores).
		SetMemoryTotal(memInfo.Total).
		SetMemoryUsed(memInfo.Used).
		SetMemoryFree(memInfo.Free).
		SetMemoryUsagePercent(memInfo.UsedPercent).
		SetDiskTotal(diskInfo.Total).
		SetDiskUsed(diskInfo.Used).
		SetDiskFree(diskInfo.Free).
		SetDiskUsagePercent(diskInfo.UsedPercent).
		SetNetworkBytesSent(bytesSent).
		SetNetworkBytesRecv(bytesRecv).
		SetOs(hostInfo.OS).
		SetPlatform(hostInfo.Platform).
		SetPlatformVersion(hostInfo.PlatformVersion).
		SetHostname(hostInfo.Hostname).
		SetGoroutinesCount(runtime.NumGoroutine()).
		SetHeapAlloc(memStats.HeapAlloc).
		SetHeapSys(memStats.HeapSys).
		SetGcCount(memStats.NumGC).
		SetNillableLoadAvg1(load1).
		SetNillableLoadAvg5(load5).
		SetNillableLoadAvg15(load15).
		SetUptime(hostInfo.Uptime).
		SetRecordedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to save system monitor record: %w", err)
	}

	return record, nil
}

// cleanupOldRecords 清理过期的监控记录
func cleanupOldRecords(retentionDays int) error {
	ctx := context.Background()
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	deleted, err := database.Client.SystemMonitor.Delete().
		Where(systemmonitor.RecordedAtLT(cutoffTime)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to cleanup old records: %w", err)
	}

	if deleted > 0 {
		logging.Info("Cleaned up %d old system monitor records\n", deleted)
	}

	return nil
}

// GetLatestSystemMonitor 获取最新的系统监控状态
func (SystemMonitorFunc) GetLatestSystemMonitor(ctx context.Context) (*models.SystemMonitorResponse, error) {
	record, err := database.Client.SystemMonitor.Query().
		Order(ent.Desc(systemmonitor.FieldRecordedAt)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest system status: %w", err)
	}

	return convertToResponse(record), nil
}

// GetSystemMonitorHistory 获取系统监控历史记录
// limit: 返回的记录数量
// hours: 查询最近多少小时的数据
func (SystemMonitorFunc) GetSystemMonitorHistory(ctx context.Context, limit int, hours int) ([]*models.SystemMonitorResponse, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	records, err := database.Client.SystemMonitor.Query().
		Where(systemmonitor.RecordedAtGTE(startTime)).
		Order(ent.Desc(systemmonitor.FieldRecordedAt)).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get system status history: %w", err)
	}

	return convertToResponseList(records), nil
}

// GetSystemMonitorByRange 根据时间范围获取系统监控数据
func (SystemMonitorFunc) GetSystemMonitorByRange(ctx context.Context, startTime, endTime time.Time) ([]*models.SystemMonitorResponse, error) {
	records, err := database.Client.SystemMonitor.Query().
		Where(
			systemmonitor.RecordedAtGTE(startTime),
			systemmonitor.RecordedAtLTE(endTime),
		).
		Order(ent.Asc(systemmonitor.FieldRecordedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get system status by time range: %w", err)
	}

	return convertToResponseList(records), nil
}

// GetSystemMonitorSummary 获取系统监控统计摘要
func (SystemMonitorFunc) GetSystemMonitorSummary(ctx context.Context, hours int) (*models.SystemMonitorSummaryResponse, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	records, err := database.Client.SystemMonitor.Query().
		Where(systemmonitor.RecordedAtGTE(startTime)).
		Order(ent.Asc(systemmonitor.FieldRecordedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get system stats summary: %w", err)
	}

	if len(records) == 0 {
		return &models.SystemMonitorSummaryResponse{
			Count: 0,
		}, nil
	}

	// 计算平均值、最大值、最小值
	var (
		totalCPU    float64
		totalMemory float64
		totalDisk   float64
		maxCPU      float64
		maxMemory   float64
		maxDisk     float64
		minCPU      = 100.0
		minMemory   = 100.0
		minDisk     = 100.0
	)

	for _, record := range records {
		totalCPU += record.CPUUsagePercent
		totalMemory += record.MemoryUsagePercent
		totalDisk += record.DiskUsagePercent

		if record.CPUUsagePercent > maxCPU {
			maxCPU = record.CPUUsagePercent
		}
		if record.CPUUsagePercent < minCPU {
			minCPU = record.CPUUsagePercent
		}

		if record.MemoryUsagePercent > maxMemory {
			maxMemory = record.MemoryUsagePercent
		}
		if record.MemoryUsagePercent < minMemory {
			minMemory = record.MemoryUsagePercent
		}

		if record.DiskUsagePercent > maxDisk {
			maxDisk = record.DiskUsagePercent
		}
		if record.DiskUsagePercent < minDisk {
			minDisk = record.DiskUsagePercent
		}
	}

	count := int64(len(records))
	duration := records[count-1].RecordedAt.Sub(records[0].RecordedAt).Hours()

	summary := &models.SystemMonitorSummaryResponse{
		Count: count,
		CPU: models.SystemMonitorMetricsSummary{
			Avg: totalCPU / float64(count),
			Max: maxCPU,
			Min: minCPU,
		},
		Memory: models.SystemMonitorMetricsSummary{
			Avg: totalMemory / float64(count),
			Max: maxMemory,
			Min: minMemory,
		},
		Disk: models.SystemMonitorMetricsSummary{
			Avg: totalDisk / float64(count),
			Max: maxDisk,
			Min: minDisk,
		},
		Period: models.SystemMonitorPeriodSummary{
			Start: records[0].RecordedAt.Format(time.RFC3339),
			End:   records[count-1].RecordedAt.Format(time.RFC3339),
			Hours: duration,
		},
	}

	return summary, nil
}

// DeleteSystemMonitor 删除系统监控记录
func (SystemMonitorFunc) DeleteSystemMonitor(ctx context.Context, id string) error {
	// 将字符串ID转换为uint64
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id format: %w", err)
	}

	err = database.Client.SystemMonitor.DeleteOneID(idUint).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("system monitor record not found")
		}
		return fmt.Errorf("failed to delete system monitor record: %w", err)
	}
	return nil
}

// DeleteSystemMonitorByRange 根据时间范围删除系统监控记录
func (SystemMonitorFunc) DeleteSystemMonitorByRange(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	deleted, err := database.Client.SystemMonitor.Delete().
		Where(
			systemmonitor.RecordedAtGTE(startTime),
			systemmonitor.RecordedAtLTE(endTime),
		).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete system monitor records: %w", err)
	}

	return int64(deleted), nil
}

// 以下是保持向后兼容的旧函数名

// GetLatestSystemStatus 获取最新的系统状态 (已弃用，使用 GetLatestSystemMonitor)
func (SystemMonitorFunc) GetLatestSystemStatus(ctx context.Context) (*ent.SystemMonitor, error) {
	record, err := database.Client.SystemMonitor.Query().
		Order(ent.Desc(systemmonitor.FieldRecordedAt)).
		First(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest system status: %w", err)
	}

	return record, nil
}

// GetSystemStatusHistory 获取系统状态历史记录
// limit: 返回的记录数量
// hours: 查询最近多少小时的数据
func (SystemMonitorFunc) GetSystemStatusHistory(ctx context.Context, limit int, hours int) ([]*ent.SystemMonitor, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	records, err := database.Client.SystemMonitor.Query().
		Where(systemmonitor.RecordedAtGTE(startTime)).
		Order(ent.Desc(systemmonitor.FieldRecordedAt)).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get system status history: %w", err)
	}

	return records, nil
}

// GetSystemStatusByTimeRange 根据时间范围获取系统状态
func (SystemMonitorFunc) GetSystemStatusByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*ent.SystemMonitor, error) {
	records, err := database.Client.SystemMonitor.Query().
		Where(
			systemmonitor.RecordedAtGTE(startTime),
			systemmonitor.RecordedAtLTE(endTime),
		).
		Order(ent.Asc(systemmonitor.FieldRecordedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get system status by time range: %w", err)
	}

	return records, nil
}

// GetSystemStatsSummary 获取系统状态统计摘要 (已弃用，使用 GetSystemMonitorSummary)
func (SystemMonitorFunc) GetSystemStatsSummary(ctx context.Context, hours int) (map[string]interface{}, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	records, err := database.Client.SystemMonitor.Query().
		Where(systemmonitor.RecordedAtGTE(startTime)).
		Order(ent.Asc(systemmonitor.FieldRecordedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get system stats summary: %w", err)
	}

	if len(records) == 0 {
		return map[string]interface{}{
			"count": 0,
		}, nil
	}

	// 计算平均值、最大值、最小值
	var (
		totalCPU    float64
		totalMemory float64
		totalDisk   float64
		maxCPU      float64
		maxMemory   float64
		maxDisk     float64
		minCPU      = 100.0
		minMemory   = 100.0
		minDisk     = 100.0
	)

	for _, record := range records {
		totalCPU += record.CPUUsagePercent
		totalMemory += record.MemoryUsagePercent
		totalDisk += record.DiskUsagePercent

		if record.CPUUsagePercent > maxCPU {
			maxCPU = record.CPUUsagePercent
		}
		if record.CPUUsagePercent < minCPU {
			minCPU = record.CPUUsagePercent
		}

		if record.MemoryUsagePercent > maxMemory {
			maxMemory = record.MemoryUsagePercent
		}
		if record.MemoryUsagePercent < minMemory {
			minMemory = record.MemoryUsagePercent
		}

		if record.DiskUsagePercent > maxDisk {
			maxDisk = record.DiskUsagePercent
		}
		if record.DiskUsagePercent < minDisk {
			minDisk = record.DiskUsagePercent
		}
	}

	count := len(records)
	summary := map[string]interface{}{
		"count": count,
		"cpu": map[string]float64{
			"avg": totalCPU / float64(count),
			"max": maxCPU,
			"min": minCPU,
		},
		"memory": map[string]float64{
			"avg": totalMemory / float64(count),
			"max": maxMemory,
			"min": minMemory,
		},
		"disk": map[string]float64{
			"avg": totalDisk / float64(count),
			"max": maxDisk,
			"min": minDisk,
		},
		"period": map[string]interface{}{
			"start": records[0].RecordedAt,
			"end":   records[count-1].RecordedAt,
			"hours": hours,
		},
	}

	return summary, nil
}

// DeleteSystemMonitorRecord 删除系统监控记录
func (SystemMonitorFunc) DeleteSystemMonitorRecord(ctx context.Context, id uint64) error {
	err := database.Client.SystemMonitor.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete system monitor record: %w", err)
	}
	return nil
}

// DeleteSystemMonitorRecordsByTimeRange 根据时间范围删除系统监控记录
func (SystemMonitorFunc) DeleteSystemMonitorRecordsByTimeRange(ctx context.Context, startTime, endTime time.Time) (int, error) {
	deleted, err := database.Client.SystemMonitor.Delete().
		Where(
			systemmonitor.RecordedAtGTE(startTime),
			systemmonitor.RecordedAtLTE(endTime),
		).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete system monitor records: %w", err)
	}

	return deleted, nil
}
