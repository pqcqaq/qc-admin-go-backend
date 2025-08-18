package handlers

import (
	"context"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/pkg/jwt"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// SendVerifyCode 发送验证码
// @Summary      发送验证码
// @Description  发送验证码到指定标识符
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.SendVerifyCodeRequest true "发送验证码请求"
// @Success      200 {object} models.SendVerifyCodeResponse
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/send-verify-code [post]
func (h *AuthHandler) SendVerifyCode(c *gin.Context) {
	var req models.SendVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求参数格式错误", err.Error()))
		return
	}

	err := funcs.SendVerificationCode(context.Background(), req.SenderType, req.Purpose, req.Identifier)
	if err != nil {
		middleware.ThrowError(c, middleware.BusinessError("发送验证码失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.SendVerifyCodeResponse{
			Message: "验证码发送成功",
		},
	})
}

// VerifyCode 验证验证码（测试接口）
// @Summary      验证验证码
// @Description  验证验证码是否正确（仅用于测试）
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.VerifyCodeRequest true "验证验证码请求"
// @Success      200 {object} models.VerifyCodeResponse
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/verify-code [post]
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req models.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求参数格式错误", err.Error()))
		return
	}

	err := funcs.VerifyCode(context.Background(), req.SenderType, req.Purpose, req.Identifier, req.Code)
	if err != nil {
		middleware.ThrowError(c, middleware.BusinessError("验证码验证失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.VerifyCodeResponse{
			Message: "验证码验证成功",
		},
	})
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录认证
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "登录请求"
// @Success      200 {object} models.LoginResponse
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求参数格式错误", err.Error()))
		return
	}

	// 参数验证
	if req.CredentialType == "password" && req.Secret == "" {
		middleware.ThrowError(c, middleware.ValidationError("密码登录必须提供密码", ""))
		return
	}
	if req.CredentialType != "password" && req.VerifyCode == "" {
		middleware.ThrowError(c, middleware.ValidationError("非密码登录必须提供验证码", ""))
		return
	}

	user, err := funcs.UserLogin(context.Background(), req.CredentialType, req.Identifier, req.Secret, req.VerifyCode)
	if err != nil {
		middleware.ThrowError(c, middleware.UnauthorizedError("登录失败", err.Error()))
		return
	}

	// 构建用户信息和Token
	userInfo, token, err := funcs.BuildUserInfoWithToken(context.Background(), user, true)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("构建用户信息失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.LoginResponse{
			User:    *userInfo,
			Token:   token,
			Message: "登录成功",
		},
	})
}

// Register 用户注册
// @Summary      用户注册
// @Description  用户注册
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RegisterRequest true "注册请求"
// @Success      200 {object} models.RegisterResponse
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      409 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求参数格式错误", err.Error()))
		return
	}

	// 参数验证
	if req.CredentialType == "password" && req.Secret == "" {
		middleware.ThrowError(c, middleware.ValidationError("密码注册必须提供密码", ""))
		return
	}
	if req.CredentialType != "password" && req.VerifyCode == "" {
		middleware.ThrowError(c, middleware.ValidationError("非密码注册必须提供验证码", ""))
		return
	}

	user, err := funcs.UserRegister(context.Background(), req.CredentialType, req.Identifier, req.Secret, req.VerifyCode, req.Username)
	if err != nil {
		if err.Error() == "用户已存在" {
			middleware.ThrowError(c, middleware.UserExistsError(err.Error()))
		} else {
			middleware.ThrowError(c, middleware.BusinessError("注册失败", err.Error()))
		}
		return
	}

	// 构建用户信息，注册时可选择是否生成Token
	userInfo, token, err := funcs.BuildUserInfoWithToken(context.Background(), user, true)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("构建用户信息失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.RegisterResponse{
			User:    *userInfo,
			Token:   token,
			Message: "注册成功",
		},
	})
}

// ResetPassword 重置密码
// @Summary      重置密码
// @Description  重置用户密码
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.ResetPasswordRequest true "重置密码请求"
// @Success      200 {object} models.ResetPasswordResponse
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      404 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求参数格式错误", err.Error()))
		return
	}

	// 参数验证
	if req.CredentialType == "password" && req.OldPassword == "" {
		middleware.ThrowError(c, middleware.ValidationError("密码重置必须提供原密码", ""))
		return
	}
	if req.CredentialType != "password" && req.VerifyCode == "" {
		middleware.ThrowError(c, middleware.ValidationError("非密码重置必须提供验证码", ""))
		return
	}

	err := funcs.ResetPassword(context.Background(), req.CredentialType, req.Identifier, req.NewPassword, req.VerifyCode, req.OldPassword)
	if err != nil {
		if err.Error() == "用户不存在" {
			middleware.ThrowError(c, middleware.UserNotFoundError(err.Error()))
		} else if err.Error() == "原密码错误" {
			middleware.ThrowError(c, middleware.UnauthorizedError("原密码错误", err.Error()))
		} else {
			middleware.ThrowError(c, middleware.BusinessError("重置密码失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.ResetPasswordResponse{
			Message: "密码重置成功",
		},
	})
}

// RefreshToken 刷新Token
// @Summary      刷新Token
// @Description  刷新JWT Token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{success=bool,data=object{token=string,message=string}}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从中间件获取当前token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 7 {
		middleware.ThrowError(c, middleware.UnauthorizedError("无效的认证头", ""))
		return
	}

	currentToken := authHeader[7:] // 去掉"Bearer "

	// 刷新token
	newToken, err := jwt.RefreshToken(currentToken)
	if err != nil {
		middleware.ThrowError(c, middleware.UnauthorizedError("Token刷新失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"token":   newToken,
			"message": "Token刷新成功",
		},
	})
}

// GetUserInfo 获取当前用户信息
// @Summary      获取当前用户信息
// @Description  获取当前登录用户的详细信息
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{success=bool,data=models.UserInfo}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      404 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/user-info [get]
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.ThrowError(c, middleware.UnauthorizedError("未找到用户信息", ""))
		return
	}

	// 获取用户信息
	user, err := funcs.GetUserByID(context.Background(), userID)
	if err != nil {
		middleware.ThrowError(c, middleware.NotFoundError("用户不存在", err.Error()))
		return
	}

	// 构建完整的用户信息（包含角色和权限，但不包含新token）
	userInfo, _, err := funcs.BuildUserInfoWithToken(context.Background(), user, false)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("构建用户信息失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    userInfo,
	})
}
