package invitecode

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type UserGetter interface {
	Get(ctx context.Context, id uint) (User, error)
	SetQuota(ctx context.Context, userID uint, quota int) error
	MarkInviteCodeUsed(ctx context.Context, userID uint) error
}

type User struct {
	ID             uint
	AccountQuota   int
	UsedInviteCode bool
}

type Service struct {
	store      Store
	userGetter UserGetter
}

func NewService(store Store, userGetter UserGetter) *Service {
	if store == nil {
		panic("invitecode.NewService: store cannot be nil")
	}
	if userGetter == nil {
		panic("invitecode.NewService: userGetter cannot be nil")
	}
	return &Service{
		store:      store,
		userGetter: userGetter,
	}
}

func (s *Service) Generate(ctx context.Context, maxUses int, expireDays int, description string, createdBy int64) (*InviteCode, error) {
	if maxUses != -1 && maxUses <= 0 {
		return nil, ErrInvalidMaxUses
	}

	code, err := generateCode()
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}

	var expireAt *time.Time
	if expireDays > 0 {
		expireTime := time.Now().AddDate(0, 0, expireDays)
		expireAt = &expireTime
	}

	inviteCode := &InviteCode{
		Code:        code,
		MaxUses:     maxUses,
		CurrentUses: 0,
		Description: description,
		ExpireAt:    expireAt,
		Status:      StatusActive,
		CreatedBy:   createdBy,
	}

	if err := s.store.Create(ctx, inviteCode); err != nil {
		return nil, fmt.Errorf("create invite code: %w", err)
	}

	return inviteCode, nil
}

func (s *Service) Activate(ctx context.Context, code string, userID uint) error {
	code = strings.ToUpper(strings.TrimSpace(code))

	if code == "" {
		return ErrInvalidCode
	}

	user, err := s.userGetter.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if user.UsedInviteCode {
		return ErrAlreadyUsed
	}

	if user.AccountQuota > 0 {
		return ErrHasQuota
	}

	inviteCode, err := s.store.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NotFoundError(code)
		}
		return fmt.Errorf("get invite code: %w", err)
	}

	if !inviteCode.IsValid() {
		if inviteCode.Status == StatusRevoked {
			return CodeRevokedError(code)
		}
		if inviteCode.IsExpired() {
			return CodeExpiredError(code)
		}
		if inviteCode.IsExhausted() {
			return CodeExhaustedError(code)
		}
		return ErrInvalidCode
	}

	existingUsage, err := s.store.GetUsageByUser(ctx, userID)
	if err == nil && existingUsage != nil {
		return ErrAlreadyUsed
	}

	if err := s.userGetter.SetQuota(ctx, userID, 1); err != nil {
		return fmt.Errorf("set user quota: %w", err)
	}

	if err := s.userGetter.MarkInviteCodeUsed(ctx, userID); err != nil {
		return fmt.Errorf("mark invite code used: %w", err)
	}

	usage := &InviteCodeUsage{
		InviteCodeID: inviteCode.ID,
		UserID:       userID,
		UsedAt:       time.Now(),
	}

	if err := s.store.RecordUsage(ctx, usage); err != nil {
		return fmt.Errorf("record usage: %w", err)
	}

	inviteCode.MarkUsed()
	if err := s.store.Update(ctx, inviteCode); err != nil {
		return fmt.Errorf("update invite code: %w", err)
	}

	return nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (*InviteCode, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	inviteCode, err := s.store.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, NotFoundError(code)
		}
		return nil, fmt.Errorf("get invite code: %w", err)
	}

	return inviteCode, nil
}

func (s *Service) GetWithUsage(ctx context.Context, code string) (*InviteCodeWithUsage, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	inviteCode, err := s.store.GetWithUsage(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, NotFoundError(code)
		}
		return nil, fmt.Errorf("get invite code with usage: %w", err)
	}

	return inviteCode, nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]*InviteCode, error) {
	codes, err := s.store.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list invite codes: %w", err)
	}
	return codes, nil
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	count, err := s.store.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count invite codes: %w", err)
	}
	return count, nil
}

func (s *Service) Revoke(ctx context.Context, code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))

	inviteCode, err := s.store.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NotFoundError(code)
		}
		return fmt.Errorf("get invite code: %w", err)
	}

	inviteCode.Revoke()

	if err := s.store.Update(ctx, inviteCode); err != nil {
		return fmt.Errorf("update invite code: %w", err)
	}

	return nil
}

func generateCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const codeLength = 8

	code := make([]byte, codeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}

	return string(code), nil
}
