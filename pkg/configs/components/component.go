package components

type ComponentConfig struct {
	Messaging MessagingConfig `mapstructure:"messaging"`
}

func SetComponentsConfigDefaults() {
	setMessagingConfigDefaults()
}
