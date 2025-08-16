//go:build sqlserver || all

package drivers

import (
	_ "github.com/denisenkom/go-mssqldb"
)

func init() {
	RegisterDriver("sqlserver", "SQL Server")
}
