// Package validator 提供参数验证工具
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// 用户名正则: 字母、数字、下划线，3-32位
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	// 邮箱正则
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidateUsername 验证用户名
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	if len(username) < 3 {
		return fmt.Errorf("用户名长度不能少于3个字符")
	}

	if len(username) > 32 {
		return fmt.Errorf("用户名长度不能超过32个字符")
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("用户名只能包含字母、数字和下划线")
	}

	return nil
}

// ValidatePassword 验证密码
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于6个字符")
	}

	if len(password) > 64 {
		return fmt.Errorf("密码长度不能超过64个字符")
	}

	return nil
}

// ValidateEmail 验证邮箱
func ValidateEmail(email string) error {
	if email == "" {
		return nil // 邮箱可选
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	return nil
}

// ValidateDays 验证天数
func ValidateDays(days int) error {
	if days <= 0 {
		return fmt.Errorf("天数必须大于0")
	}

	if days > 3650 { // 最大10年
		return fmt.Errorf("天数不能超过3650天(10年)")
	}

	return nil
}

// ValidateMaxDevices 验证最大设备数
func ValidateMaxDevices(maxDevices int) error {
	if maxDevices <= 0 {
		return fmt.Errorf("最大设备数必须大于0")
	}

	if maxDevices > 100 {
		return fmt.Errorf("最大设备数不能超过100")
	}

	return nil
}

// SanitizeUsername 清理用户名(移除前后空格，转小写)
func SanitizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// SanitizeEmail 清理邮箱(移除前后空格，转小写)
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
