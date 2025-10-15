package handlers

import (
	"fmt"
	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/pkg/logging"
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

	err := funcs.VerifyCodeFuncs{}.SendVerificationCode(middleware.GetRequestContext(c), req.SenderType, req.Purpose, req.Identifier, req.ClientCode)
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

	err := funcs.VerifyCodeFuncs{}.VerifyCode(middleware.GetRequestContext(c), req.SenderType, req.Purpose, req.Identifier, req.Code)
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

	user, err := funcs.AuthFuncs{}.UserLoginWithContext(
		middleware.GetRequestContext(c),
		c,
		req.CredentialType,
		req.Identifier,
		req.Secret,
		req.VerifyCode,
		req.ClientCode,
	)
	if err != nil {
		middleware.ThrowError(c, middleware.UnauthorizedError("登录失败", err.Error()))
		return
	}

	clientIdAny, ex := c.Get("client_device_id")
	if !ex {
		middleware.ThrowError(c, middleware.InternalServerError("找不到终端", fmt.Errorf("cannot find client_device_id in gin context")))
	}

	var clientId uint64 = clientIdAny.(uint64)

	// 构建用户信息和Token
	var rememberMe bool = false
	if req.RememberMe != nil {
		rememberMe = *req.RememberMe
	}
	userInfo, token, err := funcs.AuthFuncs{}.BuildUserInfoWithToken(middleware.GetRequestContext(c), user, &clientId, rememberMe)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("构建用户信息失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.LoginResponse{
			User:    *userInfo,
			Token:   *token,
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

	ctx := middleware.GetRequestContext(c)
	user, err := funcs.AuthFuncs{}.UserRegister(ctx, req.CredentialType, req.Identifier, req.Secret, req.VerifyCode, req.Username)
	if err != nil {
		if err.Error() == "用户已存在" {
			middleware.ThrowError(c, middleware.UserExistsError(err.Error()))
		} else {
			middleware.ThrowError(c, middleware.BusinessError("注册失败", err.Error()))
		}
		return
	}

	cd, err := funcs.ClientDeviceFuncs{}.GetClientDeviceByCodeInner(ctx, req.ClientCode)

	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("查找设备类型失败", err.Error()))
	}

	// 构建用户信息，注册时可选择是否生成Token
	userInfo, token, err := funcs.AuthFuncs{}.BuildUserInfoWithToken(ctx, user, &cd.ID, false)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("构建用户信息失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": models.RegisterResponse{
			User:    *userInfo,
			Token:   *token,
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

	err := funcs.AuthFuncs{}.ResetPassword(middleware.GetRequestContext(c), req.CredentialType, req.Identifier, req.NewPassword, req.VerifyCode, req.OldPassword)
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
// @Param        request body models.RefreshTokenRequest true "刷新Token请求"
// @Success      200 {object} object{success=bool,data=object{token=string,message=string}}
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	// 绑定并验证请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("请求参数错误", err.Error()))
		return
	}

	token, ex := middleware.GetCurrentAccessToken(c)
	var tokenStr string
	if ex {
		tokenStr = token
	} else {
		tokenStr = ""
	}
	ctx := middleware.GetRequestContext(c)

	newTokenInfo, err := funcs.AuthFuncs{}.RefreshToken(ctx, tokenStr, req.RefreshToken)

	if err != nil {
		middleware.ThrowError(c, middleware.UnauthorizedError("Token刷新失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"token":   newTokenInfo,
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
	user, err := funcs.UserFuncs{}.GetUserByID(middleware.GetRequestContext(c), userID)
	if err != nil {
		middleware.ThrowError(c, middleware.NotFoundError("用户不存在", err.Error()))
		return
	}

	// 构建完整的用户信息（包含角色和权限，但不包含新token）
	userInfo, _, err := funcs.AuthFuncs{}.BuildUserInfoWithToken(middleware.GetRequestContext(c), user, nil, false)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("构建用户信息失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    userInfo,
	})
}

// Logout 用户登出
// @Summary      用户登出
// @Description  用户登出，更新登录记录
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{success=bool,data=object{message=string}}
// @Failure      401 {object} object{success=bool,message=string}
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 尝试获取会话ID
	sessionID, exists := c.Get("session_id")
	if exists && sessionID != "" {
		// 更新登录记录的退出时间
		err := funcs.LoginRecordFuncs{}.UpdateLoginRecordLogout(middleware.GetRequestContext(c), sessionID.(string))
		if err != nil {
			// 记录错误但不影响登出响应
			logging.Warn("更新登录记录退出信息失败: %v", err)
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"message": "登出成功",
		},
	})
}

// GetUserMenuTree 获取当前用户的菜单树
// @Summary      获取当前用户的菜单树
// @Description  根据当前用户的角色和权限，返回该用户可访问的菜单树形结构
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{success=bool,data=[]object}
// @Failure      401  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /auth/user-menu-tree [get]
func (h *AuthHandler) GetUserMenuTree(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.ThrowError(c, middleware.UnauthorizedError("未找到用户信息", ""))
		return
	}

	// 获取用户的菜单树
	menuTree, err := funcs.UserFuncs{}.GetUserMenuTree(middleware.GetRequestContext(c), userID)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("获取用户菜单树失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    menuTree,
	})
}
