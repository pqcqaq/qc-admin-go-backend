package funcs

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/credential"
	"go-backend/database/ent/user"
	"go-backend/pkg/database"
	"go-backend/pkg/jwt"
	"go-backend/pkg/logging"
	"go-backend/shared/models"

	"golang.org/x/crypto/argon2"
)

const (
	// 密码哈希参数
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// 认证用途常量
const (
	PurposeLogin         = "login"          // 登录
	PurposeRegister      = "register"       // 注册
	PurposeResetPassword = "reset_password" // 重置密码
)

// 认证方式常量
const (
	CredentialTypePassword = "password"
	CredentialTypeEmail    = "email"
	CredentialTypePhone    = "phone"
	CredentialTypeOauth    = "oauth"
	CredentialTypeTotp     = "totp"
)

// hashPassword 哈希密码
func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("生成盐值失败: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// 将盐值和哈希值组合并编码
	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// verifyPassword 验证密码
func verifyPassword(password, hashedPassword string) (bool, error) {
	decoded, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false, fmt.Errorf("解码哈希密码失败: %w", err)
	}

	if len(decoded) != argonSaltLen+argonKeyLen {
		return false, fmt.Errorf("哈希密码格式错误")
	}

	salt := decoded[:argonSaltLen]
	hash := decoded[argonSaltLen:]

	newHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return subtle.ConstantTimeCompare(hash, newHash) == 1, nil
}

// UserLogin 用户登录
func UserLogin(ctx context.Context, credentialType, identifier, secret, verifyCodeStr string) (*ent.User, error) {
	// 首先查找用户认证信息
	credentialRecord, err := database.Client.Credential.Query().
		Where(
			credential.CredentialTypeEQ(credential.CredentialType(credentialType)),
			credential.Identifier(identifier),
		).
		WithUser().
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("用户不存在或认证信息无效")
		}
		return nil, fmt.Errorf("查询用户认证信息失败: %w", err)
	}

	// 检查用户状态
	userRecord := credentialRecord.Edges.User
	if userRecord == nil {
		return nil, fmt.Errorf("用户信息异常")
	}

	if userRecord.Status == user.StatusInactive {
		return nil, fmt.Errorf("用户账号已禁用")
	}

	// 检查认证是否锁定
	if credentialRecord.LockedUntil != nil && credentialRecord.LockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("账号已锁定，请稍后再试")
	}

	// 根据认证类型进行验证
	var authSuccess bool

	if credentialType == CredentialTypePassword {
		// 密码登录直接校验密码
		if credentialRecord.Secret == "" {
			return nil, fmt.Errorf("未设置密码")
		}

		match, err := verifyPassword(secret, credentialRecord.Secret)
		if err != nil {
			return nil, fmt.Errorf("密码验证失败: %w", err)
		}
		authSuccess = match
	} else {
		// 其他认证方式需要验证码
		if verifyCodeStr == "" {
			return nil, fmt.Errorf("请提供验证码")
		}

		err = VerifyCode(ctx, credentialType, PurposeLogin, identifier, verifyCodeStr)
		if err != nil {
			authSuccess = false
		} else {
			authSuccess = true
		}
	}

	// 更新认证记录
	updateBuilder := credentialRecord.Update().
		SetLastUsedAt(time.Now())

	if authSuccess {
		// 认证成功，重置失败次数
		updateBuilder = updateBuilder.
			SetFailedAttempts(0).
			ClearLockedUntil()
	} else {
		// 认证失败，增加失败次数
		failedAttempts := credentialRecord.FailedAttempts + 1
		updateBuilder = updateBuilder.SetFailedAttempts(failedAttempts)

		// 失败次数达到5次，锁定30分钟
		if failedAttempts >= 5 {
			lockUntil := time.Now().Add(30 * time.Minute)
			updateBuilder = updateBuilder.SetLockedUntil(lockUntil)
		}
	}

	_, err = updateBuilder.Save(ctx)
	if err != nil {
		// 更新失败不影响认证结果
		logging.Warn("更新认证记录失败: %v\n", err)
	}

	if !authSuccess {
		if credentialType == CredentialTypePassword {
			return nil, fmt.Errorf("用户名或密码错误")
		} else {
			return nil, fmt.Errorf("验证码错误或已过期")
		}
	}

	return userRecord, nil
}

