//go:build !sqlite && !mysql && !postgres && !clickhouse && !sqlserver && !oracle

package drivers

// 当没有指定任何数据库驱动时的默认实现
// 这会在编译时自动包含 sqlite 作为默认驱动

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	RegisterDriver("sqlite3", "SQLite3 (default)")
}
