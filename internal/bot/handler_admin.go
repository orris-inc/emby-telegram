// Package bot 管理员命令处理器
package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/account"
	"emby-telegram/internal/logger"
	"emby-telegram/internal/user"
	"emby-telegram/pkg/timeutil"
)

// handleAdmin 处理 /admin 命令
func (b *Bot) handleAdmin(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	return `🔑 <b>管理员命令</b>

<b>用户管理:</b>
/users [页码] - 列出所有用户
/setrole &lt;telegram_id&gt; &lt;admin|user&gt; - 设置用户角色
/blockuser &lt;telegram_id&gt; - 封禁用户
/unblockuser &lt;telegram_id&gt; - 解封用户

<b>账号管理:</b>
/accounts [页码] - 列出所有账号
/deleteaccount &lt;用户名&gt; - 删除账号
/suspend &lt;用户名&gt; - 暂停账号
/activate &lt;用户名&gt; - 激活账号

<b>Emby 管理:</b>
/checkemby - 检查 Emby 服务器连接状态
/syncaccount &lt;用户名&gt; &lt;密码&gt; - 手动同步账号到 Emby
/embyusers - 列出 Emby 服务器上的所有用户
/setdevicelimit &lt;用户名&gt; &lt;设备数&gt; - 设置账号设备限制
/updatepolicies - 批量更新所有非管理员用户策略

<b>统计信息:</b>
/stats - 查看系统统计
/playingstats - 查看 Emby 播放状态

<b>使用示例:</b>
<code>/users 1</code> - 查看第1页用户
<code>/setrole 123456 admin</code> - 设置用户为管理员
<code>/blockuser 123456</code> - 封禁用户
<code>/deleteaccount emby_john</code> - 删除账号
<code>/checkemby</code> - 检查 Emby 连接
<code>/setdevicelimit john 5</code> - 设置设备限制
`, nil
}

// handleListUsers 处理 /users 命令
func (b *Bot) handleListUsers(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	page := 1
	if hasArg(args, 1) {
		if p, err := strconv.Atoi(getArg(args, 0)); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	offset := (page - 1) * limit

	users, err := b.userService.List(ctx, offset, limit)
	if err != nil {
		return "", fmt.Errorf("获取用户列表失败: %w", err)
	}

	totalCount, _ := b.userService.Count(ctx)

	if len(users) == 0 {
		return "没有找到用户", nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("👥 <b>用户列表</b> (第 %d 页，共 %d 个用户)\n\n", page, totalCount))

	for i, u := range users {
		roleEmoji := "👤"
		if u.IsAdmin() {
			roleEmoji = "👑"
		}

		statusEmoji := "✅"
		if u.IsBlocked {
			statusEmoji = "🚫"
		}

		builder.WriteString(fmt.Sprintf("%d. %s %s %s\n", offset+i+1, roleEmoji, u.DisplayName(), statusEmoji))
		builder.WriteString(fmt.Sprintf("   ID: <code>%d</code> | 角色: %s\n", u.TelegramID, u.Role))
		builder.WriteString(fmt.Sprintf("   注册时间: %s\n\n", timeutil.FormatDate(u.CreatedAt)))
	}

	totalPages := (int(totalCount) + limit - 1) / limit
	if totalPages > 1 {
		builder.WriteString(fmt.Sprintf("📄 第 %d/%d 页\n", page, totalPages))
		if page < totalPages {
			builder.WriteString(fmt.Sprintf("使用 <code>/users %d</code> 查看下一页\n", page+1))
		}
	}

	return builder.String(), nil
}

// handleListAccounts 处理 /accounts 命令
func (b *Bot) handleListAccounts(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	page := 1
	if hasArg(args, 1) {
		if p, err := strconv.Atoi(getArg(args, 0)); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	offset := (page - 1) * limit

	accounts, err := b.accountService.ListAll(ctx, offset, limit)
	if err != nil {
		return "", fmt.Errorf("获取账号列表失败: %w", err)
	}

	totalCount, _ := b.accountService.Count(ctx)

	if len(accounts) == 0 {
		return "没有找到账号", nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📋 <b>所有账号列表</b> (第 %d 页，共 %d 个账号)\n\n", page, totalCount))

	for i, acc := range accounts {
		status := getStatusEmoji(string(acc.Status))
		expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

		builder.WriteString(fmt.Sprintf("%d. <b>%s</b> %s\n", offset+i+1, acc.Username, status))
		builder.WriteString(fmt.Sprintf("   用户ID: %d | 状态: %s\n", acc.UserID, acc.Status))
		builder.WriteString(fmt.Sprintf("   到期: %s\n\n", expireInfo))
	}

	totalPages := (int(totalCount) + limit - 1) / limit
	if totalPages > 1 {
		builder.WriteString(fmt.Sprintf("📄 第 %d/%d 页\n", page, totalPages))
		if page < totalPages {
			builder.WriteString(fmt.Sprintf("使用 <code>/accounts %d</code> 查看下一页\n", page+1))
		}
	}

	return builder.String(), nil
}

// handleDeleteAccount 处理 /deleteaccount 命令
func (b *Bot) handleDeleteAccount(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供用户名\n\n使用方法: <code>/deleteaccount &lt;用户名&gt;</code>", nil
	}

	username := getArg(args, 0)

	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("获取账号信息失败: %w", err)
	}

	if err := b.accountService.Delete(ctx, acc.ID); err != nil {
		return "", fmt.Errorf("删除账号失败: %w", err)
	}

	return fmt.Sprintf("✅ 账号 <b>%s</b> 已删除", acc.Username), nil
}

// handleSuspendAccount 处理 /suspend 命令
func (b *Bot) handleSuspendAccount(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供用户名\n\n使用方法: <code>/suspend &lt;用户名&gt;</code>", nil
	}

	username := getArg(args, 0)

	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("获取账号信息失败: %w", err)
	}

	if err := b.accountService.Suspend(ctx, acc.ID); err != nil {
		return "", fmt.Errorf("暂停账号失败: %w", err)
	}

	return fmt.Sprintf("⏸️ 账号 <b>%s</b> 已暂停", acc.Username), nil
}

