// Package sqlite 账号存储实现
package sqlite

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"emby-telegram/internal/account"
)

// AccountStore 账号存储实现
type AccountStore struct {
	db *gorm.DB
}

// NewAccountStore 创建账号存储实例
func NewAccountStore(db *gorm.DB) *AccountStore {
	return &AccountStore{db: db}
}

// Create 创建账号
func (s *AccountStore) Create(ctx context.Context, acc *account.Account) error {
	if err := s.db.WithContext(ctx).Create(acc).Error; err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}

// Get 根据 ID 获取账号
func (s *AccountStore) Get(ctx context.Context, id uint) (*account.Account, error) {
	var acc account.Account
	if err := s.db.WithContext(ctx).First(&acc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrNotFound
		}
		return nil, fmt.Errorf("get account: %w", err)
	}
	return &acc, nil
}

// GetByUsername 根据用户名获取账号
func (s *AccountStore) GetByUsername(ctx context.Context, username string) (*account.Account, error) {
	var acc account.Account
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrNotFound
		}
		return nil, fmt.Errorf("get account by username: %w", err)
	}
	return &acc, nil
}

// List 列出指定用户的所有账号
func (s *AccountStore) List(ctx context.Context, userID uint) ([]*account.Account, error) {
	var accounts []*account.Account
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return accounts, nil
}

// ListAll 列出所有账号(分页)
func (s *AccountStore) ListAll(ctx context.Context, offset, limit int) ([]*account.Account, error) {
	var accounts []*account.Account
	query := s.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list all accounts: %w", err)
	}
	return accounts, nil
}

// ListAllWithUser 列出所有账号及关联用户信息(分页)
func (s *AccountStore) ListAllWithUser(ctx context.Context, offset, limit int) ([]*account.AccountWithUser, error) {
	var results []*account.AccountWithUser
	query := s.db.WithContext(ctx).
		Table("accounts").
		Select("accounts.*, users.username as owner_username, users.first_name as owner_first_name, users.telegram_id as owner_telegram_id").
		Joins("LEFT JOIN users ON users.id = accounts.user_id").
		Order("accounts.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("list all accounts with user: %w", err)
	}
	return results, nil
}

// GetWithUser 根据 ID 获取账号及用户信息
func (s *AccountStore) GetWithUser(ctx context.Context, id uint) (*account.AccountWithUser, error) {
	var result account.AccountWithUser
	if err := s.db.WithContext(ctx).
		Table("accounts").
		Select("accounts.*, users.username as owner_username, users.first_name as owner_first_name, users.telegram_id as owner_telegram_id").
		Joins("LEFT JOIN users ON users.id = accounts.user_id").
		Where("accounts.id = ?", id).
		First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrNotFound
		}
		return nil, fmt.Errorf("get account with user: %w", err)
	}
	return &result, nil
}

// Update 更新账号
func (s *AccountStore) Update(ctx context.Context, acc *account.Account) error {
	if err := s.db.WithContext(ctx).Save(acc).Error; err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	return nil
}

// Delete 删除账号(硬删除)
// 使用 Unscoped() 真正从数据库中删除记录，而不是软删除
func (s *AccountStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Unscoped().Delete(&account.Account{}, id).Error; err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// Count 统计账号数量
func (s *AccountStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&account.Account{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count accounts: %w", err)
	}
	return count, nil
}

// CountByUser 统计指定用户的账号数量
func (s *AccountStore) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&account.Account{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count user accounts: %w", err)
	}
	return count, nil
}

// CountByStatus 统计指定状态的账号数量
func (s *AccountStore) CountByStatus(ctx context.Context, status account.Status) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&account.Account{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count accounts by status: %w", err)
	}
	return count, nil
}
