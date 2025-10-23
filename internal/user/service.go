// Package user 用户业务服务
package user

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Service 用户业务服务
type Service struct {
	store Store
}

// NewService 创建用户服务实例
func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// GetOrCreate 获取或创建用户
// 如果用户不存在则自动创建
func (s *Service) GetOrCreate(ctx context.Context, tgUser *tgbotapi.User) (*User, error) {
	// 先尝试获取
	user, err := s.store.GetByTelegramID(ctx, tgUser.ID)
	if err == nil {
		return user, nil
	}

	// 如果不存在，创建新用户
	if errors.Is(err, ErrNotFound) {
		user = &User{
			TelegramID: tgUser.ID,
			Username:   tgUser.UserName,
			FirstName:  tgUser.FirstName,
			LastName:   tgUser.LastName,
			Role:       RoleUser,
			IsBlocked:  false,
		}

		if err := s.store.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}

		return user, nil
	}

	return nil, fmt.Errorf("get user: %w", err)
}

// Get 获取用户
func (s *Service) Get(ctx context.Context, id uint) (*User, error) {
	user, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

// GetByTelegramID 根据 Telegram ID 获取用户
func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	user, err := s.store.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, NotFoundError(telegramID)
	}
	return user, nil
}

// List 列出所有用户(分页)
func (s *Service) List(ctx context.Context, offset, limit int) ([]*User, error) {
	users, err := s.store.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// SetRole 设置用户角色
func (s *Service) SetRole(ctx context.Context, telegramID int64, role string) error {
	// 验证角色
	var userRole Role
	switch role {
	case "admin":
		userRole = RoleAdmin
	case "user":
		userRole = RoleUser
	default:
		return InvalidRoleError(role)
	}

	user, err := s.store.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return NotFoundError(telegramID)
	}

	user.SetRole(userRole)

	if err := s.store.Update(ctx, user); err != nil {
		return fmt.Errorf("update user role: %w", err)
	}

	return nil
}

// Block 封禁用户
func (s *Service) Block(ctx context.Context, telegramID int64) error {
	user, err := s.store.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return NotFoundError(telegramID)
	}

	user.Block()

	if err := s.store.Update(ctx, user); err != nil {
		return fmt.Errorf("block user: %w", err)
	}

	return nil
}

// Unblock 解封用户
func (s *Service) Unblock(ctx context.Context, telegramID int64) error {
	user, err := s.store.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return NotFoundError(telegramID)
	}

	user.Unblock()

	if err := s.store.Update(ctx, user); err != nil {
		return fmt.Errorf("unblock user: %w", err)
	}

	return nil
}

// CheckAccess 检查用户是否可以访问
func (s *Service) CheckAccess(ctx context.Context, telegramID int64) error {
	user, err := s.store.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return NotFoundError(telegramID)
	}

	if !user.CanAccess() {
		return BlockedError(telegramID)
	}

	return nil
}

// IsAdmin 检查用户是否为管理员
func (s *Service) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	user, err := s.store.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return false, NotFoundError(telegramID)
	}

	return user.IsAdmin(), nil
}

// Count 统计用户数量
func (s *Service) Count(ctx context.Context) (int64, error) {
	count, err := s.store.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// CountByRole 统计指定角色的用户数量
func (s *Service) CountByRole(ctx context.Context, role Role) (int64, error) {
	count, err := s.store.CountByRole(ctx, role)
	if err != nil {
		return 0, fmt.Errorf("count users by role: %w", err)
	}
	return count, nil
}

// UpdateProfile 更新用户资料
func (s *Service) UpdateProfile(ctx context.Context, tgUser *tgbotapi.User) error {
	user, err := s.store.GetByTelegramID(ctx, tgUser.ID)
	if err != nil {
		return NotFoundError(tgUser.ID)
	}

	// 更新用户信息
	user.Username = tgUser.UserName
	user.FirstName = tgUser.FirstName
	user.LastName = tgUser.LastName

	if err := s.store.Update(ctx, user); err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	return nil
}

// SetQuota 设置用户账号配额
func (s *Service) SetQuota(ctx context.Context, userID uint, quota int) error {
	user, err := s.store.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if quota < 0 {
		quota = 0
	}

	user.AccountQuota = quota

	if err := s.store.Update(ctx, user); err != nil {
		return fmt.Errorf("update user quota: %w", err)
	}

	return nil
}

// GetByUsername 根据 Telegram username 获取用户
func (s *Service) GetByUsername(ctx context.Context, username string) (*User, error) {
	user, err := s.store.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

// MarkInviteCodeUsed 标记用户已使用邀请码
func (s *Service) MarkInviteCodeUsed(ctx context.Context, userID uint) error {
	user, err := s.store.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	user.UsedInviteCode = true

	if err := s.store.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}
