package models

// SystemMonitorResponse 系统监控响应结构
type SystemMonitorResponse struct {
	ID                 string   `json:"id"`
	CPUUsagePercent    float64  `json:"cpuUsagePercent"`
	CPUCores           int      `json:"cpuCores"`
	MemoryTotal        uint64   `json:"memoryTotal"`
	MemoryUsed         uint64   `json:"memoryUsed"`
	MemoryFree         uint64   `json:"memoryFree"`
	MemoryUsagePercent float64  `json:"memoryUsagePercent"`
	DiskTotal          uint64   `json:"diskTotal"`
	DiskUsed           uint64   `json:"diskUsed"`
	DiskFree           uint64   `json:"diskFree"`
	DiskUsagePercent   float64  `json:"diskUsagePercent"`
	NetworkBytesSent   uint64   `json:"networkBytesSent"`
	NetworkBytesRecv   uint64   `json:"networkBytesRecv"`
	OS                 string   `json:"os"`
	Platform           string   `json:"platform"`
	PlatformVersion    string   `json:"platformVersion"`
	Hostname           string   `json:"hostname"`
	GoroutinesCount    int      `json:"goroutinesCount"`
	HeapAlloc          uint64   `json:"heapAlloc"`
	HeapSys            uint64   `json:"heapSys"`
	GCCount            uint32   `json:"gcCount"`
	LoadAvg1           *float64 `json:"loadAvg1,omitempty"`
	LoadAvg5           *float64 `json:"loadAvg5,omitempty"`
	LoadAvg15          *float64 `json:"loadAvg15,omitempty"`
	Uptime             uint64   `json:"uptime"`
	RecordedAt         string   `json:"recordedAt"`
	CreatedAt          string   `json:"createdAt"`
	UpdatedAt          string   `json:"updatedAt,omitempty"`
}

// SystemMonitorHistoryRequest 历史记录查询请求
type SystemMonitorHistoryRequest struct {
	Limit *int `form:"limit" json:"limit" binding:"omitempty,min=1,max=1000"` // 返回记录数，默认100
	Hours *int `form:"hours" json:"hours" binding:"omitempty,min=1,max=168"`  // 查询最近多少小时，默认1小时，最大7天
}

// SystemMonitorRangeRequest 时间范围查询请求
type SystemMonitorRangeRequest struct {
	Start string `form:"start" json:"start" binding:"required"` // 开始时间 (ISO 8601)
	End   string `form:"end" json:"end" binding:"required"`     // 结束时间 (ISO 8601)
}

// SystemMonitorSummaryRequest 统计摘要请求
type SystemMonitorSummaryRequest struct {
	Hours *int `form:"hours" json:"hours" binding:"omitempty,min=1,max=720"` // 查询最近多少小时，默认24小时，最大30天
}

// SystemMonitorSummaryResponse 系统监控统计摘要响应
type SystemMonitorSummaryResponse struct {
	Count  int64                       `json:"count"`
	CPU    SystemMonitorMetricsSummary `json:"cpu"`
	Memory SystemMonitorMetricsSummary `json:"memory"`
	Disk   SystemMonitorMetricsSummary `json:"disk"`
	Period SystemMonitorPeriodSummary  `json:"period"`
}

// SystemMonitorMetricsSummary 指标统计摘要
type SystemMonitorMetricsSummary struct {
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
	Min float64 `json:"min"`
}

// SystemMonitorPeriodSummary 时间周期摘要
type SystemMonitorPeriodSummary struct {
	Start string  `json:"start"`
	End   string  `json:"end"`
	Hours float64 `json:"hours"`
}

// DeleteSystemMonitorRangeResponse 批量删除响应
type DeleteSystemMonitorRangeResponse struct {
	Deleted int64 `json:"deleted"`
}
