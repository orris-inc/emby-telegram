// Package bot 管理员回调处理
package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/account"
	"emby-telegram/internal/user"
	"emby-telegram/pkg/timeutil"
)

// handleAdminCallback 处理管理员相关回调
func (b *Bot) handleAdminCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	// 检查管理员权限
	if !currentUser.IsAdmin() {
		return CallbackResponse{
			Answer:    "此功能需要管理员权限",
			ShowAlert: true,
		}
	}

	if len(parts) < 2 {
		return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
	}

	subAction := parts[1]

	switch subAction {
	case "menu":
		return b.showAdminMenu(ctx)
	case "users":
		page := 1
		if len(parts) >= 3 {
			page = strToInt(parts[2])
		}
		return b.showUsersList(ctx, page)
	case "accounts":
		page := 1
		if len(parts) >= 3 {
			page = strToInt(parts[2])
		}
		return b.showAllAccountsList(ctx, page)
	case "stats":
		return b.showStats(ctx)
	case "emby":
		return b.showEmbyMenu(ctx)
	default:
		return CallbackResponse{Answer: "未知操作", ShowAlert: true}
	}
}

// showAdminMenu 显示管理员菜单
func (b *Bot) showAdminMenu(ctx context.Context) CallbackResponse {
	text := `🔑 <b>管理员菜单</b>

请选择管理功能：`

	keyboard := AdminMenuKeyboard()

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showUsersList 显示用户列表
func (b *Bot) showUsersList(ctx context.Context, page int) CallbackResponse {
	limit := 10
	offset := (page - 1) * limit

	users, err := b.userService.List(ctx, offset, limit)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取用户列表失败",
			ShowAlert: true,
		}
	}

	totalCount, _ := b.userService.Count(ctx)

	if len(users) == 0 {
		return CallbackResponse{
			Answer:    "没有找到用户",
			ShowAlert: true,
		}
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
	keyboard := PaginationKeyboard(CallbackAdminUsers, page, totalPages, CallbackAdminMenu)

	return CallbackResponse{
		EditText:   builder.String(),
		EditMarkup: &keyboard,
	}
}

// showAllAccountsList 显示所有账号列表
func (b *Bot) showAllAccountsList(ctx context.Context, page int) CallbackResponse {
	limit := 10
	offset := (page - 1) * limit

	accounts, err := b.accountService.ListAll(ctx, offset, limit)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号列表失败",
			ShowAlert: true,
		}
	}

	totalCount, _ := b.accountService.Count(ctx)

	if len(accounts) == 0 {
		return CallbackResponse{
			Answer:    "没有找到账号",
			ShowAlert: true,
		}
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
	keyboard := PaginationKeyboard(CallbackAdminAccounts, page, totalPages, CallbackAdminMenu)

	return CallbackResponse{
		EditText:   builder.String(),
		EditMarkup: &keyboard,
	}
}

// showStats 显示统计信息
func (b *Bot) showStats(ctx context.Context) CallbackResponse {
	// 统计用户
	totalUsers, _ := b.userService.Count(ctx)
	adminCount, _ := b.userService.CountByRole(ctx, user.RoleAdmin)
	userCount, _ := b.userService.CountByRole(ctx, user.RoleUser)

	// 统计账号
	totalAccounts, _ := b.accountService.Count(ctx)
	activeAccounts, _ := b.accountService.CountByStatus(ctx, account.StatusActive)
	suspendedAccounts, _ := b.accountService.CountByStatus(ctx, account.StatusSuspended)
	expiredAccounts, _ := b.accountService.CountByStatus(ctx, account.StatusExpired)

	avgAccounts := 0.0
	if totalUsers > 0 {
		avgAccounts = float64(totalAccounts) / float64(totalUsers)
	}

	text := fmt.Sprintf(`📊 <b>系统统计</b>

<b>用户统计:</b>
• 总用户数: %d
• 管理员: %d 👑
• 普通用户: %d 👤

<b>账号统计:</b>
• 总账号数: %d
• 激活账号: %d ✅
• 暂停账号: %d ⏸️
• 过期账号: %d ❌

<b>平均账号数:</b> %.2f 个/用户`,
		totalUsers,
		adminCount,
		userCount,
		totalAccounts,
		activeAccounts,
		suspendedAccounts,
		expiredAccounts,
		avgAccounts,
	)

	keyboard := BackButton(CallbackAdminMenu)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showEmbyMenu 显示 Emby 管理菜单
func (b *Bot) showEmbyMenu(ctx context.Context) CallbackResponse {
	// 检查 Emby 连接状态
	status := "✅ 已连接"
	if err := b.embyClient.Ping(ctx); err != nil {
		status = fmt.Sprintf("❌ 连接失败: %v", err)
	}

	text := fmt.Sprintf(`🎬 <b>Emby 管理</b>

<b>连接状态:</b> %s

<b>可用操作:</b>
• 使用 /checkemby 检查连接
• 使用 /embyusers 列出所有用户
• 使用 /syncaccount 手动同步账号`, status)

	keyboard := BackButton(CallbackAdminMenu)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}
