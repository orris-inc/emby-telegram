package invitecode

import (
	"time"

	"gorm.io/gorm"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusExpired Status = "expired"
	StatusRevoked Status = "revoked"
)

type InviteCode struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Code        string         `gorm:"uniqueIndex;size:20;not null" json:"code"`
	MaxUses     int            `gorm:"not null;default:-1" json:"max_uses"`
	CurrentUses int            `gorm:"not null;default:0" json:"current_uses"`
	Description string         `gorm:"size:200" json:"description"`
	ExpireAt    *time.Time     `json:"expire_at"`
	Status      Status         `gorm:"size:20;default:active" json:"status"`
	CreatedBy   int64          `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InviteCode) TableName() string {
	return "invite_codes"
}

func (ic *InviteCode) IsExpired() bool {
	if ic.ExpireAt == nil {
		return false
	}
	return time.Now().After(*ic.ExpireAt)
}

func (ic *InviteCode) IsExhausted() bool {
	if ic.MaxUses == -1 {
		return false
	}
	return ic.CurrentUses >= ic.MaxUses
}

func (ic *InviteCode) IsValid() bool {
	return ic.Status == StatusActive && !ic.IsExpired() && !ic.IsExhausted()
}

func (ic *InviteCode) Revoke() {
	ic.Status = StatusRevoked
}

func (ic *InviteCode) MarkUsed() {
	ic.CurrentUses++
	if ic.IsExpired() {
		ic.Status = StatusExpired
	} else if ic.IsExhausted() {
		ic.Status = StatusExpired
	}
}

type InviteCodeUsage struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	InviteCodeID uint       `gorm:"not null;index" json:"invite_code_id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	UsedAt       time.Time  `json:"used_at"`
	InviteCode   InviteCode `gorm:"foreignKey:InviteCodeID" json:"-"`
}

func (InviteCodeUsage) TableName() string {
	return "invite_code_usage"
}

type InviteCodeWithUsage struct {
	*InviteCode
	UsageRecords []*InviteCodeUsage `json:"usage_records"`
}
