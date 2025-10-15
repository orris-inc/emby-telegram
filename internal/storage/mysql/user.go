package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"emby-telegram/internal/user"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, u *user.User) error {
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

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

func (s *UserStore) Update(ctx context.Context, u *user.User) error {
	if err := s.db.WithContext(ctx).Save(u).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *UserStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&user.User{}, id).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *UserStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&user.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

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
