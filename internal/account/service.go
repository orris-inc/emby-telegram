// Package account 账号业务服务
package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"emby-telegram/internal/emby"
	"emby-telegram/internal/logger"
	"emby-telegram/pkg/crypto"
	"emby-telegram/pkg/validator"
)

// UserGetter 用户查询接口
type UserGetter interface {
	Get(ctx context.Context, id uint) (User, error)
}

// User 用户信息
type User struct {
	ID           uint
	IsAdmin      bool
	AccountQuota int
}

// Service 账号业务服务
type Service struct {
	store               Store
	userGetter          UserGetter
	embyClient          *emby.Client
	usernamePrefix      string
	defaultExpire       int
	defaultDevices      int
	passwordLength      int
	maxAccountsPerUser  int
	maxAccountsPerAdmin int
	enableSync          bool
	syncOnCreate        bool
	syncOnDelete        bool
}

// NewService 创建账号服务实例
func NewService(store Store, userGetter UserGetter, embyClient *emby.Client, usernamePrefix string, defaultExpire, defaultDevices, passwordLength, maxAccountsPerUser, maxAccountsPerAdmin int, enableSync, syncOnCreate, syncOnDelete bool) *Service {
	return &Service{
		store:               store,
		userGetter:          userGetter,
		embyClient:          embyClient,
		usernamePrefix:      usernamePrefix,
		defaultExpire:       defaultExpire,
		defaultDevices:      defaultDevices,
		passwordLength:      passwordLength,
		maxAccountsPerUser:  maxAccountsPerUser,
		maxAccountsPerAdmin: maxAccountsPerAdmin,
		enableSync:          enableSync,
		syncOnCreate:        syncOnCreate,
		syncOnDelete:        syncOnDelete,
	}
}

// checkQuota 检查用户配额
func (s *Service) checkQuota(ctx context.Context, userID uint) error {
	user, err := s.userGetter.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user info: %w", err)
	}

	// 检查是否授权
	if user.AccountQuota == 0 {
		return NotAuthorizedError()
	}

	// 查询当前账号数
	currentCount, err := s.store.CountByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("count user accounts: %w", err)
	}

	// 检查是否超配额
	if int(currentCount) >= user.AccountQuota {
		return QuotaExceededError(int(currentCount), user.AccountQuota)
	}

	return nil
}

// checkAccountLimit 检查账号数量限制
func (s *Service) checkAccountLimit(ctx context.Context, userID uint) error {
	user, err := s.userGetter.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user info: %w", err)
	}

	currentCount, err := s.store.CountByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("count user accounts: %w", err)
	}

	limit := s.maxAccountsPerUser
	if user.IsAdmin {
		limit = s.maxAccountsPerAdmin
		if limit == -1 {
			return nil
		}
	}

	if int(currentCount) >= limit {
		return AccountLimitExceededError(int(currentCount), limit)
	}

	return nil
}

// syncToEmby 同步账号到 Emby
func (s *Service) syncToEmby(ctx context.Context, acc *Account, plainPassword string) error {
	if !s.enableSync || s.embyClient == nil {
		return nil
	}

	// 创建 Emby 用户
	embyUser, err := s.embyClient.CreateUser(ctx, acc.Username, plainPassword)
	if err != nil {
		if errors.Is(err, emby.ErrUserAlreadyExists) {
			// 用户已存在,尝试获取用户信息
			existingUser, getErr := s.embyClient.GetUserByName(ctx, acc.Username)
			if getErr != nil {
				return fmt.Errorf("user exists but failed to get: %w", getErr)
			}
			acc.MarkSynced(existingUser.ID)
			return nil
		}
		acc.MarkSyncFailed(err)
		logger.Errorf("sync to emby failed for %s: %v", acc.Username, err)
		return err
	}

	// 等待 Emby 完成用户创建
	time.Sleep(100 * time.Millisecond)

	// 应用默认用户策略（包括 MaxParentalRating 等）
	defaultPolicy := emby.CreateDefaultPolicy(acc.MaxDevices)
	if err := s.embyClient.UpdateUserPolicy(ctx, embyUser.ID, defaultPolicy); err != nil {
		logger.Warnf("failed to set default policy for %s: %v", acc.Username, err)
		// 不返回错误，策略可以后续手动设置
	} else {
		logger.Infof("default policy set for %s: max_devices=%d", acc.Username, acc.MaxDevices)
	}

	acc.MarkSynced(embyUser.ID)
	return nil
}

// deleteFromEmby 从 Emby 删除账号
func (s *Service) deleteFromEmby(ctx context.Context, acc *Account) error {
	if !s.enableSync || s.embyClient == nil || acc.EmbyUserID == "" {
		return nil
	}

	if err := s.embyClient.DeleteUser(ctx, acc.EmbyUserID); err != nil {
		if !errors.Is(err, emby.ErrUserNotFound) {
			logger.Errorf("failed to delete emby user %s: %v", acc.Username, err)
			return err
		}
		// 用户不存在,视为成功
	}

	return nil
}

