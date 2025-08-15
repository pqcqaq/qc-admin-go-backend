package configs

import "github.com/spf13/viper"

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

func setDatabaseConfigDefaults() {
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "file:ent.db?cache=shared&_fk=1")
}
