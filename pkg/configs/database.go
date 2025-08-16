package configs

import "github.com/spf13/viper"

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver           string `mapstructure:"driver"`
	DSN              string `mapstructure:"dsn"`
	AutoMigrate      bool   `mapstructure:"auto_migrate"` // 是否自动迁移数据库模式
	SkipMigrateCheck bool   `mapstructure:"skip_migrate"` // 是否跳过迁移检查
}

func setDatabaseConfigDefaults() {
	// 	数据库默认配置
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "file:ent.db?cache=shared&_fk=1")
	viper.SetDefault("database.auto_migrate", false) // 默认启用自动迁移
	viper.SetDefault("database.skip_migrate", false) // 默认不跳过迁移检查
}
