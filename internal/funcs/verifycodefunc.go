package funcs

import (
	"context"
	"fmt"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/verifycode"
	vcpkg "go-backend/internal/funcs/verifycode"
	"go-backend/pkg/database"
)

// SendVerificationCode 发送验证码通用接口
func SendVerificationCode(ctx context.Context, senderType, purpose, identifier string) error {
	// 检查30秒内是否已发送过验证码
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)
	exists, err := database.Client.VerifyCode.Query().
		Where(
			verifycode.Identifier(identifier),
			verifycode.SenderTypeEQ(verifycode.SenderType(senderType)),
			verifycode.SendFor(purpose),
			verifycode.SendAtGTE(thirtySecondsAgo),
			verifycode.DeleteTimeIsNil(),
		).
		Exist(ctx)

	if err != nil {
		return fmt.Errorf("检查验证码发送记录失败: %w", err)
	}

	if exists {
		return fmt.Errorf("验证码发送过于频繁，请稍后再试")
	}

	// 生成验证码
	code := vcpkg.GenerateCode(6)

	// 获取发送器
	factory := vcpkg.DefaultSenderFactory
	sender, err := factory.GetSender(vcpkg.SenderType(senderType))
	if err != nil {
		return fmt.Errorf("不支持的发送方式: %w", err)
	}

	// 创建验证码记录
	now := time.Now()
	expiresAt := now.Add(15 * time.Minute) // 15分钟过期

	verifyCodeRecord, err := database.Client.VerifyCode.Create().
		SetCode(code).
		SetIdentifier(identifier).
		SetSenderType(verifycode.SenderType(senderType)).
		SetSendFor(purpose).
		SetExpiresAt(expiresAt).
		SetSendSuccess(false).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("创建验证码记录失败: %w", err)
	}

	// 发送验证码
	err = sender.Send(ctx, identifier, code, purpose)
	if err != nil {
		// 发送失败，保持记录但不更新发送状态
		return fmt.Errorf("发送验证码失败: %w", err)
	}

	// 更新发送状态
	_, err = verifyCodeRecord.Update().
		SetSendSuccess(true).
		SetSendAt(now).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("更新验证码发送状态失败: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码通用接口
func VerifyCode(ctx context.Context, senderType, purpose, identifier, code string) error {
	// 查询15分钟内有效的验证码
	fifteenMinutesAgo := time.Now().Add(-15 * time.Minute)
	now := time.Now()

	verifyCodeRecord, err := database.Client.VerifyCode.Query().
		Where(
			verifycode.Identifier(identifier),
			verifycode.SenderTypeEQ(verifycode.SenderType(senderType)),
			verifycode.SendFor(purpose),
			verifycode.Code(code),
			verifycode.SendSuccess(true),
			verifycode.ExpiresAtGTE(now),
			verifycode.CreateTimeGTE(fifteenMinutesAgo),
			verifycode.UsedAtIsNil(),
			verifycode.DeleteTimeIsNil(),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("验证码无效或已过期")
		}
		return fmt.Errorf("验证验证码失败: %w", err)
	}

	// 标记验证码已使用
	_, err = verifyCodeRecord.Update().
		SetUsedAt(now).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("更新验证码使用状态失败: %w", err)
	}

	return nil
}
