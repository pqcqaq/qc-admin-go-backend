//go:build clickhouse || all

package drivers

import (
	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func init() {
	RegisterDriver("clickhouse", "ClickHouse")
}
