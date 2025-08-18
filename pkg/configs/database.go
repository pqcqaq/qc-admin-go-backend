package configs

import (
	"time"

	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Debug                   bool          `mapstructure:"debug"` // 是否启用调试模式
	Driver                  string        `mapstructure:"driver"`
	DSN                     string        `mapstructure:"dsn"`
	AutoMigrate             bool          `mapstructure:"auto_migrate"`              // 是否自动迁移数据库模式
	SkipMigrateCheck        bool          `mapstructure:"skip_migrate"`              // 是否跳过迁移检查
	MaxIdleConns            int           `mapstructure:"max_idle_conns"`            // 最大空闲连接数
	MaxOpenConns            int           `mapstructure:"max_open_conns"`            // 最大打开连接数
	ConnMaxLifetime         time.Duration `mapstructure:"conn_max_lifetime"`         // 连接最大生命周期
	ConnectionCheckInterval time.Duration `mapstructure:"connection_check_interval"` // 连接检查间隔
}

func setDatabaseConfigDefaults() {
	// 	数据库默认配置
	viper.SetDefault("database.debug", false) // 默认不启用调试模式
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "file:ent.db?cache=shared&_fk=1")
	viper.SetDefault("database.auto_migrate", false) // 默认启用自动迁移
	viper.SetDefault("database.skip_migrate", false) // 默认不跳过迁移检查
	viper.SetDefault("database.max_idle_conns", 10)  // 默认最大空闲连接数
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", time.Hour)              // 默认连接最大生命周期为1小时
	viper.SetDefault("database.connection_check_interval", 30*time.Minute) // 默认连接检查间隔为1分钟
}