// updatePasswordInEmby 在 Emby 更新密码
func (s *Service) updatePasswordInEmby(ctx context.Context, acc *Account, newPassword string) error {
	if !s.enableSync || s.embyClient == nil || acc.EmbyUserID == "" {
		return nil
	}

	if err := s.embyClient.UpdatePassword(ctx, acc.EmbyUserID, newPassword); err != nil {
		acc.MarkSyncFailed(fmt.Errorf("update password failed: %w", err))
		logger.Errorf("failed to update emby password for %s: %v", acc.Username, err)
		return err
	}

	acc.MarkSynced(acc.EmbyUserID)
	return nil
}

// suspendInEmby 在 Emby 暂停账号
func (s *Service) suspendInEmby(ctx context.Context, acc *Account) error {
	if !s.enableSync || s.embyClient == nil || acc.EmbyUserID == "" {
		return nil
	}

	if err := s.embyClient.DisableUser(ctx, acc.EmbyUserID); err != nil {
		acc.MarkSyncFailed(fmt.Errorf("suspend failed: %w", err))
		logger.Errorf("failed to suspend emby user %s: %v", acc.Username, err)
		return err
	}

	acc.MarkSynced(acc.EmbyUserID)
	return nil
}

// activateInEmby 在 Emby 激活账号
func (s *Service) activateInEmby(ctx context.Context, acc *Account) error {
	if !s.enableSync || s.embyClient == nil || acc.EmbyUserID == "" {
		return nil
	}

	if err := s.embyClient.EnableUser(ctx, acc.EmbyUserID); err != nil {
		acc.MarkSyncFailed(fmt.Errorf("activate failed: %w", err))
		logger.Errorf("failed to activate emby user %s: %v", acc.Username, err)
		return err
	}

	acc.MarkSynced(acc.EmbyUserID)
	return nil
}

// Create 创建账号
// 自动生成密码，返回明文密码和账号信息
func (s *Service) Create(ctx context.Context, username string, userID uint) (*Account, string, error) {
	// 清理用户名
	username = validator.SanitizeUsername(username)

	// 验证用户名
	if err := validator.ValidateUsername(username); err != nil {
		return nil, "", ValidationError("username", err.Error())
	}

	// 检查是否已存在
	if _, err := s.store.GetByUsername(ctx, username); err == nil {
		return nil, "", AlreadyExistsError(username)
	}

	// 检查用户配额
	if err := s.checkQuota(ctx, userID); err != nil {
		return nil, "", err
	}

	// 检查账号数量限制
	if err := s.checkAccountLimit(ctx, userID); err != nil {
		return nil, "", err
	}

	// 生成随机密码
	plainPassword, err := crypto.GeneratePassword(s.passwordLength)
	if err != nil {
		return nil, "", fmt.Errorf("generate password: %w", err)
	}

	// 加密密码
	hashedPassword, err := crypto.HashPassword(plainPassword)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	// 计算过期时间
	expireAt := time.Now().AddDate(0, 0, s.defaultExpire)

	// 创建账号
	acc := &Account{
		Username:   username,
		Password:   hashedPassword,
		UserID:     userID,
		Status:     StatusActive,
		ExpireAt:   &expireAt,
		MaxDevices: s.defaultDevices,
	}

	if err := s.store.Create(ctx, acc); err != nil {
		return nil, "", fmt.Errorf("create account: %w", err)
	}

	// 同步到 Emby
	if s.syncOnCreate {
		if err := s.syncToEmby(ctx, acc, plainPassword); err != nil {
			logger.Warnf("account %s created locally but emby sync failed: %v", acc.Username, err)
			// 更新同步状态
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		} else {
			// 同步成功，更新账号信息
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update emby user id for %s: %v", acc.Username, updateErr)
			}
		}
	}

	return acc, plainPassword, nil
}

// CreateWithPassword 创建账号(指定密码)
func (s *Service) CreateWithPassword(ctx context.Context, username, password string, userID uint) (*Account, error) {
	// 清理用户名
	username = validator.SanitizeUsername(username)

	// 验证
	if err := validator.ValidateUsername(username); err != nil {
		return nil, ValidationError("username", err.Error())
	}

	if err := validator.ValidatePassword(password); err != nil {
		return nil, ValidationError("password", err.Error())
	}

	// 检查是否已存在
	if _, err := s.store.GetByUsername(ctx, username); err == nil {
		return nil, AlreadyExistsError(username)
	}

	// 检查用户配额
	if err := s.checkQuota(ctx, userID); err != nil {
		return nil, err
	}

	// 检查账号数量限制
	if err := s.checkAccountLimit(ctx, userID); err != nil {
		return nil, err
	}

	// 加密密码
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 计算过期时间
	expireAt := time.Now().AddDate(0, 0, s.defaultExpire)

	// 创建账号
	acc := &Account{
		Username:   username,
		Password:   hashedPassword,
		UserID:     userID,
		Status:     StatusActive,
		ExpireAt:   &expireAt,
		MaxDevices: s.defaultDevices,
	}

	if err := s.store.Create(ctx, acc); err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	// 同步到 Emby
	if s.syncOnCreate {
		if err := s.syncToEmby(ctx, acc, password); err != nil {
			logger.Warnf("account %s created locally but emby sync failed: %v", acc.Username, err)
			// 更新同步状态
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		} else {
			// 同步成功，更新账号信息
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update emby user id for %s: %v", acc.Username, updateErr)
			}
		}
	}

	return acc, nil
}

