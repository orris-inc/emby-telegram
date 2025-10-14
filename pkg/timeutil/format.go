// Package timeutil 提供时间处理工具
package timeutil

import (
	"fmt"
	"time"
)

const (
	// DateTimeFormat 标准日期时间格式
	DateTimeFormat = "2006-01-02 15:04:05"
	// DateFormat 标准日期格式
	DateFormat = "2006-01-02"
	// TimeFormat 标准时间格式
	TimeFormat = "15:04:05"
)

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDuration 格式化时间段为易读格式
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}

// DaysUntil 计算距离指定时间还有多少天
func DaysUntil(t time.Time) int {
	duration := time.Until(t)
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// IsExpired 检查时间是否已过期
func IsExpired(t *time.Time) bool {
	if t == nil {
		return false
	}
	return time.Now().After(*t)
}

// AddDays 添加天数
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// FormatExpireTime 格式化过期时间显示
// 如果已过期，返回 "已过期"
// 如果未过期，返回 "yyyy-MM-dd (剩余X天)"
func FormatExpireTime(t *time.Time) string {
	if t == nil {
		return "永久有效"
	}

	if IsExpired(t) {
		return "已过期"
	}

	days := DaysUntil(*t)
	return fmt.Sprintf("%s (剩余%d天)", FormatDate(*t), days)
}

// ParseDateTime 解析日期时间字符串
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(DateTimeFormat, s)
}

// ParseDate 解析日期字符串
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// Today 获取今天0点时间
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
