package funcs

import (
	"go-backend/pkg/configs"
	"sync"
	"time"
)

func Setup() {
	config := configs.GetConfig()

	monitorConfig := config.Server.Components.Monitor
	if monitorConfig.Enabled {
		interval := time.Duration(monitorConfig.Interval) * time.Second
		ent := monitorConfig.RetentionDays
		// 启动系统监控
		InitSystemMonitor(interval, ent)
	}

	// 初始化WebSocket认证缓存
	wsCacheLock = sync.RWMutex{}
	wsCache = make(map[uint64]*WsCache)
	records := queryWsList()
	wsCache = makeCache(records)
}

func Cleanup() {
	// 清理系统监控
	StopSystemMonitor()
}
