package email

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-backend/pkg/configs"

	"gopkg.in/gomail.v2"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// EmailClient 邮件客户端结构
type EmailClient struct {
	config    *configs.EmailConfig
	dialer    *gomail.Dialer
	templates map[string]*template.Template
}

// 单例相关变量
var (
	Client *EmailClient
	once   sync.Once
	mu     sync.RWMutex
)

// 全局logger实例
var logger LoggerInterface

// SetLogger 设置logger实例
func SetLogger(l LoggerInterface) {
	logger = l
}

// NewClient 创建新的邮件客户端
func NewClient(emailConfig *configs.EmailConfig) (*EmailClient, error) {
	if emailConfig.Host == "" || emailConfig.Username == "" || emailConfig.Password == "" {
		return nil, fmt.Errorf("邮件配置不完整: host, username, password 不能为空")
	}

	if emailConfig.From == "" {
		emailConfig.From = emailConfig.Username
	}

	if emailConfig.TemplateDir == "" {
		emailConfig.TemplateDir = "./templates"
	}

	// 创建SMTP拨号器
	dialer := gomail.NewDialer(
		emailConfig.Host,
		emailConfig.Port,
		emailConfig.Username,
		emailConfig.Password,
	)

	// 设置TLS/SSL
	if emailConfig.UseSSL {
		// SSL连接
		dialer.SSL = true
	} else if emailConfig.UseTLS {
		// TLS连接
		dialer.TLSConfig = nil // 使用默认TLS配置
	}

	client := &EmailClient{
		config:    emailConfig,
		dialer:    dialer,
		templates: make(map[string]*template.Template),
	}

	// 加载邮件模板
	if err := client.loadTemplates(); err != nil {
		return nil, fmt.Errorf("加载邮件模板失败: %w", err)
	}

	// 测试连接
	if err := client.TestConnection(); err != nil {
		return nil, fmt.Errorf("邮件服务连接测试失败: %w", err)
	}

	if logger != nil {
		logger.Info("邮件服务初始化成功: %s:%d", emailConfig.Host, emailConfig.Port)
	}

	return client, nil
}

// GetClient 获取单例邮件客户端
func GetClient() *EmailClient {
	mu.RLock()
	defer mu.RUnlock()
	return Client
}

// InitializeClient 初始化单例邮件客户端
func InitializeClient(emailConfig *configs.EmailConfig) error {
	var err error
	once.Do(func() {
		Client, err = NewClient(emailConfig)
	})
	return err
}

// loadTemplates 加载邮件模板
func (c *EmailClient) loadTemplates() error {
	// 检查模板目录是否存在
	if _, err := os.Stat(c.config.TemplateDir); os.IsNotExist(err) {
		if logger != nil {
			logger.Info("模板目录不存在，将使用内置模板: %s", c.config.TemplateDir)
		}
		return nil // 不是错误，将使用内置模板
	}

	// 扫描模板目录
	pattern := filepath.Join(c.config.TemplateDir, "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("扫描模板文件失败: %w", err)
	}

	// 加载每个模板文件
	for _, file := range files {
		// 获取文件名（不包含扩展名）作为模板名
		baseName := strings.TrimSuffix(filepath.Base(file), ".html")

		tmpl, err := template.ParseFiles(file)
		if err != nil {
			if logger != nil {
				logger.Error("解析模板文件失败: %s, error=%v", file, err)
			}
			continue // 跳过错误的模板文件
		}

		c.templates[baseName] = tmpl
		if logger != nil {
			logger.Info("加载邮件模板成功: %s", baseName)
		}
	}

	return nil
}

// TemplateData 模板数据结构
type TemplateData struct {
	Code    string
	Purpose string
	To      string
}

// renderTemplate 渲染邮件模板
func (c *EmailClient) renderTemplate(purpose, code, to string) (string, error) {
	// 查找对应的模板
	tmpl, exists := c.templates[purpose]
	if !exists {
		// 如果没有找到对应的模板，使用默认模板
		tmpl, exists = c.templates["default"]
		if !exists {
			// 如果连默认模板都没有，使用内置模板
			return c.renderBuiltinTemplate(purpose, code), nil
		}
	}

	// 准备模板数据
	data := TemplateData{
		Code:    code,
		Purpose: purpose,
		To:      to,
	}

	// 渲染模板
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		if logger != nil {
			logger.Error("渲染邮件模板失败: purpose=%s, error=%v", purpose, err)
		}
		// 如果渲染失败，使用内置模板
		return c.renderBuiltinTemplate(purpose, code), nil
	}

	return buf.String(), nil
}