// Get 获取账号
func (s *Service) Get(ctx context.Context, id uint) (*Account, error) {
	acc, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	return acc, nil
}

// GetByUsername 根据用户名获取账号
func (s *Service) GetByUsername(ctx context.Context, username string) (*Account, error) {
	acc, err := s.store.GetByUsername(ctx, username)
	if err != nil {
		return nil, NotFoundError(username)
	}
	return acc, nil
}

// ListByUser 列出用户的所有账号
func (s *Service) ListByUser(ctx context.Context, userID uint) ([]*Account, error) {
	accs, err := s.store.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return accs, nil
}

// ListAll 列出所有账号(分页)
func (s *Service) ListAll(ctx context.Context, offset, limit int) ([]*Account, error) {
	accs, err := s.store.ListAll(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list all accounts: %w", err)
	}
	return accs, nil
}

// ListAllWithUser 列出所有账号及关联用户信息(分页)
func (s *Service) ListAllWithUser(ctx context.Context, offset, limit int) ([]*AccountWithUser, error) {
	accs, err := s.store.ListAllWithUser(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list all accounts with user: %w", err)
	}
	return accs, nil
}

// GetWithUser 根据 ID 获取账号及用户信息
func (s *Service) GetWithUser(ctx context.Context, id uint) (*AccountWithUser, error) {
	acc, err := s.store.GetWithUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// Renew 续期账号
func (s *Service) Renew(ctx context.Context, id uint, days int) error {
	// 验证天数
	if err := validator.ValidateDays(days); err != nil {
		return ValidationError("days", err.Error())
	}

	acc, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	// 续期
	acc.Renew(days)

	if err := s.store.Update(ctx, acc); err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	return nil
}

// Delete 删除账号
func (s *Service) Delete(ctx context.Context, id uint) error {
	// 先获取账号信息
	acc, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	// 从 Emby 删除
	if s.syncOnDelete {
		if err := s.deleteFromEmby(ctx, acc); err != nil {
			logger.Warnf("failed to delete %s from emby: %v", acc.Username, err)
			// 继续删除本地记录
		}
	}

	// 删除本地记录
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, id uint, newPassword string) error {
	// 验证密码
	if err := validator.ValidatePassword(newPassword); err != nil {
		return ValidationError("password", err.Error())
	}

	acc, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	acc.SetPassword(hashedPassword)

	if err := s.store.Update(ctx, acc); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// 同步到 Emby
	if s.enableSync {
		if err := s.updatePasswordInEmby(ctx, acc, newPassword); err != nil {
			logger.Warnf("password updated locally for %s but emby sync failed: %v", acc.Username, err)
			// 更新同步状态
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		} else {
			// 同步成功，更新账号信息
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		}
	}

	return nil
}

// Suspend 暂停账号
func (s *Service) Suspend(ctx context.Context, id uint) error {
	acc, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	acc.Suspend()

	if err := s.store.Update(ctx, acc); err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	// 同步到 Emby
	if s.enableSync {
		if err := s.suspendInEmby(ctx, acc); err != nil {
			logger.Warnf("account %s suspended locally but emby sync failed: %v", acc.Username, err)
			// 更新同步状态
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		} else {
			// 同步成功，更新账号信息
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		}
	}

	return nil
}

// Activate 激活账号
func (s *Service) Activate(ctx context.Context, id uint) error {
	acc, err := s.store.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	acc.Activate()

	if err := s.store.Update(ctx, acc); err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	// 同步到 Emby
	if s.enableSync {
		if err := s.activateInEmby(ctx, acc); err != nil {
			logger.Warnf("account %s activated locally but emby sync failed: %v", acc.Username, err)
			// 更新同步状态
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		} else {
			// 同步成功，更新账号信息
			if updateErr := s.store.Update(ctx, acc); updateErr != nil {
				logger.Errorf("failed to update sync status for %s: %v", acc.Username, updateErr)
			}
		}
	}

	return nil
}

// CheckOwnership 检查账号所有权
func (s *Service) CheckOwnership(ctx context.Context, accountID, userID uint) error {
	acc, err := s.store.Get(ctx, accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get account: %w", err)
	}

	if acc.UserID != userID {
		return UnauthorizedError("account does not belong to user")
	}

	return nil
}

// Count 统计账号数量
func (s *Service) Count(ctx context.Context) (int64, error) {
	count, err := s.store.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count accounts: %w", err)
	}
	return count, nil
}

// CountByUser 统计用户的账号数量
func (s *Service) CountByUser(ctx context.Context, userID uint) (int64, error) {
	count, err := s.store.CountByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count user accounts: %w", err)
	}
	return count, nil
}

// CountByStatus 统计指定状态的账号数量
func (s *Service) CountByStatus(ctx context.Context, status Status) (int64, error) {
	count, err := s.store.CountByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("count accounts by status: %w", err)
	}
	return count, nil
}