// UserRegister 用户注册
func UserRegister(ctx context.Context, credentialType, identifier, secret, verifyCodeStr, username string) (*ent.User, error) {
	// 检查用户是否已存在
	exists, err := database.Client.Credential.Query().
		Where(
			credential.CredentialTypeEQ(credential.CredentialType(credentialType)),
			credential.Identifier(identifier),
		).
		Exist(ctx)

	if err != nil {
		return nil, fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("用户已存在")
	}

	// 如果不是密码注册，需要验证验证码
	if credentialType != CredentialTypePassword {
		if verifyCodeStr == "" {
			return nil, fmt.Errorf("请提供验证码")
		}

		err = VerifyCode(ctx, credentialType, PurposeRegister, identifier, verifyCodeStr)
		if err != nil {
			return nil, fmt.Errorf("验证码验证失败: %w", err)
		}
	}

	// 开始事务
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 创建用户
	userRecord, err := tx.User.Create().
		SetName(username).
		SetStatus(user.StatusActive).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 处理密码哈希
	var hashedSecret string
	if secret != "" {
		hashed, err := hashPassword(secret)
		if err != nil {
			return nil, fmt.Errorf("密码哈希失败: %w", err)
		}
		hashedSecret = hashed
	}

	// 创建认证记录
	credBuilder := tx.Credential.Create().
		SetUserID(userRecord.ID).
		SetCredentialType(credential.CredentialType(credentialType)).
		SetIdentifier(identifier).
		SetSecret(hashedSecret).
		SetIsVerified(true) // 注册成功即为已验证

	_, err = credBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建认证记录失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return userRecord, nil
}

// ResetPassword 重置密码
func ResetPassword(ctx context.Context, credentialType, identifier, newPassword, verifyCodeStr, oldPassword string) error {
	// 查找用户认证信息
	credentialRecord, err := database.Client.Credential.Query().
		Where(
			credential.CredentialTypeEQ(credential.CredentialType(credentialType)),
			credential.Identifier(identifier),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("用户不存在")
		}
		return fmt.Errorf("查询用户认证信息失败: %w", err)
	}

	// 根据认证类型进行验证
	if credentialType == CredentialTypePassword {
		// 密码重置需要验证旧密码
		if oldPassword == "" {
			return fmt.Errorf("请提供原密码")
		}

		if credentialRecord.Secret == "" {
			return fmt.Errorf("未设置密码")
		}

		match, err := verifyPassword(oldPassword, credentialRecord.Secret)
		if err != nil {
			return fmt.Errorf("原密码验证失败: %w", err)
		}

		if !match {
			return fmt.Errorf("原密码错误")
		}
	} else {
		// 其他认证方式需要验证码
		if verifyCodeStr == "" {
			return fmt.Errorf("请提供验证码")
		}

		err = VerifyCode(ctx, credentialType, PurposeResetPassword, identifier, verifyCodeStr)
		if err != nil {
			return fmt.Errorf("验证码验证失败: %w", err)
		}
	}

	// 哈希新密码
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("新密码哈希失败: %w", err)
	}

	// 更新密码
	_, err = credentialRecord.Update().
		SetSecret(hashedPassword).
		SetFailedAttempts(0).
		ClearLockedUntil().
		Save(ctx)

	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// BuildUserInfoWithToken 构建包含Token和角色权限的用户信息
func BuildUserInfoWithToken(ctx context.Context, user *ent.User, includeToken bool) (*models.UserInfo, string, error) {
	userInfo := &models.UserInfo{
		ID:         user.ID,
		Name:       user.Name,
		Status:     string(user.Status),
		CreateTime: user.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime: user.UpdateTime.Format("2006-01-02 15:04:05"),
	}

	// 获取用户角色
	roles, err := GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("获取用户角色失败: %w", err)
	}

	// 转换角色信息
	userInfo.Roles = make([]*models.RoleResponse, len(roles))
	for i, role := range roles {
		userInfo.Roles[i] = ConvertRoleToResponse(role)
	}

	// 获取用户权限（通过角色继承）
	permissions, err := GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("获取用户权限失败: %w", err)
	}

	// 转换权限信息
	userInfo.Permissions = make([]*models.PermissionResponse, len(permissions))
	for i, permission := range permissions {
		userInfo.Permissions[i] = ConvertPermissionToResponse(permission)
	}

	// 生成JWT Token
	var token string
	if includeToken {
		token, err = jwt.GenerateToken(user.ID)
		if err != nil {
			return nil, "", fmt.Errorf("生成JWT Token失败: %w", err)
		}
	}

	return userInfo, token, nil
}
