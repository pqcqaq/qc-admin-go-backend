package funcs

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/clientdevice"
	"go-backend/database/ent/credential"
	"go-backend/database/ent/user"
	"go-backend/pkg/database"
	"go-backend/pkg/jwt"
	"go-backend/pkg/logging"
	"go-backend/pkg/utils"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
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

type AuthFuncs struct{}

// hashPassword 哈希密码
func (AuthFuncs) hashPassword(password string) (string, string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", "", fmt.Errorf("生成盐值失败: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// 返回哈希值和盐值
	saltStr := base64.StdEncoding.EncodeToString(salt)
	hashStr := base64.StdEncoding.EncodeToString(hash)
	return hashStr, saltStr, nil
}

// verifyPassword 验证密码
func (AuthFuncs) verifyPassword(password, hashedPassword, saltStr string) (bool, error) {
	// 如果没有盐值，尝试使用旧格式（向后兼容）
	if saltStr == "" {
		return AuthFuncs{}.verifyPasswordLegacy(password, hashedPassword)
	}

	// 解码盐值和哈希值
	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return false, fmt.Errorf("解码盐值失败: %w", err)
	}

	hash, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false, fmt.Errorf("解码哈希值失败: %w", err)
	}

	// 计算新的哈希值进行比较
	newHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return subtle.ConstantTimeCompare(hash, newHash) == 1, nil
}

