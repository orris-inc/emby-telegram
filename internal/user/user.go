// Package user 提供用户领域模型和业务逻辑
package user

import (
	"time"

	"gorm.io/gorm"
)

// Role 用户角色
type Role string

const (
	// RoleUser 普通用户
	RoleUser Role = "user"
	// RoleAdmin 管理员
	RoleAdmin Role = "admin"
)

// User 用户实体
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	TelegramID   int64          `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username     string         `gorm:"size:100" json:"username"`
	FirstName    string         `gorm:"size:100" json:"first_name"`
	LastName     string         `gorm:"size:100" json:"last_name"`
	Role         Role           `gorm:"size:20;default:user" json:"role"`
	IsBlocked    bool           `gorm:"default:false" json:"is_blocked"`
	AccountQuota int            `gorm:"default:0" json:"account_quota"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsAdmin 检查是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsUser 检查是否为普通用户
func (u *User) IsUser() bool {
	return u.Role == RoleUser
}

// Block 封禁用户
func (u *User) Block() {
	u.IsBlocked = true
}

// Unblock 解封用户
func (u *User) Unblock() {
	u.IsBlocked = false
}

// SetRole 设置角色
func (u *User) SetRole(role Role) {
	u.Role = role
}

// FullName 获取全名
func (u *User) FullName() string {
	if u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.FirstName
}

// DisplayName 获取显示名称
func (u *User) DisplayName() string {
	if u.Username != "" {
		return "@" + u.Username
	}
	return u.FullName()
}

// CanAccess 检查用户是否可以访问系统
func (u *User) CanAccess() bool {
	return !u.IsBlocked
}