// handleActivateAccount 处理 /activate 命令
func (b *Bot) handleActivateAccount(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供用户名\n\n使用方法: <code>/activate &lt;用户名&gt;</code>", nil
	}

	username := getArg(args, 0)

	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("获取账号信息失败: %w", err)
	}

	if err := b.accountService.Activate(ctx, acc.ID); err != nil {
		return "", fmt.Errorf("激活账号失败: %w", err)
	}

	return fmt.Sprintf("✅ 账号 <b>%s</b> 已激活", acc.Username), nil
}

// handleSetRole 处理 /setrole 命令
func (b *Bot) handleSetRole(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 2) {
		return "❌ 参数不足\n\n使用方法: <code>/setrole &lt;telegram_id&gt; &lt;admin|user&gt;</code>\n例如: <code>/setrole 123456 admin</code>", nil
	}

	telegramIDStr := getArg(args, 0)
	role := getArg(args, 1)

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		return "❌ Telegram ID 必须是有效的数字", nil
	}

	if role != "admin" && role != "user" {
		return "❌ 角色必须是 admin 或 user", nil
	}

	if err := b.userService.SetRole(ctx, telegramID, role); err != nil {
		return "", fmt.Errorf("设置角色失败: %w", err)
	}

	roleEmoji := "👤"
	if role == "admin" {
		roleEmoji = "👑"
	}

	return fmt.Sprintf("%s 已将用户 %d 设置为 <b>%s</b>", roleEmoji, telegramID, role), nil
}

// handleBlockUser 处理 /blockuser 命令
func (b *Bot) handleBlockUser(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供 Telegram ID\n\n使用方法: <code>/blockuser &lt;telegram_id&gt;</code>", nil
	}

	telegramIDStr := getArg(args, 0)
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		return "❌ Telegram ID 必须是有效的数字", nil
	}

	if err := b.userService.Block(ctx, telegramID); err != nil {
		return "", fmt.Errorf("封禁用户失败: %w", err)
	}

	return fmt.Sprintf("🚫 已封禁用户 %d", telegramID), nil
}

// handleUnblockUser 处理 /unblockuser 命令
func (b *Bot) handleUnblockUser(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供 Telegram ID\n\n使用方法: <code>/unblockuser &lt;telegram_id&gt;</code>", nil
	}

	telegramIDStr := getArg(args, 0)
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		return "❌ Telegram ID 必须是有效的数字", nil
	}

	if err := b.userService.Unblock(ctx, telegramID); err != nil {
		return "", fmt.Errorf("解封用户失败: %w", err)
	}

	return fmt.Sprintf("✅ 已解封用户 %d", telegramID), nil
}

