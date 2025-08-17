package utils

import (
	"time"
)

const (
	DateFormat      = "2006-01-02"
	TimeFormat      = "15:04:05"
	DateTimeFormat  = "2006-01-02 15:04:05"
	TimestampFormat = "2006-01-02T15:04:05Z07:00"
)

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// Today 获取今天的日期（零点时刻）
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow 获取明天的日期（零点时刻）
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// Yesterday 获取昨天的日期（零点时刻）
func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(t time.Time) string {
	return t.Format(TimestampFormat)
}

// ParseDate 解析日期字符串
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(DateFormat, dateStr)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(TimeFormat, timeStr)
}

// ParseDateTime 解析日期时间字符串
func ParseDateTime(datetimeStr string) (time.Time, error) {
	return time.Parse(DateTimeFormat, datetimeStr)
}

// ParseTimestamp 解析时间戳字符串
func ParseTimestamp(timestampStr string) (time.Time, error) {
	return time.Parse(TimestampFormat, timestampStr)
}

// UnixToTime Unix时间戳转time.Time
func UnixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// TimeToUnix time.Time转Unix时间戳
func TimeToUnix(t time.Time) int64 {
	return t.Unix()
}

// UnixMilliToTime Unix毫秒时间戳转time.Time
func UnixMilliToTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1000, (timestamp%1000)*int64(time.Millisecond))
}

// TimeToUnixMilli time.Time转Unix毫秒时间戳
func TimeToUnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// IsToday 判断时间是否是今天
func IsToday(t time.Time) bool {
	today := Today()
	return t.Year() == today.Year() &&
		t.Month() == today.Month() &&
		t.Day() == today.Day()
}

// IsYesterday 判断时间是否是昨天
func IsYesterday(t time.Time) bool {
	yesterday := Yesterday()
	return t.Year() == yesterday.Year() &&
		t.Month() == yesterday.Month() &&
		t.Day() == yesterday.Day()
}

// IsTomorrow 判断时间是否是明天
func IsTomorrow(t time.Time) bool {
	tomorrow := Tomorrow()
	return t.Year() == tomorrow.Year() &&
		t.Month() == tomorrow.Month() &&
		t.Day() == tomorrow.Day()
}

// DaysBetween 计算两个日期之间的天数差
func DaysBetween(t1, t2 time.Time) int {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}

	date1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC)
	date2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)

	return int(date2.Sub(date1).Hours() / 24)
}

// StartOfWeek 获取一周的开始时间（周一）
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}
	days := weekday - 1
	return time.Date(t.Year(), t.Month(), t.Day()-days, 0, 0, 0, 0, t.Location())
}

// EndOfWeek 获取一周的结束时间（周日）
func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
}

// StartOfMonth 获取月份的开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取月份的结束时间
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
}

// Age 根据出生日期计算年龄
func Age(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}
