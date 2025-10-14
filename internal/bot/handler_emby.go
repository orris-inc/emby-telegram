// Package bot Emby 管理命令处理器
package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/account"
	"emby-telegram/internal/logger"
)

// handleCheckEmby 检查 Emby 服务器连接状态
func (b *Bot) handleCheckEmby(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "", err
	}

	if b.embyClient == nil {
		return "❌ Emby 同步已禁用或未配置", nil
	}

	// 测试连接
	if err := b.embyClient.Ping(ctx); err != nil {
		return fmt.Sprintf("❌ Emby 服务器连接失败\n错误: %v", err), nil
	}

	return "✅ Emby 服务器连接正常", nil
}

// handleSyncStatus 查看账号同步状态
func (b *Bot) handleSyncStatus(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if len(args) < 1 {
		return "", account.ValidationError("username", "用户名不能为空")
	}

	username := args[0]

	// 获取账号信息
	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	// 检查所有权
	userID := uint(msg.From.ID)
	if !b.isAdmin(msg.From.ID) {
		if err := b.accountService.CheckOwnership(ctx, acc.ID, userID); err != nil {
			return "", err
		}
	}

	// 构建状态消息
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<b>账号同步状态</b>\n\n"))
	builder.WriteString(fmt.Sprintf("用户名: <code>%s</code>\n", acc.Username))

	// 同步状态
	statusEmoji := map[string]string{
		"synced":  "✅",
		"pending": "⏳",
		"failed":  "❌",
	}
	emoji, ok := statusEmoji[acc.SyncStatus]
	if !ok {
		emoji = "❓"
	}
	builder.WriteString(fmt.Sprintf("同步状态: %s %s\n", emoji, acc.SyncStatus))

	// Emby User ID
	if acc.EmbyUserID != "" {
		builder.WriteString(fmt.Sprintf("Emby ID: <code>%s</code>\n", acc.EmbyUserID))
	} else {
		builder.WriteString("Emby ID: 未同步\n")
	}

	// 最后同步时间
	if acc.LastSyncAt != nil {
		builder.WriteString(fmt.Sprintf("最后同步: %s\n", acc.LastSyncAt.Format("2006-01-02 15:04:05")))
	}

	// 同步错误
	if acc.SyncError != "" {
		builder.WriteString(fmt.Sprintf("\n<b>同步错误</b>:\n<pre>%s</pre>", acc.SyncError))
	}

	return builder.String(), nil
}

// handleSyncAccount 手动同步账号到 Emby
func (b *Bot) handleSyncAccount(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "", err
	}

	if b.embyClient == nil {
		return "❌ Emby 同步已禁用或未配置", nil
	}

	if len(args) < 2 {
		return "", account.ValidationError("args", "用法: /syncaccount &lt;用户名&gt; &lt;密码&gt;")
	}

	username := args[0]
	password := args[1]

	// 获取账号信息
	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	// 如果已同步，提示
	if acc.IsSynced() {
		return fmt.Sprintf("⚠️ 账号 <code>%s</code> 已同步到 Emby (ID: %s)\n是否要重新同步？", acc.Username, acc.EmbyUserID), nil
	}

	// 尝试创建 Emby 用户
	embyUser, err := b.embyClient.CreateUser(ctx, acc.Username, password)
	if err != nil {
		logger.Errorf("手动同步账号失败: %v", err)
		return fmt.Sprintf("❌ 同步失败: %v", err), nil
	}

	// 更新设备限制
	if acc.MaxDevices > 0 {
		if err := b.embyClient.SetMaxActiveSessions(ctx, embyUser.ID, acc.MaxDevices); err != nil {
			logger.Warnf("设置设备限制失败: %v", err)
		}
	}

	// 更新账号同步状态
	acc.MarkSynced(embyUser.ID)
	// 这里账号状态已更新，记录日志
	logger.Infof("账号 %s 已同步到 Emby (ID: %s)", acc.Username, embyUser.ID)

	return fmt.Sprintf("✅ 账号已成功同步到 Emby\n用户名: <code>%s</code>\nEmby ID: <code>%s</code>", acc.Username, embyUser.ID), nil
}

// handleListEmbyUsers 列出 Emby 服务器上的所有用户
func (b *Bot) handleListEmbyUsers(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "", err
	}

	if b.embyClient == nil {
		return "❌ Emby 同步已禁用或未配置", nil
	}

	// 获取 Emby 用户列表
	users, err := b.embyClient.ListUsers(ctx)
	if err != nil {
		return fmt.Sprintf("❌ 获取 Emby 用户列表失败: %v", err), nil
	}

	if len(users) == 0 {
		return "📋 Emby 服务器上没有用户", nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<b>Emby 用户列表</b> (共 %d 个)\n\n", len(users)))

	for i, user := range users {
		status := "✅ 启用"
		if user.Policy.IsDisabled {
			status = "❌ 禁用"
		}

		builder.WriteString(fmt.Sprintf("%d. <code>%s</code>\n", i+1, user.Name))
		builder.WriteString(fmt.Sprintf("   状态: %s\n", status))
		builder.WriteString(fmt.Sprintf("   ID: <code>%s</code>\n", user.ID))

		if i < len(users)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// handleSetDeviceLimit 手动设置账号设备限制
func (b *Bot) handleSetDeviceLimit(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "", err
	}

	if b.embyClient == nil {
		return "❌ Emby 同步已禁用或未配置", nil
	}

	if len(args) < 2 {
		return "❌ 参数不足\n\n使用方法: <code>/setdevicelimit &lt;用户名&gt; &lt;设备数&gt;</code>\n例如: <code>/setdevicelimit john 3</code>", nil
	}

	username := args[0]
	limitStr := args[1]

	// 解析设备数
	var limit int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit < 0 {
		return "❌ 设备数必须是非负整数", nil
	}

	// 获取账号信息
	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return fmt.Sprintf("❌ 获取账号信息失败: %v", err), nil
	}

	// 检查是否已同步
	if acc.EmbyUserID == "" {
		return fmt.Sprintf("❌ 账号 <code>%s</code> 尚未同步到 Emby", acc.Username), nil
	}

	// 设置设备限制
	if err := b.embyClient.SetMaxActiveSessions(ctx, acc.EmbyUserID, limit); err != nil {
		logger.Errorf("设置设备限制失败: %v", err)
		return fmt.Sprintf("❌ 设置设备限制失败: %v", err), nil
	}

	logger.Infof("成功为账号 %s 设置设备限制: %d", acc.Username, limit)

	return fmt.Sprintf(`✅ <b>设备限制设置成功！</b>

账号: <code>%s</code>
Emby ID: <code>%s</code>
最大设备数: <b>%d</b>
`, acc.Username, acc.EmbyUserID, limit), nil
}
