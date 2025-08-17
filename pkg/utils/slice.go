package utils

import "reflect"

// ContainsInt 检查int切片是否包含指定元素
func ContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsString 检查string切片是否包含指定元素
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveInt 从int切片中移除指定元素
func RemoveInt(slice []int, item int) []int {
	result := make([]int, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// RemoveString 从string切片中移除指定元素
func RemoveString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// UniqueInt 去重int切片
func UniqueInt(slice []int) []int {
	keys := make(map[int]bool)
	result := make([]int, 0, len(slice))

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// UniqueString 去重string切片
func UniqueString(slice []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// ReverseInt 反转int切片
func ReverseInt(slice []int) []int {
	result := make([]int, len(slice))
	for i, j := 0, len(slice)-1; i <= j; i, j = i+1, j-1 {
		result[i], result[j] = slice[j], slice[i]
	}
	return result
}

// ReverseString 反转string切片
func ReverseString(slice []string) []string {
	result := make([]string, len(slice))
	for i, j := 0, len(slice)-1; i <= j; i, j = i+1, j-1 {
		result[i], result[j] = slice[j], slice[i]
	}
	return result
}

// ChunkInt 将int切片分割成指定大小的块
func ChunkInt(slice []int, size int) [][]int {
	if size <= 0 {
		return nil
	}

	var chunks [][]int
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// ChunkString 将string切片分割成指定大小的块
func ChunkString(slice []string, size int) [][]string {
	if size <= 0 {
		return nil
	}

	var chunks [][]string
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// IsEmpty 检查切片是否为空（使用反射，支持任意类型）
func IsSliceEmpty(slice interface{}) bool {
	if slice == nil {
		return true
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return false
	}

	return v.Len() == 0
}

// MaxInt 返回int切片中的最大值
func MaxInt(slice []int) int {
	if len(slice) == 0 {
		panic("empty slice")
	}

	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// MinInt 返回int切片中的最小值
func MinInt(slice []int) int {
	if len(slice) == 0 {
		panic("empty slice")
	}

	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// SumInt 计算int切片的和
func SumInt(slice []int) int {
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return sum
}
