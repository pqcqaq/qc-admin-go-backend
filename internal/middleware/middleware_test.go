package middleware

import (
	"encoding/json"
	"go-backend/shared/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.Use(ErrorHandler())

	// 测试panic错误
	router.GET("/panic", func(c *gin.Context) {
		PanicWithError(InternalServerError("测试panic错误", nil))
	})

	// 测试ThrowError
	router.GET("/throw", func(c *gin.Context) {
		ThrowError(c, BadRequestError("测试throw错误", nil))
	})

	// 测试正常响应
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true, "message": "正常响应"})
	})

	tests := []struct {
		name           string
		path           string
		expectedCode   int
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:          "正常请求",
			path:          "/normal",
			expectedCode:  200,
			expectedError: false,
		},
		{
			name:           "panic错误",
			path:           "/panic",
			expectedCode:   500,
			expectedError:  true,
			expectedErrMsg: "测试panic错误",
		},
		{
			name:           "throw错误",
			path:           "/throw",
			expectedCode:   400,
			expectedError:  true,
			expectedErrMsg: "测试throw错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.expectedError {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}

				if response.Success {
					t.Error("Expected success to be false")
				}

				if !strings.Contains(response.Message, tt.expectedErrMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrMsg, response.Message)
				}
			}
		})
	}
}

func TestCustomErrorTypes(t *testing.T) {
	tests := []struct {
		name         string
		errorFunc    func() *CustomError
		expectedCode models.ErrorCode
		expectedMsg  string
	}{
		{
			name:         "BadRequestError",
			errorFunc:    func() *CustomError { return BadRequestError("", nil) },
			expectedCode: ErrCodeBadRequest,
			expectedMsg:  "请求参数错误",
		},
		{
			name:         "NotFoundError",
			errorFunc:    func() *CustomError { return NotFoundError("", nil) },
			expectedCode: ErrCodeNotFound,
			expectedMsg:  "资源未找到",
		},
		{
			name:         "UserNotFoundError",
			errorFunc:    func() *CustomError { return UserNotFoundError(nil) },
			expectedCode: ErrCodeUserNotFound,
			expectedMsg:  "用户不存在",
		},
		{
			name:         "DatabaseError",
			errorFunc:    func() *CustomError { return DatabaseError("", nil) },
			expectedCode: ErrCodeDatabaseError,
			expectedMsg:  "数据库错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()

			if err.Code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d", tt.expectedCode, err.Code)
			}

			if err.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, err.Message)
			}
		})
	}
}

func TestGetHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name         string
		errorCode    models.ErrorCode
		expectedHTTP int
	}{
		{"BadRequest", ErrCodeBadRequest, 400},
		{"NotFound", ErrCodeNotFound, 404},
		{"Internal", ErrCodeInternal, 500},
		{"UserNotFound", ErrCodeUserNotFound, 404},
		{"UserExists", ErrCodeUserExists, 409},
		{"DatabaseError", ErrCodeDatabaseError, 500},
		{"ValidationError", ErrCodeValidationError, 400},
		{"UnknownError", 9999, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpCode := getHTTPStatusCode(tt.errorCode)
			if httpCode != tt.expectedHTTP {
				t.Errorf("Expected HTTP code %d, got %d", tt.expectedHTTP, httpCode)
			}
		})
	}
}
