package utils

import (
	"github.com/google/uuid"
)

// UUIDString 生成指定长度的 UUID 字符串（不足标准长度时会截断）
func UUIDString() string {
	u, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return u.String()
}
