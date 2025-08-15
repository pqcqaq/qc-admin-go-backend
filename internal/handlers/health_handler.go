package handlers

import (
	"net/http"

	"go-backend/pkg/database"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建新的健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查端点
func (h *HealthHandler) Health(c *gin.Context) {
	var dbStatus string
	if database.IsInstanceInitialized() {
		dbStatus = "Database Connected"
	} else {
		dbStatus = "Database Not Connected"
	}
	
	response := &models.HealthResponse{
		Status:  "ok",
		Message: "Server is running",
		Components: []string{
			dbStatus,
			"Cache",
			"Message Queue",
		},
	}
	c.JSON(http.StatusOK, response)
}
