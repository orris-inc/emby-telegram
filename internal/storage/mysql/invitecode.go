package mysql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"emby-telegram/internal/invitecode"
)

type InviteCodeStore struct {
	db *gorm.DB
}

func NewInviteCodeStore(db *gorm.DB) *InviteCodeStore {
	return &InviteCodeStore{db: db}
}

func (s *InviteCodeStore) Create(ctx context.Context, inviteCode *invitecode.InviteCode) error {
	return s.db.WithContext(ctx).Create(inviteCode).Error
}

func (s *InviteCodeStore) Get(ctx context.Context, id uint) (*invitecode.InviteCode, error) {
	var ic invitecode.InviteCode
	err := s.db.WithContext(ctx).First(&ic, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, invitecode.ErrNotFound
		}
		return nil, err
	}
	return &ic, nil
}

func (s *InviteCodeStore) GetByCode(ctx context.Context, code string) (*invitecode.InviteCode, error) {
	var ic invitecode.InviteCode
	err := s.db.WithContext(ctx).Where("code = ?", code).First(&ic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, invitecode.ErrNotFound
		}
		return nil, err
	}
	return &ic, nil
}

func (s *InviteCodeStore) GetWithUsage(ctx context.Context, code string) (*invitecode.InviteCodeWithUsage, error) {
	var ic invitecode.InviteCode
	err := s.db.WithContext(ctx).Where("code = ?", code).First(&ic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, invitecode.ErrNotFound
		}
		return nil, err
	}

	var usages []*invitecode.InviteCodeUsage
	if err := s.db.WithContext(ctx).Where("invite_code_id = ?", ic.ID).Order("used_at DESC").Find(&usages).Error; err != nil {
		return nil, err
	}

	return &invitecode.InviteCodeWithUsage{
		InviteCode:   &ic,
		UsageRecords: usages,
	}, nil
}

func (s *InviteCodeStore) Update(ctx context.Context, inviteCode *invitecode.InviteCode) error {
	return s.db.WithContext(ctx).Save(inviteCode).Error
}

func (s *InviteCodeStore) List(ctx context.Context, offset, limit int) ([]*invitecode.InviteCode, error) {
	var codes []*invitecode.InviteCode
	err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *InviteCodeStore) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&invitecode.InviteCode{}).Count(&count).Error
	return count, err
}

func (s *InviteCodeStore) RecordUsage(ctx context.Context, usage *invitecode.InviteCodeUsage) error {
	return s.db.WithContext(ctx).Create(usage).Error
}

func (s *InviteCodeStore) GetUsageByUser(ctx context.Context, userID uint) (*invitecode.InviteCodeUsage, error) {
	var usage invitecode.InviteCodeUsage
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&usage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &usage, nil
}
