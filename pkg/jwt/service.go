package jwt

import (
	"errors"
	"sync"
	"time"

	"go-backend/pkg/configs"
)

var (
	service *JWTService
	once    sync.Once
	mu      sync.RWMutex
)

// 错误定义
var (
	ErrServiceNotInitialized = errors.New("JWT service not initialized")
)

// InitializeService 初始化JWT服务
func InitializeService(config *configs.JWTConfig) error {
	var err error
	once.Do(func() {
		service = NewJWTService(config.SecretKey, config.Issuer)
	})
	return err
}

// GetService 获取JWT服务实例
func GetService() *JWTService {
	mu.RLock()
	defer mu.RUnlock()
	return service
}

// GenerateAccessToken 生成Token (全局函数)
func GenerateAccessToken(userID, clientId uint64, expiry time.Duration) (string, error) {
	if service == nil {
		return "", ErrServiceNotInitialized
	}
	return service.GenerateToken(userID, clientId, expiry, false)
}

// GenerateRefreshToken 生成Token (全局函数)
func GenerateRefreshToken(userID, clientId uint64, expiry time.Duration) (string, error) {
	if service == nil {
		return "", ErrServiceNotInitialized
	}
	return service.GenerateToken(userID, clientId, expiry, true)
}

// ValidateToken 验证Token (全局函数)
func ValidateToken(tokenString string) (*Claims, error) {
	if service == nil {
		return nil, ErrServiceNotInitialized
	}
	return service.ValidateToken(tokenString)
}

// RefreshToken 刷新Token (全局函数)
func RefreshToken(tokenString string, clientId uint64, expiry time.Duration) (string, error) {
	if service == nil {
		return "", ErrServiceNotInitialized
	}
	return service.RefreshToken(tokenString, clientId, expiry)
}
