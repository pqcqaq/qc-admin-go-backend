//go:build mysql || all

package drivers

import (
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	RegisterDriver("mysql", "MySQL")
}
