package utils

import (
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// IntMax 返回两个int中的最大值
func IntMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IntMin 返回两个int中的最小值
func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 返回两个int64中的最大值
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 返回两个int64中的最小值
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// AbsInt 返回int的绝对值
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AbsInt64 返回int64的绝对值
func AbsInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// AbsFloat64 返回float64的绝对值
func AbsFloat64(x float64) float64 {
	return math.Abs(x)
}

// PowInt 整数幂运算
func PowInt(base, exp int) int {
	result := 1
	for exp > 0 {
		if exp%2 == 1 {
			result *= base
		}
		base *= base
		exp /= 2
	}
	return result
}

// GCD 最大公约数
func GCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM 最小公倍数
func LCM(a, b int) int {
	return AbsInt(a*b) / GCD(a, b)
}

// IsPrime 判断是否为质数
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	sqrt := int(math.Sqrt(float64(n)))
	for i := 3; i <= sqrt; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// IsEven 判断是否为偶数
func IsEven(n int) bool {
	return n%2 == 0
}

// IsOdd 判断是否为奇数
func IsOdd(n int) bool {
	return n%2 != 0
}

// RandomInt 生成[min, max]范围内的随机整数
func RandomInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	return rand.Intn(max-min+1) + min
}

// RandomFloat64 生成[min, max)范围内的随机浮点数
func RandomFloat64(min, max float64) float64 {
	if min > max {
		min, max = max, min
	}
	return rand.Float64()*(max-min) + min
}

// Round 四舍五入到指定小数位
func Round(num float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(num*ratio) / ratio
}

// RoundUp 向上取整到指定小数位
func RoundUp(num float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Ceil(num*ratio) / ratio
}

// RoundDown 向下取整到指定小数位
func RoundDown(num float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Floor(num*ratio) / ratio
}

// InRange 检查数字是否在指定范围内
func InRange(num, min, max int) bool {
	return num >= min && num <= max
}

// InRangeFloat64 检查浮点数是否在指定范围内
func InRangeFloat64(num, min, max float64) bool {
	return num >= min && num <= max
}

// Clamp 将数字限制在指定范围内
func Clamp(num, min, max int) int {
	if num < min {
		return min
	}
	if num > max {
		return max
	}
	return num
}

// ClampFloat64 将浮点数限制在指定范围内
func ClampFloat64(num, min, max float64) float64 {
	if num < min {
		return min
	}
	if num > max {
		return max
	}
	return num
}

// Factorial 计算阶乘
func Factorial(n int) int64 {
	if n < 0 {
		return 0
	}
	if n == 0 || n == 1 {
		return 1
	}

	result := int64(1)
	for i := 2; i <= n; i++ {
		result *= int64(i)
	}
	return result
}

// Fibonacci 计算第n个斐波那契数
func Fibonacci(n int) int64 {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}

	a, b := int64(0), int64(1)
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}
