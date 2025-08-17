package handlers

import (
	"go-backend/pkg/logging"
)

// Init 初始化所有事件处理器
func Init() {
	logging.Info("Initializing event handlers...")

	// 注册角色继承检查处理器
	RegisterRoleInheritanceHandler()

	logging.Info("Event handlers initialized successfully")
}
