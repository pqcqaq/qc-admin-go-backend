package middleware

import (
	"encoding/json"
	"go-backend/internal/funcs"
	"go-backend/pkg/logging"
	"go-backend/shared/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered any) {
		handleError(c, recovered)
	})
}

// ErrorHandler 错误处理中间件（用于处理手动抛出的错误）
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleGinError(c, err.Err)
		}
	}
}

// handleError 处理panic恢复的错误
func handleError(c *gin.Context, recovered any) {
	var customErr *CustomError
	var ok bool

	// 尝试转换为自定义错误
	if customErr, ok = recovered.(*CustomError); !ok {
		// 如果不是自定义错误，创建一个内部服务器错误
		customErr = InternalServerError(
			"服务器内部错误",
			map[string]any{
				"error": recovered,
			},
		)
	}

	// 记录错误日志
	logError(c, customErr, true)

	// 响应错误
	respondWithError(c, customErr)

	// 终止请求
	c.Abort()
}

// handleGinError 处理Gin错误
func handleGinError(c *gin.Context, err error) {
	var customErr *CustomError
	var ok bool

	// 尝试转换为自定义错误
	if customErr, ok = err.(*CustomError); !ok {
		// 如果不是自定义错误，创建一个内部服务器错误
		customErr = InternalServerError(
			err.Error(),
			nil,
		)
	}

	// 记录错误日志
	logError(c, customErr, false)

	// 响应错误
	respondWithError(c, customErr)
}

// respondWithError 统一错误响应
func respondWithError(c *gin.Context, customErr *CustomError) {
	response := ErrorResponse{
		Success:   false,
		Code:      customErr.Code,
		Message:   customErr.Message,
		Data:      customErr.Data,
		Timestamp: time.Now().Format(time.RFC3339),
		Path:      c.Request.URL.Path,
	}

	// 在开发环境中包含堆栈信息
	if gin.Mode() == gin.DebugMode && customErr.Stack != "" {
		response.Stack = customErr.Stack
	}

	// 根据错误代码设置HTTP状态码
	httpCode := getHTTPStatusCode(customErr.Code)

	c.JSON(httpCode, response)
}

// logError 记录错误日志
func logError(c *gin.Context, customErr *CustomError, isPanic bool) {
	errorType := "Error"
	if isPanic {
		errorType = "Panic"
	}

	logData := map[string]any{
		"type":       errorType,
		"code":       customErr.Code,
		"message":    customErr.Message,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"user_agent": c.Request.UserAgent(),
		"ip":         c.ClientIP(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if customErr.Data != nil {
		logData["data"] = customErr.Data
	}

	if customErr.Stack != "" {
		logData["stack"] = customErr.Stack
	}

	// 序列化日志数据
	logJSON, _ := json.Marshal(logData)
	logging.Warn("[%s] %s", errorType, string(logJSON))

	level := "warn"
	if errorType == "Panic" {
		level = "fatal"
	}
	funcs.CreateAsyncLoggingFunc(
		level,
		errorType,
		customErr.Message,
		c.Request.Method,
		c.Request.URL.Path,
		c.ClientIP(),
		c.Request.URL.RawQuery,
		customErr.Code,
		c.Request.UserAgent(),
		logData,
		customErr.Stack,
	)

}

// getHTTPStatusCode 根据自定义错误代码获取HTTP状态码
func getHTTPStatusCode(errorCode models.ErrorCode) int {
	switch {
	case errorCode >= 400 && errorCode < 500:
		return int(errorCode)
	case errorCode >= 500 && errorCode < 600:
		return int(errorCode)
	case errorCode == ErrCodeUserNotFound:
		return http.StatusNotFound
	case errorCode == ErrCodeUserExists:
		return http.StatusConflict
	case errorCode == ErrCodeInvalidUserData:
		return http.StatusBadRequest
	case errorCode == ErrCodeValidationError:
		return http.StatusBadRequest
	case errorCode == ErrCodeDatabaseError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ThrowError 抛出自定义错误的便利函数
func ThrowError(c *gin.Context, customErr *CustomError) {
	c.Error(customErr)
}

// PanicWithError 使用panic抛出自定义错误
func PanicWithError(customErr *CustomError) {
	panic(customErr)
}
