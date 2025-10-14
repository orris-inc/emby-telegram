// Package emby 提供 Emby 服务器 API 客户端
package emby

import (
	"errors"
	"fmt"
)

// Emby 错误定义
var (
	// ErrServerUnavailable Emby 服务器不可达
	ErrServerUnavailable = errors.New("emby server unavailable")

	// ErrUnauthorized API Key 无效或未授权
	ErrUnauthorized = errors.New("unauthorized: invalid api key")

	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidResponse 无效的响应
	ErrInvalidResponse = errors.New("invalid response from emby server")

	// ErrSyncDisabled 同步功能未启用
	ErrSyncDisabled = errors.New("emby sync is disabled")
)

// ServerError 创建服务器错误
func ServerError(statusCode int, message string) error {
	return fmt.Errorf("emby server error (status %d): %s: %w", statusCode, message, ErrServerUnavailable)
}

// NotFoundError 创建用户不存在错误
func NotFoundError(userID string) error {
	return fmt.Errorf("user %q: %w", userID, ErrUserNotFound)
}

// AlreadyExistsError 创建用户已存在错误
func AlreadyExistsError(username string) error {
	return fmt.Errorf("user %q: %w", username, ErrUserAlreadyExists)
}
