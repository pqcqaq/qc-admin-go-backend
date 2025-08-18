package verifycode

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// SenderType 发送器类型
type SenderType string

const (
	EmailSender SenderType = "email"
	PhoneSender SenderType = "phone"
	SMSSender   SenderType = "sms"
)

// Sender 验证码发送器接口
type Sender interface {
	Send(ctx context.Context, identifier, code, purpose string) error
	GetType() SenderType
}

// SenderFactory 发送器工厂
type SenderFactory struct {
	senders map[SenderType]Sender
}

// NewSenderFactory 创建发送器工厂
func NewSenderFactory() *SenderFactory {
	factory := &SenderFactory{
		senders: make(map[SenderType]Sender),
	}

	// 注册默认发送器
	factory.RegisterSender(EmailSender, &EmailCodeSender{})
	factory.RegisterSender(PhoneSender, &PhoneCodeSender{})
	factory.RegisterSender(SMSSender, &SMSCodeSender{})

	return factory
}

// RegisterSender 注册发送器
func (f *SenderFactory) RegisterSender(senderType SenderType, sender Sender) {
	f.senders[senderType] = sender
}

// GetSender 获取发送器
func (f *SenderFactory) GetSender(senderType SenderType) (Sender, error) {
	sender, exists := f.senders[senderType]
	if !exists {
		return nil, fmt.Errorf("sender type %s not supported", senderType)
	}
	return sender, nil
}

// GenerateCode 生成验证码
func GenerateCode(length int) string {
	if length <= 0 {
		length = 6
	}

	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := range code {
		code[i] = byte('0' + rand.Intn(10))
	}
	return string(code)
}

// 全局发送器工厂实例
var DefaultSenderFactory = NewSenderFactory()
