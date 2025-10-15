// Package emby 用户策略管理 API
package emby

import (
	"context"
	"fmt"
	"net/http"
)

// GetUserPolicy 获取用户策略
// 根据官方文档：https://dev.emby.media/reference/RestAPI/UserService/getUsersById.html
// 通过 GET /Users/{Id} 获取完整用户信息，其中包含 Policy
func (c *Client) GetUserPolicy(ctx context.Context, userID string) (*UserPolicy, error) {
	// 获取完整的用户信息
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user policy: %w", err)
	}

	return &user.Policy, nil
}

// UpdateUserPolicy 更新用户策略
func (c *Client) UpdateUserPolicy(ctx context.Context, userID string, policy *UserPolicy) error {
	path := fmt.Sprintf("/Users/%s/Policy", userID)

	if err := c.doRequest(ctx, http.MethodPost, path, policy, nil); err != nil {
		return fmt.Errorf("update user policy: %w", err)
	}

	return nil
}

// SetMaxActiveSessions 设置最大活动会话数(设备数)
// 使用 SimultaneousStreamLimit 字段来限制设备数
func (c *Client) SetMaxActiveSessions(ctx context.Context, userID string, maxSessions int) error {
	policy, err := c.GetUserPolicy(ctx, userID)
	if err != nil {
		return err
	}

	policy.SimultaneousStreamLimit = int32(maxSessions)

	return c.UpdateUserPolicy(ctx, userID, policy)
}

// SetMediaLibraryAccess 设置媒体库访问权限
func (c *Client) SetMediaLibraryAccess(ctx context.Context, userID string, folderIDs []string) error {
	policy, err := c.GetUserPolicy(ctx, userID)
	if err != nil {
		return err
	}

	if len(folderIDs) == 0 {
		policy.EnableAllFolders = true
		policy.EnabledFolders = []string{}
	} else {
		policy.EnableAllFolders = false
		policy.EnabledFolders = folderIDs
	}

	return c.UpdateUserPolicy(ctx, userID, policy)
}

// CreateDefaultPolicy 创建默认用户策略
func CreateDefaultPolicy(maxDevices int) *UserPolicy {
	return &UserPolicy{
		IsAdministrator:                  false,
		IsHidden:                         true,
		IsHiddenRemotely:                 true,
		IsHiddenFromUnusedDevices:        false,
		IsDisabled:                       false,
		LockedOutDate:                    0,
		MaxParentalRating:                10,
		AllowTagOrRating:                 false,
		BlockedTags:                      []string{},
		IsTagBlockingModeInclusive:       false,
		IncludeTags:                      []string{},
		EnableUserPreferenceAccess:       true,
		AccessSchedules:                  []string{},
		BlockUnratedItems:                []string{},
		EnableRemoteControlOfOtherUsers:  false,
		EnableSharedDeviceControl:        false,
		EnableRemoteAccess:               true,
		EnableLiveTvManagement:           false,
		EnableLiveTvAccess:               false,
		EnableMediaPlayback:              true,
		EnableAudioPlaybackTranscoding:   false,
		EnableVideoPlaybackTranscoding:   false,
		EnablePlaybackRemuxing:           false,
		EnableContentDeletion:            false,
		RestrictedFeatures:               []string{},
		EnableContentDeletionFromFolders: []string{},
		EnableContentDownloading:         false,
		EnableSubtitleDownloading:        false,
		EnableSubtitleManagement:         false,
		EnableSyncTranscoding:            false,
		EnableMediaConversion:            false,
		EnabledChannels:                  []string{},
		EnableAllChannels:                true,
		EnabledFolders:                   []string{},
		EnableAllFolders:                 true,
		InvalidLoginAttemptCount:         0,
		EnablePublicSharing:              false,
		RemoteClientBitrateLimit:         0,
		AuthenticationProviderId:         "",
		ExcludedSubFolders:               []string{},
		SimultaneousStreamLimit:          int32(maxDevices),
		EnabledDevices:                   []string{},
		EnableAllDevices:                 true,
	}
}
