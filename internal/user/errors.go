// Package user 领域错误定义
package user

import (
	"errors"
	"fmt"
)

// 领域错误定义
var (
	// ErrNotFound 用户不存在
	ErrNotFound = errors.New("user not found")

	// ErrAlreadyExists 用户已存在
	ErrAlreadyExists = errors.New("user already exists")

	// ErrBlocked 用户已被封禁
	ErrBlocked = errors.New("user is blocked")

	// ErrUnauthorized 未授权操作
	ErrUnauthorized = errors.New("unauthorized operation")

	// ErrInvalidRole 无效角色
	ErrInvalidRole = errors.New("invalid role")
)

// NotFoundError 创建用户不存在错误
func NotFoundError(telegramID int64) error {
	return fmt.Errorf("user with telegram_id %d: %w", telegramID, ErrNotFound)
}

// AlreadyExistsError 创建用户已存在错误
func AlreadyExistsError(telegramID int64) error {
	return fmt.Errorf("user with telegram_id %d: %w", telegramID, ErrAlreadyExists)
}

// BlockedError 创建用户已封禁错误
func BlockedError(telegramID int64) error {
	return fmt.Errorf("user with telegram_id %d: %w", telegramID, ErrBlocked)
}

// UnauthorizedError 创建未授权错误
func UnauthorizedError(action string) error {
	return fmt.Errorf("%s: %w", action, ErrUnauthorized)
}

// InvalidRoleError 创建无效角色错误
func InvalidRoleError(role string) error {
	return fmt.Errorf("role %q: %w", role, ErrInvalidRole)
}
