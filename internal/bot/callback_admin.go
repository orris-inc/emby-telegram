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
	case "user":
		if len(parts) < 3 {
			return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
		}
		userID := strToUint(parts[2])
		page := 1
		if len(parts) >= 4 {
			page = strToInt(parts[3])
		}
		return b.showUserDetail(ctx, userID, page)
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
	case "playing":
		return b.showPlayingStats(ctx)
	case "updatepolicies":
		return b.handleUpdatePoliciesCallback(ctx)
	case "account":
		if len(parts) < 3 {
			return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
		}
		accountID := strToUint(parts[2])
		page := 1
		if len(parts) >= 4 {
			page = strToInt(parts[3])
		}
		return b.showAdminAccountDetail(ctx, accountID, page)
	case "suspend":
		if len(parts) < 3 {
			return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
		}
		accountID := strToUint(parts[2])
		return b.handleSuspendAccount(ctx, accountID)
	case "activate":
		if len(parts) < 3 {
			return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
		}
		accountID := strToUint(parts[2])
		return b.handleActivateAccount(ctx, accountID)
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
	limit := 5
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

	text := fmt.Sprintf(`👥 <b>用户列表</b>

共 %d 个用户，点击用户查看详情`, totalCount)

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, u := range users {
		roleEmoji := "👤"
		if u.IsAdmin() {
			roleEmoji = "👑"
		}

		statusEmoji := "✅"
		if u.IsBlocked {
			statusEmoji = "🚫"
		}

		buttonText := fmt.Sprintf("%s %s %s %s", roleEmoji, u.DisplayName(), statusEmoji, u.Role)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				buttonText,
				CallbackAdminUserDetail+":"+fmt.Sprintf("%d:%d", u.ID, page),
			),
		))
	}

	totalPages := (int(totalCount) + limit - 1) / limit

	if totalPages > 1 {
		var pageRow []tgbotapi.InlineKeyboardButton
		if page > 1 {
			pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ 上一页", CallbackAdminUsers+":"+fmt.Sprintf("%d", page-1)))
		}
		pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d/%d", page, totalPages),
			"page:current",
		))
		if page < totalPages {
			pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData("➡️ 下一页", CallbackAdminUsers+":"+fmt.Sprintf("%d", page+1)))
		}
		rows = append(rows, pageRow)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回", CallbackAdminMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showUserDetail 显示用户详情
func (b *Bot) showUserDetail(ctx context.Context, userID uint, page int) CallbackResponse {
	u, err := b.userService.Get(ctx, userID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取用户信息失败",
			ShowAlert: true,
		}
	}

	roleEmoji := "👤"
	if u.IsAdmin() {
		roleEmoji = "👑"
	}

	statusText := "正常"
	if u.IsBlocked {
		statusText = "已封禁"
	}

	accountCount, _ := b.accountService.CountByUser(ctx, u.ID)

	text := fmt.Sprintf(`👤 <b>用户详情</b>

%s <b>%s</b>

<b>Telegram ID:</b> <code>%d</code>
<b>用户名:</b> @%s
<b>姓名:</b> %s
<b>角色:</b> %s
<b>状态:</b> %s
<b>账号数量:</b> %d 个
<b>注册时间:</b> %s`,
		roleEmoji,
		u.DisplayName(),
		u.TelegramID,
		u.Username,
		u.FullName(),
		u.Role,
		statusText,
		accountCount,
		timeutil.FormatDateTime(u.CreatedAt),
	)

	keyboard := BackButton(CallbackAdminUsers + ":" + fmt.Sprintf("%d", page))

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showAllAccountsList 显示所有账号列表
func (b *Bot) showAllAccountsList(ctx context.Context, page int) CallbackResponse {
	limit := 5
	offset := (page - 1) * limit

	accounts, err := b.accountService.ListAllWithUser(ctx, offset, limit)
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

	text := fmt.Sprintf(`📋 <b>所有账号列表</b>

共 %d 个账号，点击账号查看详情`, totalCount)

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, acc := range accounts {
		status := getStatusEmoji(string(acc.Status))

		buttonText := fmt.Sprintf("%s %s - %s", status, acc.Username, acc.GetOwnerDisplayName())

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				buttonText,
				CallbackAdminAccountDetail+":"+fmt.Sprintf("%d:%d", acc.ID, page),
			),
		))
	}

	totalPages := (int(totalCount) + limit - 1) / limit

	if totalPages > 1 {
		var pageRow []tgbotapi.InlineKeyboardButton
		if page > 1 {
			pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ 上一页", CallbackAdminAccounts+":"+fmt.Sprintf("%d", page-1)))
		}
		pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d/%d", page, totalPages),
			"page:current",
		))
		if page < totalPages {
			pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData("➡️ 下一页", CallbackAdminAccounts+":"+fmt.Sprintf("%d", page+1)))
		}
		rows = append(rows, pageRow)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回", CallbackAdminMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return CallbackResponse{
		EditText:   text,
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
	if b.embyClient == nil {
		return CallbackResponse{
			Answer:    "Emby 同步未启用",
			ShowAlert: true,
		}
	}

	status := "✅ 已连接"
	if err := b.embyClient.Ping(ctx); err != nil {
		status = fmt.Sprintf("❌ 连接失败: %v", err)
	}

	text := fmt.Sprintf(`🎬 <b>Emby 管理</b>

<b>连接状态:</b> %s

请选择操作：`, status)

	keyboard := EmbyMenuKeyboard()

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showPlayingStats 显示播放统计
func (b *Bot) showPlayingStats(ctx context.Context) CallbackResponse {
	if b.embyClient == nil {
		return CallbackResponse{
			Answer:    "Emby 同步未启用",
			ShowAlert: true,
		}
	}

	sessions, err := b.embyClient.GetSessions(ctx)
	if err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("获取播放统计失败: %v", err),
			ShowAlert: true,
		}
	}

	var builder strings.Builder
	builder.WriteString("📊 <b>当前播放统计</b>\n\n")

	playingCount := 0
	for _, session := range sessions {
		if session.IsPlaying() {
			playingCount++
			builder.WriteString(fmt.Sprintf("👤 <b>%s</b>\n", session.UserName))
			builder.WriteString(fmt.Sprintf("📺 %s\n", session.NowPlayingItem.GetDisplayName()))
			builder.WriteString(fmt.Sprintf("💻 %s (%s)\n", session.DeviceName, session.Client))
			builder.WriteString(fmt.Sprintf("⏱ 进度: %.1f%%\n", session.GetProgress()))

			if session.TranscodingInfo != nil {
				playMethod := "直接播放"
				if !session.TranscodingInfo.IsVideoDirect || !session.TranscodingInfo.IsAudioDirect {
					playMethod = "转码中"
				}
				builder.WriteString(fmt.Sprintf("🎬 %s\n", playMethod))
			}
			builder.WriteString("\n")
		}
	}

	if playingCount == 0 {
		builder.WriteString("当前没有用户在播放内容")
	} else {
		builder.WriteString(fmt.Sprintf("共 %d 个用户正在播放", playingCount))
	}

	keyboard := BackButton(CallbackAdminEmby)

	return CallbackResponse{
		EditText:   builder.String(),
		EditMarkup: &keyboard,
	}
}

