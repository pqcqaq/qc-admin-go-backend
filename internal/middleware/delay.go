package middleware

import (
	"go-backend/pkg/configs"
	"go-backend/pkg/logging"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

// DelayMiddleware 延迟中间件，用于开发环境模拟网络波动
func DelayMiddleware() gin.HandlerFunc {
	// 获取配置的延迟时间
	delay := configs.GetConfig().Server.Middleware.Delay
	logging.WithName("delay_mock_middleware").Info("Delay middleware enabled: min=%s, max=%s", delay.Min, delay.Max)

	return func(c *gin.Context) {
		// 计算随机延迟时间（Min 到 Max 之间的随机值）
		minMs := delay.Min.Nanoseconds()
		maxMs := delay.Max.Nanoseconds()

		var randomDelay time.Duration
		if maxMs > minMs {
			randomDelay = time.Duration(rand.Int63n(maxMs-minMs) + minMs)
		} else {
			randomDelay = delay.Min
		}

		// 记录延迟信息（仅在调试模式下）
		if configs.GetConfig().Server.Debug {
			logging.WithName("delay_mock_middleware").Info(
				"path: %s, method: %s, delay_ms: %d, client_ip: %s",
				c.Request.URL.Path, c.Request.Method, randomDelay.Milliseconds(), c.ClientIP(),
			)
		}

		// 执行延迟
		time.Sleep(randomDelay)
		// 继续处理请求
		c.Next()
	}
}