// verifyPasswordLegacy 验证旧格式的密码（向后兼容）
func (AuthFuncs) verifyPasswordLegacy(password, hashedPassword string) (bool, error) {
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
func (AuthFuncs) UserLogin(ctx context.Context, credentialType, identifier, secret, verifyCodeStr, deviceCode string) (*ent.User, error) {
	return AuthFuncs{}.UserLoginWithContext(ctx, nil, credentialType, identifier, secret, verifyCodeStr, deviceCode)
}

// UserLoginWithContext 用户登录（带上下文记录）
func (AuthFuncs) UserLoginWithContext(ctx context.Context, ginCtx *gin.Context, credentialType, identifier, secret, verifyCodeStr, deviceCode string) (*ent.User, error) {
	var userRecord *ent.User
	var sessionID string
	var loginStatus = LoginStatusFailed
	var failureReason string

	// 生成会话ID
	if ginCtx != nil {
		sessionID = AuthFuncs{}.generateSessionID()
	}

	// 找设备信息
	clientDevice, err := ClientDeviceFuncs{}.GetClientDeviceByCodeInner(ctx, deviceCode)

	if err != nil {
		if ent.IsNotFound(err) {
			failureReason = "终端信息无效"
			return nil, fmt.Errorf(failureReason)
		}
		failureReason = "查询终端类型失败"
		return nil, fmt.Errorf("%s: %w", failureReason, err)
	}

	// 函数结束时记录登录日志
	defer func() {
		if ginCtx != nil && userRecord != nil {
			// 只在有用户记录时记录日志
			_, logErr := LoginRecordFuncs{}.CreateLoginRecordFromGinContext(
				ctx, ginCtx, userRecord.ID, identifier, credentialType,
				loginStatus, failureReason, sessionID, clientDevice.ID,
			)
			if logErr != nil {
				logging.Warn("记录登录日志失败: %v", logErr)
			}
		} else if ginCtx != nil {
			// 即使没有用户记录也要记录失败尝试（使用0作为用户ID）
			_, logErr := LoginRecordFuncs{}.CreateLoginRecordFromGinContext(
				ctx, ginCtx, 0, identifier, credentialType,
				loginStatus, failureReason, "", clientDevice.ID,
			)
			if logErr != nil {
				logging.Warn("记录登录失败日志失败: %v", logErr)
			}
		}
	}()

	// 查找用户认证信息
	credentialRecord, err := database.Client.Credential.Query().
		Where(
			credential.CredentialTypeEQ(credential.CredentialType(credentialType)),
			credential.Identifier(identifier),
		).
		WithUser().
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			failureReason = "用户不存在或认证信息无效"
			return nil, fmt.Errorf(failureReason)
		}
		failureReason = "查询用户认证信息失败"
		return nil, fmt.Errorf("%s: %w", failureReason, err)
	}

	// 检查用户状态
	userRecord = credentialRecord.Edges.User
	if userRecord == nil {
		failureReason = "用户信息异常"
		return nil, fmt.Errorf(failureReason)
	}

	if userRecord.Status == user.StatusInactive {
		failureReason = "用户账号已禁用"
		return nil, fmt.Errorf(failureReason)
	}

	var needRoleIds []uint64
	roles := clientDevice.Edges.Roles
	for _, role := range roles {
		needRoleIds = append(needRoleIds, role.ID)
	}

	// 如果用户角色中不存在，则不允许登录
	has, err := HasAnyRoleId(ctx, userRecord.ID, needRoleIds)

	if err != nil {
		return nil, fmt.Errorf("判断用户权限失败")
	}

	if !has {
		return nil, fmt.Errorf("用户没有权限使用这个终端进行登录")
	}

	// 检查认证是否锁定
	now := time.Now()
	if credentialRecord.LockedUntil != nil && credentialRecord.LockedUntil.After(now) {
		// 在锁定期内，记录尝试次数但不进行认证
		_, updateErr := credentialRecord.Update().
			SetLastUsedAt(now).
			Save(ctx)
		if updateErr != nil {
			logging.Warn("更新锁定期间的尝试记录失败: %v\n", updateErr)
		}

		remainingTime := credentialRecord.LockedUntil.Sub(now)
		loginStatus = LoginStatusLocked
		failureReason = fmt.Sprintf("账号已锁定，剩余时间: %v", remainingTime.Round(time.Minute))
		return nil, fmt.Errorf(failureReason)
	}

	// 如果锁定时间已过期，自动解锁并重置失败次数
	if credentialRecord.LockedUntil != nil && !credentialRecord.LockedUntil.After(now) {
		_, err = credentialRecord.Update().
			SetFailedAttempts(0).
			ClearLockedUntil().
			SetLastUsedAt(now).
			Save(ctx)
		if err != nil {
			logging.Warn("自动解锁失败: %v\n", err)
		} else {
			logging.Info("账号自动解锁: %s", credentialRecord.Identifier)
		}
		// 重新查询更新后的记录
		credentialRecord, err = database.Client.Credential.Query().
			Where(
				credential.CredentialTypeEQ(credential.CredentialType(credentialType)),
				credential.Identifier(identifier),
			).
			WithUser().
			First(ctx)
		if err != nil {
			failureReason = "重新查询认证信息失败"
			return nil, fmt.Errorf("%s: %w", failureReason, err)
		}
	}

	// 根据认证类型进行验证
	var authSuccess bool

	if credentialType == CredentialTypePassword {
		// 密码登录直接校验密码
		if credentialRecord.Secret == "" {
			failureReason = "未设置密码"
			return nil, fmt.Errorf(failureReason)
		}

		match, err := AuthFuncs{}.verifyPassword(secret, credentialRecord.Secret, credentialRecord.Salt)
		if err != nil {
			failureReason = "密码验证失败"
			return nil, fmt.Errorf("%s: %w", failureReason, err)
		}
		authSuccess = match
		if !authSuccess {
			failureReason = "用户名或密码错误"
		}
	} else {
		// 其他认证方式需要验证码
		if verifyCodeStr == "" {
			failureReason = "请提供验证码"
			return nil, fmt.Errorf(failureReason)
		}

		err = VerifyCodeFuncs{}.VerifyCode(ctx, credentialType, PurposeLogin, identifier, verifyCodeStr)
		if err != nil {
			authSuccess = false
			failureReason = "验证码错误或已过期"
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
		loginStatus = LoginStatusSuccess
		failureReason = "" // 清空失败原因
	} else {
		// 认证失败，增加失败次数
		failedAttempts := credentialRecord.FailedAttempts + 1
		updateBuilder = updateBuilder.SetFailedAttempts(failedAttempts)

		// 失败次数达到5次，锁定30分钟
		if failedAttempts >= 5 {
			lockUntil := time.Now().Add(30 * time.Minute)
			updateBuilder = updateBuilder.SetLockedUntil(lockUntil)
			loginStatus = LoginStatusLocked
			failureReason = "账号因多次失败尝试被锁定"
		}
	}

	_, err = updateBuilder.Save(ctx)
	if err != nil {
		// 更新失败不影响认证结果
		logging.Warn("更新认证记录失败: %v\n", err)
	}

	if !authSuccess {
		if failureReason == "" {
			if credentialType == CredentialTypePassword {
				failureReason = "用户名或密码错误"
			} else {
				failureReason = "验证码错误或已过期"
			}
		}
		return nil, fmt.Errorf(failureReason)
	}

	// 如果有gin上下文，将sessionID存储到上下文中，后续可以用于退出登录时更新记录
	if ginCtx != nil && sessionID != "" {
		ginCtx.Set("session_id", sessionID)
		ginCtx.Set("client_device_id", clientDevice.ID)
	}

	return userRecord, nil
}

