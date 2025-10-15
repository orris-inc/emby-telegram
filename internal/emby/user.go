// Package emby 用户管理 API
package emby

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"emby-telegram/internal/logger"
)

// CreateUser 创建 Emby 用户
// 根据官方文档：https://dev.emby.media/reference/RestAPI/UserService/postUsersNew.html
// CreateUserByName 支持在创建时直接设置密码
func (c *Client) CreateUser(ctx context.Context, name, password string) (*EmbyUser, error) {
	// 在创建时直接设置密码
	req := CreateUserRequest{
		Name:     name,
		Password: password, // 直接设置密码
	}

	var user EmbyUser
	if err := c.doRequest(ctx, http.MethodPost, "/Users/New", req, &user); err != nil {
		return nil, fmt.Errorf("create emby user: %w", err)
	}

	logger.Infof("emby user created: %s (id: %s)", user.Name, user.ID)

	// 检查密码是否已成功设置
	if password != "" && !user.HasPassword && !user.HasConfiguredPassword {
		// 如果密码未在创建时设置成功，使用备用方法
		logger.Warnf("password not set on creation, trying fallback method for user: %s", user.Name)
		if err := c.UpdatePassword(ctx, user.ID, password); err != nil {
			logger.Errorf("failed to set password for user %s: %v", user.Name, err)
			// 如果设置密码失败，尝试删除已创建的用户
			_ = c.DeleteUser(ctx, user.ID)
			return nil, fmt.Errorf("set user password: %w", err)
		}
		logger.Infof("password set successfully for user: %s", user.Name)

		// 重新获取用户信息以验证密码是否已设置
		updatedUser, err := c.GetUser(ctx, user.ID)
		if err != nil {
			logger.Warnf("failed to re-fetch user info for verification: %v", err)
		} else {
			// 更新用户信息
			user = *updatedUser
		}
	}

	return &user, nil
}

// GetUser 根据 ID 获取用户
func (c *Client) GetUser(ctx context.Context, userID string) (*EmbyUser, error) {
	var user EmbyUser
	path := fmt.Sprintf("/Users/%s", userID)

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &user); err != nil {
		return nil, fmt.Errorf("get emby user: %w", err)
	}

	return &user, nil
}

// ListUsers 列出所有用户
func (c *Client) ListUsers(ctx context.Context) ([]*EmbyUser, error) {
	var users []*EmbyUser

	if err := c.doRequest(ctx, http.MethodGet, "/Users", nil, &users); err != nil {
		return nil, fmt.Errorf("list emby users: %w", err)
	}

	return users, nil
}

// DeleteUser 删除用户
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/Users/%s", userID)

	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("delete emby user: %w", err)
	}

	return nil
}

// UpdatePassword 修改用户密码
// 根据浏览器实际使用方式：使用 application/x-www-form-urlencoded 格式
// 只发送 NewPw 参数，不发送其他参数
func (c *Client) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	// 构建表单数据（只包含 NewPw，与浏览器完全一致）
	formData := url.Values{}
	formData.Set("NewPw", newPassword)

	path := fmt.Sprintf("/Users/%s/Password", userID)

	// 使用 form-urlencoded 格式发送请求
	if err := c.doFormRequest(ctx, http.MethodPost, path, formData, nil); err != nil {
		return fmt.Errorf("update emby user password: %w", err)
	}

	return nil
}

// AuthenticateUser 测试用户认证（用于验证密码是否正确）
func (c *Client) AuthenticateUser(ctx context.Context, username, password string) (*EmbyUser, error) {
	req := map[string]string{
		"Username": username,
		"Pw":       password,
	}

	var authResponse struct {
		User         EmbyUser `json:"User"`
		SessionInfo  map[string]interface{} `json:"SessionInfo"`
		AccessToken  string `json:"AccessToken"`
		ServerId     string `json:"ServerId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, "/Users/AuthenticateByName", req, &authResponse); err != nil {
		return nil, fmt.Errorf("authenticate user: %w", err)
	}

	return &authResponse.User, nil
}

// GetUserByName 根据用户名查找用户
func (c *Client) GetUserByName(ctx context.Context, name string) (*EmbyUser, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Name == name {
			return user, nil
		}
	}

	return nil, NotFoundError(name)
}

// DisableUser 禁用用户(通过更新策略)
func (c *Client) DisableUser(ctx context.Context, userID string) error {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	policy := user.Policy
	policy.IsDisabled = true

	return c.UpdateUserPolicy(ctx, userID, &policy)
}

// EnableUser 启用用户
func (c *Client) EnableUser(ctx context.Context, userID string) error {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	policy := user.Policy
	policy.IsDisabled = false

	return c.UpdateUserPolicy(ctx, userID, &policy)
}
