// Package emby Emby HTTP 客户端
package emby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"emby-telegram/internal/logger"
)

// Client Emby HTTP 客户端
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	enabled    bool
	retryCount int
}

// NewClient 创建 Emby 客户端实例
func NewClient(baseURL, apiKey string, timeout int, retryCount int, enabled bool) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		enabled:    enabled,
		retryCount: retryCount,
	}
}

// IsEnabled 检查是否启用同步
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// doRequest 执行 HTTP 请求(带重试)
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	if !c.enabled {
		return ErrSyncDisabled
	}

	var lastErr error
	for i := 0; i <= c.retryCount; i++ {
		if i > 0 {
			logger.DebugKV("emby API retry attempt", "attempt", i, "method", method, "path", path)
			time.Sleep(time.Second * time.Duration(i)) // 指数退避
		}

		err := c.doRequestOnce(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// 某些错误不需要重试
		if err == ErrUnauthorized || err == ErrUserNotFound || err == ErrUserAlreadyExists {
			return err
		}
	}

	return lastErr
}

// doFormRequest 执行 application/x-www-form-urlencoded 格式的 HTTP 请求(带重试)
func (c *Client) doFormRequest(ctx context.Context, method, path string, formData url.Values, result interface{}) error {
	if !c.enabled {
		return ErrSyncDisabled
	}

	var lastErr error
	for i := 0; i <= c.retryCount; i++ {
		if i > 0 {
			logger.DebugKV("emby API retry attempt", "attempt", i, "method", method, "path", path)
			time.Sleep(time.Second * time.Duration(i))
		}

		err := c.doFormRequestOnce(ctx, method, path, formData, result)
		if err == nil {
			return nil
		}

		lastErr = err

		if err == ErrUnauthorized || err == ErrUserNotFound || err == ErrUserAlreadyExists {
			return err
		}
	}

	return lastErr
}

// doRequestOnce 执行单次 HTTP 请求
func (c *Client) doRequestOnce(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	// 构建完整 URL
	url := fmt.Sprintf("%s/emby%s", c.baseURL, path)

	// 序列化请求体
	var reqBody io.Reader
	var jsonData []byte
	if body != nil {
		var err error
		jsonData, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.apiKey)

	// 记录请求
	logger.InfoKV("emby API request", "method", method, "url", url)
	if body != nil {
		logger.DebugKV("request body", "length_bytes", len(jsonData))
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrServerUnavailable, err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// 记录响应
	logger.InfoKV("emby API response", "status_code", resp.StatusCode)
	logger.DebugKV("response body", "content", string(respBody))

	// 处理 HTTP 状态码
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, respBody)
	}

	// 解析响应
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		}
	}

	return nil
}

// doFormRequestOnce 执行单次 application/x-www-form-urlencoded 格式的 HTTP 请求
func (c *Client) doFormRequestOnce(ctx context.Context, method, path string, formData url.Values, result interface{}) error {
	// 构建完整 URL
	fullURL := fmt.Sprintf("%s/emby%s", c.baseURL, path)

	// 将表单数据编码为字符串
	formBody := formData.Encode()
	reqBody := strings.NewReader(formBody)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// 设置请求头（关键：使用 application/x-www-form-urlencoded）
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.apiKey)

	// 记录请求
	logger.InfoKV("emby API request (form)", "method", method, "url", fullURL)
	logger.DebugKV("form data", "length_bytes", len(formBody))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrServerUnavailable, err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// 记录响应
	logger.InfoKV("emby API response", "status_code", resp.StatusCode)
	logger.DebugKV("response body", "content", string(respBody))

	// 处理 HTTP 状态码
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, respBody)
	}

	// 解析响应
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		}
	}

	return nil
}

// handleErrorResponse 处理错误响应
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	message := string(body)

	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrUserNotFound
	case http.StatusConflict:
		return ErrUserAlreadyExists
	default:
		return ServerError(statusCode, message)
	}
}

// Ping 检查 Emby 服务器连接
func (c *Client) Ping(ctx context.Context) error {
	if !c.enabled {
		return ErrSyncDisabled
	}

	var info SystemInfo
	if err := c.doRequest(ctx, http.MethodGet, "/System/Info", nil, &info); err != nil {
		return fmt.Errorf("ping emby server: %w", err)
	}

	logger.InfoKV("emby server connected", "name", info.ServerName, "version", info.Version)
	return nil
}
