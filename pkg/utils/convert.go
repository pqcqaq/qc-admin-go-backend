package utils

import (
	"strconv"
	"strings"
)

// StringToInt 字符串转int，失败时返回默认值
func StringToInt(s string, defaultValue int) int {
	if result, err := strconv.Atoi(s); err == nil {
		return result
	}
	return defaultValue
}

// StringToInt64 字符串转int64，失败时返回默认值
func StringToInt64(s string, defaultValue int64) int64 {
	if result, err := strconv.ParseInt(s, 10, 64); err == nil {
		return result
	}
	return defaultValue
}

// StringToFloat64 字符串转float64，失败时返回默认值
func StringToFloat64(s string, defaultValue float64) float64 {
	if result, err := strconv.ParseFloat(s, 64); err == nil {
		return result
	}
	return defaultValue
}

// StringToBool 字符串转bool，失败时返回默认值
func StringToBool(s string, defaultValue bool) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "1", "yes", "on", "y", "t":
		return true
	case "false", "0", "no", "off", "n", "f":
		return false
	default:
		return defaultValue
	}
}

// IntToString int转字符串
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString int64转字符串
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Float64ToString float64转字符串
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Float64ToStringWithPrecision float64转字符串（指定精度）
func Float64ToStringWithPrecision(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

// BoolToString bool转字符串
func BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

// BoolToInt bool转int（true=1, false=0）
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// IntToBool int转bool（0=false, 非0=true）
func IntToBool(i int) bool {
	return i != 0
}

// StringSliceToIntSlice 字符串切片转int切片
func StringSliceToIntSlice(strs []string) []int {
	result := make([]int, 0, len(strs))
	for _, s := range strs {
		if i, err := strconv.Atoi(s); err == nil {
			result = append(result, i)
		}
	}
	return result
}

// IntSliceToStringSlice int切片转字符串切片
func IntSliceToStringSlice(ints []int) []string {
	result := make([]string, len(ints))
	for i, v := range ints {
		result[i] = strconv.Itoa(v)
	}
	return result
}

// JoinInts 将int切片用分隔符连接成字符串
func JoinInts(ints []int, sep string) string {
	return strings.Join(IntSliceToStringSlice(ints), sep)
}

// SplitToInts 将字符串按分隔符分割并转换为int切片
func SplitToInts(s, sep string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, sep)
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			if i, err := strconv.Atoi(part); err == nil {
				result = append(result, i)
			}
		}
	}
	return result
}

// ToByte 各种类型转byte
func ToByte(v interface{}) byte {
	switch val := v.(type) {
	case byte:
		return val
	case int:
		return byte(val)
	case int8:
		return byte(val)
	case int16:
		return byte(val)
	case int32:
		return byte(val)
	case int64:
		return byte(val)
	case uint:
		return byte(val)
	case uint16:
		return byte(val)
	case uint32:
		return byte(val)
	case uint64:
		return byte(val)
	case float32:
		return byte(val)
	case float64:
		return byte(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return byte(i)
		}
	}
	return 0
}

// ToInt 各种类型转int
func ToInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int8:
		return int(val)
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case uint:
		return int(val)
	case uint8:
		return int(val)
	case uint16:
		return int(val)
	case uint32:
		return int(val)
	case uint64:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	case bool:
		if val {
			return 1
		}
		return 0
	}
	return 0
}

// ToString 各种类型转string
func ToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case []byte:
		return string(val)
	}
	return ""
}
