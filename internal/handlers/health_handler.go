package handlers

import (
	"net/http"

	"go-backend/pkg/caching"
	"go-backend/pkg/database"
	"go-backend/pkg/s3"
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
// @Summary      健康检查
// @Description  检查服务健康状态
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	var dbStatus string
	if database.IsInstanceInitialized() {
		dbStatus = "Alive"
	} else {
		dbStatus = "Dead"
	}

	var cacheStatus string
	if caching.IsAlive() {
		cacheStatus = "Alive"
	} else {
		cacheStatus = "Dead"
	}

	var s3Status string
	if err := s3.Client.TestConnection(); err == nil {
		s3Status = "Alive"
	} else {
		s3Status = "Dead"
	}

	response := &models.HealthResponse{
		Status:  "ok",
		Message: "Server is running",
		Components: map[string]string{
			"database": dbStatus,
			"cache":    cacheStatus,
			"s3":       s3Status,
		},
	}
	c.JSON(http.StatusOK, response)
}
