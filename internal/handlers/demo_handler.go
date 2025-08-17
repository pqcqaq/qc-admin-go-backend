package handlers

import (
	"go-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type DemoHandler struct {
}

func NewDemoHandler() *DemoHandler {
	return &DemoHandler{}
}

// DemoErrorHandling 演示不同类型的错误处理
// @Summary      演示错误处理
// @Description  演示不同类型的错误处理机制
// @Tags         demo
// @Accept       json
// @Produce      json
// @Param        type  query     string  false  "错误类型"  Enums(panic,validation,database,notfound,unauthorized,forbidden,conflict,success)
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /demo/error [get]
func (h *DemoHandler) DemoErrorHandling(c *gin.Context) {
	errorType := c.Query("type")

	switch errorType {
	case "panic":
		// 演示使用panic抛出错误
		middleware.PanicWithError(middleware.InternalServerError("演示panic错误", map[string]any{
			"demo": true,
		}))
	case "validation":
		// 演示验证错误
		middleware.ThrowError(c, middleware.ValidationError("演示验证错误", map[string]any{
			"field": "demo_field",
			"value": "invalid_value",
		}))
		return
	case "notfound":
		// 演示资源未找到错误
		middleware.ThrowError(c, middleware.NotFoundError("演示资源未找到", nil))
		return
	case "business":
		// 演示业务逻辑错误
		middleware.ThrowError(c, middleware.NewCustomError(5001, "演示业务逻辑错误", map[string]any{
			"business_rule": "demo_rule_violation",
		}))
		return
	default:
		c.JSON(200, gin.H{
			"success": true,
			"message": "错误处理演示，请使用 ?type=panic|validation|notfound|business 参数",
		})
	}
}
