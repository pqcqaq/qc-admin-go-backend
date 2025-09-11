package middleware

type MiddlewareConfig struct {
	Delay DelayConfig `mapstructure:"delay"` // 延迟中间件配置
}

func SetMiddlewareConfigDefaults() {
	setDelayConfigDefaults()
}
