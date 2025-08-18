package middleware

import (
	"fmt"
	"go-backend/shared/models"
	"runtime"
)

// CustomError 自定义错误结构
type CustomError struct {
	Code    models.ErrorCode `json:"code"`
	Message string           `json:"message"`
	Data    any              `json:"data,omitempty"`
	Stack   string           `json:"stack,omitempty"`
}

// Error 实现 error 接口
func (e *CustomError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// NewCustomError 创建新的自定义错误
func NewCustomError(code models.ErrorCode, message string, data any) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewCustomErrorWithStack 创建带有堆栈信息的自定义错误
func NewCustomErrorWithStack(code models.ErrorCode, message string, data any) *CustomError {
	stack := make([]byte, 1024*8)
	stack = stack[:runtime.Stack(stack, false)]

	return &CustomError{
		Code:    code,
		Message: message,
		Data:    data,
		Stack:   string(stack),
	}
}

// ErrorResponse 统一的错误响应结构
type ErrorResponse struct {
	Success   bool             `json:"success"`
	Code      models.ErrorCode `json:"code"`
	Message   string           `json:"message"`
	Data      any              `json:"data,omitempty"`
	Timestamp string           `json:"timestamp"`
	Path      string           `json:"path"`
	Stack     string           `json:"stack,omitempty"`
}

// 预定义的错误代码

const (
	// 通用错误
	ErrCodeInternal     models.ErrorCode = 500
	ErrCodeBadRequest   models.ErrorCode = 400
	ErrCodeUnauthorized models.ErrorCode = 401
	ErrCodeForbidden    models.ErrorCode = 403
	ErrCodeNotFound     models.ErrorCode = 404
	ErrCodeConflict     models.ErrorCode = 409

	// 业务错误
	ErrCodeUserNotFound    models.ErrorCode = 1001
	ErrCodeUserExists      models.ErrorCode = 1002
	ErrCodeInvalidUserData models.ErrorCode = 1003
	ErrCodeDatabaseError   models.ErrorCode = 2001
	ErrCodeValidationError models.ErrorCode = 3001
)

// 预定义错误消息
var ErrorMessages = map[models.ErrorCode]string{
	ErrCodeInternal:        "内部服务器错误",
	ErrCodeBadRequest:      "请求参数错误",
	ErrCodeUnauthorized:    "未授权",
	ErrCodeForbidden:       "禁止访问",
	ErrCodeNotFound:        "资源未找到",
	ErrCodeConflict:        "资源冲突",
	ErrCodeUserNotFound:    "用户不存在",
	ErrCodeUserExists:      "用户已存在",
	ErrCodeInvalidUserData: "用户数据无效",
	ErrCodeDatabaseError:   "数据库错误",
	ErrCodeValidationError: "数据验证错误",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code models.ErrorCode) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return "未知错误"
}

// 便利函数创建常见错误
func BadRequestError(message string, data any) *CustomError {
	if message == "" {
		message = GetErrorMessage(ErrCodeBadRequest)
	}
	return NewCustomError(ErrCodeBadRequest, message, data)
}

func NotFoundError(message string, data any) *CustomError {
	if message == "" {
		message = GetErrorMessage(ErrCodeNotFound)
	}
	return NewCustomError(ErrCodeNotFound, message, data)
}

func UnauthorizedError(message string, data any) *CustomError {
	if message == "" {
		message = GetErrorMessage(ErrCodeUnauthorized)
	}
	return NewCustomError(ErrCodeUnauthorized, message, data)
}

func InternalServerError(message string, data any) *CustomError {
	if message == "" {
		message = GetErrorMessage(ErrCodeInternal)
	}
	return NewCustomErrorWithStack(ErrCodeInternal, message, data)
}

func DatabaseError(message string, data any) *CustomError {
	if message == "" {
		message = GetErrorMessage(ErrCodeDatabaseError)
	}
	return NewCustomErrorWithStack(ErrCodeDatabaseError, message, data)
}

func ValidationError(message string, data any) *CustomError {
	if message == "" {
		message = GetErrorMessage(ErrCodeValidationError)
	}
	return NewCustomError(ErrCodeValidationError, message, data)
}

func BusinessError(message string, data any) *CustomError {
	if message == "" {
		message = "业务逻辑错误"
	}
	return NewCustomError(ErrCodeInternal, message, data)
}

func UserNotFoundError(data any) *CustomError {
	return NewCustomError(ErrCodeUserNotFound, GetErrorMessage(ErrCodeUserNotFound), data)
}

func UserExistsError(data any) *CustomError {
	return NewCustomError(ErrCodeUserExists, GetErrorMessage(ErrCodeUserExists), data)
}
