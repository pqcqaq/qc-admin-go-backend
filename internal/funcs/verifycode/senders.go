package verifycode

import (
	"context"
	"go-backend/pkg/email"
	"go-backend/pkg/logging"
	"go-backend/pkg/sms"
)

// EmailCodeSender 邮箱验证码发送器
type EmailCodeSender struct{}

func (s *EmailCodeSender) Send(ctx context.Context, identifier, code, purpose string) error {
	// 使用邮件服务发送验证码
	err := email.SendVerificationCode(identifier, code, purpose)
	if err != nil {
		logging.Error("邮件验证码发送失败: identifier=%s, purpose=%s, error=%v", identifier, purpose, err)
		return err
	}

	logging.Info("邮件验证码发送成功: identifier=%s, purpose=%s", identifier, purpose)
	return nil
}

func (s *EmailCodeSender) GetType() SenderType {
	return EmailSender
}

// PhoneCodeSender 手机验证码发送器
type PhoneCodeSender struct{}

func (s *PhoneCodeSender) Send(ctx context.Context, identifier, code, purpose string) error {
	// 使用短信服务发送验证码
	err := sms.SendVerificationCode(identifier, code, purpose)
	if err != nil {
		logging.Error("短信验证码发送失败: identifier=%s, purpose=%s, error=%v", identifier, purpose, err)
		return err
	}

	logging.Info("短信验证码发送成功: identifier=%s, purpose=%s", identifier, purpose)
	return nil
}

func (s *PhoneCodeSender) GetType() SenderType {
	return PhoneSender
}
