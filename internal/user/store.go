// Package user 存储接口定义
package user

import "context"

// Store 用户存储接口
type Store interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error

	// Get 根据 ID 获取用户
	Get(ctx context.Context, id uint) (*User, error)

	// GetByTelegramID 根据 Telegram ID 获取用户
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)

	// List 列出所有用户(分页)
	List(ctx context.Context, offset, limit int) ([]*User, error)

	// Update 更新用户
	Update(ctx context.Context, user *User) error

	// Delete 删除用户
	Delete(ctx context.Context, id uint) error

	// Count 统计用户数量
	Count(ctx context.Context) (int64, error)

	// CountByRole 统计指定角色的用户数量
	CountByRole(ctx context.Context, role Role) (int64, error)
}
