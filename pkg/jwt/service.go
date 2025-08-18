package jwt

import (
	"errors"
	"sync"

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
		service = NewJWTService(config.SecretKey, config.Issuer, config.Expiry)
	})
	return err
}

// GetService 获取JWT服务实例
func GetService() *JWTService {
	mu.RLock()
	defer mu.RUnlock()
	return service
}

// GenerateToken 生成Token (全局函数)
func GenerateToken(userID uint64) (string, error) {
	if service == nil {
		return "", ErrServiceNotInitialized
	}
	return service.GenerateToken(userID)
}

// ValidateToken 验证Token (全局函数)
func ValidateToken(tokenString string) (*Claims, error) {
	if service == nil {
		return nil, ErrServiceNotInitialized
	}
	return service.ValidateToken(tokenString)
}

// RefreshToken 刷新Token (全局函数)
func RefreshToken(tokenString string) (string, error) {
	if service == nil {
		return "", ErrServiceNotInitialized
	}
	return service.RefreshToken(tokenString)
}