// handleStats 处理 /stats 命令
func (b *Bot) handleStats(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	// 统计用户
	totalUsers, _ := b.userService.Count(ctx)
	adminCount, _ := b.userService.CountByRole(ctx, user.RoleAdmin)
	userCount, _ := b.userService.CountByRole(ctx, user.RoleUser)

	// 统计账号
	totalAccounts, _ := b.accountService.Count(ctx)
	activeAccounts, _ := b.accountService.CountByStatus(ctx, account.StatusActive)
	suspendedAccounts, _ := b.accountService.CountByStatus(ctx, account.StatusSuspended)
	expiredAccounts, _ := b.accountService.CountByStatus(ctx, account.StatusExpired)

	return fmt.Sprintf(`📊 <b>系统统计</b>

<b>用户统计:</b>
• 总用户数: %d
• 管理员: %d
• 普通用户: %d

<b>账号统计:</b>
• 总账号数: %d
• 激活账号: %d ✅
• 暂停账号: %d ⏸️
• 过期账号: %d ❌

<b>平均账号数:</b> %.2f 个/用户
`,
		totalUsers,
		adminCount,
		userCount,
		totalAccounts,
		activeAccounts,
		suspendedAccounts,
		expiredAccounts,
		float64(totalAccounts)/float64(totalUsers),
	), nil
}

func (b *Bot) handlePlayingStats(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if b.embyClient == nil {
		return "❌ Emby 同步未启用", nil
	}

	sessions, err := b.embyClient.GetSessions(ctx)
	if err != nil {
		logger.Errorf("failed to get emby sessions: %v", err)
		return "", fmt.Errorf("获取播放状态失败: %w", err)
	}

	if len(sessions) == 0 {
		return "📺 当前没有活跃的播放会话", nil
	}

	var playingSessions []string
	var idleSessions []string

	for _, session := range sessions {
		if session.IsPlaying() {
			item := session.NowPlayingItem
			progress := session.GetProgress()

			info := fmt.Sprintf("👤 <b>%s</b>\n", session.UserName)
			info += fmt.Sprintf("   📱 %s (%s)\n", session.DeviceName, session.Client)
			info += fmt.Sprintf("   🎬 %s\n", item.GetDisplayName())
			info += fmt.Sprintf("   ⏱️ %.1f%% | %s",
				progress,
				session.PlayState.PlayMethod)

			if session.TranscodingInfo != nil && (!session.TranscodingInfo.IsVideoDirect || !session.TranscodingInfo.IsAudioDirect) {
				info += fmt.Sprintf(" | 转码中 (%.1f%%)", session.TranscodingInfo.CompletionPercentage)
			}

			playingSessions = append(playingSessions, info)
		} else if session.NowPlayingItem != nil {
			idleSessions = append(idleSessions, fmt.Sprintf("👤 <b>%s</b> - 已暂停", session.UserName))
		}
	}

	result := "📺 <b>Emby 播放状态</b>\n\n"

	if len(playingSessions) > 0 {
		result += fmt.Sprintf("<b>正在播放 (%d):</b>\n", len(playingSessions))
		for i, info := range playingSessions {
			if i > 0 {
				result += "\n"
			}
			result += info + "\n"
		}
	}

	if len(idleSessions) > 0 {
		if len(playingSessions) > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("<b>已暂停 (%d):</b>\n", len(idleSessions))
		for _, info := range idleSessions {
			result += info + "\n"
		}
	}

	result += fmt.Sprintf("\n📊 总会话数: %d", len(sessions))

	return result, nil
}

func (b *Bot) handleUpdatePolicies(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if b.embyClient == nil {
		return "❌ Emby 同步未启用", nil
	}

	updated, failed, err := b.embyClient.BatchUpdateNonAdminPolicies(ctx)
	if err != nil {
		logger.Errorf("failed to batch update policies: %v", err)
		return "", fmt.Errorf("批量更新策略失败: %w", err)
	}

	return fmt.Sprintf(`✅ <b>批量更新用户策略完成</b>

• 成功更新: %d 个用户
• 失败: %d 个用户
• 已跳过管理员账号

所有非管理员用户已应用默认安全策略。`, updated, failed), nil
}
