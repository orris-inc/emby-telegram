// Package account 领域错误定义
package account

import (
	"errors"
	"fmt"
)

// 领域错误定义
var (
	// ErrNotFound 账号不存在
	ErrNotFound = errors.New("account not found")

	// ErrAlreadyExists 账号已存在
	ErrAlreadyExists = errors.New("account already exists")

	// ErrInvalidInput 无效输入
	ErrInvalidInput = errors.New("invalid input")

	// ErrExpired 账号已过期
	ErrExpired = errors.New("account expired")

	// ErrSuspended 账号已暂停
	ErrSuspended = errors.New("account suspended")

	// ErrUnauthorized 未授权操作
	ErrUnauthorized = errors.New("unauthorized operation")

	// ErrAccountLimitExceeded 账号数量超过限制
	ErrAccountLimitExceeded = errors.New("account limit exceeded")

	// ErrNotAuthorized 用户未授权创建账号
	ErrNotAuthorized = errors.New("user not authorized to create accounts")
)

// NotFoundError 创建账号不存在错误
func NotFoundError(username string) error {
	return fmt.Errorf("account %q: %w", username, ErrNotFound)
}

// AlreadyExistsError 创建账号已存在错误
func AlreadyExistsError(username string) error {
	return fmt.Errorf("account %q: %w", username, ErrAlreadyExists)
}

// ValidationError 创建验证错误
func ValidationError(field, reason string) error {
	return fmt.Errorf("validation failed for %s: %s: %w", field, reason, ErrInvalidInput)
}

// ExpiredError 创建账号过期错误
func ExpiredError(username string) error {
	return fmt.Errorf("account %q: %w", username, ErrExpired)
}

// SuspendedError 创建账号暂停错误
func SuspendedError(username string) error {
	return fmt.Errorf("account %q: %w", username, ErrSuspended)
}

// UnauthorizedError 创建未授权错误
func UnauthorizedError(action string) error {
	return fmt.Errorf("%s: %w", action, ErrUnauthorized)
}

// AccountLimitExceededError 创建账号数量超限错误
func AccountLimitExceededError(current, limit int) error {
	return fmt.Errorf("account limit reached (%d/%d): %w", current, limit, ErrAccountLimitExceeded)
}

// NotAuthorizedError 创建用户未授权错误
func NotAuthorizedError() error {
	return fmt.Errorf("%w: please contact admin for authorization", ErrNotAuthorized)
}

// QuotaExceededError 创建配额超限错误
func QuotaExceededError(current, quota int) error {
	return fmt.Errorf("account quota exceeded (%d/%d): %w", current, quota, ErrAccountLimitExceeded)
}
