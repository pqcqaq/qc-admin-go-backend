package middleware

import (
	"context"
	"strings"

	"go-backend/internal/funcs"
	"go-backend/pkg/jwt"
	"go-backend/pkg/logging"

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

		record, exists := c.Get(string(ApiAuthRecord))

		if !exists || record == nil {
			ThrowError(c, ForbiddenError("未找到API认证权限", nil))
			c.Abort()
			return
		}

		apiAuthRecord := record.(*APIAuthRecord)

		if apiAuthRecord == nil {
			ThrowError(c, ForbiddenError("未找到API认证权限", nil))
			c.Abort()
			return
		}

		// 若设置了public，则若提供了Authorization就检查，否则不检查

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {

			// 没提供令牌，但是公开，直接放行
			if apiAuthRecord.IsPublic {
				c.Next()
				return
			}

			ThrowError(c, UnauthorizedError("未提供认证令牌", nil))
			c.Abort()
			return
		}

		// 下面是检查逻辑，在提供了令牌或者为非public时必须检查
		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {

			// 令牌格式错误
			if apiAuthRecord.IsPublic {
				logging.Warn("JWTAuthMiddleware: Authorization header format invalid but API is public, allowing request")
				c.Next()
				return
			}

			ThrowError(c, UnauthorizedError("认证令牌格式错误", nil))
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {

			// 令牌格式错误
			if apiAuthRecord.IsPublic {
				logging.Warn("JWTAuthMiddleware: Authorization token empty but API is public, allowing request")
				c.Next()
				return
			}

			ThrowError(c, UnauthorizedError("认证令牌为空", nil))
			c.Abort()
			return
		}

		// 验证token
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {

			// 令牌格式错误
			if apiAuthRecord.IsPublic {
				logging.Warn("JWTAuthMiddleware: token invalid but API is public, allowing request")
				c.Next()
				return
			}

			ThrowError(c, UnauthorizedError("认证令牌无效", err.Error()))
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("jwt_claims", claims)

		if apiAuthRecord.IsPublic {
			// 公开API，直接放行
			c.Next()
			return
		}

		// 需要认证的接口
		res, err := funcs.HasAnyPermissionsOptimized(context.Background(), claims.UserID, apiAuthRecord.Permissions)
		if err != nil {
			ThrowError(c, InternalServerError("权限检查失败", err.Error()))
			c.Abort()
			return
		}
		if !res {
			ThrowError(c, ForbiddenError("没有访问此API的权限", nil))
			c.Abort()
			return
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
