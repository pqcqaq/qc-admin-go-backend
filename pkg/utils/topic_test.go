package utils

import (
	"testing"
)

func TestMatchTopic(t *testing.T) {
	testCases := []struct {
		name     string
		sub      string
		topic    string
		expected bool
	}{
		{
			name:     "单层通配符匹配",
			sub:      "sport/+/player1",
			topic:    "sport/tennis/player1",
			expected: true,
		},
		{
			name:     "单层通配符不匹配多层",
			sub:      "sport/+/player1",
			topic:    "sport/tennis/player1/score",
			expected: false,
		},
		{
			name:     "多层通配符匹配单层",
			sub:      "sport/#",
			topic:    "sport",
			expected: true,
		},
		{
			name:     "多层通配符匹配多层",
			sub:      "sport/#",
			topic:    "sport/tennis/player1",
			expected: true,
		},
		{
			name:     "多个单层通配符匹配",
			sub:      "+/+/temperature",
			topic:    "home/bedroom/temperature",
			expected: true,
		},
		{
			name:     "多个单层通配符层级不匹配",
			sub:      "+/+/temperature",
			topic:    "home/temperature",
			expected: false,
		},
		{
			name:     "多层通配符匹配精确路径",
			sub:      "home/bedroom/#",
			topic:    "home/bedroom",
			expected: true,
		},
		{
			name:     "多层通配符匹配子路径",
			sub:      "home/bedroom/#",
			topic:    "home/bedroom/temp/sensor",
			expected: true,
		},
		{
			name:     "根级多层通配符",
			sub:      "#",
			topic:    "any/topic/here",
			expected: true,
		},
		{
			name:     "单层通配符匹配单个层级",
			sub:      "+",
			topic:    "single",
			expected: true,
		},
		{
			name:     "单层通配符不匹配多个层级",
			sub:      "+",
			topic:    "two/levels",
			expected: false,
		},
		{
			name:     "精确匹配",
			sub:      "home/bedroom/temperature",
			topic:    "home/bedroom/temperature",
			expected: true,
		},
		{
			name:     "精确不匹配",
			sub:      "home/bedroom/temperature",
			topic:    "home/kitchen/temperature",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matchTopic(tc.sub, tc.topic)
			if result != tc.expected {
				t.Errorf("matchTopic(%q, %q) = %v, want %v", tc.sub, tc.topic, result, tc.expected)
			}
		})
	}
}

func TestIsAnyMatch(t *testing.T) {
	testCases := []struct {
		name     string
		subsList []string
		topic    string
		expected bool
	}{
		{
			name:     "匹配列表中的第一个订阅",
			subsList: []string{"home/+/temperature", "home/bedroom/#", "sensor/+"},
			topic:    "home/bedroom/temperature",
			expected: true,
		},
		{
			name:     "匹配列表中的第二个订阅",
			subsList: []string{"home/+/temperature", "home/bedroom/#", "sensor/+"},
			topic:    "home/bedroom/humidity",
			expected: true,
		},
		{
			name:     "匹配列表中的第三个订阅",
			subsList: []string{"home/+/temperature", "home/bedroom/#", "sensor/+"},
			topic:    "sensor/temp",
			expected: true,
		},
		{
			name:     "不匹配任何订阅",
			subsList: []string{"home/+/temperature", "home/bedroom/#", "sensor/+"},
			topic:    "office/kitchen/humidity",
			expected: false,
		},
		{
			name:     "空订阅列表",
			subsList: []string{},
			topic:    "any/topic",
			expected: false,
		},
		{
			name:     "nil订阅列表",
			subsList: nil,
			topic:    "any/topic",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAnyMatch(tc.subsList, tc.topic)
			if result != tc.expected {
				t.Errorf("IsAnyMatch(%v, %q) = %v, want %v", tc.subsList, tc.topic, result, tc.expected)
			}
		})
	}
}

func TestIsAllMatch(t *testing.T) {
	testCases := []struct {
		name     string
		subsList []string
		topic    string
		expected bool
	}{
		{
			name:     "匹配所有订阅",
			subsList: []string{"home/bedroom/+", "home/+/temperature"},
			topic:    "home/bedroom/temperature",
			expected: true,
		},
		{
			name:     "不匹配所有订阅",
			subsList: []string{"home/bedroom/+", "home/+/temperature"},
			topic:    "sensor/temp",
			expected: false,
		},
		{
			name:     "部分匹配",
			subsList: []string{"home/bedroom/+", "home/+/humidity"},
			topic:    "home/bedroom/temperature",
			expected: false,
		},
		{
			name:     "单个订阅匹配",
			subsList: []string{"home/bedroom/temperature"},
			topic:    "home/bedroom/temperature",
			expected: true,
		},
		{
			name:     "单个订阅不匹配",
			subsList: []string{"home/bedroom/temperature"},
			topic:    "home/bedroom/humidity",
			expected: false,
		},
		{
			name:     "空订阅列表",
			subsList: []string{},
			topic:    "any/topic",
			expected: false,
		},
		{
			name:     "nil订阅列表",
			subsList: nil,
			topic:    "any/topic",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAllMatch(tc.subsList, tc.topic)
			if result != tc.expected {
				t.Errorf("IsAllMatch(%v, %q) = %v, want %v", tc.subsList, tc.topic, result, tc.expected)
			}
		})
	}
}
