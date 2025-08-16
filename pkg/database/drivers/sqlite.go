//go:build sqlite || all

package drivers

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	RegisterDriver("sqlite3", "SQLite3")
}
