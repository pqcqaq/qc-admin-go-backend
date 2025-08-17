package database

import (
	"go-backend/database/handlers"
)

// InitEventSystem 初始化事件系统
func InitEventSystem() {
	// 初始化所有事件处理器
	handlers.Init()
}
