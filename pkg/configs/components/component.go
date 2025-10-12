package components

type ComponentConfig struct {
	Messaging MessagingConfig `mapstructure:"messaging"`
	Monitor   MonitorConfig   `mapstructure:"monitor"`
}

func SetComponentsConfigDefaults() {
	setMessagingConfigDefaults()
	setMinitorConfigDefaults()
}
