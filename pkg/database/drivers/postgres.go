//go:build postgres || all

package drivers

import (
	_ "github.com/lib/pq"
)

func init() {
	RegisterDriver("postgres", "PostgreSQL")
}
