// Package account 存储接口定义
package account

import "context"

// Store 账号存储接口
// 按照 Google Go 最佳实践，接口定义在消费端(业务层)
type Store interface {
	// Create 创建账号
	Create(ctx context.Context, acc *Account) error

	// Get 根据 ID 获取账号
	Get(ctx context.Context, id uint) (*Account, error)

	// GetByUsername 根据用户名获取账号
	GetByUsername(ctx context.Context, username string) (*Account, error)

	// List 列出指定用户的所有账号
	List(ctx context.Context, userID uint) ([]*Account, error)

	// ListAll 列出所有账号(分页)
	ListAll(ctx context.Context, offset, limit int) ([]*Account, error)

	// Update 更新账号
	Update(ctx context.Context, acc *Account) error

	// Delete 删除账号
	Delete(ctx context.Context, id uint) error

	// Count 统计账号数量
	Count(ctx context.Context) (int64, error)

	// CountByUser 统计指定用户的账号数量
	CountByUser(ctx context.Context, userID uint) (int64, error)

	// CountByStatus 统计指定状态的账号数量
	CountByStatus(ctx context.Context, status Status) (int64, error)
}
