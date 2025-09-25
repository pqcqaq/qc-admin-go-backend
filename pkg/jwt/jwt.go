package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT Claims
type Claims struct {
	jwt.RegisteredClaims
	UserID         uint64 `json:"user_id"`
	ClientDeviceId uint64 `json:"clientDeviceId"`
	IsRefresh      bool   `json:"isRefresh"`
	Expiry         uint64 `json:"expity"`
}

// JWTService JWT服务
type JWTService struct {
	secretKey []byte
	issuer    string
}

// NewJWTService 创建JWT服务
func NewJWTService(secretKey, issuer string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken 生成JWT Token
func (j *JWTService) GenerateToken(userID uint64, clientId uint64, expiry time.Duration, isRefresh bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:         userID,
		ClientDeviceId: clientId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d@%d", userID, clientId),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		IsRefresh: isRefresh,
		Expiry:    uint64(time.Now().Add(time.Duration(expiry)).UnixMilli()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken 验证JWT Token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新Token
func (j *JWTService) RefreshToken(tokenString string, clientId uint64, expiry time.Duration) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果id不同则不能刷新
	if claims.ClientDeviceId != clientId {
		return "", fmt.Errorf("不允许在不同终端刷新同一token")
	}

	return j.GenerateToken(claims.UserID, clientId, expiry, false)
}
