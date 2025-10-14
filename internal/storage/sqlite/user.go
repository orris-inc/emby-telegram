// Package sqlite 用户存储实现
package sqlite

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"emby-telegram/internal/user"
)

// UserStore 用户存储实现
type UserStore struct {
	db *gorm.DB
}

// NewUserStore 创建用户存储实例
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

// Create 创建用户
func (s *UserStore) Create(ctx context.Context, u *user.User) error {
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// Get 根据 ID 获取用户
func (s *UserStore) Get(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

// GetByTelegramID 根据 Telegram ID 获取用户
func (s *UserStore) GetByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	var u user.User
	if err := s.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("get user by telegram_id: %w", err)
	}
	return &u, nil
}

// List 列出所有用户(分页)
func (s *UserStore) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	var users []*user.User
	query := s.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// Update 更新用户
func (s *UserStore) Update(ctx context.Context, u *user.User) error {
	if err := s.db.WithContext(ctx).Save(u).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// Delete 删除用户(软删除)
func (s *UserStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&user.User{}, id).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// Count 统计用户数量
func (s *UserStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&user.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// CountByRole 统计指定角色的用户数量
func (s *UserStore) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&user.User{}).
		Where("role = ?", role).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count users by role: %w", err)
	}
	return count, nil
}
