package utils

import "strings"

// matchTopic 检查消息topic是否匹配订阅pattern
func matchTopic(subPattern, msgTopic string) bool {
	// 分割topic为层级
	subLevels := strings.Split(subPattern, "/")
	msgLevels := strings.Split(msgTopic, "/")

	subLen := len(subLevels)
	msgLen := len(msgLevels)

	// 遍历订阅pattern的每一层
	for i := 0; i < subLen; i++ {
		// 如果遇到多层通配符 #
		if subLevels[i] == "#" {
			// # 必须是最后一层
			if i == subLen-1 {
				return true
			}
			return false // # 不在最后,pattern无效
		}

		// 如果消息topic层级已用完,但订阅pattern还有(且不是#)
		if i >= msgLen {
			return false
		}

		// 如果是单层通配符 +,跳过这一层
		if subLevels[i] == "+" {
			continue
		}

		// 精确匹配这一层
		if subLevels[i] != msgLevels[i] {
			return false
		}
	}

	// 所有层级都匹配完成,长度必须相等
	return subLen == msgLen
}

// IsAnyMatch 检查msgTopic是否匹配subsList中的任意一个订阅
func IsAnyMatch(subsList []string, msgTopic string) bool {
	for _, sub := range subsList {
		if matchTopic(sub, msgTopic) {
			return true
		}
	}
	return false
}

// IsAllMatch 检查msgTopic是否匹配subsList中的所有订阅
func IsAllMatch(subsList []string, msgTopic string) bool {
	if len(subsList) == 0 {
		return false // 空列表返回false
	}

	for _, sub := range subsList {
		if !matchTopic(sub, msgTopic) {
			return false
		}
	}
	return true
}
