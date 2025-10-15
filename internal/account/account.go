// Package account 提供账号领域模型和业务逻辑
package account

import (
	"time"

	"gorm.io/gorm"
)

// Status 账号状态
type Status string

const (
	// StatusActive 激活状态
	StatusActive Status = "active"
	// StatusSuspended 暂停状态
	StatusSuspended Status = "suspended"
	// StatusExpired 过期状态
	StatusExpired Status = "expired"
)

// Account 账号实体
type Account struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Username   string         `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password   string         `gorm:"size:255;not null" json:"-"` // 不序列化密码
	Email      string         `gorm:"size:100" json:"email"`
	UserID     uint           `gorm:"index;not null" json:"user_id"` // 关联的 Telegram 用户
	Status     Status         `gorm:"size:20;default:active" json:"status"`
	ExpireAt   *time.Time     `json:"expire_at,omitempty"`
	MaxDevices int            `gorm:"default:3" json:"max_devices"`

	// Emby 同步字段
	EmbyUserID string         `gorm:"size:100;index" json:"emby_user_id,omitempty"` // Emby 用户 ID
	SyncStatus string         `gorm:"size:20;default:pending" json:"sync_status"`    // synced/pending/failed
	LastSyncAt *time.Time     `json:"last_sync_at,omitempty"`
	SyncError  string         `gorm:"type:text" json:"sync_error,omitempty"` // 同步错误信息

	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "accounts"
}

// AccountWithUser 账号及关联用户信息
type AccountWithUser struct {
	Account
	OwnerUsername  string
	OwnerFirstName string
	OwnerTelegramID int64
}

// GetOwnerDisplayName 获取所有者显示名称
func (a *AccountWithUser) GetOwnerDisplayName() string {
	if a.OwnerUsername != "" {
		return "@" + a.OwnerUsername
	}
	return a.OwnerFirstName
}

// IsActive 检查账号是否激活
func (a *Account) IsActive() bool {
	return a.Status == StatusActive
}

// IsExpired 检查账号是否过期
func (a *Account) IsExpired() bool {
	if a.ExpireAt == nil {
		return false
	}
	return time.Now().After(*a.ExpireAt)
}

// IsSuspended 检查账号是否被暂停
func (a *Account) IsSuspended() bool {
	return a.Status == StatusSuspended
}

// Activate 激活账号
func (a *Account) Activate() {
	a.Status = StatusActive
}

// Suspend 暂停账号
func (a *Account) Suspend() {
	a.Status = StatusSuspended
}

// MarkExpired 标记为过期
func (a *Account) MarkExpired() {
	a.Status = StatusExpired
}

// Renew 续期账号
func (a *Account) Renew(days int) {
	now := time.Now()
	if a.ExpireAt != nil && a.ExpireAt.After(now) {
		// 如果未过期，在原有基础上延长
		newExpire := a.ExpireAt.AddDate(0, 0, days)
		a.ExpireAt = &newExpire
	} else {
		// 如果已过期，从现在开始计算
		newExpire := now.AddDate(0, 0, days)
		a.ExpireAt = &newExpire
	}
	// 续期时自动激活
	a.Status = StatusActive
}

// SetPassword 设置密码
func (a *Account) SetPassword(hashedPassword string) {
	a.Password = hashedPassword
}

// DaysUntilExpire 计算距离过期还有多少天
func (a *Account) DaysUntilExpire() int {
	if a.ExpireAt == nil {
		return -1 // 永久有效
	}
	duration := time.Until(*a.ExpireAt)
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// IsValid 检查账号是否有效(激活且未过期)
func (a *Account) IsValid() bool {
	return a.IsActive() && !a.IsExpired()
}

// IsSynced 检查是否已同步到 Emby
func (a *Account) IsSynced() bool {
	return a.SyncStatus == "synced" && a.EmbyUserID != ""
}

// MarkSynced 标记为已同步
func (a *Account) MarkSynced(embyUserID string) {
	now := time.Now()
	a.EmbyUserID = embyUserID
	a.SyncStatus = "synced"
	a.LastSyncAt = &now
	a.SyncError = ""
}

// MarkSyncFailed 标记同步失败
func (a *Account) MarkSyncFailed(err error) {
	now := time.Now()
	a.SyncStatus = "failed"
	a.LastSyncAt = &now
	if err != nil {
		a.SyncError = err.Error()
	}
}

// MarkSyncPending 标记为待同步
func (a *Account) MarkSyncPending() {
	a.SyncStatus = "pending"
	a.SyncError = ""
}
