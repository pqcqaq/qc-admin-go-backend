//go:build oracle || all

package drivers

import (
	_ "github.com/godror/godror"
)

func init() {
	RegisterDriver("godror", "Oracle")
}