// renderBuiltinTemplate 渲染内置模板（后备方案）
func (c *EmailClient) renderBuiltinTemplate(purpose, code string) string {
	var purposeText string
	var emoji string
	var bgColor string
	var borderColor string
	var codeColor string

	switch purpose {
	case "register":
		purposeText = "注册"
		emoji = "🎉"
		bgColor = "#e8f5e8"
		borderColor = "#28a745"
		codeColor = "#28a745"
	case "login":
		purposeText = "登录"
		emoji = "🔐"
		bgColor = "#e3f2fd"
		borderColor = "#007bff"
		codeColor = "#007bff"
	case "reset_password":
		purposeText = "重置密码"
		emoji = "🔑"
		bgColor = "#fff3cd"
		borderColor = "#ffc107"
		codeColor = "#856404"
	default:
		purposeText = "验证"
		emoji = "📧"
		bgColor = "#f8f9fa"
		borderColor = "#6c757d"
		codeColor = "#495057"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s验证码</title>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; color: #333; margin-bottom: 30px; }
        .code-box { background-color: %s; border: 2px dashed %s; border-radius: 6px; padding: 20px; text-align: center; margin: 20px 0; }
        .code { font-size: 32px; font-weight: bold; color: %s; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .info { color: #666; line-height: 1.6; margin: 20px 0; }
        .warning { color: #dc3545; font-size: 14px; margin-top: 20px; }
        .footer { text-align: center; color: #999; font-size: 12px; margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s %s验证码</h1>
        </div>
        <div class="info">
            <p>您好！</p>
            <p>您正在进行<strong>%s</strong>操作，验证码如下：</p>
        </div>
        <div class="code-box">
            <div class="code">%s</div>
        </div>
        <div class="info">
            <p>验证码有效期为 <strong>15分钟</strong>，请及时使用。</p>
            <p>如果这不是您本人的操作，请忽略此邮件。</p>
        </div>
        <div class="warning">
            <p>• 请勿将验证码告诉他人</p>
            <p>• 验证码仅限本次使用</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
        </div>
    </div>
</body>
</html>`, purposeText, bgColor, borderColor, codeColor, emoji, purposeText, purposeText, code)
}

// TestConnection 测试邮件服务连接
func (c *EmailClient) TestConnection() error {
	if !c.config.Enable {
		logger.Info("邮件服务未启用，跳过连接测试")
		return nil
	}
	// 创建一个测试消息但不发送
	m := gomail.NewMessage()
	m.SetHeader("From", c.config.From)
	m.SetHeader("To", c.config.From) // 发送给自己作为测试
	m.SetHeader("Subject", "连接测试")
	m.SetBody("text/plain", "这是一个连接测试消息")

	// 尝试连接但不发送
	sender, err := c.dialer.Dial()
	if err != nil {
		return err
	}
	defer sender.Close()

	return nil
}

// SendMessage 发送邮件消息
func (c *EmailClient) SendMessage(to, subject, body string) error {
	m := gomail.NewMessage()

	// 设置发件人
	if c.config.FromName != "" {
		m.SetHeader("From", m.FormatAddress(c.config.From, c.config.FromName))
	} else {
		m.SetHeader("From", c.config.From)
	}

	// 设置收件人
	m.SetHeader("To", to)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置正文 (支持HTML)
	if len(body) > 0 && body[0] == '<' {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// 发送邮件
	if err := c.dialer.DialAndSend(m); err != nil {
		if logger != nil {
			logger.Error("发送邮件失败: to=%s, error=%v", to, err)
		}
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	if logger != nil {
		logger.Info("邮件发送成功: to=%s, subject=%s", to, subject)
	}

	return nil
}

// SendVerificationCode 发送验证码邮件
func (c *EmailClient) SendVerificationCode(to, code, purpose string) error {
	subject := "验证码通知"

	// 根据用途生成不同的主题
	switch purpose {
	case "register":
		subject = "注册验证码"
	case "login":
		subject = "登录验证码"
	case "reset_password":
		subject = "重置密码验证码"
	}

	// 使用模板渲染邮件内容
	body, err := c.renderTemplate(purpose, code, to)
	if err != nil {
		if logger != nil {
			logger.Error("渲染邮件模板失败，使用内置模板: purpose=%s, error=%v", purpose, err)
		}
		// 如果模板渲染失败，使用内置模板
		body = c.renderBuiltinTemplate(purpose, code)
	}

	return c.SendMessage(to, subject, body)
}

// 全局便捷函数

// SendVerificationCode 发送验证码邮件 (全局函数)
func SendVerificationCode(to, code, purpose string) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("邮件客户端未初始化")
	}
	return client.SendVerificationCode(to, code, purpose)
}

// SendMessage 发送邮件消息 (全局函数)
func SendMessage(to, subject, body string) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("邮件客户端未初始化")
	}
	return client.SendMessage(to, subject, body)
}
