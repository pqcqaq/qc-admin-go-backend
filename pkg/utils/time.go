package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// TimeFormat 常用时间格式
const (
	LayoutDateTime = "2006-01-02 15:04:05"
	LayoutDate     = "2006-01-02"
	LayoutTime     = "15:04:05"
)

// JSONTime 自定义时间类型，便于 JSON 序列化/反序列化
type JSONTime time.Time

// MarshalJSON 实现 json 序列化
func (t JSONTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte(`""`), nil
	}
	formatted := fmt.Sprintf(`"%s"`, time.Time(t).Format(LayoutDateTime))
	return []byte(formatted), nil
}

// UnmarshalJSON 实现 json 反序列化
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	// 去掉引号
	str := string(data)
	if str == `""` || str == `null` {
		*t = JSONTime(time.Time{})
		return nil
	}

	// 先尝试标准格式
	parsed, err := time.Parse(`"`+LayoutDateTime+`"`, str)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	// 再尝试仅日期
	parsed, err = time.Parse(`"`+LayoutDate+`"`, str)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	// 最后尝试 RFC3339
	parsed, err = time.Parse(`"`+time.RFC3339+`"`, str)
	if err == nil {
		*t = JSONTime(parsed)
		return nil
	}

	return fmt.Errorf("invalid time format: %s", str)
}

// String 返回时间的字符串形式
func (t JSONTime) String() string {
	if time.Time(t).IsZero() {
		return ""
	}
	return time.Time(t).Format(LayoutDateTime)
}

// Now 返回当前时间（time.Time）
func Now() time.Time {
	return time.Now()
}

// NowString 返回当前时间字符串（YYYY-MM-DD HH:mm:ss）
func NowString() string {
	return time.Now().Format(LayoutDateTime)
}

// TimeToDateTimeString 返回时间的字符串形式（YYYY-MM-DD HH:mm:ss）
func TimeToDateTimeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutDateTime)
}

// TimeToDateString 返回时间的字符串形式（YYYY-MM-DD）
func TimeToDateString(t *time.Time) string {
	if t == nil {
		return ""
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutDate)
}

// TimeToTimeString 返回时间的字符串形式（HH:mm:ss）
func TimeToTimeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(LayoutTime)
}

// ParseTime 根据 LayoutDateTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil // 返回零值时间
	}
	return time.Parse(LayoutDateTime, s)
}

// ParseTimeWithLayout 按自定义格式解析时间字符串
func ParseTimeWithLayout(s, layout string) (time.Time, error) {
	return time.Parse(layout, s)
}

// ToTimestamp 转为 Unix 时间戳（秒）
func ToTimestamp(t time.Time) int64 {
	return t.Unix()
}

// FromTimestamp 从 Unix 时间戳（秒）转为 time.Time
func FromTimestamp(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// MustJSONTime 将 time.Time 转为 JSONTime
func MustJSONTime(t time.Time) JSONTime {
	return JSONTime(t)
}

// ExampleUsage 演示序列化和反序列化
func ExampleUsage() {
	type Demo struct {
		Name string   `json:"name"`
		Time JSONTime `json:"time"`
	}

	d := Demo{
		Name: "Test",
		Time: MustJSONTime(Now()),
	}

	// 序列化
	b, _ := json.Marshal(d)
	fmt.Println("Serialized:", string(b))

	// 反序列化
	var d2 Demo
	_ = json.Unmarshal(b, &d2)
	fmt.Println("Deserialized:", d2.Name, d2.Time.String())
}
