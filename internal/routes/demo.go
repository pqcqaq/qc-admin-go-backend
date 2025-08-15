package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func (r *Router) setupDemoRoutes(rg *gin.RouterGroup) {

	demoHandler := handlers.NewDemoHandler()

	demo := rg.Group("/demo")
	{
		demo.GET("/error-handling", demoHandler.DemoErrorHandling)
	}
}