// handleUpdatePoliciesCallback 处理批量更新策略回调
func (b *Bot) handleUpdatePoliciesCallback(ctx context.Context) CallbackResponse {
	if b.embyClient == nil {
		return CallbackResponse{
			Answer:    "Emby 同步未启用",
			ShowAlert: true,
		}
	}

	updated, failed, err := b.embyClient.BatchUpdateNonAdminPolicies(ctx)
	if err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("批量更新策略失败: %v", err),
			ShowAlert: true,
		}
	}

	text := fmt.Sprintf(`✅ <b>批量更新用户策略完成</b>

• 成功更新: %d 个用户
• 失败: %d 个用户
• 已跳过管理员账号`, updated, failed)

	keyboard := BackButton(CallbackAdminEmby)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showAdminAccountDetail 显示管理员账号详情
func (b *Bot) showAdminAccountDetail(ctx context.Context, accountID uint, page int) CallbackResponse {
	acc, err := b.accountService.GetWithUser(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	status := getStatusEmoji(string(acc.Status))
	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)
	createdAt := timeutil.FormatDateTime(acc.CreatedAt)

	syncStatus := "✅ 已同步"
	if acc.EmbyUserID == "" {
		syncStatus = "❌ 未同步"
	} else if acc.SyncError != "" {
		syncStatus = fmt.Sprintf("⚠️ 同步失败: %s", acc.SyncError)
	}

	ownerInfo := fmt.Sprintf("%s (ID: %d)", acc.GetOwnerDisplayName(), acc.OwnerTelegramID)

	text := fmt.Sprintf(`📝 <b>账号详情</b>

<b>用户名:</b> <code>%s</code>
<b>状态:</b> %s %s
<b>有效期:</b> %s
<b>最大设备数:</b> %d
<b>创建时间:</b> %s
<b>所属用户:</b> %s
<b>Emby 同步状态:</b> %s
<b>Emby 用户ID:</b> <code>%s</code>`,
		acc.Username,
		status,
		acc.Status,
		expireInfo,
		acc.MaxDevices,
		createdAt,
		ownerInfo,
		syncStatus,
		acc.EmbyUserID,
	)

	keyboard := AdminAccountActionsKeyboard(acc.ID, string(acc.Status), page)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// handleSuspendAccount 处理停用账号
func (b *Bot) handleSuspendAccount(ctx context.Context, accountID uint) CallbackResponse {
	if err := b.accountService.Suspend(ctx, accountID); err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("停用失败: %v", err),
			ShowAlert: true,
		}
	}

	acc, _ := b.accountService.Get(ctx, accountID)

	return CallbackResponse{
		Answer: "账号已停用",
		EditText: fmt.Sprintf(`✅ <b>账号已停用</b>

账号 <b>%s</b> 已被停用
如需重新激活，请点击"激活"按钮`, acc.Username),
		EditMarkup: func() *tgbotapi.InlineKeyboardMarkup {
			kb := AdminAccountActionsKeyboard(acc.ID, string(acc.Status), 1)
			return &kb
		}(),
	}
}

// handleActivateAccount 处理激活账号
func (b *Bot) handleActivateAccount(ctx context.Context, accountID uint) CallbackResponse {
	if err := b.accountService.Activate(ctx, accountID); err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("激活失败: %v", err),
			ShowAlert: true,
		}
	}

	acc, _ := b.accountService.Get(ctx, accountID)

	return CallbackResponse{
		Answer: "账号已激活",
		EditText: fmt.Sprintf(`✅ <b>账号已激活</b>

账号 <b>%s</b> 已被激活
如需停用，请点击"停用"按钮`, acc.Username),
		EditMarkup: func() *tgbotapi.InlineKeyboardMarkup {
			kb := AdminAccountActionsKeyboard(acc.ID, string(acc.Status), 1)
			return &kb
		}(),
	}
}
