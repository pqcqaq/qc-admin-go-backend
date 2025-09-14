package middleware

import (
	"context"
	"strings"

	"go-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	// UserIDKey 用户ID的context key
	UserIDKey string = "user_id"
	// JWTClaimsKey JWT Claims的context key
	JWTClaimsKey string = "jwt_claims"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ThrowError(c, UnauthorizedError("未提供认证令牌", nil))
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ThrowError(c, UnauthorizedError("认证令牌格式错误", nil))
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			ThrowError(c, UnauthorizedError("认证令牌为空", nil))
			c.Abort()
			return
		}

		// 验证token
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			ThrowError(c, UnauthorizedError("认证令牌无效", err.Error()))
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选的JWT认证中间件（不强制要求token）
func OptionalJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != "" {
				if claims, err := jwt.ValidateToken(tokenString); err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("jwt_claims", claims)
				}
			}
		}
		c.Next()
	}
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint64, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint64); ok {
			return uid, true
		}
	}
	return 0, false
}

// GetJWTClaims 从上下文中获取JWT Claims
func GetJWTClaims(c *gin.Context) (*jwt.Claims, bool) {
	if claims, exists := c.Get("jwt_claims"); exists {
		if jwtClaims, ok := claims.(*jwt.Claims); ok {
			return jwtClaims, true
		}
	}
	return nil, false
}

// RequireAuth 确保用户已认证的助手函数
func RequireAuth(c *gin.Context) (uint64, bool) {
	userID, exists := GetCurrentUserID(c)
	if !exists {
		ThrowError(c, UnauthorizedError("需要认证", nil))
		c.Abort()
		return 0, false
	}
	return userID, true
}

// GetRequestContext 从gin.Context获取带有用户信息的context.Context
// 这是统一处理context传递的核心函数
func GetRequestContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// 如果有用户ID，将其添加到context中
	if userID, exists := GetCurrentUserID(c); exists {
		ctx = context.WithValue(ctx, UserIDKey, userID)
	}

	// 如果有JWT Claims，将其添加到context中
	if claims, exists := GetJWTClaims(c); exists {
		ctx = context.WithValue(ctx, JWTClaimsKey, claims)
	}

	return ctx
}

// GetUserIDFromContext 从context中获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint64, bool) {
	if userID, ok := ctx.Value(UserIDKey).(uint64); ok {
		return userID, true
	}
	return 0, false
}

// GetJWTClaimsFromContext 从context中获取JWT Claims
func GetJWTClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	if claims, ok := ctx.Value(JWTClaimsKey).(*jwt.Claims); ok {
		return claims, true
	}
	return nil, false
}
