package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"emby-telegram/internal/account"
)

type AccountStore struct {
	db *gorm.DB
}

func NewAccountStore(db *gorm.DB) *AccountStore {
	return &AccountStore{db: db}
}

func (s *AccountStore) Create(ctx context.Context, acc *account.Account) error {
	if err := s.db.WithContext(ctx).Create(acc).Error; err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}

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

func (s *AccountStore) Update(ctx context.Context, acc *account.Account) error {
	if err := s.db.WithContext(ctx).Save(acc).Error; err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	return nil
}

func (s *AccountStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Unscoped().Delete(&account.Account{}, id).Error; err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

func (s *AccountStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&account.Account{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count accounts: %w", err)
	}
	return count, nil
}

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
