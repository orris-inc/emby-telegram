// Package emby Emby API 数据结构定义
package emby

import "time"

// EmbyUser Emby 用户结构
type EmbyUser struct {
	ID                       string     `json:"Id"`
	Name                     string     `json:"Name"`
	ServerId                 string     `json:"ServerId"`
	HasPassword              bool       `json:"HasPassword"`
	HasConfiguredPassword    bool       `json:"HasConfiguredPassword"`
	HasConfiguredEasyPassword bool      `json:"HasConfiguredEasyPassword"`
	EnableAutoLogin          bool       `json:"EnableAutoLogin"`
	LastLoginDate            *time.Time `json:"LastLoginDate"`
	LastActivityDate         *time.Time `json:"LastActivityDate"`
	Policy                   UserPolicy `json:"Policy"`
}

// UserPolicy 用户策略
// 根据官方文档：https://dev.emby.media/reference/RestAPI/UserService/postUsersByIdPolicy.html#MediaBrowser_Model_Users_UserPolicy
type UserPolicy struct {
	IsAdministrator                 bool     `json:"IsAdministrator"`
	IsHidden                        bool     `json:"IsHidden"`
	IsHiddenRemotely                bool     `json:"IsHiddenRemotely"`
	IsHiddenFromUnusedDevices       bool     `json:"IsHiddenFromUnusedDevices"`
	IsDisabled                      bool     `json:"IsDisabled"`
	LockedOutDate                   int64    `json:"LockedOutDate"`
	MaxParentalRating               int32    `json:"MaxParentalRating"`
	AllowTagOrRating                bool     `json:"AllowTagOrRating"`
	BlockedTags                     []string `json:"BlockedTags"`
	IsTagBlockingModeInclusive      bool     `json:"IsTagBlockingModeInclusive"`
	IncludeTags                     []string `json:"IncludeTags"`
	EnableUserPreferenceAccess      bool     `json:"EnableUserPreferenceAccess"`
	AccessSchedules                 []string `json:"AccessSchedules"`
	BlockUnratedItems               []string `json:"BlockUnratedItems"`
	EnableRemoteControlOfOtherUsers bool     `json:"EnableRemoteControlOfOtherUsers"`
	EnableSharedDeviceControl       bool     `json:"EnableSharedDeviceControl"`
	EnableRemoteAccess              bool     `json:"EnableRemoteAccess"`
	EnableLiveTvManagement          bool     `json:"EnableLiveTvManagement"`
	EnableLiveTvAccess              bool     `json:"EnableLiveTvAccess"`
	EnableMediaPlayback             bool     `json:"EnableMediaPlayback"`
	EnableAudioPlaybackTranscoding  bool     `json:"EnableAudioPlaybackTranscoding"`
	EnableVideoPlaybackTranscoding  bool     `json:"EnableVideoPlaybackTranscoding"`
	EnablePlaybackRemuxing          bool     `json:"EnablePlaybackRemuxing"`
	EnableContentDeletion           bool     `json:"EnableContentDeletion"`
	RestrictedFeatures              []string `json:"RestrictedFeatures"`
	EnableContentDeletionFromFolders []string `json:"EnableContentDeletionFromFolders"`
	EnableContentDownloading        bool     `json:"EnableContentDownloading"`
	EnableSubtitleDownloading       bool     `json:"EnableSubtitleDownloading"`
	EnableSubtitleManagement        bool     `json:"EnableSubtitleManagement"`
	EnableSyncTranscoding           bool     `json:"EnableSyncTranscoding"`
	EnableMediaConversion           bool     `json:"EnableMediaConversion"`
	EnabledChannels                 []string `json:"EnabledChannels"`
	EnableAllChannels               bool     `json:"EnableAllChannels"`
	EnabledFolders                  []string `json:"EnabledFolders"`
	EnableAllFolders                bool     `json:"EnableAllFolders"`
	InvalidLoginAttemptCount        int32    `json:"InvalidLoginAttemptCount"`
	EnablePublicSharing             bool     `json:"EnablePublicSharing"`
	RemoteClientBitrateLimit        int32    `json:"RemoteClientBitrateLimit"`
	AuthenticationProviderId        string   `json:"AuthenticationProviderId"`
	ExcludedSubFolders              []string `json:"ExcludedSubFolders"`
	SimultaneousStreamLimit         int32    `json:"SimultaneousStreamLimit"`
	EnabledDevices                  []string `json:"EnabledDevices"`
	EnableAllDevices                bool     `json:"EnableAllDevices"`
}

// CreateUserRequest 创建用户请求
// 根据 Emby API: CreateUserByName 可以包含 Password 字段
type CreateUserRequest struct {
	Name     string `json:"Name"`
	Password string `json:"Password,omitempty"` // 可选：直接在创建时设置密码
}

// UpdatePasswordRequest 修改密码请求
// 根据 Emby API 文档: https://dev.emby.media/reference/RestAPI/UserService/postUsersByIdPassword.html
// 以及社区讨论: https://emby.media/community/index.php?/topic/110010-change-password/
type UpdatePasswordRequest struct {
	Id            string `json:"Id,omitempty"`     // 用户 ID（可选，URL 中已包含）
	CurrentPw     string `json:"CurrentPw"`        // 当前密码（新用户设置密码时为空字符串）
	NewPw         string `json:"NewPw"`            // 新密码
	ResetPassword bool   `json:"ResetPassword"`    // 重置标志
}

// SystemInfo 系统信息
type SystemInfo struct {
	ID                 string `json:"Id"`
	ServerName         string `json:"ServerName"`
	Version            string `json:"Version"`
	OperatingSystem    string `json:"OperatingSystem"`
	OperatingSystemDisplayName string `json:"OperatingSystemDisplayName"`
	LocalAddress       string `json:"LocalAddress"`
	WanAddress         string `json:"WanAddress"`
	HasPendingRestart  bool   `json:"HasPendingRestart"`
	HasUpdateAvailable bool   `json:"HasUpdateAvailable"`
}

// AuthenticationResult 认证结果
type AuthenticationResult struct {
	User        EmbyUser `json:"User"`
	AccessToken string   `json:"AccessToken"`
	ServerId    string   `json:"ServerId"`
}