// generateSessionID 生成会话ID
func (AuthFuncs) generateSessionID() string {
	randomStr, err := utils.GenerateRandomString(16)
	if err != nil {
		// 如果生成随机字符串失败，使用时间戳作为替代
		randomStr = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("session_%d_%s", time.Now().UnixNano(), randomStr)
}

// UserRegister 用户注册
func (AuthFuncs) UserRegister(ctx context.Context, credentialType, identifier, secret, verifyCodeStr, username string) (*ent.User, error) {
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

		err = VerifyCodeFuncs{}.VerifyCode(ctx, credentialType, PurposeRegister, identifier, verifyCodeStr)
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
	var hashedSecret, saltStr string
	if secret != "" {
		hash, salt, err := AuthFuncs{}.hashPassword(secret)
		if err != nil {
			return nil, fmt.Errorf("密码哈希失败: %w", err)
		}
		hashedSecret = hash
		saltStr = salt
	}

	// 创建主认证记录（用户注册的认证方式）
	credBuilder := tx.Credential.Create().
		SetUserID(userRecord.ID).
		SetCredentialType(credential.CredentialType(credentialType)).
		SetIdentifier(identifier).
		SetIsVerified(true) // 注册成功即为已验证

	// 如果有盐值，设置盐值字段
	if saltStr != "" && hashedSecret != "" && credentialType == CredentialTypePassword {
		credBuilder = credBuilder.SetSecret(hashedSecret)
		credBuilder = credBuilder.SetSalt(saltStr)
	}

	_, err = credBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建认证记录失败: %w", err)
	}

	// 如果提供了密码且注册方式不是密码注册，创建额外的密码认证记录
	if secret != "" && credentialType != CredentialTypePassword {
		_, err = tx.Credential.Create().
			SetUserID(userRecord.ID).
			SetCredentialType(credential.CredentialTypePassword).
			SetIdentifier(username). // 使用用户名作为密码登录的标识符
			SetSecret(hashedSecret).
			SetSalt(saltStr).
			SetIsVerified(true).
			SetFailedAttempts(0).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("创建密码认证记录失败: %w", err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return userRecord, nil
}

// ResetPassword 重置密码
func (AuthFuncs) ResetPassword(ctx context.Context, credentialType, identifier, newPassword, verifyCodeStr, oldPassword string) error {
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 查找用户认证信息
	credentialRecord, err := tx.Credential.Query().
		Where(
			credential.CredentialTypeEQ(credential.CredentialType(credentialType)),
			credential.Identifier(identifier),
		).Only(ctx)

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

		match, err := AuthFuncs{}.verifyPassword(oldPassword, credentialRecord.Secret, credentialRecord.Salt)
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

		err = VerifyCodeFuncs{}.VerifyCode(ctx, credentialType, PurposeResetPassword, identifier, verifyCodeStr)
		if err != nil {
			return fmt.Errorf("验证码验证失败: %w", err)
		}
	}

	// 哈希新密码
	hashedPassword, saltStr, err := AuthFuncs{}.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("新密码哈希失败: %w", err)
	}

	// 查找用户的密码认证记录
	passwordCredential, err := tx.Credential.Query().
		Where(
			credential.UserIDEQ(credentialRecord.UserID),
			credential.CredentialTypeEQ(credential.CredentialTypePassword),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// 如果不存在密码认证记录，则创建一个新的
			// identifier是登录用户 的用户名
			user, err := tx.User.Get(ctx, credentialRecord.UserID)
			if err != nil {
				return fmt.Errorf("查询用户信息失败: %w", err)
			}
			_, err = tx.Credential.Create().
				SetUserID(credentialRecord.UserID).
				SetCredentialType(credential.CredentialTypePassword).
				SetIdentifier(user.Name).
				SetSecret(hashedPassword).
				SetSalt(saltStr).
				SetIsVerified(true).
				SetFailedAttempts(0).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("创建密码认证记录失败: %w", err)
			}
		} else {
			return fmt.Errorf("查询密码认证记录失败: %w", err)
		}
	} else {
		// 如果存在密码认证记录，则更新
		_, err = passwordCredential.Update().
			SetSecret(hashedPassword).
			SetSalt(saltStr).
			SetFailedAttempts(0).
			ClearLockedUntil().
			Save(ctx)
		if err != nil {
			return fmt.Errorf("更新密码失败: %w", err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// BuildUserInfoWithToken 构建包含Token和角色权限的用户信息
func (AuthFuncs) BuildUserInfoWithToken(ctx context.Context, user *ent.User, clientId *uint64, rememberMe bool) (*models.UserInfo, *models.TokenInfo, error) {
	userInfo := &models.UserInfo{
		ID:         utils.ToString(user.ID),
		Name:       user.Name,
		Status:     string(user.Status),
		CreateTime: user.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime: user.UpdateTime.Format("2006-01-02 15:04:05"),
		Age:        user.Age,
		Sex:        string(user.Sex),
	}

	// 查询头像
	avatar, err := database.Client.User.QueryAvatar(user).Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, nil, fmt.Errorf("查询头像信息失败: %w", err)
		}
	} else if avatar != nil {
		userInfo.Avatar = avatar.URL
	}

	// 获取用户角色
	roles, err := UserFuncs{}.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取用户角色失败: %w", err)
	}

	// 转换角色信息
	userInfo.Roles = make([]*models.RoleResponse, len(roles))
	for i, role := range roles {
		userInfo.Roles[i] = RoleFuncs{}.ConvertRoleToResponse(role)
	}

	// 获取用户权限（通过角色继承）
	permissions, err := UserFuncs{}.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取用户权限失败: %w", err)
	}

	// 转换权限信息
	userInfo.Permissions = make([]*models.PermissionResponse, len(permissions))
	for i, permission := range permissions {
		userInfo.Permissions[i] = PermissionFuncs{}.ConvertPermissionToResponse(permission)
	}

	tokenInfo := models.TokenInfo{}

	// 生成JWT Token
	if clientId != nil {

		client, err := database.Client.ClientDevice.Query().Where(clientdevice.IDEQ(*clientId)).Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				return nil, nil, fmt.Errorf("找不到对应id的client")
			}
			return nil, nil, fmt.Errorf("查找终端类型失败")
		}

		tokenInfo.RefreshExpiredIn = uint64(time.Now().Add(time.Duration(client.RefreshTokenExpiry) * time.Millisecond).UnixMilli())
		tokenInfo.AccessExpiredIn = uint64(time.Now().Add(time.Duration(client.AccessTokenExpiry) * time.Millisecond).UnixMilli())
		timeoutAccess := time.Duration(client.AccessTokenExpiry) * time.Millisecond
		timeoutRefresh := time.Duration(client.RefreshTokenExpiry) * time.Millisecond

		tokenInfo.AccessToken, err = jwt.GenerateAccessToken(user.ID, client.ID, timeoutAccess)
		if err != nil {
			return nil, nil, fmt.Errorf("生成JWT Token失败: %w", err)
		}

		tokenInfo.RefreshToken, err = jwt.GenerateRefreshToken(user.ID, client.ID, timeoutRefresh, rememberMe)

		if err != nil {
			return nil, nil, fmt.Errorf("生成JWT Token失败: %w", err)
		}
	}

	return userInfo, &tokenInfo, nil
}
func (AuthFuncs) RefreshToken(ctx context.Context, accessToken, refreshToken string) (*models.TokenInfo, error) {
	// 如果accessToken没有过期，则不允许刷新
	if accessToken != "" {
		claims, err := jwt.ValidateToken(accessToken)
		if err == nil && claims.ExpiresAt != nil && claims.ExpiresAt.After(time.Now()) {
			return nil, fmt.Errorf("access Token未过期，无需刷新")
		}
	}

	// 验证refresh token的有效性
	claims, err := jwt.ValidateToken(refreshToken)
	if err != nil {
		logging.Error("验证refresh Token失败: %v", err)
		return nil, fmt.Errorf("验证Token失败")
	}

	// 获取客户端设备配置信息
	client, err := ClientDeviceFuncs{}.GetClientDeviceByIdInner(ctx, claims.ClientDeviceId)
	if err != nil {
		return nil, fmt.Errorf("获取设备类型失败")
	}

	// 生成新的access token
	timeoutAccess := time.Duration(client.AccessTokenExpiry) * time.Millisecond
	newToken, err := jwt.RefreshToken(refreshToken, client.ID, timeoutAccess)
	if err != nil {
		return nil, fmt.Errorf("token刷新失败: %w", err)
	}

	tokenInfo := models.TokenInfo{
		AccessToken:     newToken,
		AccessExpiredIn: uint64(time.Now().Add(timeoutAccess).UnixMilli()),
	}

	// 如果勾选了"记住我"功能，需要判断是否刷新refresh token
	if claims.RememberMe {
		timeoutRefresh := time.Duration(client.RefreshTokenExpiry) * time.Millisecond

		// 计算refresh token的剩余有效时间
		now := time.Now()
		refreshExpiresAt := time.UnixMilli(int64(claims.Expiry))
		remainingTime := refreshExpiresAt.Sub(now)
		totalTime := timeoutRefresh

		// 只有在refresh token过期时间过半时才重新生成，避免退化成单token模式
		if remainingTime < totalTime/2 {
			// 重新生成refresh token
			newRefreshToken, err := jwt.GenerateRefreshToken(claims.UserID, client.ID, timeoutRefresh, claims.RememberMe)
			if err != nil {
				return nil, fmt.Errorf("refresh Token刷新失败: %w", err)
			}
			tokenInfo.RefreshToken = newRefreshToken
			tokenInfo.RefreshExpiredIn = uint64(now.Add(timeoutRefresh).UnixMilli())
		} else {
			// 未过半，继续使用原refresh token
			tokenInfo.RefreshToken = refreshToken
			tokenInfo.RefreshExpiredIn = claims.Expiry
		}
	} else {
		// 未勾选"记住我"，保持原refresh token不变
		tokenInfo.RefreshToken = refreshToken
		tokenInfo.RefreshExpiredIn = claims.Expiry
	}

	return &tokenInfo, nil
}
