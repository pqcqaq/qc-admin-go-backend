package configs

import "github.com/spf13/viper"

// EmailConfig 邮件配置
type EmailConfig struct {
	Enable      bool   `mapstructure:"enable"`       // 是否启用邮件功能
	Host        string `mapstructure:"host"`         // SMTP服务器地址
	Port        int    `mapstructure:"port"`         // SMTP端口
	Username    string `mapstructure:"username"`     // 用户名
	Password    string `mapstructure:"password"`     // 密码
	From        string `mapstructure:"from"`         // 发件人邮箱
	FromName    string `mapstructure:"from_name"`    // 发件人名称
	UseTLS      bool   `mapstructure:"use_tls"`      // 是否使用TLS
	UseSSL      bool   `mapstructure:"use_ssl"`      // 是否使用SSL
	TemplateDir string `mapstructure:"template_dir"` // 模板目录
}

// setEmailConfigDefaults 设置邮件默认配置
func setEmailConfigDefaults() {
	viper.SetDefault("email.enable", false)
	viper.SetDefault("email.host", "smtp.gmail.com")
	viper.SetDefault("email.port", 587)
	viper.SetDefault("email.username", "")
	viper.SetDefault("email.password", "")
	viper.SetDefault("email.from", "")
	viper.SetDefault("email.from_name", "系统通知")
	viper.SetDefault("email.use_tls", true)
	viper.SetDefault("email.use_ssl", false)
	viper.SetDefault("email.template_dir", "./templates")
}
