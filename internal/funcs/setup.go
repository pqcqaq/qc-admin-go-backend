package funcs

import (
	"go-backend/pkg/configs"
	"time"
)

func Setup() {
	config := configs.GetConfig()

	monitorConfig := config.Server.Components.Monitor
	if monitorConfig.Enabled {
		interval := time.Duration(monitorConfig.Interval)
		ent := monitorConfig.RetentionDays
		// 启动系统监控
		InitSystemMonitor(interval, ent)
	}
}

func Cleanup() {
	// 清理系统监控
	StopSystemMonitor()
}
