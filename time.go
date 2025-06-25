package nova

import (
	"time"
)

// TimeFormat 时间格式化
const (
	TimeFormat      = "2006-01-02 15:04:05"
	DateFormat      = "2006-01-02"
	TimeFormatShort = "2006-01-02 15:04"
	TimeFormatLong  = "2006-01-02 15:04:05.000"
)

// FormatTime 格式化时间
func FormatTime(t time.Time, format string) string {
	return t.Format(format)
}

// ParseTime 解析时间字符串
func ParseTime(s string, format string) (time.Time, error) {
	return time.Parse(format, s)
}

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// Today 获取今天的开始时间
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Yesterday 获取昨天的开始时间
func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

// Tomorrow 获取明天的开始时间
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// ThisWeek 获取本周的开始时间
func ThisWeek() time.Time {
	now := time.Now()
	offset := int(now.Weekday())
	if offset == 0 {
		offset = 7
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -offset+1)
}

// LastWeek 获取上周的开始时间
func LastWeek() time.Time {
	return ThisWeek().AddDate(0, 0, -7)
}

// NextWeek 获取下周的开始时间
func NextWeek() time.Time {
	return ThisWeek().AddDate(0, 0, 7)
}

// ThisMonth 获取本月的开始时间
func ThisMonth() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// LastMonth 获取上月的开始时间
func LastMonth() time.Time {
	return ThisMonth().AddDate(0, -1, 0)
}

// NextMonth 获取下月的开始时间
func NextMonth() time.Time {
	return ThisMonth().AddDate(0, 1, 0)
}

// ThisYear 获取本年的开始时间
func ThisYear() time.Time {
	now := time.Now()
	return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
}

// LastYear 获取上年的开始时间
func LastYear() time.Time {
	return ThisYear().AddDate(-1, 0, 0)
}

// NextYear 获取下年的开始时间
func NextYear() time.Time {
	return ThisYear().AddDate(1, 0, 0)
}

// IsToday 判断是否为今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsYesterday 判断是否为昨天
func IsYesterday(t time.Time) bool {
	yesterday := Yesterday()
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

// IsTomorrow 判断是否为明天
func IsTomorrow(t time.Time) bool {
	tomorrow := Tomorrow()
	return t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day()
}

// IsThisWeek 判断是否为本周
func IsThisWeek(t time.Time) bool {
	thisWeek := ThisWeek()
	nextWeek := NextWeek()
	return t.After(thisWeek) && t.Before(nextWeek)
}

// IsLastWeek 判断是否为上周
func IsLastWeek(t time.Time) bool {
	lastWeek := LastWeek()
	thisWeek := ThisWeek()
	return t.After(lastWeek) && t.Before(thisWeek)
}

// IsNextWeek 判断是否为下周
func IsNextWeek(t time.Time) bool {
	thisWeek := ThisWeek()
	nextWeek := NextWeek()
	return t.After(thisWeek) && t.Before(nextWeek)
}

// IsThisMonth 判断是否为本月
func IsThisMonth(t time.Time) bool {
	thisMonth := ThisMonth()
	nextMonth := NextMonth()
	return t.After(thisMonth) && t.Before(nextMonth)
}

// IsLastMonth 判断是否为上月
func IsLastMonth(t time.Time) bool {
	lastMonth := LastMonth()
	thisMonth := ThisMonth()
	return t.After(lastMonth) && t.Before(thisMonth)
}

// IsNextMonth 判断是否为下月
func IsNextMonth(t time.Time) bool {
	thisMonth := ThisMonth()
	nextMonth := NextMonth()
	return t.After(thisMonth) && t.Before(nextMonth)
}

// IsThisYear 判断是否为本年
func IsThisYear(t time.Time) bool {
	thisYear := ThisYear()
	nextYear := NextYear()
	return t.After(thisYear) && t.Before(nextYear)
}

// IsLastYear 判断是否为上年
func IsLastYear(t time.Time) bool {
	lastYear := LastYear()
	thisYear := ThisYear()
	return t.After(lastYear) && t.Before(thisYear)
}

// IsNextYear 判断是否为下年
func IsNextYear(t time.Time) bool {
	thisYear := ThisYear()
	nextYear := NextYear()
	return t.After(thisYear) && t.Before(nextYear)
}
