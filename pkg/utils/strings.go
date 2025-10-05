package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// IsEmpty 检查字符串是否为空或只包含空白字符
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// DefaultIfEmpty 如果字符串为空则返回默认值
func DefaultIfEmpty(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Truncate 截断字符串到指定长度
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Contains 检查字符串是否包含任意一个子字符串
func Contains(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsAll 检查字符串是否包含所有子字符串
func ContainsAll(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// CamelToSnake 驼峰转下划线
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SnakeToCamel 下划线转驼峰
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// MD5 计算字符串的MD5哈希值
func MD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// SHA256 计算字符串的SHA256哈希值
func SHA256(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhone 验证手机号格式（中国手机号）
func IsValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// StringToByte 将字符串转换为字节切片（零拷贝，高性能）
// 警告：返回的 []byte 与原字符串共享底层数据，不可修改
// 如果需要修改字节切片，请使用标准的 []byte(s) 转换
func StringToByte(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// ByteToString 将字节切片转换为字符串（零拷贝，高性能）
// 警告：如果修改了原 []byte，字符串内容也会改变
func ByteToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// StringShorten 使用哈希算法生成固定长度的字符串指纹
// 参数:
//
//	s: 原始字符串
//	length: 目标长度
//
// 返回:
//
//	固定长度的字符串指纹（使用 Base64 编码，包含大小写字母、数字和+/）
func StringShorten(s string, length int) string {
	if len(s) == 0 || length <= 0 {
		return ""
	}

	// 使用 SHA-256 计算哈希值
	hash := sha256.Sum256([]byte(s))

	// 转换为 Base64 编码（更高的信息密度）
	encoded := base64.StdEncoding.EncodeToString(hash[:])

	// 如果需要的长度小于等于哈希长度，直接截取
	if length <= len(encoded) {
		return encoded[:length]
	}

	// 如果需要更长的输出，使用多轮哈希
	result := encoded
	counter := 1

	for len(result) < length {
		// 用原始字符串 + 计数器生成新的哈希
		nextHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", s, counter)))
		nextEncoded := base64.StdEncoding.EncodeToString(nextHash[:])
		result += nextEncoded
		counter++
	}

	return result[:length]
}

// IsValidUTF8 检查字节数组是否是有效的UTF-8编码
func IsValidUTF8(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return utf8.Valid(data)
}
